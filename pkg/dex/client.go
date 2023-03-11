package dex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
)

type Client struct {
	sockPath string
}

func NewClient(sockPath string) (*Client, error) {
	return &Client{sockPath: sockPath}, nil
}

func (c *Client) httpClient() (*http.Client, error) {
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", c.sockPath)
			},
		},
	}

	return &httpc, nil
}

func (c *Client) Get(contentIdentifier string) (*GetResult, error) {
	httpc, err := c.httpClient()
	if err != nil {
		return nil, err
	}

	resp, err := httpc.Get(fmt.Sprintf("http://unix/dex/driver/get/%s", contentIdentifier))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get request returned %s: %s", resp.Status, string(body))
	}

	var res GetResult
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) Schema(hash string) (*SchemaResult, error) {
	httpc, err := c.httpClient()
	if err != nil {
		return nil, err
	}

	resp, err := httpc.Get(fmt.Sprintf("http://unix/dex/driver/schema/%s", hash))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("schema request returned %s: %s", resp.Status, string(body))
	}

	var res SchemaResult
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) Interface(hash string) (*InterfaceResult, error) {
	httpc, err := c.httpClient()
	if err != nil {
		return nil, err
	}

	resp, err := httpc.Get(fmt.Sprintf("http://unix/dex/driver/interface/%s", hash))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("interface request returned %s: %s", resp.Status, string(body))
	}

	var res InterfaceResult
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) Implementations(interfaceHash string) (*ImplementationsResult, error) {
	httpc, err := c.httpClient()
	if err != nil {
		return nil, err
	}

	resp, err := httpc.Get(fmt.Sprintf("http://unix/dex/driver/implementations/%s", interfaceHash))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("implementations request returned %s: %s", resp.Status, string(body))
	}

	var res ImplementationsResult
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
