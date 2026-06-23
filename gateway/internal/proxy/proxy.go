package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// ProxyHandler 统一反向代理处理器
type ProxyHandler struct {
	platformURL *url.URL
	userURL     *url.URL
	platform    *httputil.ReverseProxy
	user        *httputil.ReverseProxy
}

func NewProxyHandler(platformUpstream, userUpstream string) http.HandlerFunc {
	platformURL, _ := url.Parse(platformUpstream)
	userURL, _ := url.Parse(userUpstream)

	ph := &ProxyHandler{
		platformURL: platformURL,
		userURL:     userURL,
		platform:    httputil.NewSingleHostReverseProxy(platformURL),
		user:        httputil.NewSingleHostReverseProxy(userURL),
	}

	return ph.ServeHTTP
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/user"):
		ph.user.ServeHTTP(w, r)

	case strings.HasPrefix(r.URL.Path, "/api/platform"):
		ph.platform.ServeHTTP(w, r)

	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"gateway"}`))
	}
}
