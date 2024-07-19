package mongo

import (
	"log"
	"loveair/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDB) VerifyEmailExist(email string) error {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"_id": 1,
	}

	creds := new(models.User)

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)
	err := collection.FindOne(ctx, bson.M{"email": email}, options.FindOne().SetProjection(projection)).Decode(&creds)

	return err
}

func (m *MongoDB) AddUser(usr *models.User) error {
	ctx, cancel := getContext()
	defer cancel()

	id := primitive.NewObjectID()

	data := primitive.M{
		"_id":                              id,
		"id":                               usr.ID,
		"verification":                     usr.Verification,
		"is_paused":                        usr.IsPaused,
		"is_active":                        usr.IsActive,
		"first_name":                       usr.FirstName,
		"last_name":                        usr.LastName,
		"email":                            usr.Email,
		"password":                         usr.Password,
		"is_onboarded":                     usr.IsOnboarded,
		"stage_ID":                         usr.StageID,
		"joined_at":                        usr.JoinedAt,
		"preference":                       usr.Preference,
		"address":                          usr.Address,
		"rose_count":                       usr.RoseCount,
		"religiom":                         usr.Religion,
		"subscription":                     usr.Subscription,
		"notification":                     usr.Notification,
		"free_trial_count":                 usr.FreeTrialCount,
		"free_trial_count_issue_timestamp": usr.FreeTrialCountIssueTimestamp,
	}

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)
	_, err := collection.InsertOne(ctx, data)
	return err
}

func (m *MongoDB) GetCredential(email string) (*models.User, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"password":                         1,
		"is_onboarded":                     1,
		"email":                            1,
		"id":                               1,
		"first_name":                       1,
		"profile_picture":                  bson.M{"$arrayElemAt": bson.A{"$photos", 0}},
		"is_paused":                        1,
		"phone":                            1,
		"subscription":                     1,
		"verification":                     1,
		"is_active":                        1,
		"deactivated_by":                   1,
		"deactivation_date":                1,
		"free_trial_count":                 1,
		"free_trial_count_issue_timestamp": 1,
	}

	creds := new(models.User)

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	err := collection.FindOne(ctx, bson.M{"email": email},
		options.FindOne().SetProjection(projection)).Decode(&creds)

	return creds, err
}

func (m *MongoDB) AddNewDevice(d *models.Device, email string) error {
	filter := bson.M{"email": email}
	//This update adds the device to the devices array and removes the oldest deveice in the array if the array is full.
	update := bson.M{
		"$push": bson.M{
			"devices": bson.M{
				"$each": []interface{}{*d}, "$slice": -3}}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) GetDevice(email, did string) (*models.Device, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"devices": 1,
	}

	filter := bson.M{
		"email":   email,
		"devices": bson.M{"$elemMatch": bson.M{"device_id": did}},
	}

	creds := new(models.User)

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	err := collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&creds)

	if len(creds.Devices) == 0 {
		return nil, err

	}

	return &creds.Devices[0], nil
}

func (m *MongoDB) GetUserPushNotificationIDs(id string) ([]string, string, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"first_name": 1,
		"devices":    1,
	}

	filter := bson.M{
		"id": id,
	}

	creds := new(models.User)

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	err := collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&creds)

	var pIDs []string

	for _, device := range creds.Devices {
		pIDs = append(pIDs, device.PushTkn)
	}

	return pIDs, creds.FirstName, err
}

