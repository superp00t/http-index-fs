package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/superp00t/etc"
)

// hFile provides a FUSE file interface for a staticly sized HTTP resource.
type hFile struct {
	c    *http.Client
	url  string
	size int64
	nodefs.File
}

func (f *hFile) String() string {
	return f.url
}

func (f *hFile) GetAttr(out *fuse.Attr) fuse.Status {
	out.Mode = fuse.S_IFREG | 0644
	out.Size = f.size
	return fuse.OK
}

func NewDataFile(url string, c *http.Client) hFile {
	f := new(hFile)
	f.File = nodefs.NewDefaultFile()
	return f
}

func (f *dataFile) Read(buf []byte, off int64) (res fuse.ReadResult, code fuse.Status) {
	r, err := http.NewRequest("GET", f)
	if err != nil {
		return nil, fuse.ENOENT
	}

	ln := int64(len(buf))

	r.Header.Set("Range", fmt.Sprintf("%d-%d", off, off+ln))

	rsp, err := f.c.Do()
	if err != nil {
		return nil, fuse.ENOENT
	}

	b := etc.NewBuffer()
	io.Copy(b, rsp.Body)

	return fuse.ReadResultData(b.Bytes()), fuse.OK
}
