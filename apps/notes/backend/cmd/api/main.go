package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/eagraf/habitat/pkg/fs"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func main() {
	var fsUrl string
	if len(os.Args) < 2 {
		// TODO: eventually we want this: (or rather, pass in a community name)
		// log.Fatal().Msg("no local ipfs url specified as first argument, exiting")
		// for now:
		fsUrl = "http://localhost:5001/api/v0"
	} else {
		fsUrl = os.Args[1]
	}

	log.Info().Msg("starting notes api")
	fs, err := fs.NewFS("ipfs", fsUrl)
	if err != nil {
		log.Fatal().Msgf("failed to get file system driver: %s", err)
	}

	s := &FileSystemService{
		fs: fs,
	}

	r := mux.NewRouter()
	r.HandleFunc("/ls", s.ListFilesHandler)

	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Info().Msg("notes api listening on localhost:8000")
	log.Fatal().Err(srv.ListenAndServe())
}

type FileSystemService struct {
	fs fs.FS
}

type ListFilesResponse struct {
	DirEntries []*ListFilesDirEntry `json:"dir_entries"`
}

type ListFilesDirEntry struct {
	Name    string            `json:"name"`
	IsDir   bool              `json:"is_dir"`
	DirType int               `json:"filemode"`
	Info    ListFilesFileInfo `json:"info"`
}

type ListFilesFileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Mode    int       `json:"mode"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
}

func (s *FileSystemService) ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	files, err := s.fs.ReadDir("/")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	// TODO this is jank. this conversion happens because FS returns DirEntry interfaces, which json
	// does not know how to interpret. need a better way of deeling with this. probably concrete types in FS module
	// Another possibility is creating custom marshalers and unmarshalers

	resFiles := make([]*ListFilesDirEntry, len(files))
	for i, f := range files {
		info, _ := f.Info()
		resFiles[i] = &ListFilesDirEntry{
			Name:    f.Name(),
			IsDir:   f.IsDir(),
			DirType: int(f.Type()),
			Info: ListFilesFileInfo{
				Name:    info.Name(),
				Size:    info.Size(),
				Mode:    int(info.Mode()),
				ModTime: info.ModTime(),
				IsDir:   info.IsDir(),
			},
		}
	}

	marshaled, err := json.Marshal(resFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.WriteHeader(http.StatusOK)
	w.Write(marshaled)
}