func (m *MongoDB) DeleteDevice(email, did string) error {
	filter := bson.M{"email": email}
	update := bson.M{"$pull": bson.M{"devices": bson.M{"device_id": did}}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) UpdatePreference(userID string, pref models.Preference, addr string, vicinity string, utcOffset int) error {
	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{
		"preference": pref,
		"address":    addr,
		"vicinity":   vicinity,
		"utc_offset": utcOffset,
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) GetPreference(id string) (models.Preference, string, string, int, int, string, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"preference":   1,
		"address":      1,
		"vicinity":     1,
		"utc_offset":   1,
		"rose_count":   1,
		"subscription": 1,
		"presence":     1,
	}

	creds := new(models.User)

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	err := collection.FindOne(ctx, bson.M{"id": id},
		options.FindOne().SetProjection(projection)).Decode(&creds)

	return creds.Preference, creds.Address, creds.Vicinity, creds.UTCOffset, creds.RoseCount, creds.Subscription, err
}

func (m *MongoDB) UpdateLocation(userID string, loc models.Location) error {
	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{
		"location":                  loc,
		"address":                   loc.Address,
		"vicinity":                  loc.Vicinity,
		"utc_offset":                0,
		"preference.geo_circle.lat": loc.Lat,
		"preference.geo_circle.lon": loc.Lon,
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) UpdateNotification(userID string, noti models.Notification) error {
	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{
		"notification.email": noti.Email,
		"notification.push":  noti.Push,
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) UpdateProfile(userID string, usr models.User) error {
	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{
		"first_name":             usr.FirstName,
		"last_name":              usr.LastName,
		"gender":                 usr.Gender,
		"dob":                    usr.DOB,
		"relationship_intention": usr.RelationshipIntention,
		"interests":              usr.Interests,
		"location":               usr.Location,
		"religion":               usr.Religion,

		"intro_type":      usr.IntroType,
		"intro_video_uri": usr.IntroVideoUri,
		"intro_audio_uri": usr.IntroAudioUri,
		"photos":          usr.Photos,
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) UpdateAccount(userID string, usr models.User) error {
	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{
		//! commented out because both need email verification.
		// "email":     usr.Email,
		"phone":     usr.Phone,
		"is_paused": usr.IsPaused,
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) DeactivateAccount(userID, by string) error {
	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{
		"is_paused":         true,
		"is_active":         false,
		"deactivated_by":    by,
		"deactivation_date": time.Now().UTC(),
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) ReactivateAccount(id string) error {
	filter := bson.M{"id": id}
	update := bson.M{"$set": bson.M{
		"is_paused":         false,
		"is_active":         true,
		"deactivated_by":    "",
		"deactivation_date": time.Time{},
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

//!reactivate account must reset "deactivation_date"

func (m *MongoDB) UpdatePassword(email, password string) error {
	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{
		"password": password,
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

// HydratePotentialMatches --------------------------------------------
func GetUserByID(users []models.User, userID string) (*models.User, bool) {
	for i, user := range users {
		if user.ID == userID {
			return &users[i], true
		}
	}
	return nil, false
}

func (m *MongoDB) HydratePotentialMatches(ids []string, usrs []models.User) ([]models.User, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"id":                     1,
		"first_name":             1,
		"gender":                 1,
		"dob":                    1,
		"relationship_intention": 1,
		"interests":              1,
		"intro_type":             1,
		"intro_video_uri":        1,
		"intro_audio_uri":        1,
		"photos":                 1,
		"joined_at":              1,
		"location":               1,
	}

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	filter := bson.M{"id": bson.M{"$in": ids}}
	cursor, err := collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}

	var users []models.User

	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}

		if usr, ok := GetUserByID(usrs, user.ID); ok {
			user.Presence = usr.Presence
			user.LastSeen = usr.LastSeen
			user.MutualInterest = usr.MutualInterest
			user.ExclusiveInterest = usr.ExclusiveInterest
		}

		users = append(users, user)
	}
	return users, err
}

func (m *MongoDB) GetPotentialMatch(id string) (models.User, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"id":                     1,
		"first_name":             1,
		"last_name":              1,
		"gender":                 1,
		"dob":                    1,
		"relationship_intention": 1,
		"interests":              1,
		"intro_type":             1,
		"intro_video_uri":        1,
		"intro_audio_uri":        1,
		"photos":                 1,
		"joined_at":              1,
		"location":               1,
		"religion":               1,
		"verification":           1,
		"notification":           1,
		"subscription":           1,
	}

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	var user models.User

	err := collection.FindOne(ctx, bson.M{"id": id},
		options.FindOne().SetProjection(projection)).Decode(&user)

	return user, err
}

func GetMrByID(mrs []models.MeetRequest, userID string) (*models.MeetRequest, bool) {
	for i, mr := range mrs {
		if mr.User.ID == userID {
			return &mrs[i], true
		}
	}
	return nil, false
}

func (m *MongoDB) HydrateMeetRequests(ids []string, mrs []models.MeetRequest) ([]models.MeetRequest, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"id":                     1,
		"first_name":             1,
		"gender":                 1,
		"dob":                    1,
		"relationship_intention": 1,
		"interests":              1,
		"intro_type":             1,
		"intro_video_uri":        1,
		"intro_audio_uri":        1,
		"photos":                 1,
		"joined_at":              1,
		"location":               1,
		"verification":           1,
	}

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	filter := bson.M{"id": bson.M{"$in": ids}}
	cursor, err := collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}

	var meetRequests []models.MeetRequest

	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}

		mr, ok := GetMrByID(mrs, user.ID)
		if ok {
			mr.User = user
			mr.User.LastSeen = mr.LastSeen
			mr.User.Presence = mr.Presence
			mr.User.MutualInterest = mr.MutualInterest
			mr.User.ExclusiveInterest = mr.ExclusiveInterest
		}

		meetRequests = append(meetRequests, *mr)
	}
	return meetRequests, err
}

// Chat
func (m *MongoDB) AddChat(chat *models.Chat) error {
	ctx, cancel := getContext()
	defer cancel()

	_id := primitive.NewObjectID()

	data := primitive.M{
		"_id":        _id,
		"id":         chat.ID,
		"recipients": chat.Recipients,
		"messages":   chat.Messages,
		"matched_at": chat.MatchedAt,
	}

	database := m.client.Database(LADB)
	collection := database.Collection(ChatCLX)
	_, err := collection.InsertOne(ctx, data)

	return err
}

func (m *MongoDB) GetChats(userID string) (*[]models.Chat, error) {
	ctx, cancel := getContext()
	defer cancel()

	database := m.client.Database(LADB)
	collection := database.Collection(ChatCLX)

	projection := bson.M{
		"id":             1,
		"status":         1,
		"non_recipients": 1,
		// return only the document that match the id, incase user choose to leave chat or erase chat.
		// "recipients": bson.M{"$elemMatch": bson.M{"id": userID}},
		"recipients": 1,
		"matched_at": 1,
		// If the array document is less than whats available
		// mongo db returns the entire array document.
		"messages":     bson.M{"$slice": -20},
		"unmatched_at": 1,
	}

	cur, err := collection.Find(ctx, bson.M{"recipients.id": userID}, options.Find().SetProjection(projection))
	if err != nil {
		log.Fatal(err)
	}

	defer cur.Close(ctx)
	chats := []models.Chat{}
	for cur.Next(ctx) {
		var chat models.Chat
		if err = cur.Decode(&chat); err != nil {
			log.Fatal(err)
		}

		chats = append(chats, chat)
	}
	return &chats, err
}

func GetChatByID(chats []models.Chat, user models.User) (models.Chat, bool) {
	for i, chat := range chats {
		if len(chat.Recipients) == 1 {
			if chat.NonRecipient.ID == user.ID {
				chats[i].NonRecipient = user
				return chats[i], true
			}
			// skip this chat you unmatch the user.

		} else {
			for j, recipient := range chat.Recipients {
				if recipient.ID == user.ID {
					// Update the recipient to the provided user
					chats[i].Recipients[j] = user
					return chats[i], true
				}
			}
		}

	}
	return models.Chat{}, false // Return false if no chat is found
}

func (m *MongoDB) HydrateChats(ids []string, chats *[]models.Chat) ([]models.Chat, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"id":                     1,
		"first_name":             1,
		"gender":                 1,
		"dob":                    1,
		"relationship_intention": 1,
		"intro_type":             1,
		"intro_video_uri":        1,
		"intro_audio_uri":        1,
		"interests":              1,
		"photos":                 1,
		"joined_at":              1,
		"vicinity":               1,
		"location":               1,
		"verification":           1,
	}

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	filter := bson.M{"id": bson.M{"$in": ids}}
	cursor, err := collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}

	var newChats []models.Chat

	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}

		chat, ok := GetChatByID(*chats, user)
		if ok {
			newChats = append(newChats, chat)
		}

	}
	return newChats, err
}

