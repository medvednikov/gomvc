package ezweb

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Controller is the core type of ezweb
type Controller struct {
	Request *http.Request
	Out     http.ResponseWriter

	// args is a list of all query string values:
	// example.com/?a=1&b=2 => []string{ "1", "2" }
	// It's used only for mapping method arguments: func Index(a, b string) {}
	args []string

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
}

type Form struct{}

// Debug is used to determine how to display error messages. Default is true,
// set to false when deploying. One of the easy ways to do that automatically
// is to parse machine's hostname.
var Debug bool

// GetHandler generates a net/http handler func from a controller making it
// handle all incoming requests.
// Example:
// http.HandleFunc("/Account/", ez.GetHandler(&AccountController{}))
func GetHandler(obj interface{}) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Show a general error message on production
		if !Debug {
			defer func() {
				if r := recover(); r != nil {
					// TODO Custom error templates
					fmt.Fprintln(w,
						`An unhandled error has occurred,
				we have been notified about it. Sorry for the inconvenience.`)
					fmt.Println("Error: ", r)
				}
			}()
		}

		// Set HTTP headers
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Get base object (ez.Controller), initialize it and update it
		val := reflect.ValueOf(obj)
		// The base Controller object may be several 'parents' away
		parentVal := val.Elem().Field(0)
		for parentVal.Type().Name() != "Controller" {
			parentVal = parentVal.Field(0) // TODO error if nothing was found
		}
		c := parentVal.Interface().(Controller)
		c.ControllerName = reflect.TypeOf(obj).Elem().Name()
		c.initValues(w, r)
		parentVal.Set(reflect.ValueOf(c))

		// Run the 'before run' action if it exists
		beforeRun := val.MethodByName("BeforeRun_")
		if beforeRun.IsValid() {
			beforeRun.Call([]reflect.Value{})
		}

		// Run the actual method. This can't be defined as type's
		// method since it needs an interface{} for reflection
		run(obj, &c)

		// Run the 'after run' action if it exists
		afterRun := val.MethodByName("AfterRun_")
		if afterRun.IsValid() {
			afterRun.Call([]reflect.Value{})
		}
	}
}

// Route is a helper method that runs http.HandleFunc for a given path and
// controller
func Route(path string, controller interface{}) {
	http.HandleFunc(path, GetHandler(controller))
}

// Index defines a default action
func (c *Controller) Index() {
	c.Say(`Welcome to ezweb! Define your own Index action:

    type Home struct {
        ezweb.Controller
    }

    func (c *Home) Index() {
        c.Write("Hello world!")
    }
    `)
}

