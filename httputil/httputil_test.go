package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpHandle(t *testing.T) {
	RegisterUrlStat("/", "/abc")
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("empty response"))
	})
	h := HttpHandler(fn)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://baidu.com/", nil)
	if err != nil {
		t.FailNow()
	}
	for i := 0; i < 10; i++ {
		h.ServeHTTP(w, r)
	}
	t.Log("\n", UrlStatResult())
}
