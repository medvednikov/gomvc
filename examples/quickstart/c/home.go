package controllers

import "github.com/medvednikov/gomvc"

type Home struct {
	gomvc.Controller
}

func (c *Home) Index(name string) {
	if name == "" {
		name = "stranger"
	}
	c.Write("Hello, ", name, "! :)")
}
