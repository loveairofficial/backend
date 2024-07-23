package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"loveair/models"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	stream "github.com/GetStream/stream-chat-go/v6"
	"github.com/cloudinary/cloudinary-go/api"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/schema"

	// "github.com/houseme/mobiledetect/ua"

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type TokenResponse struct {
	Token string `json:"token"`
}

func (re *Rest) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	// Check if there is an already existing account with the email.
	if err := re.dbase.VerifyEmailExist(email); err != mongo.ErrNoDocuments {
		//login user
		re.writeJSON(w, Response{
			Status:     "200",
			StatusCode: http.StatusOK,
			Message:    "Login",
		})
		re.sLogger.Log.Errorln(err)
		return
	}

	// generate 4 digit pin
	pin, err := GenerateRandomPin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	// cache it on redis
	err = re.cbaseIf.SetPin(email, pin, time.Minute*10)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	// send pin to client via email
	resStatus, err := re.emailIf.SendEmailVerificationPin(email, pin)
	if err != nil || resStatus != 202 {
		w.WriteHeader(http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Email verification successfull",
	})
}

func (re *Rest) VerifyEmailVerificationPin(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	pin := r.URL.Query().Get("pin")

	//retrieve pin from cache
	res, err := re.cbaseIf.GetPin(email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// validate pin
	if res != pin {
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else {
		//delete cache email & pin
		err = re.cbaseIf.DeletePin(email)
		if err != nil {
			re.sLogger.Log.Errorln(err)
		}
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Pin verification successfull",
	})
}

// Reset Password
func (re *Rest) HandleSendPasswordResetPin(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	// Check if there is an already existing account with the email.
	if err := re.dbase.VerifyEmailExist(email); err == mongo.ErrNoDocuments {
		w.WriteHeader(http.StatusNotFound)
		re.sLogger.Log.Errorln(err)
		return
	}

	// generate 4 digit pin
	pin, err := GenerateRandomPin()
	if err != nil {
		fmt.Println("Error generating PIN:", err)
		return
	}

	// cache it on redis
	err = re.cbaseIf.SetPin(email, pin, time.Minute*10)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// send pin to client via email
	resStatus, err := re.emailIf.SendPasswordResetPin(email, pin)
	if err != nil || resStatus != 202 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//sign-up user
	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Password reset pin sent successfully ",
	})
}

