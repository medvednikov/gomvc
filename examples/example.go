package main

import (
	//ez "github.com/medvednikov/ezweb"
	ez "medved/ezweb"
	"net/http"
)

type BaseC struct {
	ez.Controller
}

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
