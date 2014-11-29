package main

import (
	. "./controllers"
	ez "github.com/medvednikov/ezweb"
)

func main() {
	ez.Route("/", &Home{})
	ez.Run(":8088", true)
}
