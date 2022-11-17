package transport

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"sync"

	"github.com/eagraf/habitat/pkg/compass"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
)

var (
	// ErrTransportShutdown is returned when operations on a transport are
	// invoked after it's been terminated.
	ErrTransportShutdown = errors.New("transport shutdown")
)

// LibP2PTransport is a transport using libp2p streams to communicate
type LibP2PTransport struct {
	consumeCh chan raft.RPC

	heartbeatFn     func(raft.RPC)
	heartbeatFnLock sync.Mutex

	logger hclog.Logger

	shutdown     bool
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex

	host       host.Host
	protocol   protocol.ID
	publicAddr raft.ServerAddress
}

// LibP2PTransportConfig encapsulates configuration for the network transport layer.
type LibP2PTransportConfig struct {
	Logger hclog.Logger

	Host       host.Host
	Protocol   protocol.ID
	PublicAddr raft.ServerAddress
}

// NewLibP2PTransportWithConfig creates a new network transport with the given config struct
func NewLibP2PTransportWithConfig(
	config *LibP2PTransportConfig,
) *LibP2PTransport {
	if config.Logger == nil {
		config.Logger = hclog.New(&hclog.LoggerOptions{
			Name:   "raft-net",
			Output: hclog.DefaultOutput,
			Level:  hclog.DefaultLevel,
		})
	}
	trans := &LibP2PTransport{
		consumeCh:  make(chan raft.RPC),
		logger:     config.Logger,
		shutdownCh: make(chan struct{}),
		host:       config.Host,
		protocol:   config.Protocol,
		publicAddr: config.PublicAddr,
	}

	trans.host.SetStreamHandler(trans.protocol, trans.streamHandler)

	return trans
}

// NewLibP2PTransport creates a new network transport with the given dialer
// and listener. The maxPool controls how many connections we will pool. The
// timeout is used to apply I/O deadlines. For InstallSnapshot, we multiply
// the timeout by (SnapshotSize / TimeoutScale).
func NewLibP2PTransport(
	host host.Host,
	protocol protocol.ID,
	publicAddr raft.ServerAddress,
) *LibP2PTransport {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "raft-net",
		Level: hclog.DefaultLevel,
	})
	config := &LibP2PTransportConfig{
		Host:       host,
		Protocol:   protocol,
		PublicAddr: publicAddr,
		Logger:     logger,
	}
	return NewLibP2PTransportWithConfig(config)
}

// SetHeartbeatHandler is used to setup a heartbeat handler
// as a fast-pass. This is to avoid head-of-line blocking from
// disk IO.
func (t *LibP2PTransport) SetHeartbeatHandler(cb func(rpc raft.RPC)) {
	t.heartbeatFnLock.Lock()
	defer t.heartbeatFnLock.Unlock()
	t.heartbeatFn = cb
}

// Close is used to stop the network transport.
func (t *LibP2PTransport) Close() error {
	t.shutdownLock.Lock()
	defer t.shutdownLock.Unlock()

	if !t.shutdown {
		close(t.shutdownCh)
		t.host.Close()
		t.shutdown = true
	}
	return nil
}

// Consumer implements the Transport interface.
func (t *LibP2PTransport) Consumer() <-chan raft.RPC {
	return t.consumeCh
}

// LocalAddr implements the Transport interface.
func (t *LibP2PTransport) LocalAddr() raft.ServerAddress {
	// TODO fixme
	return t.publicAddr
}

// AppendEntriesPipeline returns an interface that can be used to pipeline
// AppendEntries requests.
func (t *LibP2PTransport) AppendEntriesPipeline(id raft.ServerID, target raft.ServerAddress) (raft.AppendPipeline, error) {
	return nil, raft.ErrPipelineReplicationNotSupported
}

// AppendEntries implements the Transport interface.
func (t *LibP2PTransport) AppendEntries(id raft.ServerID, target raft.ServerAddress, args *raft.AppendEntriesRequest, resp *raft.AppendEntriesResponse) error {
	return t.genericRPC(id, target, rpcAppendEntries, args, resp)
}

// RequestVote implements the Transport interface.
func (t *LibP2PTransport) RequestVote(id raft.ServerID, target raft.ServerAddress, args *raft.RequestVoteRequest, resp *raft.RequestVoteResponse) error {
	return t.genericRPC(id, target, rpcRequestVote, args, resp)
}

