package rest

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"loveair/base/cache"
	"loveair/base/data"
	"loveair/base/meta"
	"loveair/email"
	"loveair/log"
	"loveair/models"
	"math/big"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/schema"
	"github.com/rs/xid"
)

var decoder = schema.NewDecoder()

var (
	accessTknExpiration  = 24 * time.Hour
	refreshTknExpiration = 168 * time.Hour
)

type Claims struct {
	Email string
	DID   string
	jwt.StandardClaims
}

type Response struct {
	Status     string `json:"status,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
	Message    string `json:"message,omitempty"`
	Data       Data   `json:"data"`
}

type Data struct {
	AccessTkn             string               `json:"access_tkn"`
	RefreshTkn            string               `json:"refresh_tkn"`
	StreamTkn             string               `json:"stream_tkn"`
	IsOnboarded           bool                 `json:"is_onboarded"`
	Email                 string               `json:"email"`
	ID                    string               `json:"id"`
	FirstName             string               `json:"first_name"`
	LastName              string               `json:"last_name"`
	StageID               int                  `json:"stage_id"`
	Gender                string               `json:"gender"`
	DOB                   string               `json:"dob"`
	RelationshipIntention string               `json:"relationship_intention"`
	Interests             []string             `json:"interests"`
	Signature             string               `json:"signature"`
	Timestamp             string               `json:"timestamp"`
	IntroType             string               `json:"intro_type"`
	IntroVideoUri         string               `json:"intro_video_uri"`
	IntroAudioUri         string               `json:"intro_audio_uri"`
	Images                []models.Photo       `json:"images"`
	Preference            models.Preference    `json:"preference"`
	Address               string               `json:"address"`
	Vicinity              string               `json:"vicinity"`
	UTCOffset             int                  `json:"utc_offset"`
	Users                 []models.User        `json:"users"`
	RoseCount             int                  `json:"rose_count"`
	CallID                string               `json:"call_id"`
	MeetRequests          []models.MeetRequest `json:"meet_requests"`
	Subscription          string               `json:"subscription"`
	Chats                 []models.Chat        `json:"chats"`
	User                  models.User          `json:"user"`
	ProfilePicture        models.Photo         `json:"profilePicture"`
}

type Rest struct {
	secret  string
	dbase   data.Interface
	mbase   meta.Interface
	cbaseIf cache.Interface
	emailIf email.Interface
	sLogger log.SLoger
}

func InitRest(secret string, dbase data.Interface, mbase meta.Interface, cbaseIf cache.Interface, emailIf email.Interface, sLogger log.SLoger) *Rest {
	return &Rest{
		secret,
		dbase,
		mbase,
		cbaseIf,
		emailIf,
		sLogger,
	}
}

func (re *Rest) writeJSON(w http.ResponseWriter, v interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		re.sLogger.Log.Errorln(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		re.sLogger.Log.Errorln(err)
		return
	}
}

func (re *Rest) generateUID() string {
	// Generate unique ID.
	uid := xid.New()
	return uid.String()
}

func (re *Rest) generateAccessTkn(duration time.Duration, email string) (string, error) {
	expirationTime := time.Now().Add(duration)

	claims := &Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tknString, err := tkn.SignedString([]byte(re.secret))
	if err != nil {
		return "", err
	}

	return tknString, nil
}

func (re *Rest) generateRefreshTkn(duration time.Duration, email, did string) (string, string, error) {
	expirationTime := time.Now().Add(duration)

	if did == "" {
		did = re.generateUID()
	}

	claims := &Claims{
		Email: email,
		DID:   did,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tknString, err := tkn.SignedString([]byte(re.secret))

	if err != nil {
		return "", "", err
	}
	return tknString, did, nil
}

func GenerateRandomPin() (string, error) {
	// The maximum value for a 4-digit PIN is 9999
	max := big.NewInt(10000) // The upper limit is exclusive, so use 10000

	// Generate a random number between 0 and 9999
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	// Convert the number to a 4-digit string with leading zeros if necessary
	pin := fmt.Sprintf("%04d", n.Int64())

	return pin, nil
}
