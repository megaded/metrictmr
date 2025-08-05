package middleware

import (
	"net"
	"net/http"
)

const RealIp string = "X-Real-IP"

func TrustedSubnet(subnet net.IPNet) func(h http.Handler) http.Handler {
	fn := func(h http.Handler) http.Handler {
		subNetFn := func(w http.ResponseWriter, r *http.Request) {
			hw := w

			realip := r.Header.Get(RealIp)
			if realip == "" {
				hw.WriteHeader(http.StatusForbidden)
				return
			}
			ip := net.ParseIP(realip)
			if !subnet.Contains(ip) {
				hw.WriteHeader(http.StatusForbidden)
				return
			}

			h.ServeHTTP(hw, r)
		}
		return http.HandlerFunc(subNetFn)
	}
	return fn
}
