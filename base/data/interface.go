package data

import (
	"loveair/models"
	"time"
)

type Interface interface {
	//~ Client
	VerifyEmailExist(string) error
	AddUser(*models.User) error
	GetCredential(string) (*models.User, error)

	//Device
	AddNewDevice(*models.Device, string) error
	GetDevice(string, string) (*models.Device, error)
	DeleteDevice(string, string) error
	GetUserPushNotificationIDs(id string) ([]string, error)

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
	UpdateNotification(string, string, models.Notification) error
	UpdateProfile(string, models.User) error
	UpdateAccount(string, models.User) error
	UpdatePassword(string, string) error
	DeactivateAccount(string, string) error
	ReactivateAccount(string) error

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

	//Feedback
	AddFeedback(models.Feedback) error

	//Subscription
	UpdateSubscription(string, string) error
	AddTransaction(models.WebhookPayload) error

	UpdateFreeTrialCount(string, int, time.Time) error

	//Config
	GetLatestStableBuildNumber() (int, error)
	GetTerms() (string, error)
	GetPrivacyPolicy() (string, error)
	GetHowLoveairWorks() (string, error)

	//~ Admin
	//Users
	GetAdminCredential(string) (*models.Admin, error)
	CheckAdminCredential(string) error
	GetUsers(int64, int64) (*[]models.User, int64, error)

	// Roles
	GetRoles() (*[]models.Role, error)

	//Admins
	AddAdmin(*models.Admin) error
	GetAdmins() (*[]models.Admin, error)
	SuppressAccount(string) error
	UnSuppressAccount(string) error
}
