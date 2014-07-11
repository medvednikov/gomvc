# ezweb - a tiny Go web framework #

This is an early release of a tiny and simple MVC-ish web framework for Go.

It's basically a small wrapper around Go's net/http.

It supports Postgres via gorp, but more databases will be added in the future.

This is an alpha release missing several key features and is not recommended for use in production.

## Installation ##

    # Install dependencies:
    go get github.com/coopernurse/gorp

    # Install the package:
    go get github.com/medvednikov/ezweb
	    
    // Use in your code:
    import "github.com/medvednikov/ezweb"


## API Documentation ##

Full godoc output from the latest code in master is available here:

http://godoc.org/github.com/medvednikov/ezweb


## Quick start ##

```go
// run with 'go run examples/example.go' and visit http://localhost:8088/
package main

import (
	ez "github.com/medvednikov/ezweb"
	"net/http"
)

type Home struct {
	ez.Controller
}

func (c *Home) Index(name string) {
	if name == "" {
		name = "stranger"
	}
	c.Write("Hello, ", name, "! :)")
}

func main() {
	ez.Route("/", &Home{}) // or
	// http.HandleFunc("/", ez.GetHandler(&Home{}))

	http.ListenAndServe(":8088", nil)
}
```

## Examples ##

```go

```







 
