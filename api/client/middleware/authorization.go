package middleware

import (
	"loveair/log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

// var (
// 	accessTknExpiration  = 24 * time.Hour
// 	refreshTknExpiration = 168 * time.Hour
// )

// Struct that will be encoded into a JWT
// jwt.StandardClaims was added as an embedded type, to provide fields like expiry time.
type Claims struct {
	Email string
	DID   string
	jwt.StandardClaims
}

func Authorization(secret string, serviceLogger log.SLoger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tk string

			// Retrieve the Authorization header from the request
			if tk = r.Header.Get("Authorization"); tk == "" {
				// Retrieve the access tkn string from url for websocket
				if tk = r.URL.Query().Get("access_token"); tk == "" {
					http.Error(w, "access_token is not found in URL and Authorization header.", http.StatusUnauthorized)
					serviceLogger.Log.Errorln("access_token is not found in URL and Authorization header.")
					return
				}
			}

			claim := &Claims{}
			tkn, err := jwt.ParseWithClaims(tk, claim, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err == nil && tkn.Valid {
				next.ServeHTTP(w, r)
			} else {
				// Its expired or tampered with, relogin.
				http.Error(w, "Token Expired", http.StatusUnauthorized)
				serviceLogger.Log.Errorln("Token Expired")
				return
			}

		})
	}
}

// func Authorization(secret string, database data.Interface, serviceLogger log.SLoger) mux.MiddlewareFunc {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 			atkn, err := r.Cookie("access_tkn")
// 			if err != nil {
// 				if err == http.ErrNoCookie {
// 					w.WriteHeader(http.StatusUnauthorized)
// 					serviceLogger.Log.Errorln(err)
// 					return
// 				}
// 				w.WriteHeader(http.StatusBadRequest)
// 				return
// 			}

// 			claim := &Claims{}
// 			tkn, err := jwt.ParseWithClaims(atkn.Value, claim, func(token *jwt.Token) (interface{}, error) {
// 				return []byte(secret), nil
// 			})

// 			if err == nil && tkn.Valid {
// 				serviceLogger.Log.Info("access_tkn valid, continue.")
// 				next.ServeHTTP(w, r)
// 			} else if ve, ok := err.(*jwt.ValidationError); ok { // Check if jwt has expired.
// 				if ve.Errors&jwt.ValidationErrorExpired != 0 {
// 					//~ Access tkn has expired, handle accordingly (check refresh tkn)
// 					rtkn, err := r.Cookie("refresh_tkn")
// 					if err != nil {
// 						if err == http.ErrNoCookie {
// 							w.WriteHeader(http.StatusUnauthorized)
// 							serviceLogger.Log.Errorln(err)
// 							return
// 						}
// 						w.WriteHeader(http.StatusBadRequest)
// 						return
// 					}

// 					claim := &Claims{}
// 					tkn, err := jwt.ParseWithClaims(rtkn.Value, claim, func(token *jwt.Token) (interface{}, error) {
// 						return []byte(secret), nil
// 					})

// 					if err == nil && tkn.Valid {
// 						//~ Check to make sure the device was the one issued the tkn.
// 						device, err := database.GetDevice(claim.Email, claim.DID)

// 						if err == mongo.ErrNoDocuments {
// 							http.Error(w, "Unauthorized, relogin!", http.StatusUnauthorized)
// 							serviceLogger.Log.Errorln(err)
// 							return
// 						} else if err != nil {
// 							w.WriteHeader(http.StatusInternalServerError)
// 							serviceLogger.Log.Errorln(err)
// 							return
// 						}

// 						usrAgent := ua.New(r.UserAgent())

// 						if usrAgent.OSInfo().Name != device.OSName || usrAgent.OSInfo().Version != device.OSVersion || usrAgent.Name() != device.BrowserName {
// 							serviceLogger.Log.Errorln("Device does not match the details of the device that was issued this refresh_tkn, device will be deleted becuse of potential threat.")
// 							err = database.DeleteDevice(claim.Email, claim.DID)
// 							if err != nil {
// 								w.WriteHeader(http.StatusInternalServerError)
// 								serviceLogger.Log.Errorln(err)
// 								return
// 							}

// 							http.Error(w, "Unauthorized (potential threat), relogin!", http.StatusUnauthorized)
// 							return
// 						}

// 						// Generate jwt access token.
// 						tknString, err := generateAccessTkn(accessTknExpiration, claim.Email, secret)
// 						if err != nil {
// 							w.WriteHeader(http.StatusInternalServerError)
// 							serviceLogger.Log.Errorln(err)
// 							return
// 						}

// 						http.SetCookie(w, &http.Cookie{
// 							Name:     "access_tkn",
// 							Value:    tknString,
// 							Expires:  time.Now().Add(1000 * time.Hour),
// 							HttpOnly: true,
// 							Path:     "/clr",
// 						})

