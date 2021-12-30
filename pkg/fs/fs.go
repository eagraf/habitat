package fs

import (
	"fmt"
	"time"
)

type FileMode uint32

// File interface loosely based off of os/fs File in go 1.16
type File interface {
	Stat() (FileInfo, error)
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
}

// FileInfo interface loosely based off of os/fs FileInfo in go 1.16
type FileInfo interface {
	Name() string       // base name of the file
	Size() int64        // length in bytes for regular files; system-dependent for others
	Mode() FileMode     // file mode bits
	ModTime() time.Time // modification time
	IsDir() bool        // abbreviation for Mode().IsDir()
	//	Sys() interface{}   // underlying data source (can return nil)
}

type DirEntry interface {
	Name() string
	IsDir() bool
	Type() FileMode
	Info() (FileInfo, error)
}

// FS interface loosely based off of os/fs FS interface in go 1.16
type FS interface {
	Open(name string) (File, error)
	ReadDir(name string) ([]DirEntry, error)
}

const (
	IPFSType = "ipfs"
)

func NewFS(fsType string, fsUrl string) (FS, error) {
	switch fsType {
	case IPFSType:
		// TODO this api url is hardcoded right now, we need to create a better way for configuring it
		return NewIPFS(fsUrl)
	default:
		return nil, fmt.Errorf("invalid FS type: %s", fsType)
	}
}
