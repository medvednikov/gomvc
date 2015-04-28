package gomvc

import "net/http"

type JSON struct {
	Model interface{}
}

type View struct {
	Model interface{}
}

func (c *Controller) JSON(model interface{}) JSON { return JSON{model} }
func (c *Controller) View(model interface{}) View { return View{model} }

func (c *Controller) JSONError(errorMsg string) JSON {
	c.SetContentType("application/json")
	c.Out.WriteHeader(http.StatusBadRequest) // 400
	return JSON{struct{ ErrorMsg string }{errorMsg}}
}

func (c *Controller) JSONRedirect(url string) JSON {
	return JSON{struct{ RedirectUrl string }{url}}
}
