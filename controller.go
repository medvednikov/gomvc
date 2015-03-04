package gomvc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Controller is the core type of gomvc
type Controller struct {
	Request *http.Request
	Out     http.ResponseWriter

	// Params is a map of all query string key-vales:
	// example.com/?a=1&b=2 => map[string]string{ "a":"1", "b":"2" }
	Params map[string]string

	// Form is a map with form values submitted by a POST request
	Form map[string]string

	// Uri contains current path:
	// example.com/Account/Unsubscribe?email=1 => "Account/Unsubscribe"
	Uri string

	// ActionName is the name of the running action (method)
	ActionName string

	// ControllerName is the name of the controller subtype
	ControllerName string

	// CustomTemplateFuncs defines extra html/template functions that can
	// be run in all html templates used in this controller
	CustomTemplateFuncs template.FuncMap

	// PageTitle defines the title of the HTML page and is set in the action
	PageTitle string

	stopped bool
}

var (
	// Debug is used to determine how to display error messages. Default is
	// true, set to false when deploying. One of the easy ways to do that
	// automatically is to parse machine's hostname.
	Debug bool = true

	// Gorilla router. Used for parsing url variables like /member/{id}
	router *mux.Router = mux.NewRouter()

	// A global map with all actions' argument names. They are fetched from
	// the source files since it's impossible to get argument names via
	// reflect. Example:
	// func (c *Home) Register(name string, email string)
	// ActionArgs["Home"]["Register"] = [ "name", "email" ]
	ActionArgs map[string]map[string][]string

	TimeStamp int64

	// Template cache. Once a template file is parsed, the result is saved
	// for future use to improve performance
	templateCache = make(map[string]*template.Template, 0)

	allTemplates *template.Template

	AssetFunc func(string) ([]byte, error)
)

// GetHandler generates a net/http handler func from a controller type.
// A new controller instance is created to handle incoming requests.
// Example:
// http.HandleFunc("/Account/", ez.GetHandler(&AccountController{}))
func GetHandler(obj interface{}) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Show a general error message on production
		if !Debug {
			defer func() {
				if r := recover(); r != nil {
					// TODO Custom error templates
					fmt.Fprintln(w, `
An unhandled error has occurred,
we have been notified about it. Sorry for the inconvenience.`)
					log.Println("gomvc Error: ", r)
					log.Println(string(debug.Stack()))
				}
			}()
		}

		// Set HTTP headers
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Fetch the type of the controller (e.g. "Home")
		typ := reflect.Indirect(reflect.ValueOf(obj)).Type()

		// Create a new controller of this type for this request
		val := reflect.New(typ)

		// Get base object c (ez.Controller), initialize it and update
		// it. It can be several 'parents' away.
		parentVal := val.Elem().Field(0)
		for parentVal.Type().Name() != "Controller" {
			parentVal = parentVal.Field(0) // TODO error if nothing was found
		}
		c := parentVal.Interface().(Controller)

		c.ControllerName = typ.Name()
		c.InitValues(w, r)
		// Since c is copy, not a pointer, need to manually update the
		// parent controller object TODO
		parentVal.Set(reflect.ValueOf(c))

		// Run the 'before action' action if it exists
		beforeAction := val.MethodByName("BeforeAction_")
		if beforeAction.IsValid() {
			beforeAction.Call([]reflect.Value{})
		}

		// c contained a copy of the parent controller, so we need to
		// re-fetch it in case it was updated in BeforeAction.
		// TODO this is ugly, maybe possible to make it a pointer
		c = parentVal.Interface().(Controller)

		// Run the actual method
		method := val.MethodByName(c.ActionName)
		runMethod(method, &c)

		// Run the 'after run' action if it exists
		afterAction := val.MethodByName("AfterAction_")
		if afterAction.IsValid() {
			afterAction.Call([]reflect.Value{})
		}
	}
}

