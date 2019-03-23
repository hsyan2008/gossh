package core

import hfw "github.com/hsyan2008/hfw2"

type Controller struct {
	hfw.Controller
}

func (_ *Controller) Before(httpCtx *hfw.HTTPContext) {
}
