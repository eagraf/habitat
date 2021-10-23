package ipfs

import (
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

func (c *Client) getEndpointURL(endpointPath string) *url.URL {
	url, err := c.apiURL.Parse("./" + endpointPath)
	if err != nil {
		//TODO handle bad path
	}
	return url
}

func (c *Client) PostRequest(endpointPath string, contentType string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest("POST", c.getEndpointURL(endpointPath).String(), body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}
	req.Header.Set("Content-Type", contentType)

	client := http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error doing request: %s", err)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %s", err)
	}

	return buf, nil
}