func (re *Rest) HandleVerifyPasswordResetPin(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	pin := r.URL.Query().Get("pin")

	//retrieve pin from cache
	res, err := re.cbaseIf.GetPin(email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// validate pin
	if res != pin {
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else {
		//delete cache email & pin
		err = re.cbaseIf.DeletePin(email)
		if err != nil {
			re.sLogger.Log.Errorln(err)
		}
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Pin verification successfull",
	})
}

func (re *Rest) HandlePasswordReset(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	password := r.URL.Query().Get("password")

	if bytes, err := bcrypt.GenerateFromPassword([]byte(password), 5); err == nil {
		password = string(bytes)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	// Store credentials to database.
	err := re.dbase.UpdatePassword(email, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Pin verification successfull",
	})
}

func (re *Rest) SignUp(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Error(err)
		return
	}

	// Create a new account.
	usr := new(models.User)

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	err = decoder.Decode(usr, r.Form)
	if err != nil {
		w.WriteHeader(500)
		re.sLogger.Log.Errorln(err)
		return
	}

	usr.ID = re.generateUID()
	usr.JoinedAt = time.Now().UTC()
	usr.StageID = 1
	fmt.Println(usr)

	// Generate hash from raw password string and store.
	if bytes, err := bcrypt.GenerateFromPassword([]byte(usr.Password), 5); err == nil {
		usr.Password = string(bytes)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	//Init User
	usr.Verification = false
	usr.IsOnboarded = false
	usr.IsActive = true
	usr.IsPaused = true
	usr.Preference = models.Preference{
		InterestedIn:          []string{"Open to all"},
		RelationshipIntention: []string{"Dating"},
		AgeRange: models.Range{
			Min: 18,
			Max: 35,
		},
		Global:   true,
		Presence: "Online",
		GeoCircle: models.GeoCircle{
			Lat:    0.0,
			Lon:    0.0,
			Radius: 100,
			Unit:   "mi",
		},
		Religion: []string{"Open to all"},
	}
	usr.Address = ""
	usr.RoseCount = 0
	usr.Subscription = "Free"
	usr.Religion = ""
	usr.Notification = models.Notification{
		Email: true,
		Push:  true,
	}

	usr.FreeTrialCount = 5
	usr.FreeTrialCountIssueTimestamp = time.Now().UTC()
	usr.IsSuppressed = false

	// Store credentials to database.
	err = re.dbase.AddUser(usr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	// Store credentials to metabase.
	err = re.mbase.AddUser(*usr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	// Generate jwt access token.
	atknString, err := re.generateAccessTkn(accessTknExpiration, r.PostForm.Get("email"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	// Set the token and expiration time as a header
	aTokenWithExpiration := fmt.Sprintf("%s|%d", atknString, time.Now().Add(accessTknExpiration).Unix())

	// Generate jwt refresh token.
	rtknString, did, err := re.generateRefreshTkn(refreshTknExpiration, r.PostForm.Get("email"), "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	// Set the token and expiration time as a header
	rTokenWithExpiration := fmt.Sprintf("%s|%d", rtknString, time.Now().Add(refreshTknExpiration).Unix())

	device := new(models.Device)

	// Parse the devices JSON string
	if err := json.Unmarshal([]byte(r.PostForm.Get("device")), device); err != nil {
		re.sLogger.Log.Errorln(err)
	}

	device.DeviceID = did

	// add new device to database.
	err = re.dbase.AddNewDevice(device, usr.Email)
	if err != nil {
		re.sLogger.Log.Errorln(err)
	}

	//~ Generate jwt stream token
	//! do not hardcode credentials!!!
	//! add token to env
	client, err := stream.NewClient("vj79fb5bcmwt", "w82x6tnpjwjumdjqraj267vhskpgs34ptp8ydue8jzfg2rwye7dxab27f8jkgcub")
	if err != nil {
		http.Error(w, "Error creating Stream client", http.StatusInternalServerError)
		return
	}

	//! userID should be username
	sToken, err := client.CreateToken(usr.ID, time.Now().Add(streamTknExpiration))
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	resStatus, err := re.emailIf.SendWelcomeEmail(usr.Email, usr.FirstName)
	if err != nil || resStatus != 202 {
		re.sLogger.Log.Errorln(err)
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Client signup successful",
		Data: Data{
			AccessTkn:    aTokenWithExpiration,
			RefreshTkn:   rTokenWithExpiration,
			StreamTkn:    sToken,
			IsOnboarded:  usr.IsOnboarded,
			Email:        usr.Email,
			ID:           usr.ID,
			FirstName:    usr.FirstName,
			Subscription: usr.Subscription,
		},
	})
}

func (re *Rest) SignIn(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Error(err)
		return
	}

	creds, err := re.dbase.GetCredential(r.PostForm.Get("email"))
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(creds.Password), []byte(r.PostForm.Get("password"))); err != nil {
		http.Error(w, "Email or password incorrect", http.StatusUnauthorized)
		re.sLogger.Log.Errorln(err)
		return
	}

	//check is account is active
	//! Create a cron job to delete accounts that are deactivated by Users after 30 days.
	if !creds.IsActive {
		if creds.DeactivatedBy == "User" {
			// Deactivation date is less than 30 days old, ask user to reactivate to sign in.
			http.Error(w, "Account is deactivated, user needs to reactivate account to sign-in", http.StatusForbidden)
			re.sLogger.Log.Infoln("Account is deactivated, user needs to reactivate account to sign-in")
			return
		} else {
			// account deactivated by admin, ask user to contact support for appesl.
			http.Error(w, "Account is deactivated", http.StatusGone)
			re.sLogger.Log.Infoln("Account is deactivated by admin")
			return
		}
	}

	// Generate jwt access token.
	atknString, err := re.generateAccessTkn(accessTknExpiration, r.PostForm.Get("email"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	fmt.Println("access: ", atknString)

	// Set the token and expiration time as a header
	aTokenWithExpiration := fmt.Sprintf("%s|%d", atknString, time.Now().Add(accessTknExpiration).Unix())

	// Generate jwt refresh token.
	rtknString, did, err := re.generateRefreshTkn(refreshTknExpiration, r.PostForm.Get("email"), "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	fmt.Println("refresh: ", rtknString)

	// Set the token and expiration time as a header
	rTokenWithExpiration := fmt.Sprintf("%s|%d", rtknString, time.Now().Add(refreshTknExpiration).Unix())

	// add new device to database.
	device := new(models.Device)

	// Parse the devices JSON string
	if err := json.Unmarshal([]byte(r.PostForm.Get("device")), device); err != nil {
		re.sLogger.Log.Errorln(err)
	}

	device.DeviceID = did

	// add new device to database.
	err = re.dbase.AddNewDevice(device, creds.Email)
	if err != nil {
		re.sLogger.Log.Errorln(err)
	}

	//~ Generate jwt stream token
	//! do not hardcode credentials!!!
	client, err := stream.NewClient("vj79fb5bcmwt", "w82x6tnpjwjumdjqraj267vhskpgs34ptp8ydue8jzfg2rwye7dxab27f8jkgcub")
	if err != nil {
		http.Error(w, "Error creating Stream client", http.StatusInternalServerError)
		return
	}

	//! userID should be username
	sToken, err := client.CreateToken(creds.ID, time.Now().Add(streamTknExpiration))
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Login successful",
		Data: Data{
			AccessTkn:      aTokenWithExpiration,
			RefreshTkn:     rTokenWithExpiration,
			StreamTkn:      sToken,
			IsOnboarded:    creds.IsOnboarded,
			Email:          creds.Email,
			ID:             creds.ID,
			FirstName:      creds.FirstName,
			ProfilePicture: creds.ProfilePicture,
			Subscription:   creds.Subscription,
		},
	})
}

func (re *Rest) ReactivateAccount(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Error(err)
		return
	}

	creds, err := re.dbase.GetCredential(r.PostForm.Get("email"))
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// creds, err := re.dbase.GetCredential(r.PostForm.Get("email"))
	err = re.dbase.ReactivateAccount(creds.ID)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		re.writeJSON(w, Response{
			Status:     "500",
			StatusCode: http.StatusInternalServerError,
			Message:    "Error saving data to database",
		})
		return
	}

	// Store credentials to metabase.
	err = re.mbase.ReactivateAccount(creds.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	// Generate jwt access token.
	atknString, err := re.generateAccessTkn(accessTknExpiration, r.PostForm.Get("email"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	fmt.Println("access: ", atknString)

	// Set the token and expiration time as a header
	aTokenWithExpiration := fmt.Sprintf("%s|%d", atknString, time.Now().Add(accessTknExpiration).Unix())

	// Generate jwt refresh token.
	rtknString, did, err := re.generateRefreshTkn(refreshTknExpiration, r.PostForm.Get("email"), "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	fmt.Println("refresh: ", rtknString)

	// Set the token and expiration time as a header
	rTokenWithExpiration := fmt.Sprintf("%s|%d", rtknString, time.Now().Add(refreshTknExpiration).Unix())

	// add new device to database.
	device := new(models.Device)

	// Parse the devices JSON string
	if err := json.Unmarshal([]byte(r.PostForm.Get("device")), device); err != nil {
		re.sLogger.Log.Errorln(err)
	}

	device.DeviceID = did

	// add new device to database.
	err = re.dbase.AddNewDevice(device, creds.Email)
	if err != nil {
		re.sLogger.Log.Errorln(err)
	}

	//~ Generate jwt stream token
	//! do not hardcode credentials!!!
	client, err := stream.NewClient("vj79fb5bcmwt", "w82x6tnpjwjumdjqraj267vhskpgs34ptp8ydue8jzfg2rwye7dxab27f8jkgcub")
	if err != nil {
		http.Error(w, "Error creating Stream client", http.StatusInternalServerError)
		return
	}

	//! userID should be username
	sToken, err := client.CreateToken(creds.ID, time.Now().Add(streamTknExpiration))
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Login successful",
		Data: Data{
			AccessTkn:      aTokenWithExpiration,
			RefreshTkn:     rTokenWithExpiration,
			StreamTkn:      sToken,
			IsOnboarded:    creds.IsOnboarded,
			Email:          creds.Email,
			ID:             creds.ID,
			FirstName:      creds.FirstName,
			ProfilePicture: creds.ProfilePicture,
			Subscription:   creds.Subscription,
		},
	})
}

func (re *Rest) SignOut(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	var rtk string

	// Retrieve the Authorization header from the request
	if rtk = r.Header.Get("Refresh-Authorization"); rtk == "" {
		http.Error(w, "access_token is not found in Authorization header.", http.StatusUnauthorized)
		return
	}

	fmt.Println(email, rtk)

	claim := &Claims{}
	_, err := jwt.ParseWithClaims(rtk, claim, func(token *jwt.Token) (interface{}, error) {
		return []byte(re.secret), nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}
	fmt.Println(claim)
	err = re.dbase.DeleteDevice(email, claim.DID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200 OK",
		StatusCode: 200,
		Message:    "Sign-out successful",
	})
}

func (re *Rest) GetSignature(w http.ResponseWriter, r *http.Request) {
	pID := r.URL.Query().Get("public_id")
	overwrite := r.URL.Query().Get("overwrite")
	upload_preset := r.URL.Query().Get("upload_preset")
	folder := r.URL.Query().Get("folder")
	timestamp := time.Now().Unix()

	fmt.Println(pID, overwrite, upload_preset)

	ParamsToSign := make(url.Values)
	ParamsToSign["overwrite"] = []string{overwrite}
	ParamsToSign["public_id"] = []string{pID}
	ParamsToSign["timestamp"] = []string{fmt.Sprintf("%d", timestamp)}
	ParamsToSign["upload_preset"] = []string{upload_preset}
	ParamsToSign["folder"] = []string{folder}

	// !dont hardcode secret
	signature, err := api.SignParameters(ParamsToSign, os.Getenv("CLOUDINARY_SECRET"))

	fmt.Println(signature, err)

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Data: Data{
			Signature: signature,
			Timestamp: fmt.Sprintf("%d", timestamp),
		},
	})
}

func (re *Rest) UpdatePreference(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	fmt.Println(r.PostForm.Get("utcOffset"), r.PostFormValue("preference"), r.PostFormValue("address"), r.PostFormValue("vicinity"))

	utcOffset, err := strconv.Atoi(r.PostForm.Get("utcOffset"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	prefStr := r.PostFormValue("preference")

	var pref models.Preference
	if err := json.Unmarshal([]byte(prefStr), &pref); err != nil {
		http.Error(w, "Invalid interests data", http.StatusBadRequest)
		return
	}

	err = re.dbase.UpdatePreference(id, pref, r.PostFormValue("address"), r.PostFormValue("vicinity"), utcOffset)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		re.writeJSON(w, Response{
			Status:     "500",
			StatusCode: http.StatusInternalServerError,
			Message:    "Error saving data to database",
		})
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "",
	})
}

func (re *Rest) GetPreference(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	pref, addr, vicinity, utcOffset, roseCount, subscription, err := re.dbase.GetPreference(id)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Println(pref, addr, vicinity, utcOffset)

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Data: Data{
			Preference:   pref,
			Address:      addr,
			Vicinity:     vicinity,
			UTCOffset:    utcOffset,
			RoseCount:    roseCount,
			Subscription: subscription,
		},
	})
}