func (m *MongoDB) AddMessage(message *models.Message) error {
	id := primitive.NewObjectID()

	data := primitive.M{
		"_id":         id,
		"id":          message.ID,
		"chat_id":     message.ChatID,
		"status":      message.Status,
		"receiver_id": message.RecieverID,
		"sender_id":   message.SenderID,
		"content":     message.Content,
		"timestamp":   message.Timestamp,
		"type":        message.Type,
	}

	filter := bson.M{"id": message.ChatID}
	update := bson.M{"$addToSet": bson.M{"messages": data}}

	err := m.Updater(ChatCLX, filter, update)

	return err
}

func (m *MongoDB) UpdateMessageStatus(chatID string, msgIDs []string) error {
	ctx, cancel := getContext()
	defer cancel()

	filter := bson.M{"id": chatID}
	update := bson.M{
		"$set": bson.M{
			"messages.$[elem].status": "Read",
		},
	}

	arrayFilters := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.id": bson.M{"$in": msgIDs}},
		},
	})

	database := m.client.Database(LADB)
	collection := database.Collection(ChatCLX)

	_, err := collection.UpdateOne(ctx, filter, update, arrayFilters)

	return err
}

func (m *MongoDB) RemoveUserFromChat(chatID, senderID string) error {
	filter := bson.M{"id": chatID}

	update := bson.M{
		"$set": bson.M{
			"status":         "unmatched",
			"non_recipients": models.User{ID: senderID},
			"unmatched_at":   time.Now().UTC(),
		},

		"$pull": bson.M{
			"recipients": bson.M{"id": senderID},
		},
	}

	err := m.Updater(ChatCLX, filter, update)

	return err
}

func (m *MongoDB) MergeCachedSession(msgSlice []models.Message) error {
	ctx, cancel := getContext()
	defer cancel()

	filter := bson.M{"id": msgSlice[0].ChatID}
	update := bson.M{"$push": bson.M{"messages": bson.M{"$each": msgSlice}}}

	database := m.client.Database(LADB)
	collection := database.Collection(ChatCLX)

	_, err := collection.UpdateOne(
		ctx,
		filter,
		update,
	)
	return err
}

