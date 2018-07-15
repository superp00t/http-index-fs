*Warning, this software is experimental may contain yet undetected vulnerabilities. Use at your own risk.*

# Usage

```bash
# install
$ go get -u -v github.com/superp00t/http-index-fs

$ http-index-fs http://localhost:80/ /path/to/mountpoint
```

# Caveats

- I wrote this for the purpose of exposing my media server to Kodi. If it doesn't work for your use case, open an issue or submit a Pull Request.
- Generates **a lot** of HTTP requests. You probably don't want to be using this on any infrastructure you don't own.
- Tested only on NGINX (Almost certainly won't work with Apache.)
- May behave improperly when your any folder in your source HTTP directory contains an index.html file, which will prevent http-index-fs from accessing all the folder's contents.

# Library credits

- [Go Programming Language](https://golang.org/)
- [FUSE Go bindings](https://github.com/hanwen/go-fuse) provides the filesystem virtualization interface
- [GoQuery](https://github.com/PuerkitoBio/goquery) extracts the anchor links in Web index pages
- [Pflag](https://github.com/ogier/pflag, makes CLI programming easier)
