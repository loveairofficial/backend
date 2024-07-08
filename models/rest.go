package models

import "time"

type User struct {
	ID                    string    `json:"id" bson:"id"`
	IsActive              bool      `json:"isActive" bson:"is_active"`
	IsDeactivated         bool      `json:"isDeactivated" bson:"is_deactivated"`
	IsPaused              bool      `json:"isPaused" bson:"is_paused"`
	Verification          bool      `json:"verification" bson:"verification"`
	FirstName             string    `json:"firstName" bson:"first_name"`
	LastName              string    `json:"lastName" bson:"last_name"`
	Email                 string    `json:"email" bson:"email"`
	Provider              string    `json:"provider" bson:"provider"` //Authentication provider eg. loveair, google or apple
	Phone                 string    `json:"phone" bson:"phone"`
	Password              string    `json:"password" bson:"password"`
	IsOnboarded           bool      `json:"isOnboarded" bson:"is_onboarded"`
	StageID               int       `json:"stageID" bson:"stage_ID"`
	Subscription          string    `json:"subscription,omitempty" bson:"subscription"`
	Gender                string    `json:"gender,omitempty" bson:"gender"`
	DOB                   time.Time `json:"dob,omitempty" bson:"dob"`
	RelationshipIntention string    `json:"relationshipIntention,omitempty" bson:"relationship_intention"`
	Interests             []string  `json:"interests,omitempty" bson:"interests"`
	Religion              string    `json:"religion" bson:"religion"`
	ProfilePicture        Photo     `json:"profilePicture" bson:"profile_picture"`

	IntroType           string   `json:"introType,omitempty" bson:"intro_type"`
	IntroVideoUri       string   `json:"introVideoUri,omitempty" bson:"intro_video_uri"`
	IntroVideoThumbnail string   `json:"introVideoThumbnail,omitempty" bson:"intro_video_thumbnail"`
	IntroAudioUri       string   `json:"introAudioUri,omitempty" bson:"intro_audio_uri"`
	Photos              []Photo  `json:"photos,omitempty" bson:"photos"`
	Location            Location `json:"location,omitempty" bson:"location"`

	Address   string `json:"address,omitempty" bson:"address"`
	Vicinity  string `json:"vicinity,omitempty" bson:"vicinity"`
	UTCOffset int    `json:"utcOffset,omitempty" bson:"utc_offset"`

	Devices             []Device   `json:"devices,omitempty" bson:"devices"`
	JoinedAt            time.Time  `json:"joinedAt" bson:"joined_at"`
	Preference          Preference `json:"preference" bson:"preference"`
	Presence            string     `json:"presence" bson:"presence"`
	MutualInterestCount int64      `json:"mutualInterestCount"`
	MutualInterest      []string   `json:"mutualInterest"`
	ExclusiveInterest   []string   `json:"exclusiveInterest"`
	RoseCount           int        `json:"roseCount" bson:"rose_count"`
}

type Location struct {
	Lat      float64 `json:"lat" bson:"lat"`
	Lon      float64 `json:"lon" bson:"lon"`
	Address  string  `json:"address,omitempty" bson:"address"`
	Vicinity string  `json:"vicinity,omitempty" bson:"vicinity"`
}

// User info stored in neo4j
type MetaUser struct{}

type Device struct {
	// Device ID is the unique identifier for the device assigned the token.
	DeviceID string `json:"device_id,omitempty" bson:"device_id"`

	Device   string `json:"device,omitempty" bson:"device"`
	Platform string `json:"platform,omitempty" bson:"platform"`
	// OS name (Eg: “Windows”)
	OSName string `json:"os_name,omitempty" bson:"os_name"`
	// OS version (e.g. "Android", "iOS"))
	OSVersion string `json:"os_version,omitempty" bson:"os_version"`
	//Browser name (Eg: “Chrome”)
	BrowserName string `json:"browser_name,omitempty" bson:"browser_name"`
}

type Preference struct {
	InterestedIn          []string  `json:"interestedIn" bson:"interested_in"`
	RelationshipIntention []string  `json:"relationshipIntention" bson:"relationship_intention"`
	AgeRange              Range     `json:"ageRange" bson:"age_range"`
	GeoCircle             GeoCircle `json:"geoCircle" bson:"geo_circle"`
	Global                bool      `json:"global" bson:"global"`
	Religion              []string  `json:"religion" bson:"religion"`
	Presence              string    `json:"presence" bson:"presence"`
}

type Range struct {
	Min int `json:"min" bson:"min"`
	Max int `json:"max" bson:"max"`
}

type GeoCircle struct {
	Lat    float64 `json:"lat" bson:"lat"`
	Lon    float64 `json:"lon" bson:"lon"`
	Radius float64 `json:"radius" bson:"radius"`
	Unit   string  `json:"unit" bson:"unit"`
}

// Convert a date of birth to an int
// dob := time.Date(1990, 8, 15, 0, 0, 0, 0, time.UTC)
// dobInt := int(time.Since(zeroDate).Hours() / 24)

type Photo struct {
	Key      string `json:"key"`
	ID       int    `json:"id"`
	URI      string `json:"uri"`
	PublicID string `json:"publicID"`
	IsEmpty  bool   `json:"isEmpty"`
}

type Intro struct {
	URI       string `json:"uri"`
	IntroType string `json:"introType"`
}

// Mutual interest extraction
type Interest struct {
	Name string `json:"name"`
}

type UserInterest struct {
	ID           int      `json:"id"`
	Type         int      `json:"type"`
	DetailsID    string   `json:"details_id"`
	InterestType string   `json:"interest_type"`
	Interest     Interest `json:"interest"`
}

type Report struct {
	ID          string    `json:"id" bson:"id"`
	Type        string    `json:"type" bson:"type,"`
	Status      string    `json:"status" bson:"status,"` // eg. pending, resolved
	SenderID    string    `json:"senderid" bson:"sender_id,"`
	RecipientID string    `json:"recipientid," bson:"recipient_id"`
	Timestamp   time.Time `json:"timestamp" bson:"timestamp,"`
}

type Feedback struct {
	ID          string    `json:"id" bson:"id"`
	Content     string    `json:"content" bson:"content,"`
	Status      string    `json:"status" bson:"status,"` // eg. pending, resolved
	SenderID    string    `json:"senderid" bson:"sender_id,"`
	RecipientID string    `json:"recipientid," bson:"recipient_id"`
	Timestamp   time.Time `json:"timestamp" bson:"timestamp,"`
}
