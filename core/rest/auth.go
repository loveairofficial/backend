package rest

import (
	"encoding/json"
	"fmt"
	"loveair/models"
	"net/http"
	"time"

	stream "github.com/GetStream/stream-chat-go/v6"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/mongo"
)

func (re *Rest) Refresh(w http.ResponseWriter, r *http.Request) {
	re.sLogger.Log.Infoln("refreshing tkn")
	id := r.URL.Query().Get("id")

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Error(err)
		return
	}

	//~ Access tkn has expired, handle accordingly (check refresh tkn)
	var tk string

	// Retrieve the Authorization header from the request
	if tk = r.Header.Get("Authorization"); tk == "" {
		http.Error(w, "refresh_token is not found in Authorization header.", http.StatusUnauthorized)
		re.sLogger.Log.Errorln("refresh_token is not found in Authorization header.")
		return
	}

	claim := &Claims{}
	tkn, err := jwt.ParseWithClaims(tk, claim, func(token *jwt.Token) (interface{}, error) {
		return []byte(re.secret), nil
	})

	if err == nil && tkn.Valid {
		//~ Check to make sure the device was the one issued the tkn.
		savedDevice, err := re.dbase.GetDevice(claim.Email, claim.DID)

		if err == mongo.ErrNoDocuments {
			http.Error(w, "Unauthorized, relogin!", http.StatusUnauthorized)
			re.sLogger.Log.Errorln(err)
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			re.sLogger.Log.Errorln(err)
			return
		}

		device := new(models.Device)

		// Parse the devices JSON string
		if err := json.Unmarshal([]byte(r.PostForm.Get("device")), device); err != nil {
			re.sLogger.Log.Errorln(err)
		}

		if device.OSName != savedDevice.OSName || device.Brand != savedDevice.Brand || device.ModelName != savedDevice.ModelName {
			re.sLogger.Log.Errorln("Device does not match the details of the device that was issued this refresh_tkn, device will be deleted because of potential threat.")
			err = re.dbase.DeleteDevice(claim.Email, claim.DID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				re.sLogger.Log.Errorln(err)
				return
			}

			http.Error(w, "Unauthorized (potential threat), relogin!", http.StatusUnauthorized)
			re.sLogger.Log.Errorln("Unauthorized (potential threat), relogin!")

			return
		}

		// Generate jwt access token.
		atknString, err := re.generateAccessTkn(accessTknExpiration, claim.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			re.sLogger.Log.Errorln(err)
			return
		}

		re.sLogger.Log.Infoln("access: ", atknString)

		// Set the token and expiration time as a header
		aTokenWithExpiration := fmt.Sprintf("%s|%d", atknString, time.Now().Add(accessTknExpiration).Unix())

		// Generate jwt refresh token.
		rtknString, did, err := re.generateRefreshTkn(refreshTknExpiration, claim.Email, claim.DID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			re.sLogger.Log.Errorln(err)
			return
		}

		re.sLogger.Log.Infoln("refresh: ", rtknString)

		// Set the token and expiration time as a header
		rTokenWithExpiration := fmt.Sprintf("%s|%d", rtknString, time.Now().Add(refreshTknExpiration).Unix())

		device.DeviceID = did

		// add new device to database. //! do not add new device old device is still valid.
		// err = re.dbase.AddNewDevice(device, claim.Email)
		// if err != nil {
		// 	re.sLogger.Log.Errorln(err)
		// }

		//~ Generate jwt stream token
		//! do not hardcode credentials!!!
		client, err := stream.NewClient("vj79fb5bcmwt", "w82x6tnpjwjumdjqraj267vhskpgs34ptp8ydue8jzfg2rwye7dxab27f8jkgcub")
		if err != nil {
			http.Error(w, "Error creating Stream client", http.StatusInternalServerError)
			return
		}

		//! userID should be username
		sToken, err := client.CreateToken(id, time.Now().Add(streamTknExpiration))
		if err != nil {
			http.Error(w, "Error creating token", http.StatusInternalServerError)
			return
		}

		re.writeJSON(w, Response{
			Status:     "200",
			StatusCode: http.StatusOK,
			Message:    "Login successful",
			Data: Data{
				AccessTkn:  aTokenWithExpiration,
				RefreshTkn: rTokenWithExpiration,
				StreamTkn:  sToken,
			},
		})
	} else {
		// Token not valid delete device & initiate relogin.
		re.sLogger.Log.Errorln("Refresh_tkn has expired or is invalid, device will be deleted, user must relogin.")
		err = re.dbase.DeleteDevice(claim.Email, claim.DID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			re.sLogger.Log.Errorln(err)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)

		return
	}
}
