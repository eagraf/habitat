package ipfs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
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

func (c *Client) getEndpointURL(endpointPath string) *url.URL {
	urlCopy := *c.apiURL
	urlCopy.Path = path.Join(urlCopy.Path, endpointPath)

	return &urlCopy
}

func (c *Client) postRequest(endpointPath string, body, res interface{}) error {
	resp, err := http.Post(c.getEndpointURL(endpointPath).String(), "raw/json", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
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
