package httputil

import (
	"net/http"
	"time"
)

type CookieOption struct {
	Key, Val, Path, Domain string
	Expire                 time.Time
	MaxAge                 int
}

func NewHttpCookie(opt *CookieOption) *http.Cookie {
	s := opt.Key + "=" + opt.Val
	return &http.Cookie{
		opt.Key,
		opt.Val,
		opt.Path,
		opt.Domain,
		opt.Expire,
		opt.Expire.Format(time.UnixDate),
		opt.MaxAge,
		false,
		false,
		s,
		[]string{s},
	}
}
