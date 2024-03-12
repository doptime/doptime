package httpserve

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/yangkequn/goflow/config"
	"github.com/yangkequn/goflow/permission"
)

// listten to a port and start http server
func httpStart(path string, port int64) {
	//get item
	router := http.NewServeMux()
	router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		var (
			result     interface{}
			b          []byte
			s          string
			ok         bool
			err        error
			httpStatus int = http.StatusOK
			svcCtx     *HttpContext
		)
		if CorsChecked(r, w) {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*12000)
		defer cancel()
		if svcCtx, err = NewHttpContext(ctx, r, w); err != nil || svcCtx == nil {
			httpStatus = http.StatusBadRequest
		} else if r.Method == "GET" {
			result, err = svcCtx.GetHandler()
		} else if r.Method == "POST" {
			result, err = svcCtx.PostHandler()
		} else if r.Method == "PUT" {
			result, err = svcCtx.PutHandler()
		} else if r.Method == "DELETE" {
			result, err = svcCtx.DelHandler()
		}

		if len(config.Cfg.Http.CORES) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", config.Cfg.Http.CORES)
		}

		if err == nil {
			if b, ok = result.([]byte); ok {
			} else if s, ok = result.(string); ok {
				b = []byte(s)
			} else {
				if b, err = json.Marshal(result); err == nil {
					//json Compact b
					var dst *bytes.Buffer = bytes.NewBuffer([]byte{})
					if err = json.Compact(dst, b); err == nil {
						b = dst.Bytes()
					}
				}
			}
		}
		//this err may be from json.marshal, so don't move it to the above else if
		if err != nil {
			if b = []byte(err.Error()); bytes.Contains(b, []byte("JWT")) {
				httpStatus = http.StatusUnauthorized
			} else if httpStatus == http.StatusOK {
				// this if is needed, because  httpStatus may have already setted as StatusBadRequest
				httpStatus = http.StatusInternalServerError
			}
		}

		//set Content-Type
		if svcCtx != nil && len(svcCtx.ResponseContentType) > 0 {
			svcCtx.Rsb.Header().Set("Content-Type", svcCtx.ResponseContentType)
		}
		w.WriteHeader(httpStatus)
		w.Write(b)
	})

	server := &http.Server{
		Addr:              ":" + strconv.FormatInt(port, 10),
		Handler:           router,
		ReadTimeout:       50 * time.Second,
		ReadHeaderTimeout: 50 * time.Second,
		WriteTimeout:      50 * time.Second, //10ms Redundant time
		IdleTimeout:       15 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Error().Err(err).Msg("http server ListenAndServe error")
		return
	}
	log.Info().Any("port", port).Any("path", path).Msg("GoFlow http server started!")
}

// Start the http server
// This function will be halt on listen and serve http server. call go Start() if you don't want this behavior
func Start() {
}

func init() {
	//wait 1s, till the permission table is loaded
	for i := 0; !permission.ConfigurationLoaded && i < 100; i++ {
		time.Sleep(time.Millisecond * 10)
	}
	//wait, till all the apis are loaded
	log.Info().Any("port", config.Cfg.Http.Port).Any("path", config.Cfg.Http.Path).Msg("GoFlow http server is starting")
	go httpStart(config.Cfg.Http.Path, config.Cfg.Http.Port)
}
