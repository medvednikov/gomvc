package gomvc

import (
	"net/http"
	"strings"
)

type JSON struct {
	Model interface{}
}

type View struct {
	Model interface{}
}

func (c *Controller) JSON(model interface{}) JSON { return JSON{model} }

func (c *Controller) View(model interface{}) View {
	c.SetContentType("text/html")
	return View{model}
}

// Redirect performs an HTTP redirect to another action in the same controller
func (c *Controller) Redirect(action string) View {
	c.cleanUp()
	if !strings.HasPrefix(action, "http") {
		action = "/" + action
	}
	http.Redirect(c.Out, c.Request, action, 302)
	return View{}
}

func (c *Controller) JSONError(errorMsg string) JSON {
	c.SetContentType("application/json")
	c.Out.WriteHeader(http.StatusBadRequest) // 400
	return JSON{struct{ ErrorMsg string }{errorMsg}}
}

func (c *Controller) JSONRedirect(url string) JSON {
	return JSON{struct{ RedirectUrl string }{url}}
}
