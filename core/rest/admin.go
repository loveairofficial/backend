package rest

import (
	"encoding/json"
	"fmt"
	"loveair/models"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func (re *Rest) AdminLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	creds, err := re.dbase.GetAdminCredential(r.PostForm.Get("email"))
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(creds.Password), []byte(r.PostForm.Get("password"))); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		re.sLogger.Log.Errorln(err)
		return
	}

	// Generate jwt access token.
	atknString, err := re.generateAdminAccessTkn(accessTknExpiration, r.PostForm.Get("email"), creds.Role)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	// Set the token and expiration time as a header
	tokenWithExpiration := fmt.Sprintf("%s|%d", atknString, time.Now().Add(accessTknExpiration).Unix())

	re.writeJSON(w, Response{
		Status:     "200 OK",
		StatusCode: 200,
		Message:    "Login Successfull",
		AdminData: AdminData{
			AccessTkn:      tokenWithExpiration,
			Name:           creds.Name,
			ProfilePicture: creds.ProfilePicture,
			Role:           creds.Role.Name,
		},
	})
}

// Users
func (re *Rest) GetUsers(w http.ResponseWriter, r *http.Request) {
	count, err := strconv.ParseInt(r.URL.Query().Get("count"), 10, 64)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		return
	}

	offset, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		return
	}

	// save data to database
	users, usersCount, err := re.dbase.GetUsers(count, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Errorln(err)
		return
	}

	fmt.Println(users)

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		AdminData: AdminData{
			Users:      *users,
			UsersCount: usersCount,
		},
	})
}

func (re *Rest) SuppressAccount(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	// Update database.
	err := re.dbase.SuppressAccount(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	// Update metabase.
	err = re.mbase.SuppressAccount(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	resStatus, err := re.emailIf.SendAccountSuppressionEmail(r.FormValue("email"), r.FormValue("firstName"))
	if err != nil || resStatus != 202 {
		re.sLogger.Log.Errorln(err)
	}

	re.writeJSON(w, Response{
		Status:     "201",
		StatusCode: http.StatusCreated,
		Message:    "Account Suppressed Successfully",
	})
}

func (re *Rest) UnSuppressAccount(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	// Update database.
	err := re.dbase.UnSuppressAccount(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	// Update metabase.
	err = re.mbase.UnSuppressAccount(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	re.writeJSON(w, Response{
		Status:     "201",
		StatusCode: http.StatusCreated,
		Message:    "Account UnSuppressed Successfully",
	})
}

// Roles
func (re *Rest) GetRoles(w http.ResponseWriter, r *http.Request) {
	// save data to database
	roles, err := re.dbase.GetRoles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Errorln(err)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		AdminData: AdminData{
			Roles: *roles,
		},
	})
}

// Admins
func (re *Rest) AddAdmin(w http.ResponseWriter, r *http.Request) {
	data := r.FormValue("data")
	var a models.Admin

	err := json.Unmarshal([]byte(data), &a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Errorln(err)
		return
	}

	// Check if there is an already existing account with the email.
	if err = re.dbase.CheckAdminCredential(a.Email); err != mongo.ErrNoDocuments {
		w.WriteHeader(http.StatusConflict)
		re.writeJSON(w, Response{
			Status:     "409",
			StatusCode: http.StatusConflict,
			Message:    "Account already exist with this email",
		})
		re.sLogger.Log.Errorln(err)
		return
	}

	// Generate hash from raw password string and store.
	if bytes, err := bcrypt.GenerateFromPassword([]byte(a.Password), 5); err == nil {
		a.Password = string(bytes)
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Errorln(err)
		return
	}

	a.IsActive = true
	a.Activities = []models.Activity{}
	a.Joined = time.Now().Format("2006-01-02 3:4:5 pm")

	// Store credentials to database.
	err = re.dbase.AddAdmin(&a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		re.sLogger.Log.Errorln(err)
		return
	}

	re.writeJSON(w, Response{
		Status:     "201",
		StatusCode: http.StatusCreated,
		Message:    "Admin Created Successfully",
	})
}

func (re *Rest) GetAdmins(w http.ResponseWriter, r *http.Request) {
	// save data to database
	admins, err := re.dbase.GetAdmins()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Errorln(err)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		AdminData: AdminData{
			Admins: *admins,
		},
	})
}
