package middleware

import (
	"fmt"
	"loveair/log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

// Struct that will be encoded into a JWT
// jwt.StandardClaims was added as an embedded type, to provide fields like expiry time.
type Claims struct {
	Email       string
	Role        string
	Permissions map[string]map[string]bool
	jwt.StandardClaims
}

func RBAC(secret string, scope string, permission string, serviceLogger log.SLoger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tk string

			// Retrieve the Authorization header from the request
			if tk = r.Header.Get("Authorization"); tk == "" {
				if tk = r.URL.Query().Get("access_token"); tk == "" {
					http.Error(w, "access_token is not found in URL and Authorization header.", http.StatusUnauthorized)
					return
				}
			}

			fmt.Println(tk)

			claim := &Claims{}
			tkn, err := jwt.ParseWithClaims(tk, claim, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil {
				if err == jwt.ErrSignatureInvalid {
					w.WriteHeader(http.StatusUnauthorized)
					serviceLogger.Log.Errorln(err)
					return
				}
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			fmt.Println(claim.Permissions[scope][permission])

			if claim.Role == "Manager" {
				next.ServeHTTP(w, r)
				return
			}

			//
			if !tkn.Valid {
				w.WriteHeader(http.StatusUnauthorized)
				serviceLogger.Log.Errorln(err)
				return
			}

			if claim.Permissions[scope][permission] {
				next.ServeHTTP(w, r)
				return
			}

		})
	}
}
