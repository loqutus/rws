package web

import (
	"github.com/loqutus/rws/pkg/server/utils"
	"github.com/thedevsaddam/renderer"
	"net/http"
)

var rnd *renderer.Render

func init() {
	opts := renderer.Options{
		ParseGlobPattern: "../web/tpl/*.html",
	}
	rnd = renderer.New(opts)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	err := rnd.HTML(w, http.StatusOK, "index", nil)
	if err != nil {
		utils.Fail("index.html render error", err, w)
	}
}
