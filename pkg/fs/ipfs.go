package fs

import (
	"errors"
	"time"

	"github.com/eagraf/habitat/pkg/ipfs"
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
	_, err = client.GetVersion()
	if err != nil {
		return nil, err
	}

	return &IPFS{
		client: client,
	}, nil
}

func (ipfs *IPFS) Open(name string) (File, error) {
	return nil, errors.New("unimplimented")
}

func (ipfs *IPFS) ReadDir(name string) ([]DirEntry, error) {
	files, err := ipfs.client.ListFiles()
	if err != nil {
		return nil, err
	}

	res := make([]DirEntry, len(files.Entries))
	for i, f := range files.Entries {
		res[i] = &IPFSDirEntry{
			name:    f.Name,
			isDir:   f.Type == 0,
			dirType: 0,
			info: &IPFSFileInfo{
				name:  f.Name,
				size:  f.Size,
				isDir: f.Type == 0,
			},
		}
	}

	return res, nil
}
