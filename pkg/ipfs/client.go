package ipfs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Client struct {
	apiURL *url.URL
}

func NewClient(apiURL string) (*Client, error) {
	url, err := url.ParseRequestURI(apiURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		apiURL: url,
	}, nil
}

func (c *Client) getEndpointURL(endpointPath string) string {
	return c.apiURL.String() + endpointPath
}

func (c *Client) postRequest(endpointPath string, body, res interface{}) error {
	resp, err := http.Post(c.getEndpointURL(endpointPath), "raw/json", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("got exit code %s from IPFS: %s", resp.Status, body)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %s", err)
	}

	err = json.Unmarshal(buf, res)
	if err != nil {
		// assume that IPFS api will always return properly marshaled json, and that an unmarshaling error
		// indicates that there was some sort of error with our request. return body text in error
		return fmt.Errorf("ipfs client error: %s", string(buf))
	}

	return nil
}

type VersionResponse struct {
	Commit  string
	Golang  string
	Repo    string
	System  string
	Version string
}

func (c *Client) GetVersion() (*VersionResponse, error) {
	var res VersionResponse
	err := c.postRequest("version", nil, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

type ListFilesResponseEntry struct {
	Hash string
	Name string
	Size int64
	Type int
}

type ListFilesResponse struct {
	Entries []*ListFilesResponseEntry
}

func (c *Client) ListFiles() (*ListFilesResponse, error) {
	var res ListFilesResponse
	err := c.postRequest("files/ls", nil, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

type AddPeerResponse struct {
	ID     string
	Status string
}

func (c *Client) AddPeer(peerAddr string) (*AddPeerResponse, error) {
	var res AddPeerResponse
	err := c.postRequest(fmt.Sprintf("/swarm/peering/add?arg=%s", peerAddr), nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
