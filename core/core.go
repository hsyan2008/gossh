package core

import "github.com/hsyan2008/hfw"

type Controller struct {
	hfw.Controller
}

func (_ *Controller) Before(httpCtx *hfw.HTTPContext) {
}