// 						// Generate jwt refresh token.
// 						tknString, _, err = generateRefreshTkn(refreshTknExpiration, claim.Email, claim.DID, secret)
// 						if err != nil {
// 							w.WriteHeader(http.StatusInternalServerError)
// 							serviceLogger.Log.Errorln(err)
// 							return
// 						}

// 						http.SetCookie(w, &http.Cookie{
// 							Name:     "refresh_tkn",
// 							Value:    tknString,
// 							Expires:  time.Now().Add(1000 * time.Hour),
// 							HttpOnly: true,
// 							Path:     "/clr",
// 						})

// 						next.ServeHTTP(w, r)
// 					} else {
// 						// Token not valid delete device & initiate relogin.
// 						serviceLogger.Log.Errorln("Refresh_tkn has expired or is invalid, device will be deleted, user must relogin.")
// 						err = database.DeleteDevice(claim.Email, claim.DID)
// 						if err != nil {
// 							w.WriteHeader(http.StatusInternalServerError)
// 							serviceLogger.Log.Errorln(err)
// 							return
// 						}
// 						w.WriteHeader(http.StatusUnauthorized)
// 						return
// 					}

// 				} else {
// 					serviceLogger.Log.Errorln("Other validation error, potential threat via token tampering")
// 					fmt.Println(claim.Email, claim.DID)
// 					err = database.DeleteDevice(claim.Email, claim.DID)
// 					if err != nil {
// 						w.WriteHeader(http.StatusUnauthorized)
// 						serviceLogger.Log.Errorln(err)
// 						return
// 					}
// 				}
// 			} else {
// 				// Delete device
// 				serviceLogger.Log.Errorln("Other parsing error, relogin")
// 				err = database.DeleteDevice(claim.Email, claim.DID)
// 				if err != nil {
// 					w.WriteHeader(http.StatusInternalServerError)
// 					serviceLogger.Log.Errorln(err)
// 					return
// 				}

// 				w.WriteHeader(http.StatusBadRequest)
// 				return
// 			}
// 		})
// 	}
// }

// func generateAccessTkn(duration time.Duration, email, secret string) (string, error) {
// 	expirationTime := time.Now().Add(duration)

// 	claims := &Claims{
// 		Email: email,
// 		StandardClaims: jwt.StandardClaims{
// 			// In JWT, the expiry time is expressed as unix milliseconds
// 			ExpiresAt: expirationTime.Unix(),
// 		},
// 	}

// 	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	tknString, err := tkn.SignedString([]byte(re.secret))
// 	if err != nil {
// 		return "", err
// 	}

// 	return tknString, nil
// }

// func generateRefreshTkn(duration time.Duration, email, did, secret string) (string, string, error) {
// 	expirationTime := time.Now().Add(duration)

// 	if did == "" {
// 		did = re.generateUID()
// 	}

// 	claims := &Claims{
// 		Email: email,
// 		DID:   did,
// 		StandardClaims: jwt.StandardClaims{
// 			// In JWT, the expiry time is expressed as unix milliseconds
// 			ExpiresAt: expirationTime.Unix(),
// 		},
// 	}

// 	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	tknString, err := tkn.SignedString([]byte(re.secret))

// 	if err != nil {
// 		return "", "", err
// 	}
// 	return tknString, did, nil
// }

// func generateUID() string {
// 	// Generate unique ID.
// 	uid := xid.New()
// 	return uid.String()
// }

// func generateAccessTkn(duration time.Duration, email, secret string) (string, error) {
// 	expirationTime := time.Now().Add(duration)

// 	claims := &Claims{
// 		Email: email,
// 		StandardClaims: jwt.StandardClaims{
// 			// In JWT, the expiry time is expressed as unix milliseconds
// 			ExpiresAt: expirationTime.Unix(),
// 		},
// 	}

// 	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	tknString, err := tkn.SignedString([]byte(secret))
// 	if err != nil {
// 		return "", err
// 	}

// 	return tknString, nil
// }

// func generateRefreshTkn(duration time.Duration, email, did, secret string) (string, string, error) {
// 	expirationTime := time.Now().Add(duration)

// 	if did == "" {
// 		did = generateUID()
// 	}

// 	claims := &Claims{
// 		Email: email,
// 		DID:   did,
// 		StandardClaims: jwt.StandardClaims{
// 			// In JWT, the expiry time is expressed as unix milliseconds
// 			ExpiresAt: expirationTime.Unix(),
// 		},
// 	}

// 	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	tknString, err := tkn.SignedString([]byte(secret))

// 	if err != nil {
// 		return "", "", err
// 	}
// 	return tknString, did, nil
// }
