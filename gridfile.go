package sabercat

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"os"
	"time"
)

type gridFile struct {
	*mgo.GridFile
}

// Implementation of os.FileInfo
func (gf *gridFile) Name() string {
	return gf.GridFile.Name()
}

func (gf *gridFile) Size() int64 {
	return gf.GridFile.Size()
}

func (gf *gridFile) Mode() os.FileMode {
	return os.FileMode(0755)
}

func (gf *gridFile) ModTime() time.Time {
	return gf.GridFile.UploadDate()
}

func (gf *gridFile) IsDir() bool {
	return false
}

func (gf *gridFile) Sys() interface{} {
	return nil
}

// Implementation of net/http.File
func (gf *gridFile) Stat() (os.FileInfo, error) {
	return gf, nil
}

func (gf *gridFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("directory listing not implemented")
}