// ! MilesToMeters converts miles to meters
func MilesToMeters(miles float64) float64 {
	return miles * 1609.34
}

// ! KilometersToMeters converts kilometers to meters
func KilometersToMeters(kilometers float64) float64 {
	return kilometers * 1000
}

func getIDs(users []models.User) []string {
	var ids []string
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return ids
}

func (re *Rest) GetPotentialMatches(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	// Get the 'preference' query parameter
	preferenceParam := r.URL.Query().Get("preference")
	if preferenceParam == "" {
		http.Error(w, "preference parameter is missing", http.StatusBadRequest)
		return
	}

	// URL decode the parameter
	decodedPreference, err := url.QueryUnescape(preferenceParam)
	if err != nil {
		http.Error(w, "failed to decode preference parameter", http.StatusInternalServerError)
		return
	}

	// Parse the JSON string into a Preference struct
	preference := new(models.Preference)
	err = json.Unmarshal([]byte(decodedPreference), preference)
	if err != nil {
		http.Error(w, "failed to parse JSON preference", http.StatusInternalServerError)
		return
	}

	fmt.Println(preference)

	if preference.GeoCircle.Unit == "mi" {
		preference.GeoCircle.Radius = MilesToMeters(preference.GeoCircle.Radius)
	} else {
		preference.GeoCircle.Radius = KilometersToMeters(preference.GeoCircle.Radius)
	}

	fmt.Println(preference.GeoCircle.Radius)

	//use the preference to query neo4j for potential matches
	pms, err := re.mbase.GetPotentialMatches(id, preference)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	if len(pms) > 0 {
		// Get slice of IDs
		ids := getIDs(pms)

		HydratedPms, err := re.dbase.HydratePotentialMatches(ids, pms)
		if err != nil {
			re.sLogger.Log.Errorln(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		re.writeJSON(w, Response{
			Status:     "200",
			StatusCode: http.StatusOK,
			Data: Data{
				Users: HydratedPms,
			},
		})

		//TODO: Cache this users
	} else {
		re.writeJSON(w, Response{
			Status:     "200",
			StatusCode: http.StatusOK,
			Message:    "No potential matches",
		})
	}
}

func getIDs2(mrs []models.MeetRequest) []string {
	var ids []string
	for _, mr := range mrs {
		ids = append(ids, mr.User.ID)
	}
	return ids
}

func (re *Rest) GetMeetRequests(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	//use the preference to query neo4j for potential matches
	mrs, err := re.mbase.GetMeetRequests(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	if len(mrs) > 0 {
		ids := getIDs2(mrs)

		HydratedMrs, err := re.dbase.HydrateMeetRequests(ids, mrs)
		if err != nil {
			re.sLogger.Log.Errorln(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Println(HydratedMrs[0])
		re.writeJSON(w, Response{
			Status:     "200",
			StatusCode: http.StatusOK,
			Data: Data{
				MeetRequests: HydratedMrs,
			},
		})
	} else {
		re.writeJSON(w, Response{
			Status:     "200",
			StatusCode: http.StatusOK,
			Message:    "No meet requests",
		})
	}
}

func getRecipientIDs(chats []models.Chat, userID string) []string {
	recipientIDs := make([]string, 0)

	for _, chat := range chats {
		if len(chat.Recipients) == 1 {
			if chat.Recipients[0].ID == userID {
				recipientIDs = append(recipientIDs, chat.NonRecipient.ID)
			}
			// skip this chat you unmatch the user.

		} else {
			for _, recipient := range chat.Recipients {
				if recipient.ID != userID {
					recipientIDs = append(recipientIDs, recipient.ID)
				}
			}
		}

	}
	return recipientIDs
}

func (re *Rest) CheckFreeTrialAvailability(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	creds, err := re.dbase.GetCredential(email)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	duration := time.Now().UTC().Sub(creds.FreeTrialCountIssueTimestamp)

	if duration > 24*time.Hour {
		// Reset
		err := re.dbase.UpdateFreeTrialCount(email, 5, time.Now().UTC())
		if err != nil {
			re.sLogger.Log.Errorln(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		re.writeJSON(w, Response{
			Status:     "200",
			StatusCode: http.StatusOK,
			Data: Data{
				FreeTrialResetTime: 24,
				FreeTrialCount:     5,
			},
		})
		return
	} else {
		durationInHours := 24 - math.Round(duration.Hours())

		// Ensure duration is not negative
		if durationInHours < 0 {
			durationInHours = 0
		}

		//! add to another code decrease trial count.
		// err := re.dbase.UpdateFreeTrialCount(email, creds.FreeTrialCount-1, creds.FreeTrialCountIssueTimestamp)
		// if err != nil {
		// 	re.sLogger.Log.Errorln(err)
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }

		re.writeJSON(w, Response{
			Status:     "200",
			StatusCode: http.StatusOK,
			Data: Data{
				FreeTrialResetTime: durationInHours,
				FreeTrialCount:     creds.FreeTrialCount,
			},
		})
		return

	}
}

// this returns all the ids of the other party of the chats for hydration, ut also takes note of chat with only the user due to unmatching and it returns them too.
// if recipients is 1 and its my id then get the id from former_recipeint but if the id is not mine skip.

func (re *Rest) GetChats(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	//use the preference to query neo4j for potential matches
	chats, err := re.dbase.GetChats(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	ids := getRecipientIDs(*chats, id)

	fmt.Println("----------------------------", ids)
	fmt.Println("----------------------------", chats)

	var HydratedChats []models.Chat

	if len(ids) > 0 {
		fmt.Println("----------------------------Brooo")
		HydratedChats, err = re.dbase.HydrateChats(ids, chats)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			re.sLogger.Log.Errorln(err)
			return
		}
	}

	fmt.Println("----------------------------HydratedChats", HydratedChats)

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Data: Data{
			Chats: HydratedChats,
		},
	})
}

func (re *Rest) GetProfile(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	// !  update this to call its seperate get profile
	usr, err := re.dbase.GetPotentialMatch(id)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Data: Data{
			User: usr,
		},
	})
}

func (re *Rest) GetAccount(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	usr, err := re.dbase.GetCredential(email)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	parts := strings.Split(usr.Email, "@")
	if len(parts) != 2 {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusBadRequest)
	}

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) > 1 {
		censoredLocalPart := string(localPart[0]) + strings.Repeat("*", len(localPart)-1)
		usr.Email = fmt.Sprintf("%s@%s", censoredLocalPart, domain)
	}

	fmt.Println(usr.Email)

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Data: Data{
			User: *usr,
		},
	})
}

