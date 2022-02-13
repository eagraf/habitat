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
	ClientHost = "localhost"
)

type Client struct {
	conn net.Conn
}

func NewClient(port string) (*Client, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", ClientHost, port))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
	}, nil
}

func (c *Client) WriteRequest(message *ctl.Request) error {
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

func (c *Client) ReadResponse() (*ctl.Response, error) {
	buf, err := bufio.NewReader(c.conn).ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}

	var res ctl.Response
	err = json.Unmarshal(decoded, &res)
	if err != nil {
		return nil, err
	}

	c.conn.Close()

	return &res, nil
}