// Route is a helper method that runs http.HandleFunc for a given path and
// controller
func Route(path string, controller interface{}) {
	if strings.Index(path, "{") == -1 {
		// General routes without variables. Ensure Gorilla mux matches
		// all children of path:
		// Route("/", ...) will also match "/Register", "/User" etc
		router.PathPrefix(path).HandlerFunc(GetHandler(controller))
	} else {
		// Custom routes with variables, no need to match children:
		// Route("/member/{id}", ...)
		router.HandleFunc(path, GetHandler(controller))
	}
}

// Index defines a default action
func (c *Controller) Index() {
	c.Say(`Welcome to gomvc! Define your own Index action:

    type Home struct {
        gomvc.Controller
    }

    func (c *Home) Index() {
        c.Write("Hello world!")
    }
    `)
}

// View executes a template corresponding to the current controller method
func (c *Controller) Render(data interface{}) {
	if c.stopped {
		return
	}

	var t *template.Template
	var err error
	templatePath := "v/" + c.ControllerName + "/" +
		stripMethodType(c.ActionName) + ".html"

	// Fetch the template from cache, if it's not there - open the file
	// and parse it
	if _, ok := templateCache[templatePath]; ok {
		//if _, ok := templateCache[templatePath]; ok && !Debug {
		t = templateCache[templatePath]
		fmt.Println("!!!! USE CACHE")
	} else {
		t0 := time.Now()
		t, err = parseTemplate(templatePath, c)
		fmt.Println("parse time", time.Now().Sub(t0))
		if err != nil {
			log.Println("Template error: ", err)
			if Debug {
				c.Write("Template error: ", err)
			}
			return
		}
		//templateCache[templatePath] = t
	}

	t0 := time.Now()

	err = t.Execute(c.Out, data)
	fmt.Println("execute time", time.Now().Sub(t0))
	if err != nil {
		log.Println("Template execution error:", err)
		if Debug {
			c.Write("Template error:", err)
		}
	}
}

// Say prints a string with a newline to http response
func (c *Controller) Say(s ...interface{}) {
	fmt.Fprint(c.Out, s...)
	fmt.Fprintln(c.Out)
}

// Write prints a string to http response
func (c *Controller) Write(s ...interface{}) {
	fmt.Fprint(c.Out, s...)
}

// EmptyHandler returns an empty handler for http.HandleFunc
// This is used to explicitely leave certain routes unprocessed.
func EmptyHandler(w http.ResponseWriter, r *http.Request) {

}

// Redirect performs an HTTP redirect to another action in the same controller
func (c *Controller) Redirect(action string) {
	http.Redirect(c.Out, c.Request, "/"+action, 302)
}

// GetCookie returns a value of the cookie with a specified key.
// If no such cookie was found, an empty string is returned.
func (c *Controller) GetCookie(key string) string {
	res, _ := c.Request.Cookie(key)
	if res == nil {
		return ""
	}
	return res.Value
}

// SetCookie creates a new cookie valid for 10 days
func (c *Controller) SetCookie(key string, value string) {
	http.SetCookie(c.Out, &http.Cookie{
		Name:    key,
		Value:   value,
		Expires: time.Now().Add(10 * 24 * time.Hour),
		Path:    "/",
	})
}

func (c *Controller) DeleteCookie(key string) {
	http.SetCookie(c.Out, &http.Cookie{Name: key, Value: "", MaxAge: -1})
}

func (c *Controller) SetContentType(ct string) {
	c.Out.Header().Set("Content-Type", ct)
}

func (c *Controller) SetHeader(header, value string) {
	c.Out.Header().Set(header, value)
}

func (c *Controller) IsAjax() bool {
	h := c.Request.Header["X-Requested-With"]
	return len(h) > 0 && h[0] == "XMLHttpRequest"
}

func (c *Controller) RenderError(msg string, code int) {
	http.Error(c.Out, msg, code)
	c.stopped = true
}

// Abort stops execution of the current action immediately
func (c *Controller) Abort() {
	c.stopped = true
}

// ReturnJson returns a marshaled json object with content type 'application/json'.
// This is usually used for responding to AJAX requests.
func (c *Controller) RenderJson(model interface{}) {
	if c.stopped {
		return
	}

	c.SetContentType("application/json")

	obj, err := json.Marshal(model)
	if err != nil {
		log.Println(err)
		return
	}

	c.Write(string(obj))
}