func (re *Rest) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	fmt.Println(r.PostFormValue("location"))

	locStr := r.PostFormValue("location")

	var loc models.Location
	if err := json.Unmarshal([]byte(locStr), &loc); err != nil {
		http.Error(w, "Invalid interests data", http.StatusBadRequest)
		return
	}

	err = re.dbase.UpdateLocation(id, loc)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		re.writeJSON(w, Response{
			Status:     "500",
			StatusCode: http.StatusInternalServerError,
			Message:    "Error saving data to database",
		})
		return
	}

	if loc.Lat != 0.0 && loc.Lon != 0.0 {
		// Store credentials to metabase.
		err = re.mbase.UpdateUserLocation(id, loc.Lat, loc.Lon)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			re.sLogger.Log.Errorln(err)
			return
		}
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "",
	})
}

func (re *Rest) UpdateNotification(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	fmt.Println(r.PostFormValue("location"))

	notiStr := r.PostFormValue("notification")

	var noti models.Notification
	if err := json.Unmarshal([]byte(notiStr), &noti); err != nil {
		http.Error(w, "Invalid interests data", http.StatusBadRequest)
		return
	}

	err = re.dbase.UpdateNotification(id, noti)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		re.writeJSON(w, Response{
			Status:     "500",
			StatusCode: http.StatusInternalServerError,
			Message:    "Error saving data to database",
		})
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "",
	})
}

