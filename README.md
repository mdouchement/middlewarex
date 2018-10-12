# Middlewarex

[![Go Report Card](https://goreportcard.com/badge/github.com/mdouchement/middlewarex)](https://goreportcard.com/report/github.com/mdouchement/middlewarex)
[![License](https://img.shields.io/github/license/mdouchement/middlewarex.svg)](http://opensource.org/licenses/MIT)

Bunch of middlewares for [Labstack Echo](https://github.com/labstack/echo).

## Requirements

- Echo 3.x.x

## Usage
### CRUD

Used for creating RESTful API entrypoints according the given struct.


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
