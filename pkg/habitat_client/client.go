package client

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/eagraf/habitat/pkg/p2p"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	HabitatServiceAddr = "localhost:2040"
)

type Client struct {
	conn net.Conn
}

func NewClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
	}, nil
}

func (c *Client) WriteRequest(message *ctl.RequestWrapper) error {
	buf, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// base64 encode to make sure newlines are not present in bytes sent
	encoded := base64.StdEncoding.EncodeToString(buf)

	msg := append([]byte(encoded), '\n')

	_, err = c.conn.Write(msg)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ReadResponse() (*ctl.ResponseWrapper, error) {
	buf, err := bufio.NewReader(c.conn).ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}

	var res ctl.ResponseWrapper
	err = json.Unmarshal(decoded, &res)
	if err != nil {
		return nil, err
	}

	c.conn.Close()

	return &res, nil
}

func PostRequest(req, res interface{}, route string) (error, error) {
	return PostRequestToAddress(fmt.Sprintf("http://%s/%s", HabitatServiceAddr, route), req, res)
}

// PostRequestToAddress posts to the Habitat API. The first error returned is
// if for if the client is somehow unable to successfully make the request.
// The second error is if the request is successfully made, but the server responds
// with an error.
func PostRequestToAddress(address string, req, res interface{}) (error, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshaling POST request body: %s", err), nil
	}

	r, err := http.Post(address, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err, nil
	}

	resBody, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %s", err), nil
	}
	// The request was fine, but we got an error back from server
	if r.StatusCode != http.StatusOK {
		return nil, errors.New(string(resBody))
	}

	err = json.Unmarshal(resBody, res)
	if err != nil {
		return fmt.Errorf("error unmarshaling response body into result struct: %s", err), nil
	}

	return nil, nil
}

func PostLibP2PRequestToAddress(addr ma.Multiaddr, route string, peerID peer.ID, req, res interface{}) (error, error) {

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshaling POST request body: %s", err), nil
	}

	p2pReq, err := http.NewRequest("POST", "", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("error constructing HTTP request: %s", err), nil
	}

	p2pRes, err := p2p.LibP2PHTTPRequestWithRandomClient(addr, route, peerID, p2pReq)
	if err != nil {
		return err, nil
	}

	resBody, err := io.ReadAll(p2pRes.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %s", err), nil
	}

	// The request was fine, but we got an error back from server
	if p2pRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", p2pRes.Status, string(resBody))
	}

	err = json.Unmarshal(resBody, res)
	if err != nil {
		return fmt.Errorf("error unmarshaling response body into result struct: %s", err), nil
	}

	return nil, nil
}
