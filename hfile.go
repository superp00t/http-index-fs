package main

import (
	"fmt"
	"io"
	"net/http"
	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse"
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
	out.Size = uint64(f.size)
	return fuse.OK
}

func (f *hFile) Read(buf []byte, off int64) (res fuse.ReadResult, code fuse.Status) {
	r, err := http.NewRequest("GET", f.url, nil)
	if err != nil {
		yo.Fatal(err)
		return nil, fuse.ENOENT
	}

	ln := int64(len(buf))

	r.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", off, off+ln))

	yo.Println("Requesting bytes", f.url, off, "-", off+ln)
	rsp, err := f.c.Do(r)
	if err != nil {
		yo.Fatal(err)
		return nil, fuse.ENOENT
	}

	b := etc.NewBuffer()
	_, err = io.Copy(b, rsp.Body)		
	if err != nil {
		yo.Fatal(err)
	}
	
	copy(buf, b.Bytes())

	return fuse.ReadResultData(buf), fuse.OK
}
