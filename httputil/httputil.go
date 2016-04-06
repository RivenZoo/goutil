// specify http handle func define
package httputil

import (
	"fmt"
	"strings"

	"github.com/RivenZoo/goutil/debug"

	log "github.com/cihub/seelog"
)
import (
	syslog "log"
	"net/http"
	"os"
	"time"
)

var statStub = NewStatStub()

type HandlerMap map[string]http.Handler

// add new handlers to src handler map
func (hm HandlerMap) Join(handlers HandlerMap) {
	for k, h := range handlers {
		if _, ok := hm[k]; !ok {
			hm[k] = h
		}
	}
}

type UrlHandler interface {
	UrlHandlers() HandlerMap
}

func UrlRegister(h UrlHandler, mux *http.ServeMux) {
	for url, uh := range h.UrlHandlers() {
		mux.Handle(url, uh)
	}
}

// http handler wrap to handle error and count url query
func HttpHandler(fn http.HandlerFunc) http.Handler {
	handleFunc := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				http.Error(w, "inner error.", http.StatusInternalServerError)
				frames := debug.TraceCallFrame(2, 10)
				sz := len(frames)
				s := make([]string, sz)
				for i := 0; i < sz; i++ {
					s[i] = fmt.Sprintf("\t%s", frames[i])
				}
				log.Errorf("INNER-PANIC: serving:%s %v\n%s", r.RemoteAddr, e, strings.Join(s, "\n"))
			}
		}()
		// stat url path query count
		statStub.IncrCounter(r.URL.Path)
		fn(w, r)
	}
	return http.HandlerFunc(handleFunc)
}

func RegisterUrlStat(urlPath ...string) {
	statStub.RegistStat(urlPath...)
}

func UrlStatResult() string {
	return statStub.Status()
}

type BaseReqHandler struct {
	Log log.LoggerInterface
}

func newHttpServer(addr string, timeout int, mux *http.ServeMux) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  time.Duration(timeout) * time.Second,
		WriteTimeout: time.Duration(timeout) * time.Second,
		ErrorLog:     syslog.New(os.Stderr, "", syslog.LstdFlags|syslog.Lshortfile),
	}
}

type service struct {
	server *http.Server
	h      UrlHandler
}

type ServiceMgr struct {
	serv []service
}

// addr: service address
// timeout: read and write timeout, second
func (mgr *ServiceMgr) AddService(addr string, timeout int, h UrlHandler) {
	serveMux := http.NewServeMux()
	UrlRegister(h, serveMux)

	server := newHttpServer(addr, timeout, serveMux)

	mgr.serv = append(mgr.serv, service{server: server, h: h})
}

func (mgr *ServiceMgr) AllServers() []*http.Server {
	var servers []*http.Server
	for i := range mgr.serv {
		servers = append(servers, mgr.serv[i].server)
	}
	return servers
}

func (mgr *ServiceMgr) String() string {
	var str []string
	for _, s := range mgr.serv {
		var patterns []string
		m := s.h.UrlHandlers()
		for k, _ := range m {
			patterns = append(patterns, fmt.Sprintf(`"%s"`, k))
		}
		str = append(str, fmt.Sprintf(`{"addr":"%s","path":[%s]}`, s.server.Addr, strings.Join(patterns, ",")))
	}
	return strings.Join(str, "\n")
}
