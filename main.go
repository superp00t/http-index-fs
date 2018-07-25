package main

import (
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/davecgh/go-spew/spew"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/ogier/pflag"
	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
)

type IndexFS struct {
	pathfs.FileSystem

	sizes *sync.Map
	hpath string
	c     *http.Client
}

func (i *IndexFS) loadSize(name string) int64 {
	s, ok := i.sizes.Load(name)
	if !ok {
		return -1
	}

	return s.(int64)
}

/* HTTP Methods
   ============
*/
func (i *IndexFS) head(url string) int64 {
	yo.Println("HEAD", url)
	r, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		yo.Println(err)
		return -1
	}

	rsp, err := i.c.Do(r)
	if err != nil {
		yo.Println(err)
		return -1
	}

	yo.Println(url, rsp.Status)

	if rsp.StatusCode == 301 {
		yo.Println("Redirect")
		return -2
	}

	if rsp.StatusCode != 200 {
		yo.Println("invalid code", rsp.Status)
		return -1
	}

	n := rsp.ContentLength
	yo.Println("Content length", n)

	if rsp.StatusCode == 200 && n == -1 {
		return -2
	}

	return n
}

func pathEscape(str string) string {
	return strings.Replace(
		strings.Replace(
			url.QueryEscape(str),
			"%2F",
			"/",
			-1,
		), "+", "%20", -1)
}

func (i *IndexFS) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	trp := i.hpath + "/" + pathEscape(name)
	g := i.head(trp)

	if g == -1 {
		g = i.head(trp + "/")
	}

	yo.Println("G1", g, name)

	dirAttr := &fuse.Attr{
		Mode: fuse.S_IFDIR | 0644,
	}

	if name == "" {
		return dirAttr, fuse.OK
	}

	if g == -2 {
		return dirAttr, fuse.OK
	}

	if g < 0 {
		return nil, fuse.ENOENT
	}

	return &fuse.Attr{
		Mode: fuse.S_IFREG | 0644,
		Size: uint64(g),
	}, fuse.OK
}

func parseList(s string) []int64 {
	i := etc.FromString(s)

	o := []int64{}

	for {
		in, _, err := i.ReadRune()
		if err != nil {
			break
		}
	}
}

func (me *IndexFS) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
	trp := me.hpath + "/" + pathEscape(name)

	r, err := http.NewRequest("GET", trp, nil)
	if err != nil {
		return nil, fuse.ENOENT
	}

	yo.Println("(OpenDir) GET", trp)

	rsp, err := me.c.Do(r)
	if err != nil {
		return nil, fuse.ENOENT
	}

	dirBuffer := etc.NewBuffer()

	io.Copy(dirBuffer, rsp.Body)

	sizes := parseList(dirBuffer.ToString())

	var de []fuse.DirEntry

	cz, err := goquery.NewDocumentFromReader(dirBuffer)
	if err != nil {
		yo.Fatal(err)
	}

	ttl := ""

	cz.Find("title").Each(func(i int, s *goquery.Selection) {
		ttl = s.Text()
	})

	if !strings.HasPrefix(ttl, "Index of ") {
		return nil, fuse.ENOENT
	}

	cz.Find("a").Each(func(i int, s *goquery.Selection) {
		yo.Println("warn", s.Text())
		u, ok := s.Attr("href")
		if !ok {
			return
		}

		last := strings.Split(u, "/")

		lastU, err := url.QueryUnescape(last[len(last)-1])
		if err != nil {
			yo.Println(err)
			return
		}

		nname := lastU

		if s.Text() != "../" {
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
	g := me.head(me.hpath + name)
	yo.Println("(Open)", name, g)
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

	srcURL, err := url.Parse(pflag.Arg(0))
	if err != nil {
		yo.Fatal(err)
	}

	if strings.HasSuffix(srcURL.Path, "/") {
		srcURL.Path = strings.TrimRight(srcURL.Path, "/")
	}

	yo.Println("Mounting", srcURL, "to", pflag.Arg(1))
	nfs := pathfs.NewPathNodeFs(&IndexFS{
		sizes:      new(sync.Map),
		FileSystem: pathfs.NewDefaultFileSystem(),
		hpath:      srcURL.String(),
		c:          &http.Client{},
	}, nil)

	server, _, err := nodefs.MountRoot(pflag.Arg(1), nfs.Root(), nil)
	if err != nil {
		yo.Fatal("Mount fail:", err)
	}

	server.Serve()
}
