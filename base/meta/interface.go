package meta

import (
	"loveair/models"
	"time"
)

// Base is used to interface (plug and play) multiple Meta Database.
type Interface interface {
	AddUser(models.User) error
	UpdateUserInfo(string, *models.User) error
	UpdateUserLocation(string, float64, float64) error
	UpdateProfile(string, models.User) error
	UpdateAccount(string, bool) error

	//Potential matches
	GetPotentialMatches(string, *models.Preference) ([]models.User, error)
	UpdateUserPresence(string, string) error
	AddRequestedToMeetRelationship(*models.MeetRequest) error
	AddMatchRelationship(time.Time, string, string) error
	AddPassRelationship(time.Time, string, string) error
	GetMeetRequests(string) ([]models.MeetRequest, error)
	AddUnmatchRelationship(time.Time, string, string) error

	//Subscription Boost
	UpdateUserBoost(string, int) error
}