func (c *Controller) RenderJsonError(errorMsg string) {
	if c.stopped {
		return
	}

	c.SetContentType("application/json")

	j := struct {
		ErrorMsg string
		Status   string
	}{
		ErrorMsg: errorMsg,
		Status:   "FAIL",
	}

	obj, err := json.Marshal(j) // TODO
	if err != nil {
		fmt.Println(err)
		return
	}

	c.Write(string(obj))
}

func (c *Controller) RenderJsonRedirect(redirectUrl string) {
	c.SetContentType("application/json")

	j := struct {
		RedirectUrl string
		Status      string
	}{
		Status:      "OK",
		RedirectUrl: redirectUrl,
	}

	obj, _ := json.Marshal(j)
	c.Write(string(obj))
}

func staticPrefix(prefix, dir string) http.Handler {
	return http.StripPrefix("/"+prefix+"/",
		http.FileServer(http.Dir(dir)))
}

func ServeStatic(prefix, dir string) {
	http.Handle("/"+prefix+"/", staticPrefix(prefix, dir))
}

// Run initializes starts the web server
func Run(port string, isDebug bool) {
	Debug = isDebug
	TimeStamp = time.Now().Unix()
	fmt.Println("Starting a gomvc app on port ", port, " with debug=", Debug)
	getActionsFromSourceFiles()
	http.Handle("/", router)
	if port != "" {
		fmt.Println(http.ListenAndServe(port, nil))
	}
}

//////// private methods ////////

// getActionFromUri fetches an action name from uri:
// "AccountController/Settings" => "Settings"
// "Index" => "Index"
// "" => "Index"
// "Home/Register" => "Register"
// "Forum/Topic/Hello-world/234242 => "Topic"
func getActionFromUri(uri, controller string) string {
	// Root action
	if uri == "" {
		return "Index"
	}

	values := strings.Split(strings.Trim(uri, "/"), "/")
	actionName := values[0]

	// http://example.com/Controller/Action
	if len(values) > 1 { // TODO this is ugly
		if controller == "Home" {
			actionName = values[0] // Save action, controller is skipped

		} else {
			actionName = values[1]
		}
	} else if len(values) == 1 && capitalize(actionName) == controller {
		// /Action => /Action/Index
		actionName = "Index"
	}

	// Capitalize and remove unallowed characters
	actionName = capitalize(actionName)
	actionName = strings.Replace(actionName, ".", "", -1)
	actionName = replaceDashes(actionName)

	return actionName
}

// initValues parses the http.Request object and fetches all necessary values
// for gomvc.Controller
func (c *Controller) InitValues(w http.ResponseWriter, r *http.Request) {
	c.Out = w
	c.Request = r
	values := r.URL.Query()
	c.Uri = r.URL.Path[1:]
	c.ActionName = getActionFromUri(c.Uri, c.ControllerName)
	if r.Method != "GET" {
		c.ActionName += "_" + r.Method
	}
	c.PageTitle = ""

	// Generate query string map (Params)
	c.Params = make(map[string]string)
	for key, _ := range values {
		c.Params[key] = values.Get(key)
	}

	// Assign routing variables to Params
	for key, value := range mux.Vars(r) {
		c.Params[key] = value
	}

	// Generate form data
	c.Form = make(map[string]string)
	c.Request.ParseForm()
	for key, _ := range c.Request.PostForm {
		c.Form[key] = c.Request.PostForm.Get(key)
	}
}

func (c *Controller) checkMethodType() bool {
	types := []string{"POST", "PUT", "DELETE"}
	for _, t := range types {
		if strings.Index(c.ActionName, t) > -1 &&
			c.Request.Method != t {
			c.Write(t, " expected")
			return false
		}
	}
	return true
}

