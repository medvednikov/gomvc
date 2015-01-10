package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

func main() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println(flag.Args())
		log.Fatalln("Usage: repogen [type]")

	}

	typ := flag.Args()[0]

	s := `package m

import "github.com/medvednikov/gomvc/repo"

type TYPE struct {
	Id int
}

type TYPEs []*TYPE

func RetrieveTYPE(id int) *TYPE {
	var o *TYPE
	repo.Retrieve(&o, id)
	return o
}

func AddTYPE(o *TYPE) {
	repo.Add(o)
}

func UpdateTYPE(o *TYPE) {
	repo.Update(o)
}
`
	s = strings.Replace(s, "TYPE", typ, -1)
	fmt.Println(s)

}
