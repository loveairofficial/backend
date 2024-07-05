package rest

import (
	"encoding/json"
	"loveair/models"
	"net/http"
	"strconv"
	"time"
)

func (re *Rest) GetStage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	stageID, err := re.dbase.GetStage(id)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Saved Successful",
		Data: Data{
			StageID: stageID,
		},
	})
}

func (re *Rest) HandleStageOne(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	id, err := strconv.Atoi(r.PostForm.Get("stageID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	err = re.dbase.SaveStageOne(id, r.PostForm.Get("gender"), r.PostForm.Get("userID"))
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Saved Successful",
	})
}

func (re *Rest) GetStageOne(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	gender, err := re.dbase.GetStageOne(id)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Saved Successful",
		Data: Data{
			Gender: gender,
		},
	})
}

func (re *Rest) HandleStageTwo(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	id, err := strconv.Atoi(r.PostForm.Get("stageID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	dateOfBirthStr := r.PostFormValue("date")

	// Parse the date of birth
	dateOfBirth, err := time.Parse(time.RFC3339, dateOfBirthStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	err = re.dbase.SaveStageTwo(id, dateOfBirth, r.PostForm.Get("userID"))
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Saved Successful",
	})
}

func (re *Rest) GetStageTwo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	dob, err := re.dbase.GetStageTwo(id)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Format the time object to ISO 8601 string format for the frontend.
	dateOfBirthStr := dob.Format(time.RFC3339)

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Saved Successful",
		Data: Data{
			DOB: dateOfBirthStr,
		},
	})
}

func (re *Rest) HandleStageThree(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	id, err := strconv.Atoi(r.PostForm.Get("stageID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	err = re.dbase.SaveStageThree(id, r.PostForm.Get("relationshipIntention"), r.PostForm.Get("userID"))
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Saved Successful",
	})
}

func (re *Rest) GetStageThree(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	ri, err := re.dbase.GetStageThree(id)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Saved Successful",
		Data: Data{
			RelationshipIntention: ri,
		},
	})
}

func (re *Rest) HandleStageFour(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	id, err := strconv.Atoi(r.PostForm.Get("stageID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	// Retrieve the interests array from the form data
	interestsStr := r.PostFormValue("interests")

	var interests []string
	if err := json.Unmarshal([]byte(interestsStr), &interests); err != nil {
		http.Error(w, "Invalid interests data", http.StatusBadRequest)
		return
	}

	err = re.dbase.SaveStageFour(id, interests, r.PostForm.Get("userID"))
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Saved Successful",
	})
}

func (re *Rest) GetStageFour(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	interests, err := re.dbase.GetStageFour(id)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Saved Successful",
		Data: Data{
			Interests: interests,
		},
	})
}

func (re *Rest) HandleStageFive(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	id, err := strconv.Atoi(r.PostForm.Get("stageID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	// Retrieve the interests array from the form data
	introStr := r.PostFormValue("intro")

	var intro models.Intro

	if err := json.Unmarshal([]byte(introStr), &intro); err != nil {
		http.Error(w, "Invalid intro data", http.StatusBadRequest)
		return
	}

	err = re.dbase.SaveStageFive(id, intro, r.PostForm.Get("userID"))
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Saved Successful",
	})
}

func (re *Rest) GetStageFive(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	videoInroUri, audioIntroUri, introType, err := re.dbase.GetStageFive(id)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Saved Successful",
		Data: Data{
			IntroVideoUri: videoInroUri,
			IntroAudioUri: audioIntroUri,
			IntroType:     introType,
		},
	})
}

func (re *Rest) HandleStageSix(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	id, err := strconv.Atoi(r.PostForm.Get("stageID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		re.sLogger.Log.Error(err)
		return
	}

	// Retrieve the interests array from the form data
	photosStr := r.PostFormValue("images")

	var photos []models.Photo
	if err := json.Unmarshal([]byte(photosStr), &photos); err != nil {
		http.Error(w, "Invalid interests data", http.StatusBadRequest)
		return
	}

	err = re.dbase.SaveStageSix(id, photos, r.PostForm.Get("userID"))
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Saved Successful",
	})
}

func (re *Rest) GetStageSix(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	images, err := re.dbase.GetStageSix(id)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Saved Successful",
		Data: Data{
			Images: images,
		},
	})
}

func (re *Rest) HandleStageCompletion(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	//!Add user onboading details to metabase.
	usr, err := re.dbase.GetUserInfo(id)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// set it to metabase
	err = re.mbase.UpdateUserInfo(id, usr)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = re.dbase.HandleStageCompletion(id)
	if err != nil {
		re.sLogger.Log.Errorln(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	re.writeJSON(w, Response{
		Status:     "200",
		StatusCode: http.StatusOK,
		Message:    "Saved Successful",
		Data: Data{
			IsOnboarded: true,
		},
	})
}
