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
	// Debug is used to determine how to display error messages. Default is
	// true, set to false when deploying. One of the easy ways to do that
	// automatically is to parse machine's hostname.
	isDev bool = true

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

	assetFunc  func(string) ([]byte, error)
	assetNames []string

	sessionId     string
	sessionSecret string
)

type Config struct {
	Port       string
	IsDev      bool
	AssetFunc  func(string) ([]byte, error)
	AssetNames []string

	SessionId     string
	SessionSecret string
}

// Run initializes starts the web server
func Run(config *Config) {
	fmt.Println("Starting a gomvc app on port ", config.Port,
		" with isdev=", config.IsDev)
	isDev = config.IsDev
	assetFunc = config.AssetFunc
	assetNames = config.AssetNames
	sessionId = config.SessionId
	if sessionId == "" {
		sessionId = "gomvc_session"
	}
	sessionSecret = config.SessionSecret

	TimeStamp = time.Now().Unix()
	getActionsFromSourceFiles()
	cookieStore = sessions.NewCookieStore([]byte(sessionSecret))
	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // Default session lasts 30 days
		HttpOnly: true,       // Do not allow the cookie to be read from JS
		Secure:   !isDev,     // Use secure store in production only
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
		if !isDev {
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
		//w.Header().Set("Access-Control-Allow-Origin", "*")

		// Fetch the type of the controller (e.g. "Home")
		typ := reflect.Indirect(reflect.ValueOf(obj)).Type()

		// Create a new controller of this type for this request
		val := reflect.New(typ)

		// Get base object c (gomvc.Controller), initialize it and update
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
