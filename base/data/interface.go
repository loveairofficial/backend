package data

import (
	"loveair/models"
	"time"
)

type Interface interface {
	VerifyEmailExist(string) error
	AddUser(*models.User) error
	GetCredential(string) (*models.User, error)

	//Device
	AddNewDevice(*models.Device, string) error
	DeleteDevice(string, string) error

	// Onboarding
	GetStage(string) (int, error)

	SaveStageOne(int, string, string) error
	GetStageOne(string) (string, error)

	SaveStageTwo(int, time.Time, string) error
	GetStageTwo(string) (time.Time, error)

	SaveStageThree(int, string, string) error
	GetStageThree(string) (string, error)

	SaveStageFour(int, []string, string) error
	GetStageFour(string) ([]string, error)

	SaveStageFive(int, models.Intro, string) error
	GetStageFive(string) (string, string, string, error)

	SaveStageSix(int, []models.Photo, string) error
	GetStageSix(string) ([]models.Photo, error)

	HandleStageCompletion(string) error
	GetUserInfo(string) (*models.User, error)

	//Profile
	UpdateLocation(string, models.Location) error
	UpdateProfile(string, models.User) error
	UpdateAccount(string, models.User) error

	// Preference
	UpdatePreference(string, models.Preference, string, string, int) error
	GetPreference(string) (models.Preference, string, string, int, int, string, error)

	//Potential Matches
	HydratePotentialMatches([]string, []models.User) ([]models.User, error)
	GetPotentialMatch(string) (models.User, error)

	//Meet Requests
	HydrateMeetRequests([]string, []models.MeetRequest) ([]models.MeetRequest, error)

	//Chat
	AddChat(*models.Chat) error
	GetChats(string) (*[]models.Chat, error)
	HydrateChats([]string, *[]models.Chat) ([]models.Chat, error)
	AddMessage(*models.Message) error
	UpdateMessageStatus(string, []string) error
	RemoveUserFromChat(string, string) error
	MergeCachedSession([]models.Message) error

	//Report
	AddReport(models.Report) error
}