*Warning, this software is experimental may contain yet undetected vulnerabilities. Use at your own risk.*

# Usage

```bash
# install
$ go get -u -v github.com/superp00t/http-index-fs

$ http-index-fs http://localhost:80/ /path/to/mountpoint
```

# Caveats

- Tested only on NGINX.
- May behave improperly when your any folder in your source HTTP directory contains an index.html file, which will prevent http-index-fs from accessing all the folder's contents.

# Library credits

- [Go Programming Language](https://golang.org/)
- [FUSE Go bindings](https://github.com/hanwen/go-fuse)
- [Pflag](https://github.com/ogier/pflag)
