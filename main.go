package main

import (
	"errors"
	"labix.org/v2/mgo"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Dir struct {
	name  string
	fs    *mgo.GridFS
	flat  bool
	cache map[string]*File
}

func (d Dir) Open(name string) (http.File, error) {
	if d.flat && filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 {
		return nil, errors.New("http: invalid character in file path")
	}

	filename := filepath.Join(d.name, filepath.FromSlash(path.Clean(name)))
	f, err := d.fs.Open(filename)
	if err != nil {
		return nil, err
	}

	file := &File{f}

	return file, nil
}

type File struct {
	file *mgo.GridFile
}

func (f *File) Name() string {
	return f.file.Name()
}

func (f *File) Size() int64 {
	return f.file.Size()
}

func (f *File) Mode() os.FileMode {
	return os.FileMode(0755)
}

func (f *File) ModTime() time.Time {
	return f.file.UploadDate()
}

func (f *File) IsDir() bool {
	return false
}

func (f *File) Sys() interface{} {
	return nil
}

func (f *File) Close() error {
	return f.file.Close()
}

func (f *File) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *File) Readdir(count int) ([]os.FileInfo, error) {
	return nil, errors.New("http: directory listing not implemented")
}

func (f *File) Read(b []byte) (int, error) {
	return f.file.Read(b)
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

func main() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}

	db := session.DB("fs")
	fs := db.GridFS("fs")
	d := &Dir{"/", fs, false, map[string]*File{}}
	http.Handle("/", http.FileServer(d))

	err = http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		panic(err)
	}
}
