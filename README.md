# Middlewarex

[![CircleCI](https://circleci.com/gh/mdouchement/middlewarex/tree/master.svg?style=shield)](https://circleci.com/gh/mdouchement/middlewarex/tree/master)
[![cover.run](https://cover.run/go/github.com/mdouchement/middlewarex.svg?style=flat&tag=golang-1.10)](https://cover.run/go?tag=golang-1.10&repo=github.com%2Fmdouchement%2Fmiddlewarex)
[![Go Report Card](https://goreportcard.com/badge/github.com/mdouchement/middlewarex)](https://goreportcard.com/report/github.com/mdouchement/middlewarex)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/mdouchement/middlewarex)
[![License](https://img.shields.io/github/license/mdouchement/middlewarex.svg)](http://opensource.org/licenses/MIT)

Bunch of middlewares for [Labstack Echo](https://github.com/labstack/echo).

## Requirements

- Echo 3.x.x

## Usage
### CRUD

Used for creating RESTful API entrypoints according the given struct.

### Versioning

This middleware must be set as a _pre_ middleware.

Doing the following request:
```
X-Application-Version: vnd.github.v3
GET /toto
```

will be rewritten as:
```
GET /v3/toto
```

Bechmarks
```
[middlewarex]>> go test -bench .
goos: darwin
goarch: amd64
pkg: github.com/mdouchement/middlewarex
BenchmarkVersioningRW-8     	  300000	     50608 ns/op   // With header rewriting
BenchmarkVersioningVRwM-8   	 5000000	       362 ns/op   // Just versioned routes with the Versioning middleware present
BenchmarkVersioningVR-8     	10000000	       215 ns/op   // Just versioned routes without the Versioning middleware
PASS
ok  	github.com/mdouchement/middlewarex	19.850s
```

If the header is not specified, no rewrittes are applied.

## License

**MIT**


## Contributing

All PRs are welcome.

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request

As possible, run the following commands to format and lint the code:

```sh
# Format
find . -name '*.go' -not -path './vendor*' -exec gofmt -s -w {} \;

# Lint
golangci-lint run -c .golangci.yml
```