// View executes a template corresponding to the current controller method
func (c *Controller) View(data interface{}) {
	template, err := parseTemplate("templates/"+c.ControllerName+"/"+c.ActionName+".html", c)
	if err != nil {
		fmt.Println("Template error: ", err)
		if Debug {
			c.Write("Template error: ", err)
		}
		return
	}

	err = template.Execute(c.Out, data)
	if err != nil {
		fmt.Println("Template execution error:", err)
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

func (c *Controller) SetCookie(key string, value string) {
	http.SetCookie(c.Out, &http.Cookie{Name: key, Value: value})
}

func (c *Controller) DeleteCookie(key string) {
	http.SetCookie(c.Out, &http.Cookie{Name: key, Value: "", MaxAge: -1})
}

func (c *Controller) SetContentType(ct string) {
	c.Out.Header().Set("Content-Type", ct)
}

// ReturnJson returns a marshaled json object with content type 'application/json'.
// This is usually used for responding to AJAX requests.
func (c *Controller) ReturnJson(model interface{}) {
	c.SetContentType("application/json")

	j := struct {
		Model  interface{}
		Status string
	}{
		Model:  model,
		Status: "OK",
	}

	obj, err := json.Marshal(j)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.Write(string(obj))
}

func (c *Controller) ReturnJsonFail(errorMsg string) {
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

func (c *Controller) JsonRedirect(redirectUrl string) {
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

//////// private methods ////////

// getActionAndArgsFromUri fetches an action name and query string args that
// follow from uri:
// "AccountController/Settings" => "Settings", []
// "Index" => "Index", []
// "" => "Index", []
// "Home/Register" => "Register", []
// "Forum/Topic/Hello-world/234242 => "Topic", []string{"Hello-world","234242"}
func getActionAndArgsFromUri(uri string, isIndex bool) (string, []string) {
	// Root action
	if uri == "" {
		return "Index", []string{}
	}
	actionName := uri // example.com/Action

	qsArgs := make([]string, 0)
	values := strings.Split(uri, "/")

	// http://example.com/Controller/Action/Arg1/Arg2
	if len(values) > 1 { // TODO this is ugly
		if isIndex {
			actionName = values[0] // Save action, controller is skipped
			qsArgs = values[1:]    // Save args
		} else {
			actionName = values[1]
			qsArgs = values[2:]
		}
	}

	// Capitalize and remove unallowed characters
	actionName = capitalize(actionName)
	actionName = strings.Replace(actionName, ".", "", -1)

	return actionName, qsArgs
}

// initValues parses the http.Request object and fetches all necessary values
// for ezweb.Controller
func (c *Controller) initValues(w http.ResponseWriter, r *http.Request) {
	c.Out = w
	c.Request = r
	values := r.URL.Query()
	c.Uri = r.URL.Path[1:]
	actionName, qsArgs := getActionAndArgsFromUri(c.Uri, c.ControllerName == "Home")
	c.ActionName = actionName
	c.PageTitle = ""

	// Generate query string map
	c.Params = make(map[string]string)
	c.args = make([]string, 0)
	for key, _ := range values {
		c.Params[key] = values.Get(key)
		c.args = append(c.args, values.Get(key))
	}
	// TODO assign routing values to c.Params
	c.args = append(qsArgs, c.args...) // append extra args in /arg1/arg2

	// Generate form data
	c.Form = make(map[string]string)
	c.Request.ParseForm()
	for key, _ := range c.Request.PostForm {
		c.Form[key] = c.Request.PostForm.Get(key)
	}
}

// run fetches an action name from URI and launches the action
func run(c interface{}, baseC *Controller) {
	// Fetch the method (action) that needs to be run
	method := reflect.ValueOf(c).MethodByName(baseC.ActionName)
	if !method.IsValid() {
		baseC.Write("Unknown action '" + baseC.ActionName +
			"' (controller: '" + baseC.ControllerName + "')")
		return
	}

	// Run it via reflect
	nrMethodArgs := method.Type().NumIn()
	values := make([]reflect.Value, 0)

	// Loop thru all method args and assign query string parameters to them
	for i := 0; i < nrMethodArgs; i++ {
		stringValue := ""

		// If there's no query string parameter for an arg, it will
		// default to an empty string
		if i < len(baseC.args) {
			// Get value from query string, obviously the order has
			// to be the same: register(name, password string) =>
			// /Register?name=a;password=b
			// TODO allow any args order
			stringValue = baseC.args[i]
		}

		// Convert this argument to a value of a certain type (Form,
		// string, int)
		argType := method.Type().In(i)
		values = append(values, baseC.argToValue(stringValue, argType))
	}

	method.Call(values)
}

// argToValue generates a reflect.Value from an argument type and its
// corresponding query string or form value
func (c *Controller) argToValue(stringValue string, argType reflect.Type) reflect.Value {
	// Handle a struct, this must be a form
	if argType.Kind() == reflect.Struct {
		// Create a new form object
		newFormObj := reflect.New(argType).Elem()

		// Set all its fields
		for i := 0; i < argType.NumField(); i++ {
			field := newFormObj.Field(i)
			fieldName := argType.Field(i).Name // e.g. "Id", "Title"
			formValue := c.Form[decapitalize(fieldName)]

			if field.Type().Name() == "int" {
				field.SetInt(int64(toint(formValue)))
			} else {
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

// parseTemplate parses a provided html template file and returns a
// *template.Template object
func parseTemplate(file string, c *Controller) (*template.Template, error) {
	// Read layout.html
	layout, err := ioutil.ReadFile("templates/layout.html")
	if err != nil {
		fmt.Println("Template layout not found")
	}

	layoutStr := string(layout)

	// Read template file
	b, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Template '", file, "' is not found")
	}
	s := string(b)

	// TODO read from compiled cache
	//data, _ := Asset("temp/template.html")

	// Embed the template into layout
	if layoutStr != "" {
		s = strings.Replace(layoutStr, "$BODY", s, -1)
	}

	// Title
	s = strings.Replace(s, "$_ez_TITLE", c.PageTitle, -1)
	s = strings.Replace(s, "\n\n", "\n", -1)
	s = strings.Replace(s, "\t", "", -1)
	s = strings.Replace(s, "  ", "", -1)

	// @if a
	// ===> {{ if a }}
	r := regexp.MustCompile("@(if|else|end|range)(.*?)\n")
	s = r.ReplaceAllString(s, "{{ $1 $2 }}\n")

	// @Name (always starts with a capital letter)
	// ===> {{ .Name }}
	r = regexp.MustCompile("@([A-Z][a-zA-Z\\.]+)")
	s = r.ReplaceAllString(s, "{{.$1}}")

	// @func "param" (always starts with a small letter)
	// ===> {{ func "param" }}
	r = regexp.MustCompile(`@([a-z]+ "[^"]+")`)
	s = r.ReplaceAllString(s, "{{ $1 }}")

	// $translation_tag
	// ===> {{ T "translation_tag" }}
	r = regexp.MustCompile("\\!([a-zA-Z_]+)")
	s = r.ReplaceAllString(s, `{{ T "$1" }}`)

	// Custom funcs
	t := template.New(file).Funcs(template.FuncMap{
		"eq":  reflect.DeepEqual,
		"inc": func(n int) int { return n + 1 },
		//"hex":  func(id bson.ObjectId) string { return id.Hex() },
	}).Funcs(c.CustomTemplateFuncs)

	t2, err := t.Parse(s)
	return t2, err
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

func init() {
	Debug = true
}
