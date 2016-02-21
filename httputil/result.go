package httputil
import (
	"net/http"
	"fmt"
)

type HandleResult struct {
	Error       error
	RetHttpCode int
	RetCode     int
	RetMsg      string
}

func (r HandleResult) Ok() bool {
	return r.Error == nil
}

func (r *HandleResult) SetResult(err error, httpCode, ret int, msg string) {
	r.Error = err
	r.RetCode = ret
	r.RetHttpCode = httpCode
	r.RetMsg = msg
}

// write header with http code
// write response msg in json format {"ret":0,"msg":"retmsg"}
func (r *HandleResult) WriteJSON(w http.ResponseWriter) {
	w.WriteHeader(r.RetHttpCode)
	fmt.Fprintf(w, `{"ret":%d,"msg":"%s"}`, r.RetCode, r.RetMsg)
}

func RedirectTo(uri string, w http.ResponseWriter) {
	w.Header().Add("Location", uri)
	w.WriteHeader(http.StatusFound)
}