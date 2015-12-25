# gomvc - a tiny MVC web framework #

This is a simple MVC-ish web framework, which is basically a small
wrapper around Go's net/http.

## Key features ##

Ease of use, lack of clutter, and a very simple
way to quickly define actions and parameters without extra routing and
configuration files: a function declaration is enough.

Compare using net/http, beego, and gomvc to implement a simple user search page:

```go
// gomvc
func (c *Home) UserSearch(name string, age int) gomvc.View {
	user := findByNameAndAge(name, age)
	return c.View(user)
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
	http.HandleFunc("/user-search", userSearch)
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
	beego.Router("/user-search", &MainController{})
	beego.Run()
}
```

## Installation ##

    # Install the package:
    go get github.com/medvednikov/gomvc
    
    # Install the command-line tool
    go install github.com/medvednikov/gomvc/cmd/gomvc
	    
    // Use in your code:
    import "github.com/medvednikov/gomvc"

## Quick start ##

```
gomvc new mywebapp
cd mywebapp
go run cmd/main.go
```

Now visit [http://localhost:8080](http://localhost:8080)



## API Documentation ##

Full godoc output from the latest code in master is available here:

http://godoc.org/github.com/medvednikov/gomvc



## Examples ##

More examples will be added here soon...




 
