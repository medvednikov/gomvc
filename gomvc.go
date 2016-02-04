package gomvc

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var (
	// Gorilla router. Used for parsing url variables like /member/{id}
	router *mux.Router = mux.NewRouter()

	// A global map with all actions' argument names. They are fetched from
	// the source files since it's impossible to get argument names via
	// reflect. Example:
	// func (c *Home) Register(name string, email string)
	// ActionArgs["Home"]["Register"] = [ "name", "email" ]
	ActionArgs map[string]map[string][]string

	TimeStamp int64

	cookieStore *sessions.CookieStore

	config *Config
)

type Config struct {
	// IsDev is used to determine how to display error messages. Default is
	// true, set to false when deploying. One of the easy ways to do that
	// automatically is to parse machine's hostname.
	IsDev bool

	Port      string
	AssetFunc func(string) ([]byte, error)

	DelimLeft  string
	DelimRight string

	SessionID     string
	SessionSecret string
}

// Run initializes and starts the web server
func Run(c *Config) {
	config = c
	fmt.Println("Starting a gomvc app on port ", config.Port,
		" with isdev=", config.IsDev)
	if config.SessionID == "" {
		config.SessionID = "gomvc_session"
	}
	TimeStamp = time.Now().Unix()
	getActionsFromSourceFiles()
	cookieStore = sessions.NewCookieStore([]byte(config.SessionSecret))
	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // Default session lasts 30 days
		HttpOnly: true,       // Do not allow the cookie to be read from JS
		//Secure:   !config.IsDev, // Use secure store in production only
		Secure: false,
	}
	http.Handle("/", router)
	if config.Port != "" {
		fmt.Println(http.ListenAndServe(":"+config.Port, nil))
	}
}

// GetHandler generates a net/http handler func from a controller type.
// A new controller instance is created to handle incoming requests.
// Example:
// http.HandleFunc("/Account/", gomvc.GetHandler(&AccountController{}))
func GetHandler(obj interface{}) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Show a general error message on production
		if !config.IsDev {
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
		// Fetch the type of the controller (e.g. "Home")
		typ := reflect.Indirect(reflect.ValueOf(obj)).Type()
		// Create a new controller of this type for this request
		val := reflect.New(typ)
		// Create the *gomvc.Controller base
		base := reflect.New(reflect.TypeOf(Controller{}))
		// For assigning base to
		parentval := val.Elem().Field(0)
		// It can be one parent away
		if parentval.Type().String() != "*gomvc.Controller" {
			parentval = parentval.Field(0)
		}
		// Now initialize the base
		c := base.Interface().(*Controller)
		c.ControllerName = typ.Name()
		c.InitValues(w, r)
		// Assign the *gomvc.Controller base
		parentval.Set(base)
		// Run the 'before action' action if it exists
		beforeAction := val.MethodByName("BeforeAction_")
		if beforeAction.IsValid() {
			beforeAction.Call([]reflect.Value{})
		}
		// Run the actual method
		method := val.MethodByName(c.ActionName)
		//parentval.Set(base)
		runMethod(method, c)
		// Run the 'after run' action if it exists
		afterAction := val.MethodByName("AfterAction_")
		if afterAction.IsValid() {
			afterAction.Call([]reflect.Value{})
		}
		c.cleanUp()
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

func ServeStatic(prefix, dir string) {
	http.Handle("/"+prefix+"/", staticPrefix(prefix, dir))
}
