package controllers

import ez "github.com/medvednikov/ezweb"

type Home struct {
	ez.Controller
}

func (c *Home) Index(name string) {
	if name == "" {
		name = "sstranger"
	}
	c.Write("Hello, ", name, "! :)")
}