func (re *Rest) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	usrStr := r.PostFormValue("user")

	var usr models.User
	if err := json.Unmarshal([]byte(usrStr), &usr); err != nil {
		re.sLogger.Log.Errorln(err)
		http.Error(w, "Invalid interests data", http.StatusBadRequest)
		return
	}

	dateOfBirthStr := r.PostFormValue("dob")

	// Parse the date of birth
	usr.DOB, err = time.Parse(time.RFC3339, dateOfBirthStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	fmt.Println(id, usr)

	err = re.dbase.UpdateProfile(id, usr)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		re.writeJSON(w, Response{
			Status:     "500",
			StatusCode: http.StatusInternalServerError,
			Message:    "Error saving data to database",
		})
		return
	}

	// Store credentials to metabase.
	err = re.mbase.UpdateProfile(id, usr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "",
	})
}

func (re *Rest) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	usrStr := r.PostFormValue("user")

	var usr models.User
	if err := json.Unmarshal([]byte(usrStr), &usr); err != nil {
		re.sLogger.Log.Errorln(err)
		http.Error(w, "Invalid interests data", http.StatusBadRequest)
		return
	}

	err = re.dbase.UpdateAccount(id, usr)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		re.writeJSON(w, Response{
			Status:     "500",
			StatusCode: http.StatusInternalServerError,
			Message:    "Error saving data to database",
		})
		return
	}

	// Store credentials to metabase.
	err = re.mbase.UpdateAccount(id, usr.IsPaused)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "",
	})
}

