package ipfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
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

	if len(buf) > 0 {
		err = json.Unmarshal(buf, res)
		if err != nil {
			// assume that IPFS api will always return properly marshaled json, and that an unmarshaling error
			// indicates that there was some sort of error with our request. return body text in error
			return fmt.Errorf("ipfs client error: %s (%s)", err, string(buf))
		}
	}

	return nil
}

func (c *Client) postFile(endpointPath string, filename string, file io.Reader, res interface{}) error {
	if file == nil {
		return fmt.Errorf("no file supplied")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("error copying file contents: %s", err)
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.getEndpointURL(endpointPath), body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to get response to %s", req.URL)
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

	if len(buf) > 0 {
		err = json.Unmarshal(buf, res)
		if err != nil {
			// assume that IPFS api will always return properly marshaled json, and that an unmarshaling error
			// indicates that there was some sort of error with our request. return body text in error
			return fmt.Errorf("ipfs client error: %s (%s)", err, string(buf))
		}
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
	err := c.postRequest("/files/ls", nil, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

type MkdirResponse struct{}

func (c *Client) Mkdir(path string) (*MkdirResponse, error) {
	var res MkdirResponse
	err := c.postRequest(fmt.Sprintf("/files/mkdir?arg=%s", path), nil, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) ReadFile(path string) (io.Reader, bool, error) {
	resp, err := http.Post(c.getEndpointURL(fmt.Sprintf("/files/read?arg=%s", path)), "raw/json", nil)
	if err != nil {
		return nil, false, fmt.Errorf("error creating request: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, false, err
		}

		// Hax
		if strings.Contains(string(body), "file does not exist") {
			return nil, false, nil
		}

		return nil, false, fmt.Errorf("got exit code %s from IPFS: %s", resp.Status, body)
	}

	return resp.Body, true, nil
}

type WriteFileResponse struct{}

func (c *Client) WriteFile(path string, filename string, file io.Reader) (*WriteFileResponse, error) {
	var res AddFileResponse
	err := c.postFile(fmt.Sprintf("/files/write?arg=%s&create=true&parents=true&truncate=true", path), filename, file, &res)
	if err != nil {
		return nil, err
	}
	return &WriteFileResponse{}, nil

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

type AddFileResponse struct {
	Bytes int64
	Hash  string
	Name  string
	Size  string
}

func (c *Client) AddFile(filename string, file io.Reader) (*AddFileResponse, error) {
	var res AddFileResponse
	err := c.postFile("/add", filename, file, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) CatFile(path string) (io.Reader, error) {
	resp, err := http.Post(c.getEndpointURL(fmt.Sprintf("/cat?arg=%s", path)), "raw/json", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("got exit code %s from IPFS: %s", resp.Status, body)
	}

	return resp.Body, nil
}
