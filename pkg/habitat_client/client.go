package client

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"

	"github.com/eagraf/habitat/structs/ctl"
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

// SendRequest sends a request to the default address of the habitat service
func SendRequest(req interface{}) (*ctl.ResponseWrapper, error) {
	return SendRequestToAddress(HabitatServiceAddr, req)
}

func SendRequestToAddress(addr string, req interface{}) (*ctl.ResponseWrapper, error) {
	client, err := NewClient(addr)
	if err != nil {
		fmt.Println("Error: couldn't connect to habitat service")
		return nil, err
	}

	reqWrapper, err := ctl.NewRequestWrapper(req)
	if err != nil {
		return nil, err
	}

	err = client.WriteRequest(reqWrapper)
	if err != nil {
		fmt.Printf("Error creating request to habitat service: %s", err)
	}

	res, err := client.ReadResponse()
	if err != nil {
		fmt.Printf("Error: couldn't read response from habitat service: %s\n", err)
	}
	return res, err
}
