package main

import (
	"github.com/superp00t/etc"
	"flag"
	"log"
	"net/http"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type IndexFS struct {
	pathfs.FileSystem

	hpath string
	c     *http.Client
}

/* HTTP Methods
   ============
*/
func (i *IndexFS) head(url string) int64 {
	r, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return -1
	}

	rsp, err := i.c.Do(r)
	if err != nil {
		return -1
	}

	if strings.HasSuffix(url, "/") && strings.Contains(rsp.Header.Get("Content-Type"), "text/html") {
		return -2
	}

	if rsp.Status != 200 {
		return -1
	}

	n := rsp.ContentLength

	return n
}

func (i *IndexFS) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	g := i.head(i.hpath + name)

	if g == -2 {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0644
		}, fuse.OK
	}

	if g < 0 {
		return nil, fuse.ENOENT
	}

	return &fuse.Attr {
		Mode: fuse.S_IFREG | 0644, 
		Size: uint64(g),
	}, fuse.OK
}

func (me *IndexFS) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
	r, err := http.NewRequest("GET", hpath + name, nil)
	if err != nil {
		return nil, fuse.ENOENT
	}

	rsp, err := i.c.Do(r)
	if err != nil {
		return  nil, fuse.ENOENT
	}

	e := etc.NewBuffer()
	io.Copy(rsp.Body, e)

	c := e.ToString()

	var de []fuse.DirEntry

	cz, err := goquery.NewDocumentFromString(c)
	if err != nil {
		yo.Fatal(err)
	}

	cz.Find("a").Each(func(i int, s *goquery.Selection) {
		nname := s.Text()
		if nname != "../" {
			if strings.HasSuffix(nname, "/") {
				de = append(de, fuse.DirEntry{
					Name: name,
					Mode: fuse.S_IFDIR,
				})
			} else {
				de = append(de, fuse.DirEntry{
					Name: name,
					Mode: fuse.S_IFREG,
				})
			}

		}
	})

	return de, fuse.OK
}

func (me *IndexFs) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	g := me.head(me.url + name)
	if g < 0 {
		return nil, fuse.ENOENT
	}

	if flags&fuse.O_ANYWRITE != 0 {
		return nil, fuse.EPERM
	}

	f := new(hFile)
	f.url = me.url + name
	f.c = me.c
	f.size = g
	f.File = nodefs.NewDefaultFile()

	return f, fuse.OK
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Usage:\n  hello MOUNTPOINT")
	}
	nfs := pathfs.NewPathNodeFs(&HelloFs{FileSystem: pathfs.NewDefaultFileSystem()}, nil)
	server, _, err := nodefs.MountRoot(flag.Arg(0), nfs.Root(), nil)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}
	server.Serve()
}
