package httpserve

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/doptime/config/cfghttp"
)

func CorsChecked(r *http.Request, w http.ResponseWriter) bool {
	if r.Method == "OPTIONS" {
		// Allow all methods and headers for OPTIONS requests
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Accept-Language, X-CSRF-Token, Authorization, Rt, Origin, Refer, User-Agent")

		// Handle multiple allowed origins or wildcard
		origin := r.Header.Get("Origin")
		if origin != "" {
			if cfghttp.CORES == "*" {
				// Allow all origins
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if strings.Contains(cfghttp.CORES, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		w.Header().Set("Access-Control-Max-Age", strconv.Itoa(30*86400)) // 30 days
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}
