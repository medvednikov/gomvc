package gomvc

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func dump(val interface{}) string {
	return fmt.Sprintf("%#v", val)
}

func toint(s string) int {
	res, _ := strconv.Atoi(s)
	return res
}

func tofloat(s string) float64 {
	res, _ := strconv.ParseFloat(s, 64)
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
	pos := strings.LastIndex(action, "POST")
	if pos == -1 {
		return action
	}
	return action[:pos]
}

func staticPrefix(prefix, dir string) http.Handler {
	return http.StripPrefix("/"+prefix+"/",
		http.FileServer(http.Dir(dir)))
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
				res.WriteString(
					strings.ToUpper(string(action[i+1])))
				i++
			}
		} else {
			res.WriteString(string(action[i]))
		}
	}
	return res.String()
}

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

func handle(err error) {
	if err != nil {
		panic(err)
	}
}
