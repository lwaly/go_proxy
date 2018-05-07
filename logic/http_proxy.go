package logic

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"reverse_proxy/common"
)

type HandleProxy struct {
}

var host string

func init() {
	host, _ = common.Conf.GetValue("outhttp", "addr")
}

//实现Handler的接口
func (h *HandleProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ip, err := GetHttpRouteServer(r.Header.Get("RouteKey"))

	remote, err := url.Parse("http://" + host)
	fmt.Printf("%s\n", ip)
	if err != nil {
		common.Errorf("Parse url%v", err)
		return
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			//设置主机
			req.Host = host
			req.URL.Host = ip
			req.URL.Scheme = remote.Scheme
			//设置路径
			req.URL.Path = singleJoiningSlash(remote.Path, r.URL.Path)
			//设置参数
			req.PostForm = r.PostForm
			req.URL.RawQuery = r.URL.RawQuery
			req.Form = r.Form

		},
	}

	proxy.ServeHTTP(w, r)
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
