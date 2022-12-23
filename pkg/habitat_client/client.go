package client

import (
	"bufio"
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/identity"
	"github.com/eagraf/habitat/pkg/p2p"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/gorilla/websocket"
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
	return PostRequestToAddress(fmt.Sprintf("%s%s", compass.DefaultHabitatAPIAddr(), route), req, res)
}

// PostRequestToAddress posts to the Habitat API. The first error returned is
// for when the client is somehow unable to successfully make the request.
// The second error is if the request is successfully made, but the server responds
// with an error.
// TODO refactor so everything takes in a http.Client. This allows us to deduplicate libp2p and regular transport requests
func PostRequestToAddress(address string, req, res interface{}) (error, error) {
	resBody, cliErr, apiErr := PostRequestRaw(address, req)

	if cliErr != nil {
		return cliErr, nil
	} else if apiErr != nil {
		return nil, apiErr
	}

	err := json.Unmarshal(resBody, res)
	if err != nil {
		return fmt.Errorf("error unmarshaling response body into result struct: %s", err), nil
	}

	return nil, nil
}

// PostRequestRaw posts the request to the given address. It first JSON marshals the request.
// It returns the result in []byte form
func PostRequestRaw(address string, req interface{}) ([]byte, error, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling POST request body: %s", err), nil
	}

	r, err := http.Post(address, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error on http.Post: %s", err.Error()), nil
	}

	resBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err), nil
	}
	// The request was fine, but we got an error back from server
	if r.StatusCode != http.StatusOK {
		return nil, nil, errors.New(string(resBody))
	}

	return resBody, nil, nil
}

func PostFileToAddress(address string, client *http.Client, file *os.File, res interface{}) (error, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return fmt.Errorf("error creating form file: %s", err), nil
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("error copying file contents into request: %s", err), nil
	}
	err = writer.Close()
	if err != nil {
		return err, nil
	}

	req, err := http.NewRequest("POST", address, body)
	if err != nil {
		return err, nil
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return err, nil
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %s", err), nil
	}
	// The request was fine, but we got an error back from server
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(resBody))
	}

	err = json.Unmarshal(resBody, res)
	if err != nil {
		return fmt.Errorf("error unmarshaling response body into result struct: %s", err), nil
	}

	return nil, nil
}

func PostRetrieveFileFromAddress(address string, req interface{}) (io.Reader, error, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling POST request body: %s", err), nil
	}

	r, err := http.Post(address, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err, nil
	}
	if r.StatusCode != http.StatusOK {
		resBody, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err, nil
		}
		return nil, nil, errors.New(string(resBody))
	}

	return r.Body, nil, nil
}

func PostLibP2PRequestToAddress(node *p2p.Node, proxyAddr string, route string, req, res interface{}) (error, error) {

	peerID, addr, err := compass.LibP2PHabitatAPIAddr(proxyAddr)
	if err != nil {
		return fmt.Errorf("error decomposing multiaddr: %s", err), nil
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshaling POST request body: %s", err), nil
	}

	p2pReq, err := http.NewRequest("POST", "", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("error constructing HTTP request: %s", err), nil
	}

	var p2pRes *http.Response
	if node == nil {
		apiErr, err := PostLibP2PRequestToAddress(nil, addr.String(), route, p2pReq, p2pRes)
		if err != nil {
			return err, nil
		} else if apiErr != nil {
			return apiErr, nil
		}
	} else {
		nodeRes, err := node.PostHTTPRequest(addr, route, peerID, p2pReq)
		if err != nil {
			return err, nil
		}
		p2pRes = nodeRes
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

func GetWebsocketConn(addr, route string) (*websocket.Conn, error) {
	wsURL := addr + route

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func WebsocketKeySigningExchange(conn *websocket.Conn, userIdentity *identity.UserIdentity) error {
	var pubKeyMsg ctl.SigningPublicKeyMsg
	err := conn.ReadJSON(&pubKeyMsg)
	if err != nil {
		return err
	}
	if werr := pubKeyMsg.GetError(); werr != nil {
		return werr
	}

	pubKey, err := x509.ParsePKCS1PublicKey(pubKeyMsg.PublicKey)
	if err != nil {
		return err
	}

	// TODO propogate proper node IDs through
	cert, err := identity.GenerateMemberNodeCertificate(pubKeyMsg.NodeID, userIdentity, pubKey)
	if err != nil {
		return err
	}

	// PEM encode the cert
	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	certMsg := &ctl.SigningCertMsg{
		UserCertificate: userIdentity.CertBytes,
		NodeCertificate: certPEM.Bytes(),
	}

	err = conn.WriteJSON(certMsg)
	if err != nil {
		return err
	}
	return nil
}
