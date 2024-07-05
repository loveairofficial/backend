package rest

import (
	"encoding/json"
	"fmt"
	"loveair/models"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	stream "github.com/GetStream/stream-chat-go/v6"
	"github.com/cloudinary/cloudinary-go/api"
	"github.com/dgrijalva/jwt-go"
	"github.com/houseme/mobiledetect/ua"
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

	//sign-up user
	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Sign-Up",
	})
}

func (re *Rest) SignUp(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	// Create a new account.
	usr := new(models.User)

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
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	usrAgent := ua.New(r.UserAgent())

	// add new device to database.
	err = re.dbase.AddNewDevice(&models.Device{
		DeviceID:    did,
		Device:      usrAgent.Device(),
		Platform:    usrAgent.Platform(),
		OSName:      usrAgent.OSInfo().Name,
		OSVersion:   usrAgent.OSInfo().Version,
		BrowserName: usrAgent.UserAgentBrowser().Name,
	}, usr.Email)

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
	sToken, err := client.CreateToken(usr.ID, time.Now().Add(accessTknExpiration))
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	creds, err := re.dbase.GetCredential(r.PostForm.Get("email"))
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.sLogger.Log.Infoln("------------------", creds)

	if err = bcrypt.CompareHashAndPassword([]byte(creds.Password), []byte(r.PostForm.Get("password"))); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		http.Error(w, "Email or password incorrect", http.StatusUnauthorized)
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

	usrAgent := ua.New(r.UserAgent())

	// add new device to database.
	err = re.dbase.AddNewDevice(&models.Device{
		DeviceID:    did,
		Device:      usrAgent.Device(),
		Platform:    usrAgent.Platform(),
		OSName:      usrAgent.OSInfo().Name,
		OSVersion:   usrAgent.OSInfo().Version,
		BrowserName: usrAgent.UserAgentBrowser().Name,
	}, creds.Email)

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
	sToken, err := client.CreateToken(creds.ID, time.Now().Add(accessTknExpiration))
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
	if rtk = r.Header.Get("Authorization"); rtk == "" {
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
	signature, err := api.SignParameters(ParamsToSign, "YoVyOQ-uoCP3CVhqB0CoohPIxT0")

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

		re.writeJSON(w, Response{
			Status:     "200",
			StatusCode: http.StatusOK,
			Data: Data{
				MeetRequests: HydratedMrs,
			},
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

// uniqueValues returns the values that are not common between two slices of strings

/**
Step1: check if email exist
step2: if it exist, let frontend know so user can add password and login/authenticate.
step3: if it doesnt send user verification code and start account signup process
**/