func (re *Rest) DeactivateAccount(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	err := re.dbase.DeactivateAccount(id, "User")
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		re.writeJSON(w, Response{
			Status:     "500",
			StatusCode: http.StatusInternalServerError,
			Message:    "Error saving data to database",
		})
		return
	}

	// Store credentials to metabase.
	err = re.mbase.DeactivateAccount(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "",
	})
}

func (re *Rest) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	cp := r.FormValue("current-password")
	np := r.FormValue("new-password")

	creds, err := re.dbase.GetCredential(r.URL.Query().Get("email"))
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(creds.Password), []byte(cp)); err != nil {
		http.Error(w, "Email or password incorrect", http.StatusUnauthorized)
		re.sLogger.Log.Errorln(err)
		return
	}

	// Generate hash from raw password string and store.
	if bytes, err := bcrypt.GenerateFromPassword([]byte(np), 5); err == nil {
		np = string(bytes)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	err = re.dbase.UpdatePassword(r.URL.Query().Get("email"), np)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Password updated successfully",
	})
}

// uniqueValues returns the values that are not common between two slices of strings

func (re *Rest) HandleGlassfyWebhook(w http.ResponseWriter, r *http.Request) {
	// Get the bearer token from environment variable
	expectedToken := os.Getenv("GLASSFY_WEBHOOK_AUTH_BEARER")

	// Extract the token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Verify the token
	if token != expectedToken {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var payload models.WebhookPayload
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, "can't parse JSON", http.StatusBadRequest)
		return
	}

	// Handle the webhook payload
	fmt.Printf("Received webhook: %+v\n", payload)

	// Process event type
	switch payload.Type {
	case 5001:
		fmt.Println("Subscription Initial Buy event")
		err := re.dbase.UpdateSubscription(payload.CustomID, "Premium")
		if err != nil {
			re.sLogger.Log.Errorln(err)
			return
		}

		err = re.mbase.UpdateUserBoost(payload.CustomID, 10)
		if err != nil {
			re.sLogger.Log.Errorln(err)
			return
		}

		err = re.dbase.AddTransaction(payload)
		if err != nil {
			re.sLogger.Log.Errorln(err)
			return
		}

	//set user to premium and boost user in meta db and store tx in tx clx
	case 5002:
		fmt.Println("Subscription Restarted event")
		err := re.dbase.UpdateSubscription(payload.CustomID, "Premium")
		if err != nil {
			re.sLogger.Log.Errorln(err)
			return
		}

		err = re.mbase.UpdateUserBoost(payload.CustomID, 10)
		if err != nil {
			re.sLogger.Log.Errorln(err)
			return
		}

		err = re.dbase.AddTransaction(payload)
		if err != nil {
			re.sLogger.Log.Errorln(err)
			return
		}
	case 5003:
		//set user to premium and boost user in meta db and store tx in tx clx
		fmt.Println("Subscription automatically renewed")

		err = re.dbase.AddTransaction(payload)
		if err != nil {
			re.sLogger.Log.Errorln(err)
			return
		}

	case 5004:
		fmt.Println("Subscription Expired event")
		err := re.dbase.UpdateSubscription(payload.CustomID, "Free")
		if err != nil {
			re.sLogger.Log.Errorln(err)
			return
		}

		err = re.mbase.UpdateUserBoost(payload.CustomID, 0)
		if err != nil {
			re.sLogger.Log.Errorln(err)
			return
		}
	// set user to free and remove boost from meta bd
	// case 5005:
	// 	fmt.Println("Subscription Did Change Renewal Status event")
	case 5006:
		fmt.Println("User is in Billing Retry Period event")
		err := re.dbase.UpdateSubscription(payload.CustomID, "Free")
		if err != nil {
			re.sLogger.Log.Errorln(err)
			return
		}

		err = re.mbase.UpdateUserBoost(payload.CustomID, 0)
		if err != nil {
			re.sLogger.Log.Errorln(err)
			return
		}
		// case 5007:
		// 	fmt.Println("Subscription Product Change event")
		// case 5008:
		// 	fmt.Println("In App Purchase event")
		// case 5009:
		// 	fmt.Println("Subscription Refund event")
		// case 5010:
		// 	fmt.Println("Subscription Paused event")
		// case 5011:
		// 	fmt.Println("Subscription Resumed event")
		// case 5012:
		// 	fmt.Println("Connect License event")
		// case 5013:
		// 	fmt.Println("Disconnect License event")
	default:
		fmt.Println("Unknown event type")
	}

	w.WriteHeader(http.StatusOK)
}

