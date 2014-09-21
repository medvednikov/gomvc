package main

import (
	ez "github.com/medvednikov/ezweb"
	"github.com/medvednikov/ezweb/examples/quickstart"
)

func main() {
	ez.Route("/", &quickstart.Home{})
	ez.Start(":8088", true)
}
