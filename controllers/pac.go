package controllers

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hsyan2008/gossh/core"
	hfw "github.com/hsyan2008/hfw2"
	"github.com/hsyan2008/hfw2/pac"
)

type Pac struct {
	core.Controller
}

func (ctl *Pac) Index(httpCtx *hfw.HTTPContext) {
	// httpCtx.TemplateFile = "pac.html"

	list := pac.GetAll()
	if len(list) == 0 {
		list = map[string]bool{
			"google.com": true,
		}
	}

	txt := ""
	for k, _ := range list {
		txt += fmt.Sprintf(`%s"%s": 1,%s`, "\t", k, "\n")
	}

	txt = strings.TrimRight(txt, ",\n")

	//直接输出，双引号会被转义，所以手动处理
	context, err := ioutil.ReadFile("pac.html")
	httpCtx.ThrowCheck(500, err)

	httpCtx.Template = fmt.Sprintf(string(context), txt)
}
