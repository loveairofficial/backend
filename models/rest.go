package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                    string    `json:"id" bson:"id"`
	IsActive              bool      `json:"isActive" bson:"is_active"`
	DeactivatedBy         string    `json:"deactivatedBy" bson:"deactivated_by"`
	DeactivationDate      time.Time `json:"deactivationDate" bson:"deactivation_date"`
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
	FreeDailyTrialCount   int       `json:"free_daily_trial_count,omitempty" bson:"free_daily_trial_count"`
	Gender                string    `json:"gender,omitempty" bson:"gender"`
	DOB                   time.Time `json:"dob,omitempty" bson:"dob"`
	RelationshipIntention string    `json:"relationshipIntention,omitempty" bson:"relationship_intention"`
	Interests             []string  `json:"interests,omitempty" bson:"interests"`
	Religion              string    `json:"religion" bson:"religion"`
	ProfilePicture        Photo     `json:"profilePicture" bson:"profile_picture"`

	IntroType           string       `json:"introType,omitempty" bson:"intro_type"`
	IntroVideoUri       string       `json:"introVideoUri,omitempty" bson:"intro_video_uri"`
	IntroVideoThumbnail string       `json:"introVideoThumbnail,omitempty" bson:"intro_video_thumbnail"`
	IntroAudioUri       string       `json:"introAudioUri,omitempty" bson:"intro_audio_uri"`
	Photos              []Photo      `json:"photos,omitempty" bson:"photos"`
	Location            Location     `json:"location,omitempty" bson:"location"`
	Notification        Notification `json:"notification,omitempty" bson:"notification"`

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

type Notification struct {
	Email bool `json:"email" bson:"email"`
	Push  bool `json:"push" bson:"push"`
}

// User info stored in neo4j
type MetaUser struct{}

type Device struct {
	// Device ID is the unique identifier for the device assigned the token.
	DeviceID string `json:"device_id,omitempty" bson:"device_id"`
	PushTkn  string `json:"push_tkn" bson:"push_tkn"`
	// OS name (Eg: “Windows”)
	OSName    string `json:"os_name,omitempty" bson:"os_name"`
	Brand     string `json:"brand,omitempty" bson:"brand"`
	ModelName string `json:"model_name,omitempty" bson:"model_name"`
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
	Context     string    `json:"context" bson:"context"`
	SenderID    string    `json:"senderid" bson:"sender_id,"`
	RecipientID string    `json:"recipientid," bson:"recipient_id"`
	Timestamp   time.Time `json:"timestamp" bson:"timestamp,"`
}

type Feedback struct {
	ID        string    `json:"id" bson:"id"`
	Content   string    `json:"content" bson:"content,"`
	Status    string    `json:"status" bson:"status,"` // eg. pending, resolved
	SenderID  string    `json:"senderid" bson:"sender_id,"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp,"`
}

// type WebhookPayload struct {
// 	AdjustID                 string  `json:"adjustid"`
// 	AID                      string  `json:"aid"`
// 	AppVersion               string  `json:"app_version"`
// 	AppID                    string  `json:"appid"`
// 	AppsFlyerID              string  `json:"appsflyerid"`
// 	ASID                     string  `json:"asid"`
// 	AutoRenewProductID       string  `json:"auto_renew_product_id"`
// 	AutoRenewStatus          bool    `json:"auto_renew_status"`
// 	BundleVersion            string  `json:"bundle_version"`
// 	CountryCode              string  `json:"country_code"`
// 	CurrencyCode             string  `json:"currency_code"`
// 	CustomID                 string  `json:"customid"`
// 	DateMS                   int64   `json:"date_ms"`
// 	Device                   string  `json:"device"`
// 	Environment              string  `json:"environment"`
// 	Estimated                int     `json:"estimated"`
// 	EventDate                int64   `json:"event_date"`
// 	ExpirationIntent         string  `json:"expiration_intent"`
// 	ExpireDateMS             int64   `json:"expire_date_ms"`
// 	GAID                     string  `json:"gaid"`
// 	GracePeriodExpiresDateMS int64   `json:"grace_period_expires_date_ms"`
// 	GroupIdentifier          string  `json:"group_identifier"`
// 	ID                       string  `json:"id"`
// 	IDFA                     string  `json:"idfa"`
// 	IDFV                     string  `json:"idfv"`
// 	IP                       string  `json:"ip"`
// 	IsInBillingRetryPeriod   bool    `json:"is_in_billing_retry_period"`
// 	LicenseCode              string  `json:"licensecode"`
// 	OfferCodeRefName         string  `json:"offer_code_ref_name"`
// 	OfferingID               string  `json:"offeringid"`
// 	OriginalPurchaseDateMS   int64   `json:"original_purchase_date_ms"`
// 	OriginalTransactionID    string  `json:"original_transaction_id"`
// 	PackageName              string  `json:"packagename"`
// 	Price                    float64 `json:"price"`
// 	PriceConsentStatus       string  `json:"price_consent_status"`
// 	PriceUSD                 float64 `json:"price_usd"`
// 	ProductID                string  `json:"productid"`
// 	Quantity                 int     `json:"quantity"`
// 	SDKVersion               string  `json:"sdk_version"`
// 	SortDateMS               int64   `json:"sort_date_ms"`
// 	Source                   string  `json:"source"`
// 	Store                    string  `json:"store"`
// 	SubPlatform              string  `json:"sub_platform"`
// 	SubscriberID             string  `json:"subscriberid"`
// 	SystemVersion            string  `json:"system_version"`
// 	TransactionID            string  `json:"transaction_id"`
// 	Type                     int     `json:"type"`
// 	UserUnknown              bool    `json:"userunknown"`
// 	VendorID                 string  `json:"vendorid"`
// 	WebOrderLineItemID       string  `json:"web_order_line_item_id"`
// }

type WebhookPayload struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	EventID            string             `bson:"event_id,omitempty" json:"event_id,omitempty"`
	Type               int                `bson:"type" json:"type"`
	ExpireDateMS       int64              `bson:"expire_date_ms" json:"expire_date_ms"`
	AutoRenewProductID string             `bson:"auto_renew_product_id,omitempty" json:"auto_renew_product_id,omitempty"`
	ProductID          string             `bson:"product_id,omitempty" json:"product_id,omitempty"`
	TransactionID      string             `bson:"transaction_id,omitempty" json:"transaction_id,omitempty"`
	SubscriberID       string             `bson:"subscriber_id,omitempty" json:"subscriber_id,omitempty"`
	CustomID           string             `bson:"custom_id,omitempty" json:"custom_id,omitempty"`
	DateMS             int64              `bson:"date_ms,omitempty" json:"date_ms,omitempty"`
	Price              float64            `bson:"price,omitempty" json:"price,omitempty"`
	PriceUSD           float64            `bson:"price_usd,omitempty" json:"price_usd,omitempty"`
	CurrencyCode       string             `bson:"currency_code,omitempty" json:"currency_code,omitempty"`
	CountryCode        string             `bson:"country_code,omitempty" json:"country_code,omitempty"`
	Store              string             `bson:"store,omitempty" json:"store,omitempty"`
	Estimated          int                `bson:"estimated,omitempty" json:"estimated,omitempty"`
	Environment        string             `bson:"environment,omitempty" json:"environment,omitempty"`
	Source             string             `bson:"source,omitempty" json:"source,omitempty"`
}

type Config struct {
	Value int `bson:"value" json:"value"`
}