// Report
func (m *MongoDB) AddReport(report models.Report) error {
	ctx, cancel := getContext()
	defer cancel()

	_id := primitive.NewObjectID()

	data := primitive.M{
		"_id":          _id,
		"id":           report.ID,
		"type":         report.Type,
		"status":       report.Status,
		"context":      report.Context,
		"sender_id":    report.SenderID,
		"recipient_id": report.RecipientID,
		"timestamp":    report.Timestamp,
	}

	database := m.client.Database(LADB)
	collection := database.Collection(ReportCLX)
	_, err := collection.InsertOne(ctx, data)

	return err
}

// Feedback
func (m *MongoDB) AddFeedback(feedback models.Feedback) error {
	ctx, cancel := getContext()
	defer cancel()

	_id := primitive.NewObjectID()

	data := primitive.M{
		"_id":       _id,
		"id":        feedback.ID,
		"status":    feedback.Status,
		"content":   feedback.Content,
		"sender_id": feedback.SenderID,
		"timestamp": feedback.Timestamp,
	}

	database := m.client.Database(LADB)
	collection := database.Collection(FeedbackCLX)
	_, err := collection.InsertOne(ctx, data)

	return err
}

// Subscription
func (m *MongoDB) UpdateSubscription(userID, status string) error {
	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{
		"subscription": status,
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) AddTransaction(payload models.WebhookPayload) error {
	ctx, cancel := getContext()
	defer cancel()

	_id := primitive.NewObjectID()

	data := primitive.M{
		"_id":                   _id,
		"type":                  payload.Type,
		"expire_date_ms":        payload.ExpireDateMS,
		"auto_renew_product_id": payload.AutoRenewProductID,
		"product_id":            payload.ProductID,
		"transaction_id":        payload.TransactionID,
		"subscriber_id":         payload.SubscriberID,
		"custom_id":             payload.CustomID,
		"date_ms":               time.Unix(0, payload.DateMS*int64(time.Millisecond)),
		"price":                 payload.Price,
		"price_usd":             payload.PriceUSD,
		"currency_code":         payload.CurrencyCode,
		"country_code":          payload.CountryCode,
		"store":                 payload.Store,
		"estimated":             payload.Estimated,
		"environment":           payload.Environment,
		"source":                payload.Source,
		"auto_renew_status":     payload.AutoRenewStatus,
	}

	database := m.client.Database(LADB)
	collection := database.Collection(TransactionCLX)
	_, err := collection.InsertOne(ctx, data)

	return err
}

func (m *MongoDB) UpdateFreeTrialCount(email string, count int, d time.Time) error {
	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{
		"free_trial_count":                 count,
		"free_trial_count_issue_timestamp": d,
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

// Config
func (m *MongoDB) GetLatestStableBuildNumber() (int, error) {
	ctx, cancel := getContext()
	defer cancel()

	// Create an instance of the struct
	config := new(models.Config)

	projection := bson.M{
		"value": 1,
	}

	database := m.client.Database(LADB)
	collection := database.Collection(ConfigCLX)

	err := collection.FindOne(ctx, bson.M{"name": "latest_stable_build_number"},
		options.FindOne().SetProjection(projection)).Decode(&config)

	return config.Value, err
}

func (m *MongoDB) GetTerms() (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	// Create an instance of the struct
	config := new(models.Config)

	projection := bson.M{
		"terms": 1,
	}

	database := m.client.Database(LADB)
	collection := database.Collection(ConfigCLX)

	err := collection.FindOne(ctx, bson.M{"name": "terms"},
		options.FindOne().SetProjection(projection)).Decode(&config)

	return config.Terms, err
}

func (m *MongoDB) GetPrivacyPolicy() (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	// Create an instance of the struct
	config := new(models.Config)

	projection := bson.M{
		"privacy_policy": 1,
	}

	database := m.client.Database(LADB)
	collection := database.Collection(ConfigCLX)

	err := collection.FindOne(ctx, bson.M{"name": "privacy_policy"},
		options.FindOne().SetProjection(projection)).Decode(&config)

	return config.PrivacyPolicy, err
}

func (m *MongoDB) GetHowLoveairWorks() (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	// Create an instance of the struct
	config := new(models.Config)

	projection := bson.M{
		"how_loveair_works": 1,
	}

	database := m.client.Database(LADB)
	collection := database.Collection(ConfigCLX)

	err := collection.FindOne(ctx, bson.M{"name": "how_loveair_works"},
		options.FindOne().SetProjection(projection)).Decode(&config)

	return config.HowLoveairWorks, err
}
