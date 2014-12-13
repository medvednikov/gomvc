package gomvc

import (
	"reflect"
	"regexp"
	"strings"
)

type Rule string

const (
	Required  Rule = "Required"
	MinLength Rule = "MinLength"
	MaxLength Rule = "MaxLength"
)

var rules = []Rule{Required, MinLength, MaxLength}

// FormIsValid validates the form via fields' tags  containing rules like
// "Required", "MinLength", etc
func FormIsValid(f interface{}) (ok bool, errormsg string) {
	typ := reflect.TypeOf(f).Elem()
	val := reflect.ValueOf(f).Elem()

	// Check all attributes
	for i := 0; i < typ.NumField(); i++ {
		tag := string(typ.Field(i).Tag)
		field := val.Field(i).String()
		//fieldType := field.Type().Name()

		// Test for all possible rules
		for _, rule := range rules {
			if ok, msg := test(tag, rule, field); !ok {
				return false, msg
			}
		}

	}
	return true, ""
}

// test runs the validation for the rule if it's specified. If the test was passed,
// true is returned. Otherwise it's false and the error message set in the field tag.
func test(tag string, rule Rule, val string) (bool, string) {
	lines := strings.Split(tag, "\n")
	for _, line := range lines {
		// Search for a line that contains the rule specified
		if strings.Index(line, string(rule)) == -1 {
			continue
		}

		// Parse the rules. Some examples:
		// `Required(error_msg)`
		// `MinLength=5(error_msg)`
		r := regexp.MustCompile(`[a-zA-Z]+(=[0-9]+)?\(([a-z_]+)\)`)
		matches := r.FindAllStringSubmatch(line, -1)

		// Get integer argument if it's present
		// In the above MinLength example, the argument is 5 (chars)
		arg := matches[0][1]
		intArg := 0
		if arg != "" {
			intArg = toint(arg[1:])
		}

		// Get the error message in the parenthesis
		errorMsg := matches[0][2]

		// Now perform the check
		switch rule {
		case Required:
			return val != "", errorMsg
		case MinLength:
			return len(val) >= intArg, errorMsg
		case MaxLength:
			return len(val) <= intArg, errorMsg
		}
	}

	// Specified rule was not found, so the test can't fail
	return true, ""
}
