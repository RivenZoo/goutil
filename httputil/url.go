package httputil

import "net/url"

func JoinQueryParam(uri string, param url.Values) (u *url.URL, err error) {
	u, err = url.Parse(uri)
	if err != nil {
		return
	}
	dstQuery := u.Query()
	for k, vals := range param {
		if dstQuery.Get(k) == "" {
			for _, v := range vals {
				dstQuery.Add(k, v)
			}
		}
	}
	return
}
