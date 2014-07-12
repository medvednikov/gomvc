# ezweb - a tiny Go web framework #

This is an early release of a simple MVC-ish web framework for Go.

It's basically a small wrapper around Go's net/http.

It supports PostgreSQL via [gorp](https://github.com/coopernurse/gorp), but more databases will be
added in the future.

This is an alpha release missing several key features and is not recommended for use in production.


## Quick start ##

```go
// run with 'go run examples/quickstart.go' and visit http://localhost:8088/
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


## Key features ##

The key features of ezweb are ease of use, lack of clutter, and a very simple way to quickly define
actions and parameters without extra routing and configuration files: a function declaration is
enough.

Compare using net/http, Beego, and ezweb to implement a simple user search page:

```go
// net/http
func userSearch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	age, _ := strconv.Atoi(r.URL.Query().Get("age"))

	user := usersRepo.FindByNameAndAge(name, age)

	t, _ := template.ParseFiles("usersearch.html")
	t.Execute(w, user)
}

func main() {
	http.HandleFunc("/UserSearch", userSearch)
	http.ListenAndServe(":8088", nil)
}


// Beego
func (this *MainController) Get() {
	name := this.GetString("name")
	age := int(this.GetInt("age"))
	user := usersRepo.FindByNameAndAge(name, age)
	this.Data["user"] = user
}

func main() {
	beego.Router("/UserSearch", &MainController{})
	beego.Run()
}


// ezweb
func (c *Home) UserSearch(name string, age int) {
	user := usersRepo.FindByNameAndAge(name, age)
	c.View(user)
}

func main() {
	ez.Route("/", &Home{})
	http.ListenAndServe(":8088", nil)
}
```


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



## Examples ##

More examples will be added here soon...




 
