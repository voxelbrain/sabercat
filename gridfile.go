package sabercat

import (
	"fmt"
	"labix.org/v2/mgo"
	"os"
	"time"
)

type GridFile struct {
	file *mgo.GridFile
}

// os.FileInfo

func (gf *GridFile) Name() string {
	return gf.file.Name()
}

func (gf *GridFile) Size() int64 {
	return gf.file.Size()
}

func (gf *GridFile) Mode() os.FileMode {
	return os.FileMode(0755)
}

func (gf *GridFile) ModTime() time.Time {
	return gf.file.UploadDate()
}

func (gf *GridFile) IsDir() bool {
	return false
}

func (gf *GridFile) Sys() interface{} {
	return nil
}

// http.File

func (gf *GridFile) Close() error {
	return gf.file.Close()
}

func (gf *GridFile) Stat() (os.FileInfo, error) {
	return gf, nil
}

func (gf *GridFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("directory listing not implemented")
}

func (gf *GridFile) Read(b []byte) (int, error) {
	return gf.file.Read(b)
}

func (gf *GridFile) Seek(offset int64, whence int) (int64, error) {
	return gf.file.Seek(offset, whence)
}