// runMethod runs a specified controller action (method)
func runMethod(method reflect.Value, c *Controller) {
	if !method.IsValid() {
		http.NotFound(c.Out, c.Request)
		if Debug {
			c.Write("Unknown action '" + c.ActionName +
				"' (controller: '" + c.ControllerName + "')")
		}
		return
	}

	if !c.checkMethodType() {
		return
	}

	if c.stopped {
		return
	}

	// Run it via reflect
	values := make([]reflect.Value, 0)

	// Loop thru all method args and assign query string parameters to them
	for i, argName := range ActionArgs[c.ControllerName][c.ActionName] {
		// Get value from the query string (params)
		// Register(name, password string) => /Register?name=a;password=b
		stringValue := c.Params[argName]

		// Convert this argument to a value of a certain type (Form,
		// string, int)
		argType := method.Type().In(i)
		values = append(values, c.argToValue(stringValue, argType))
	}

	// TODO handle empty values
	//fmt.Println(c.ControllerName, c.ActionName, values, dump(ActionArgs))
	method.Call(values)
}

// argToValue generates a reflect.Value from an argument type and its
// corresponding query string or form value
func (c *Controller) argToValue(stringValue string, argType reflect.Type) reflect.Value {
	// Handle a struct pointer, this must be a form
	if argType.Kind() == reflect.Ptr && argType.Elem().Kind() == reflect.Struct {
		// Dereference the form
		argType = argType.Elem()

		// Create a new form object
		newFormObj := reflect.New(argType)

		// Set all its fields
		for i := 0; i < argType.NumField(); i++ {
			field := newFormObj.Elem().Field(i)
			fieldName := argType.Field(i).Name // e.g. "Id", "Title"
			formValue := c.Form[decapitalize(fieldName)]

			switch field.Type().Name() {
			case "int":
				field.SetInt(int64(toint(formValue)))
			case "string":
				field.SetString(formValue)
			}
		}

		return newFormObj
	} else if argType.Name() == "int" {
		// Convert to int if this argument is an int, otherwise leave
		// it as a string TODO more types?
		return reflect.ValueOf(toint(stringValue))
	} else {
		return reflect.ValueOf(stringValue)
	}
	return reflect.Value{}
}

// Custom html/template functions
var defaultFuncs = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	"mul": func(a, b int) int { return a * b },
	"inc": func(n int) int { return n + 1 },
	"tojson": func(i interface{}) template.JS {
		out, _ := json.Marshal(i)
		res := template.JS(out)
		return res
	},
	"js": func(file string) template.HTML {
		if strings.Index(file, "//") == -1 {
			file = "/js/" + file
		}
		file += fmt.Sprintf("?%d", TimeStamp)
		return template.HTML("<script src='" + file + "'></script>")
	},
	"css": func(file string) template.HTML {
		if strings.Index(file, "//") == -1 {
			file = "/css/" + file
		}
		file += fmt.Sprintf("?%d", TimeStamp)
		return template.HTML("<link href='" + file + "' rel='stylesheet'>")
	},
	"staticcss": func(file string) template.HTML {
		if strings.Index(file, "//") == -1 {
			file = "/css/" + file
		}
		return template.HTML("<link href='" + file + "' rel='stylesheet'>")
	},
	"staticjs": func(file string) template.HTML {
		if strings.Index(file, "//") == -1 {
			file = "/js/" + file
		}
		return template.HTML("<script src='" + file + "'></script>")
	},
}

func ConvertTemplates(indir, outdir string) error {
	//in, err := os.Open(indir)
	//if err != nil {
	//return err
	//}

	//fis, err := in.Readdir(-1)
	//if err != nil {
	//return err
	//}

	filepath.Walk(indir, func(path string, fi os.FileInfo, err error) error {
		fmt.Println("PATH=", path)
		if fi.IsDir() {
			os.Mkdir(outdir+"/"+path, 0755)
			return nil
		}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		s := string(b)

		s = convertTemplate(s)

		out, err := os.Create(outdir + "/" + path)
		if err != nil {
			return err
		}
		out.Write([]byte(s))
		out.Close()

		return nil
	})

	/*
		for _, fi := range fis {
			fmt.Println("FI NAME=", fi.Name())

			if fi.IsDir() {
				os.Mkdir(outdir+"/"+fi.Name(), 0755)
				continue
			}
			b, err := ioutil.ReadFile(indir + "/" + fi.Name())
			if err != nil {
				return err
			}
			s := string(b)

			s = convertTemplate(s)

			out, err := os.Create(outdir + "/" + fi.Name())
			if err != nil {
				return err
			}
			out.Write([]byte(s))
			out.Close()

		}
	*/
	return nil

}

