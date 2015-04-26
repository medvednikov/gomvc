package gomvc

type JSON struct {
	Model interface{}
}

type View struct {
	Model interface{}
}

func (c *Controller) JSON(model interface{}) JSON { return JSON{model} }
func (c *Controller) View(model interface{}) View { return View{model} }
