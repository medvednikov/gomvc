# gomvc - a tiny MVC web framework for Go #

This is a simple MVC-ish web framework for Go, which is basically a small
wrapper around Go's net/http.

## Key features ##

The key features of gomvc are ease of use, lack of clutter, and a very simple
way to quickly define actions and parameters without extra routing and
configuration files: a function declaration is enough.

Compare using net/http, beego, and gomvc to implement a simple user search page:

```go
// gomvc
func (c *Home) UserSearch(name string, age int) {
	user := findByNameAndAge(name, age)
	c.Render(user)
}

func main() {
	gomvc.Route("/", &Home{})
	gomvc.Run(":8088", true)
}
```

```go
// net/http
func userSearch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	age, _ := strconv.Atoi(r.URL.Query().Get("age"))

	user := findByNameAndAge(name, age)

	t, _ := template.ParseFiles("usersearch.html")
	t.Execute(w, user)
}

func main() {
	http.HandleFunc("/UserSearch", userSearch)
	http.ListenAndServe(":8088", nil)
}
```

```go
// beego
func (this *MainController) Get() {
	name := this.GetString("name")
	age := int(this.GetInt("age"))
	user := findByNameAndAge(name, age)
	this.Data["user"] = user
}

func main() {
	beego.Router("/UserSearch", &MainController{})
	beego.Run()
}
```

## Installation ##

    # Install the package:
    go get github.com/medvednikov/gomvc
	    
    // Use in your code:
    import "github.com/medvednikov/gomvc"

## Quick start ##
You can run the quick start example with

```
cd $GOPATH/src/github.com/medvednikov/gomvc/examples/quickstart &&
go run quickstart.go
```

Now visit [http://localhost:8088](http://localhost:8088)

```go
// c/home.go
package c

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

// main.go
package main

import (
	"github.com/medvednikov/gomvc"
	. "./c"
)

func main() {
	gomvc.Route("/", &Home{})
	gomvc.Run(":8088", true)
}
```


## API Documentation ##

Full godoc output from the latest code in master is available here:

http://godoc.org/github.com/medvednikov/gomvc



## Examples ##

More examples will be added here soon...




 
