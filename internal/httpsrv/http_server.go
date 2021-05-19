package httpsrv

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/forrestjgq/gomark/gmi"

	"github.com/forrestjgq/glog"

	"github.com/forrestjgq/gomark/internal/gm"

	"github.com/gorilla/mux"
)

type httpServer struct {
	port     int
	r        *mux.Router
	template *template.Template
}

var server *httpServer

func init() {
	server = &httpServer{}
	var err error
	server.template, err = template.New("bvar").Parse(htmlContent)
	if err != nil {
		panic(err)
	}
}

func Start(port int) {
	server.port = port

	r := mux.NewRouter()
	r.HandleFunc("/vars/js/{script}", procJs)
	r.HandleFunc("/vars", procVar)
	r.HandleFunc("/vars/{var}", procVar)
	r.HandleFunc("/vars/debug", procDebug)
	r.HandleFunc("/metrics", procMetrics)
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
	case gmi.RouteMetrics:
		return serveMetrics(req)
	default:
		return &gmi.Response{
			Status: 404,
		}
	}
}

func procMetrics(w http.ResponseWriter, r *http.Request) {
	proc(gmi.RouteMetrics, w, r)
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
	} else if !strings.Contains(v, "curl/") {
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

var metricsLock sync.Mutex

type writer bytes.Buffer

func (w *writer) write(strs ...string) *writer {
	b := (*bytes.Buffer)(w)
	for _, s := range strs {
		b.WriteString(s)
	}
	return w
}

// labels
type nv struct {
	name, value string
}

type labels struct {
	isDg       bool
	hasLabel   bool
	l          []nv
	metricName string
	typ        string
}

func (lb *labels) isComplete() bool {
	return len(lb.metricName) > 0
}

var dgItems = make(map[string]*labels)

// summary is now not implemented

// summary
//const (
//	NrPercentiles = 6
//)
//
//type summary struct {
//	latencyPercentiles            [NrPercentiles]string
//	latencyAvg, count, metricName string
//}
//
//func (s *summary) isComplete() bool {
//	return len(s.metricName) > 0
//}

// dumper
type metricsDump struct {
	b        writer
	lastName string
}

func (m *metricsDump) Dump(name, desc string) bool {
	if len(desc) > 0 && desc[0] == '"' {
		// there is no necessary to monitor string in prometheus
		return true
	}

	if m.dumpLabels(name, desc) {
		return true
	}
	if m.dumpLatencyRecorderSuffix(name, desc) {
		// Has encountered name with suffix exposed by LatencyRecorder,
		// Leave it to dumpLatencyRecorderSuffix to output Summary.
		return true
	}

	m.b.write("# HELP ", name, "\n").
		write("# TYPE ", name, " gauge\n").
		write(name, " ", desc, "\n")
	return true
}

func (m *metricsDump) dumpLabels(name string, desc string) bool {
	metricsLock.Lock()
	defer metricsLock.Unlock()

	si := m.parseLabel(name)
	//glog.Errorf("parsed label %+v", si)
	if si == nil || !si.isDg || !si.isComplete() {
		return false
	}

	//glog.Errorf("last %s, current %s", m.lastName, si.metricName)
	if m.lastName != si.metricName {
		m.lastName = si.metricName
		m.b.write("# HELP ", si.metricName, "\n").
			write("# TYPE ", si.metricName, " ", si.typ, "\n")
	}

	if si.hasLabel {
		m.b.write(si.metricName, "{")
		for i, v := range si.l {
			if i > 0 {
				m.b.write(",")
			}
			m.b.write(v.name, "=\"", v.value, "\"")
		}
		m.b.write("} ", desc, "\n")
	} else {
		m.b.write(si.metricName, " ", desc, "\n")
	}

	return true
}

func (m *metricsDump) dumpLatencyRecorderSuffix(name string, desc string) bool {
	// reserved for system vars
	return false
}

func (m *metricsDump) valueOf(name string) (metric, typ string) {
	end := strings.Index(name, "_")
	if end < 0 {
		metric = name
		return
	}
	metric = name[end+1:]
	typ = name[0:end]
	return
}
func (m *metricsDump) parseLabel(name string) *labels {
	//glog.Errorf("parse label %s", name)
	if si, ok := dgItems[name]; ok {
		return si
	}
	if !strings.HasPrefix(name, "t_") {
		return nil
	}

	item := &labels{
		isDg:       true,
		hasLabel:   false,
		l:          nil,
		metricName: "",
		typ:        "",
	}
	dgItems[name] = item

	metric := name[2:]

	metric, item.typ = m.valueOf(metric)
	if item.typ == "latency" {
		switch {
		case strings.HasSuffix(metric, "count"):
			item.typ = "counter"
		case strings.HasSuffix(metric, "max_latency") || strings.HasSuffix(metric, "qps"):
			item.typ = "gauge"
		case strings.HasSuffix(metric, "latency"):
			item.typ = "histogram"
			item.hasLabel = true
			item.l = append(item.l, nv{
				name:  "quantile",
				value: "0",
			})
		case strings.HasSuffix(metric, "latency_80"):
			metric = metric[:len(metric)-3]
			item.typ = "histogram"
			item.hasLabel = true
			item.l = append(item.l, nv{
				name:  "quantile",
				value: "80",
			})
		case strings.HasSuffix(metric, "latency_90"):
			metric = metric[:len(metric)-3]
			item.typ = "histogram"
			item.hasLabel = true
			item.l = append(item.l, nv{
				name:  "quantile",
				value: "90",
			})
		case strings.HasSuffix(metric, "latency_99"):
			metric = metric[:len(metric)-3]
			item.typ = "histogram"
			item.hasLabel = true
			item.l = append(item.l, nv{
				name:  "quantile",
				value: "99",
			})
		case strings.HasSuffix(metric, "latency_999"):
			metric = metric[:len(metric)-4]
			item.typ = "histogram"
			item.hasLabel = true
			item.l = append(item.l, nv{
				name:  "quantile",
				value: "999",
			})
		case strings.HasSuffix(metric, "latency_9999"):
			metric = metric[:len(metric)-5]
			item.typ = "histogram"
			item.hasLabel = true
			item.l = append(item.l, nv{
				name:  "quantile",
				value: "9999",
			})
		}
	}

	if strings.HasPrefix(metric, "l_") {
		metric = metric[2:]
		item.hasLabel = true
		var n, v string
		metric, n = m.valueOf(metric)
		metric, v = m.valueOf(metric)
		item.l = append(item.l, nv{
			name:  n,
			value: v,
		})
	}

	item.metricName = metric

	return item
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
	case gmi.RouteMetrics:
		rsp = serveMetrics(req)
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

func serveMetrics(req *gmi.Request) (rsp *gmi.Response) {
	rsp = &gmi.Response{
		Status: 200,
	}
	rsp.SetHeader("Content-Type", "text/plain")

	dump := &metricsDump{}
	n, err := gm.Dump(dump, nil)
	if err != nil {
		rsp.Body = []byte(err.Error())
	} else if n <= 0 {
		rsp.Body = []byte("Fail to dump metrics")
	} else {
		b := (*bytes.Buffer)(&dump.b)
		rsp.Body = b.Bytes()
	}
	return
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