// ~ Config
func (re *Rest) GetLatestStableBuildNumber(w http.ResponseWriter, r *http.Request) {

	lsbn, err := re.dbase.GetLatestStableBuildNumber()
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Data: Data{
			LatestStableBuildNumber: lsbn,
		},
	})
}

func (re *Rest) GetTerms(w http.ResponseWriter, r *http.Request) {

	terms, err := re.dbase.GetTerms()
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Data: Data{
			Terms: terms,
		},
	})
}

func (re *Rest) GetPrivacyPolicy(w http.ResponseWriter, r *http.Request) {

	pp, err := re.dbase.GetPrivacyPolicy()
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Data: Data{
			PrivacyPolicy: pp,
		},
	})
}

func (re *Rest) GetHowLoveairWorks(w http.ResponseWriter, r *http.Request) {

	howLoveairWorks, err := re.dbase.GetHowLoveairWorks()
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Data: Data{
			HowLoveairWorks: howLoveairWorks,
		},
	})
}

// CustomID
// Price
// ProductID
// IsInBillingRetryPeriod
// CountryCode
// CurrencyCode
// AutoRenewStatus
// GracePeriodExpiresDateMS
// ExpireDateMS
// ExpirationIntent

// i need to know if the event was a subscription event
// i need to know if its a subscription renewal event
// i need to know if i was a expiration without renewal event
