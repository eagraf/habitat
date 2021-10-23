package fs

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/eagraf/habitat/pkg/ipfs"
	"github.com/rs/zerolog/log"
)

type IPFSFile struct {
	info FileInfo
}

func (file *IPFSFile) Stat() (FileInfo, error) {
	return file.info, nil
}

func (file *IPFSFile) Read(p []byte) (int, error) {
	return 0, errors.New("unimplimented")
}

func (file *IPFSFile) Write(p []byte) (int, error) {
	return 0, errors.New("unimplimented")
}

func (file *IPFSFile) Close() error {
	return errors.New("unimplimented")
}

type IPFSFileInfo struct {
	name    string
	size    int64
	mode    FileMode
	modTime time.Time
	isDir   bool
}

func (fi *IPFSFileInfo) Name() string {
	return fi.name
}

func (fi *IPFSFileInfo) Size() int64 {
	return fi.size
}

func (fi *IPFSFileInfo) Mode() FileMode {
	return fi.mode
}

func (fi *IPFSFileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi *IPFSFileInfo) IsDir() bool {
	return fi.isDir
}

type IPFSDirEntry struct {
	name    string
	isDir   bool
	dirType FileMode
	info    FileInfo
}

func (de *IPFSDirEntry) Name() string {
	return de.name
}

func (de *IPFSDirEntry) IsDir() bool {
	return de.isDir
}

func (de *IPFSDirEntry) Type() FileMode {
	return de.dirType
}

func (de *IPFSDirEntry) Info() (FileInfo, error) {
	return de.info, nil
}

type IPFS struct {
	client *ipfs.Client
}

func NewIPFS(apiURL string) (*IPFS, error) {
	// validate that the url is valid and talks to an API instance
	client, err := ipfs.NewClient(apiURL)
	if err != nil {
		return nil, err
	}

	// get ipfs version to check api
	buf, err := client.PostRequest("version", "application/x-www-form-urlencoded", nil)
	if err != nil {
		return nil, err
	}
	var version struct {
		Commit  string
		Golang  string
		Repo    string
		System  string
		Version string
	}

	err = json.Unmarshal(buf, &version)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("using IPFS HTTP API " + version.Version + " " + version.System)

	return &IPFS{
		client: client,
	}, nil
}

func (ipfs *IPFS) Open(name string) ([]byte, error) {
	buf, err := ipfs.client.PostRequest("files/read?arg="+name, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (ipfs *IPFS) ReadDir(name string) ([]DirEntry, error) {
	buf, err := ipfs.client.PostRequest("files/ls", "application/x-www-form-urlencoded", nil)
	if err != nil {
		return nil, err
	}

	var files struct {
		Entries []*struct {
			Hash string
			Name string
			Size int64
			Type int
		}
	}

	err = json.Unmarshal(buf, &files)
	if err != nil {
		return nil, err
	}

	res := make([]DirEntry, len(files.Entries))
	for i, f := range files.Entries {
		res[i] = &IPFSDirEntry{
			name:    f.Name,
			isDir:   f.Type == 2,
			dirType: 0,
			info: &IPFSFileInfo{
				name:  f.Name,
				size:  f.Size,
				isDir: f.Type == 2,
			},
		}
	}

	return res, nil
}

func (ipfs *IPFS) Write(name string, body io.Reader, contentType string) ([]byte, error) {
	buf, err := ipfs.client.PostRequest("files/write?create=true&arg="+name, contentType, body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
