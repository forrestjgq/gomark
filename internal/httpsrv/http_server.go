package httpsrv

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/forrestjgq/gomark/gmi"

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
	r.HandleFunc("/vars/js/{script}", procJs)
	r.HandleFunc("/vars", procVar)
	r.HandleFunc("/vars/{var}", procVar)
	r.HandleFunc("/vars/debug", procDebug)
	server.r = r

	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(port), r)
		glog.Info("server stops, error: ", err)
	}()
}

func RequestHTTP(req *gmi.Request) *gmi.Response {
	switch req.Router {
	case gmi.RouteJs:
		return serveJs(req)
	case gmi.RouteDebug:
		return serveDebug(req)
	case gmi.RouteVars:
		return serveVar(req)
	default:
		return &gmi.Response{
			Status: 404,
		}
	}
}

func procDebug(w http.ResponseWriter, r *http.Request) {
	proc(gmi.RouteDebug, w, r)
}
func serveDebug(req *gmi.Request) (rsp *gmi.Response) {
	rsp = &gmi.Response{
		Status: 200,
	}
	p := req.GetParam("perf")
	if p == "1" {
		gm.EnableInternalVariables()
		rsp.Body = []byte("internal variables enabled")
	} else if p == "0" {
		gm.DisableInternalVariables()
		rsp.Body = []byte("internal variables disabled")
	} else {
		rsp.Status = 404
	}

	return
}

var lastModified = "Wed, 16 Sep 2015 01:25:30 GMT"

func procJs(w http.ResponseWriter, r *http.Request) {
	proc(gmi.RouteJs, w, r)
}
func serveJs(req *gmi.Request) (rsp *gmi.Response) {
	rsp = &gmi.Response{
		Status: 200,
		Body:   nil,
	}

	rsp.SetHeader("content-type", "application/javascript")

	if v, ok := req.Params["script"]; ok {
		if req.GetHeader("If-Modified-Since") == lastModified {
			rsp.Status = 304
			return
		}
		rsp.SetHeader("Last-Modified", lastModified)
		if v == "jquery_min" {
			rsp.Body = []byte(jqueryMinJs)
		} else if v == "flot_min" {
			rsp.Body = []byte(flotMinJs)
		} else {
			rsp.Status = 404
		}
	} else {
		rsp.Status = 404
	}
	return
}

func useHtml(req *gmi.Request) bool {
	v := req.GetHeader("console")
	if len(v) > 0 {
		return v == "0"
	}

	v = req.GetHeader("user-agent")
	if len(v) == 0 {
		return false
	} else if strings.Index(v, "curl/") < 0 {
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

func procVar(w http.ResponseWriter, r *http.Request) {
	proc(gmi.RouteVars, w, r)
}
func proc(route gmi.Route, w http.ResponseWriter, r *http.Request) {
	req := &gmi.Request{
		Params: make(map[string]string),
	}
	vars := mux.Vars(r)
	for k, v := range vars {
		req.Params[k] = v
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	for k := range r.Form {
		req.Params[k] = r.Form.Get(k)
	}

	for k := range r.Header {
		req.SetHeader(k, r.Header.Get(k))
	}

	var rsp *gmi.Response
	switch route {
	case gmi.RouteVars:
		rsp = serveVar(req)
	case gmi.RouteDebug:
		rsp = serveDebug(req)
	case gmi.RouteJs:
		rsp = serveJs(req)
	default:
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(rsp.Status)
	for k, v := range rsp.GetHeaders() {
		w.Header().Add(k, v)
	}
	if len(rsp.Body) > 0 {
		_, _ = w.Write(rsp.Body)
	}
}
func serveVar(req *gmi.Request) (rsp *gmi.Response) {
	rsp = &gmi.Response{
		Status: 200,
	}

	buf := &bytes.Buffer{}
	if req.HasParam("series") {
		varName := ""
		if varName = req.GetParam("var"); len(varName) == 0 {
			rsp.Status = 400
			rsp.Body = []byte("var name not present")
			return
		}

		opt := gm.SeriesOption{}
		err := gm.DescribeVarSeries(varName, buf, opt)
		if err != nil {
			rsp.Status = 400
			rsp.Body = []byte(err.Error())
		} else {
			rsp.Body = buf.Bytes()
		}
		return
	}

	html := useHtml(req)
	tabs := false
	if html && !req.HasParam("dataonly") {
		tabs = true
	}
	if html {
		rsp.SetHeader("content-type", "text/html")
	} else {
		rsp.SetHeader("content-type", "text/plain")
	}

	dumper := &dumpImpl{html: html}
	opt := &gm.DumpOption{}
	opt.QuestionMark = '$'
	opt.DisplayFilter = gm.DisplayOnPlainText
	if html {
		opt.DisplayFilter = gm.DisplayOnHTML
	}
	opt.WhiteWildcards = req.GetParam("var")
	n, err := gm.Dump(dumper, opt)
	if err != nil {
		rsp.Status = 400
		rsp.Body = []byte(err.Error())
		return
	}

	if len(opt.WhiteWildcards) > 0 && n == 0 {
		rsp.Status = 400
		rsp.Body = []byte("fail to find any var")
		return
	}

	if tabs {
		in := &Input{
			Content: dumper.b.String(),
		}
		if req.HasParam("expand") {
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

		w := &bytes.Buffer{}
		err = server.template.Execute(w, in)
		if err != nil {
			rsp.Status = 400
			rsp.Body = []byte(err.Error())
		} else {
			rsp.Body = w.Bytes()
		}
	} else {
		rsp.Body = dumper.b.Bytes()
	}

	return
}