func convertTemplate(s string) string {
	// Title
	//s = strings.Replace(s, "$_ez_TITLE", c.PageTitle, -1)
	s = strings.Replace(s, "\n\n", "\n", -1)
	s = strings.Replace(s, "\t", "", -1)
	s = strings.Replace(s, "  ", "", -1)

	// Comments
	// @* .... *@ ==> {{/*}} .... {{*/}}
	r := regexp.MustCompile(`@\*(.*?)\*@`)
	s = r.ReplaceAllString(s, "")

	// @.
	r = regexp.MustCompile(`@\.`)
	s = r.ReplaceAllString(s, "{{.}}")

	// @if a
	// ===> {{ if a }}
	r = regexp.MustCompile("@(if|else|end|range|template|define)(.*?)\n")
	s = r.ReplaceAllString(s, "{{ $1 $2 }}\n")

	// @Name (always starts with a capital letter)
	// ===> {{ .Name }}
	r = regexp.MustCompile("@([A-Z][a-zA-Z\\.]+)")
	s = r.ReplaceAllString(s, "{{.$1}}")

	// @func "param" (always starts with a small letter)
	// ===> {{ func "param" }}
	r = regexp.MustCompile(`@([a-z][a-zA-Z\\.]+( "[^"]+")*)`)
	s = r.ReplaceAllString(s, "{{ $1 }}")

	// %translation_tag
	// ===> {{ T "translation_tag" }}
	r = regexp.MustCompile("%([a-zA-Z_0-9]+)")
	s = r.ReplaceAllString(s, `{{ T "$1" }}`)

	return s
}

// parseTemplate parses a provided html template file, applies all custom
// structures and functions, and returns a *template.Template object
func parseTemplate(file string, c *Controller) (*template.Template, error) {
	curdir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	// Read layout.html unless it's an AJAX request
	layoutStr := ""
	if !c.IsAjax() {
		layout, err := AssetFunc("v/layout.html")
		if err != nil || Debug {
			layout, err = ioutil.ReadFile("v/layout.html")
		}
		if err != nil {
			log.Println("Template layout not found", curdir)
		}

		layoutStr = string(layout)
	}

	s := ""

	data, err := AssetFunc(file)
	//fmt.Println("GOT ASSET", len(data), err)
	if err == nil && len(data) > 0 && !Debug {
		fmt.Println("GOT FROM BINARY")
		s = string(data)
	} else {

		// Read template file
		b, err := ioutil.ReadFile(file)
		if err != nil {
			log.Println("Template '", file, "' is not found!", curdir)
		}
		s = string(b)
	}

	// Embed the template into layout
	t0 := time.Now()
	if layoutStr != "" {
		s = strings.Replace(layoutStr, "$BODY", s, -1)
	}

	// Title
	s = strings.Replace(s, "$_ez_TITLE", c.PageTitle, -1)
	s = strings.Replace(s, "\n\n", "\n", -1)
	s = strings.Replace(s, "\t", "", -1)
	s = strings.Replace(s, "  ", "", -1)

	// Comments
	// @* .... *@ ==> {{/*}} .... {{*/}}
	r := regexp.MustCompile(`@\*(.*?)\*@`)
	s = r.ReplaceAllString(s, "")

	// @.
	r = regexp.MustCompile(`@\.`)
	s = r.ReplaceAllString(s, "{{.}}")

	// @if a
	// ===> {{ if a }}
	r = regexp.MustCompile("@(if|else|end|range|template|define)(.*?)\n")
	s = r.ReplaceAllString(s, "{{ $1 $2 }}\n")

	// @Name (always starts with a capital letter)
	// ===> {{ .Name }}
	r = regexp.MustCompile("@([A-Z][a-zA-Z\\.]+)")
	s = r.ReplaceAllString(s, "{{.$1}}")

	// @func "param" (always starts with a small letter)
	// ===> {{ func "param" }}
	r = regexp.MustCompile(`@([a-z][a-zA-Z\\.]+( "[^"]+")*)`)
	s = r.ReplaceAllString(s, "{{ $1 }}")

	// $translation_tag
	// ===> {{ T "translation_tag" }}
	r = regexp.MustCompile("%([a-zA-Z_0-9]+)")
	s = r.ReplaceAllString(s, `{{ T "$1" }}`)

	// Custom funcs
	t := template.New(file).Funcs(defaultFuncs).Funcs(c.CustomTemplateFuncs)
	//t := template.New(file).Funcs(defaultFuncs)
	fmt.Println("T replaces", time.Now().Sub(t0))

	t2, err := t.Parse(s)
	fmt.Println("T replaces", time.Now().Sub(t0))
	return t2, err
}