// genericRPC handles a simple request/response RPC.
func (t *LibP2PTransport) genericRPC(id raft.ServerID, target raft.ServerAddress, rpcType uint8, args interface{}, resp interface{}) error {
	// Dial a new connection
	peerID, err := peer.Decode(string(id))
	if err != nil {
		return err
	}

	_, addr, err := compass.DecomposeNodeMultiaddr(string(target))
	if err != nil {
		return err
	}

	t.host.Peerstore().AddAddr(peerID, addr, peerstore.PermanentAddrTTL)

	stream, err := t.host.NewStream(context.Background(), peerID, t.protocol)
	if err != nil {
		return err
	}
	defer stream.Close()
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	buf, err := json.Marshal(args)
	if err != nil {
		return err
	}

	req := &RaftRequest{
		RPCType: rpcType,
		Args:    buf,
	}

	// yolo try to json marshal without using proper serialization format
	marshaled, err := json.Marshal(req)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(marshaled)
	encoded = encoded + "\n"

	_, err = rw.Write([]byte(encoded))
	if err != nil {
		stream.Reset()
		return err
	}
	rw.Flush()

	encodedResp, err := rw.ReadBytes('\n')
	if err != nil {
		return err
	}

	marshaledResp, err := base64.StdEncoding.DecodeString(string(encodedResp))
	if err != nil {
		return err
	}

	var htResp RaftResponse
	err = json.Unmarshal(marshaledResp, &htResp)
	if err != nil {
		return err
	}

	return unmarshalRPCResponse(rpcType, htResp.Resp, resp)
}

// InstallSnapshot implements the Transport interface.
func (t *LibP2PTransport) InstallSnapshot(id raft.ServerID, target raft.ServerAddress, args *raft.InstallSnapshotRequest, resp *raft.InstallSnapshotResponse, data io.Reader) error {
	return t.genericRPC(id, target, rpcTimeoutNow, args, resp)
}

// EncodePeer implements the Transport interface.
func (t *LibP2PTransport) EncodePeer(id raft.ServerID, address raft.ServerAddress) []byte {
	return []byte(address)
}

// DecodePeer implements the Transport interface.
func (t *LibP2PTransport) DecodePeer(buf []byte) raft.ServerAddress {
	return raft.ServerAddress(buf)
}

// TimeoutNow implements the Transport interface.
func (t *LibP2PTransport) TimeoutNow(id raft.ServerID, target raft.ServerAddress, args *raft.TimeoutNowRequest, resp *raft.TimeoutNowResponse) error {
	return t.genericRPC(id, target, rpcTimeoutNow, args, resp)
}

func (t *LibP2PTransport) streamHandler(s network.Stream) {
	// add the peer's address to our address book if we don't have it already
	t.host.Peerstore().AddAddr(s.Conn().RemotePeer(), s.Conn().RemoteMultiaddr(), peerstore.PermanentAddrTTL)

	defer s.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	if err := t.handleCommand(rw); err != nil {
		s.Reset()
		if err != io.EOF {
			t.logger.Error("failed to decode incoming command", "error", err)
		}
		return
	}
}

// handleCommand is used to decode and dispatch a single command.
func (t *LibP2PTransport) handleCommand(rw *bufio.ReadWriter) error {
	encoded, err := rw.ReadBytes('\n')
	if err != nil {
		return err
	}

	marshaled, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		return err
	}

	var req RaftRequest
	err = json.Unmarshal(marshaled, &req)
	if err != nil {
		return err
	}

	command, err := unmarshalRPCRequest(req.RPCType, req.Args)
	if err != nil {
		return err
	}

	// If we got an AppendEntriesRequest, check if it is a heartbeat
	isHeartbeat := false
	if req.RPCType == rpcAppendEntries {
		aeReq := command.(*raft.AppendEntriesRequest)
		if aeReq.Term != 0 && aeReq.Leader != nil &&
			aeReq.PrevLogEntry == 0 && aeReq.PrevLogTerm == 0 &&
			len(aeReq.Entries) == 0 && aeReq.LeaderCommitIndex == 0 {
			isHeartbeat = true
		}
	}

	// Create the RPC object
	respCh := make(chan raft.RPCResponse, 1)
	rpc := raft.RPC{
		Command:  command,
		RespChan: respCh,
	}

	// Check for heartbeat fast-path
	if isHeartbeat {
		t.heartbeatFnLock.Lock()
		fn := t.heartbeatFn
		t.heartbeatFnLock.Unlock()
		if fn != nil {
			fn(rpc)
			goto RESP
		}
	}

	// Dispatch the RPC
	select {
	case t.consumeCh <- rpc:
	case <-t.shutdownCh:
		return ErrTransportShutdown
	}

	// Wait for response
RESP:
	// we will differentiate the heartbeat fast path from normal RPCs with labels
	select {
	case resp := <-respCh:
		// Send the error first
		var respBody []byte
		if resp.Error != nil {
			respBody = []byte(resp.Error.Error())
		} else {
			// Serialize raft response and serialize into HTTP response
			buf, err := json.Marshal(resp.Response)
			if err != nil {
				return err
			}
			respBody = buf
		}

		httpRespBody := &RaftResponse{
			RPCType: req.RPCType,
			Resp:    []byte(respBody),
		}

		marshaledBody, err := json.Marshal(httpRespBody)
		if err != nil {
			return err
		}

		encoded := base64.StdEncoding.EncodeToString(marshaledBody)
		encoded = encoded + "\n"

		_, err = rw.Write([]byte(encoded))
		if err != nil {
			return err
		}
		rw.Flush()

	case <-t.shutdownCh:
		return ErrTransportShutdown
	}
	return nil
}
