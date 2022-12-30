package fs

import (
	"fmt"
	"io"
	"net/http"

	"github.com/eagraf/habitat/cmd/habitat/api"
	"github.com/eagraf/habitat/pkg/ipfs"
	"github.com/eagraf/habitat/structs/ctl"
)

// CAFS stands for Content-Addressable File System. This interface is an abstraction
// of file systems like IPFS
type CAFS interface {
	Add(io.Reader) (string, error)
	Get(string) (io.Reader, error)
}

type IPFS struct {
	client *ipfs.Client
}

func (i *IPFS) Add(r io.Reader) (string, error) {
	res, err := i.client.AddFile("file", r)
	if err != nil {
		return "", err
	}

	return res.Hash, nil
}

func (i *IPFS) Get(contentID string) (io.Reader, error) {
	r, err := i.client.CatFile(contentID)
	if err != nil {
		return nil, err
	}
	return r, nil
}

type FS struct {
	cafs CAFS
}

func NewFS(ipfsClient *ipfs.Client) *FS {
	return &FS{
		cafs: &IPFS{
			client: ipfsClient,
		},
	}
}

func (fs *FS) AddHandler(w http.ResponseWriter, r *http.Request) {
	// Limit files to 10 MB
	// TODO better limiting on upload sizes to prevent upload overwhelming resource usage
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("error reading form file: %s", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentID, err := fs.cafs.Add(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	res := &ctl.AddFileResponse{
		ContentID: contentID,
	}
	api.WriteResponse(w, res)
}

func (fs *FS) GetHandler(w http.ResponseWriter, r *http.Request) {
	var req ctl.GetFileRequest
	err := api.BindPostRequest(r, &req)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	body, err := fs.cafs.Get(req.ContentID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	_, err = io.Copy(w, body)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
}