// getActionsFromSourceFiles parses all controller source files and fetches
// data about action functions
func getActionsFromSourceFiles() {
	// Parse files only on development box (when Debug is true). Production
	// boxes should not have source files and the extra overhead. The
	// actions are 'cached'.
	if !Debug {
		return
	}

	ActionArgs = make(map[string]map[string][]string, 0)

	// Parse the controllers directory (it should be in the same directory)
	curdir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	files, err := ioutil.ReadDir("c/")
	if err != nil {
		panic(`
Can't find the controllers directory in debug mode.
Make sure you are running the application from the directory where it's located
Current directory: ` + curdir)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".go") &&
			!strings.HasSuffix(file.Name(), "_test.go") {
			getActionsFromSourceFile(file.Name())
		}
	}

	// Cache actions data for production use
	out, _ := os.Create("autogen/autogen.go")
	fmt.Fprintln(out, `
// This file has been generated automatically. Do not modify it.
package autogen

import "github.com/medvednikov/gomvc"

func init() {
	gomvc.ActionArgs = `+dump(ActionArgs)+`
}`)
	out.Close()
}

func getActionsFromSourceFile(sourceFile string) {
	b, err := ioutil.ReadFile("c/" + sourceFile)
	handle(err)
	source := string(b)

	pos := strings.Index(sourceFile, ".go")
	if pos == -1 {
		return
	}

	controllerName := capitalize(sourceFile[:pos])
	ActionArgs[controllerName] = make(map[string][]string, 0)

	// Search for "func (...) ActionName(...) {"
	r := regexp.MustCompile(`func \([a-zA-Z]+ \*` + controllerName + `\) (.*?)\((.*?)*\) {`)

	matches := r.FindAllStringSubmatch(source, -1)
	for _, match := range matches {
		functionName := match[1]
		if functionName == "" {
			continue
		}
		args := []string{}

		// match[2] contains arguments (["a int", "b string"])
		if len(match) > 2 && match[2] != "" {
			args = strings.Split(match[2], ", ")

			// Get rid of type (for now)
			for i, arg := range args {
				if arg == "" {
					continue
				}
				args[i] = strings.Split(arg, " ")[0]
			}
		}

		ActionArgs[controllerName][functionName] = args
	}
}

//////// helper functions////////

func dump(val interface{}) string {
	return fmt.Sprintf("%#v", val)
}

func toint(s string) int {
	res, _ := strconv.Atoi(s)
	return res
}

// capitalize capitalizes a string: 'test' => 'Test'
func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// decapitalize does the opposite of capitalize(): 'Test' => 'test'
func decapitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func stripMethodType(action string) string {
	return strings.Replace(action, "_POST", "", -1)
}

// hello-world => helloWorld
func replaceDashes(action string) string {
	if strings.Index(action, "-") == -1 {
		return action
	}

	var res bytes.Buffer
	for i := 0; i < len(action); i++ {
		if action[i] == '-' {
			if i < len(action)-1 {
				res.WriteString(strings.ToUpper(
					string(action[i+1])))
				i++
			}
		} else {
			res.WriteString(string(action[i]))
		}
	}
	return res.String()
}

func handle(err error) {
	if err != nil {
		panic(err)
	}
}
