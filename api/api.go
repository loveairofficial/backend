package api

import (
	"loveair/api/client"
	"loveair/base/cache"
	"loveair/base/data"
	"loveair/base/meta"
	"loveair/core/rest"
	"loveair/core/websocket/gorilla"
	"loveair/email"
	"loveair/log"
	"net/http"
	"time"

	"loveair/core/websocket/router"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func requestLog(serviceLogger log.SLoger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(t)
			if duration.Milliseconds() > 100 {
				serviceLogger.Log.WithFields(logrus.Fields{
					"Duration":   duration,
					"Method":     r.Method,
					"URL":        r.URL,
					"User Agent": r.UserAgent(),
					"Raddr":      r.RemoteAddr,
					"Host":       r.Host,
					// "Status":     w.StatusCode,
				}).Warningf("Request Log")
				return
			}
			serviceLogger.Log.Infof("Duration: %v\n", duration)
		})
	}
}

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ServeAPI is responsible for activating the Restful API logic
func Start(secret string,
	endpoint string,
	dbaseIf data.Interface,
	mbaseIf meta.Interface,
	cbaseIf cache.Interface,
	sRouter *router.Router,
	emailIf email.Interface,
	sLogger log.SLoger,
) error {
	logrus.Infoln("Service Listening On " + endpoint)

	socket := gorilla.InitWebsocket(dbaseIf, mbaseIf, cbaseIf, sRouter, emailIf, sLogger)
	socket.Daemon()

	rest := rest.InitRest(secret, dbaseIf, mbaseIf, cbaseIf, emailIf, sLogger)
	// rest.Daemon()

	r := mux.NewRouter()

	// Middleware to handle CORS headers
	r.Use(Cors)
	r.Use(requestLog(sLogger))

	// Client route
	cr := r.PathPrefix("/clr").Subrouter()
	client.Route(cr, rest, secret, socket, sLogger)

	// Admin route
	// ar := r.PathPrefix("/ar").Subrouter()
	// admin.Route(ar, rest, websocket, secret, mediabaseIf, serviceLogger)

	return http.ListenAndServe(":"+endpoint, r)
}
