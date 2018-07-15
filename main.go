package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/davecgh/go-spew/spew"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/ogier/pflag"
	"github.com/superp00t/etc/yo"
	"net/http"
	"os/exec"
	"strings"
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
	yo.Println("HEAD", url)
	r, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return -1
	}

	rsp, err := i.c.Do(r)
	if err != nil {
		return -1
	}

	if rsp.StatusCode == 301 {
		return -2
	}

	if rsp.StatusCode != 200 {
		return -1
	}

	n := rsp.ContentLength

	return n
}

func (i *IndexFS) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	g := i.head(i.hpath + name)

	if g == -1 {
		g = i.head(i.hpath + name + "/")
	}

	if g == -2 {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0644,
		}, fuse.OK
	}

	if g < 0 {
		return nil, fuse.ENOENT
	}

	return &fuse.Attr{
		Mode: fuse.S_IFREG | 0644,
		Size: uint64(g),
	}, fuse.OK
}

func (me *IndexFS) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
	r, err := http.NewRequest("GET", me.hpath+name, nil)
	if err != nil {
		return nil, fuse.ENOENT
	}

	yo.Println("(OpenDir) GET", me.hpath+name)

	rsp, err := me.c.Do(r)
	if err != nil {
		return nil, fuse.ENOENT
	}

	var de []fuse.DirEntry

	cz, err := goquery.NewDocumentFromReader(rsp.Body)
	if err != nil {
		yo.Fatal(err)
	}

	cz.Find("a").Each(func(i int, s *goquery.Selection) {
		yo.Println("warn", s.Text())
		nname := s.Text()
		if nname != "../" {
			if strings.HasSuffix(nname, "/") {
				nname = strings.TrimRight(nname, "/")
				de = append(de, fuse.DirEntry{
					Name: nname,
					Mode: fuse.S_IFDIR,
				})
			} else {
				de = append(de, fuse.DirEntry{
					Name: nname,
					Mode: fuse.S_IFREG,
				})
			}

		}
	})

	yo.Println("(OpenDir)", spew.Sdump(de))

	return de, fuse.OK
}

func (me *IndexFS) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	yo.Println("(Open)", name)
	g := me.head(me.hpath + name)
	if g < 0 {
		return nil, fuse.ENOENT
	}

	if flags&fuse.O_ANYWRITE != 0 {
		return nil, fuse.EPERM
	}

	f := new(hFile)
	f.url = me.hpath + name
	f.c = me.c
	f.size = g
	f.File = nodefs.NewDefaultFile()

	return f, fuse.OK
}

func main() {
	pflag.Parse()
	if len(pflag.Args()) < 1 {
		yo.Fatal("Usage: http-index-fs (http url) (mount point)")
	}

	exec.Command("/bin/fusermount", "-uz", pflag.Arg(1)).Run()

	srcURL := pflag.Arg(0)
	if !strings.HasSuffix(srcURL, "/") {
		srcURL += "/"
	}

	yo.Println("Mounting", srcURL, "to", pflag.Arg(1))
	nfs := pathfs.NewPathNodeFs(&IndexFS{
		FileSystem: pathfs.NewDefaultFileSystem(),
		hpath:      srcURL,
		c:          &http.Client{},
	}, nil)
	server, _, err := nodefs.MountRoot(pflag.Arg(1), nfs.Root(), nil)
	if err != nil {
		yo.Fatal("Mount fail:", er)
	}
	server.Serve()
}
