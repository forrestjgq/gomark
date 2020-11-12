package httpsrv

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/golang/glog"

	"github.com/forrestjgq/gomark/internal/gm"

	"github.com/gorilla/mux"
)

type httpServer struct {
	port     int
	r        *mux.Router
	template *template.Template
}

var server *httpServer

func Start(port int) {
	var err error
	server = &httpServer{port: port}
	server.template, err = template.New("bvar").Parse(htmlContent)
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/js/{script}", procJs)
	r.HandleFunc("/vars", procVar)
	r.HandleFunc("/vars/{var}", procVar)
	r.HandleFunc("/vars/{var}/{wildcards}", procVar)
	server.r = r

	go func() {
		http.ListenAndServe(":"+strconv.Itoa(port), r)
		glog.Info("server stops")
	}()
}

var lastModified = "Wed, 16 Sep 2015 01:25:30 GMT"

func procJs(w http.ResponseWriter, r *http.Request) {
	glog.Info("js")
	vars := mux.Vars(r)
	w.Header().Add("content-type", "application/javascript")

	if v, ok := vars["script"]; ok {
		m := r.Header.Get("If-Modified-Since")
		if m == lastModified {
			glog.Info("no modify")
			w.WriteHeader(304)
			return
		}

		w.Header().Add("Last-Modified", lastModified)
		if v == "jquery_min" {
			_, _ = w.Write([]byte(jqueryMinJs))
		} else if v == "flot_min" {
			_, _ = w.Write([]byte(flotMinJs))
		} else {
			w.WriteHeader(404)
		}
	} else {
		w.WriteHeader(404)
	}
}

func useHtml(h *http.Header) bool {
	if h == nil {
		return true
	}

	v := h.Get("console")
	if len(v) > 0 {
		return v == "0"
	}

	v = h.Get("user-agent")
	if len(v) == 0 {
		return false
	}

	if strings.Index(v, "curl/") < 0 {
		return true
	}

	return false
}

type dumpImpl struct {
	b    bytes.Buffer
	html bool
}
type Input struct {
	Click      string
	ValueStart string
	ValueEnd   string
	Content    string
}

const VarSep string = " : "

func (d *dumpImpl) Dump(name, desc string) bool {
	plot := false
	if d.html {
		opt := gm.SeriesOption{
			FixedLength: true,
			TestOnly:    true,
		}
		err := gm.DescribeVarSeriesSafe(name, &d.b, opt)
		if err == nil {
			plot = true
			d.b.WriteString("<p class=\"variable\">")
		} else {
			d.b.WriteString("<p class=\"nonplot-variable\">")
		}
	}

	d.b.WriteString(name)
	d.b.WriteString(VarSep)
	if d.html {
		d.b.WriteString(fmt.Sprintf("<span id=\"value-%s\">", name))
	}
	d.b.WriteString(desc)
	if d.html {
		d.b.WriteString("</span></p>\n")
		if plot {
			d.b.WriteString(fmt.Sprintf("<div class=\"detail\"><div id=\"%s\" class=\"flot-placeholder\"></div></div>", name))
		}
	} else {
		d.b.WriteString("\r\n")
	}
	return true
}

func hasParam(key string, r *http.Request) bool {
	_, exist := r.Form[key]
	return exist
}
func procVar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	buf := &bytes.Buffer{}
	if hasParam("series", r) {
		ok := false
		varName := ""
		if varName, ok = vars["var"]; ok && len(varName) == 0 {
			w.WriteHeader(400)
			_, _ = w.Write([]byte("var name not present"))
			return
		}
		opt := gm.SeriesOption{}
		err = gm.DescribeVarSeries(varName, buf, opt)
		if err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
		} else {
			_, _ = w.Write(buf.Bytes())
		}
		return
	}

	html := useHtml(&r.Header)
	tabs := false
	if html && !hasParam("dataonly", r) {
		tabs = true
	}
	if html {
		w.Header().Add("content-type", "text/html")
	} else {
		w.Header().Add("content-type", "text/plain")
	}

	dumper := &dumpImpl{html: html}
	opt := &gm.DumpOption{}
	opt.QuestionMark = '$'
	opt.DisplayFilter = gm.DisplayOnPlainText
	if html {
		opt.DisplayFilter = gm.DisplayOnHTML
	}
	opt.WhiteWildcards = vars["wildcards"]
	n, err := gm.Dump(dumper, opt)
	if err != nil {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if len(opt.WhiteWildcards) > 0 && n == 0 {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("fail to find any var"))
		return
	}
	if tabs {
		in := &Input{
			Content: dumper.b.String(),
		}
		if hasParam("expand", r) {
			in.Click = "  $(\".variable\").click();\n"
		} else {
			in.Click = "  $(\".default_expand\").click();\n"
		}

		if gm.FlagQuoteVector {
			in.ValueStart = "        var newValue = '\"[';\n"
			in.ValueEnd = "        newValue += ']\"';\n"
		} else {
			in.ValueStart = "        var newValue = '[';\n"
			in.ValueEnd = "        newValue += ']';\n"
		}

		err = server.template.Execute(w, in)
		if err != nil {
			w.WriteHeader(400)
			return
		}
	} else {
		_, _ = w.Write(dumper.b.Bytes())
	}

}
