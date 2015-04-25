package main

import (
	. "./c"
	"github.com/medvednikov/gomvc"
)

func main() {
	gomvc.Route("/", &Home{})
	gomvc.Run(&gomvc.Config{
		Port:  "8088",
		IsDev: true,
	})
}
