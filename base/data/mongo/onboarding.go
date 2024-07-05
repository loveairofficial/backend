package mongo

import (
	"loveair/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDB) GetStage(id string) (int, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"stage_ID": 1,
	}

	creds := new(models.User)

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	err := collection.FindOne(ctx, bson.M{"id": id},
		options.FindOne().SetProjection(projection)).Decode(&creds)

	return creds.StageID, err
}

func (m *MongoDB) SaveStageOne(id int, gender, userID string) error {
	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{
		"stage_ID": id,
		"gender":   gender,
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) GetStageOne(id string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"gender": 1,
	}

	creds := new(models.User)

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	err := collection.FindOne(ctx, bson.M{"id": id},
		options.FindOne().SetProjection(projection)).Decode(&creds)

	return creds.Gender, err
}

func (m *MongoDB) SaveStageTwo(id int, dob time.Time, userID string) error {
	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{
		"stage_ID": id,
		"dob":      dob,
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) GetStageTwo(id string) (time.Time, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"dob": 1,
	}

	creds := new(models.User)

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	err := collection.FindOne(ctx, bson.M{"id": id},
		options.FindOne().SetProjection(projection)).Decode(&creds)

	return creds.DOB, err
}

func (m *MongoDB) SaveStageThree(id int, relationshipIntention, userID string) error {
	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{
		"stage_ID":               id,
		"relationship_intention": relationshipIntention,
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) GetStageThree(id string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"relationship_intention": 1,
	}

	creds := new(models.User)

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	err := collection.FindOne(ctx, bson.M{"id": id},
		options.FindOne().SetProjection(projection)).Decode(&creds)

	return creds.RelationshipIntention, err
}

func (m *MongoDB) SaveStageFour(id int, interests []string, userID string) error {
	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{
		"stage_ID":  id,
		"interests": interests,
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) GetStageFour(id string) ([]string, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"interests": 1,
	}

	creds := new(models.User)

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	err := collection.FindOne(ctx, bson.M{"id": id},
		options.FindOne().SetProjection(projection)).Decode(&creds)

	return creds.Interests, err
}

func (m *MongoDB) SaveStageFive(id int, intro models.Intro, userID string) error {
	filter := bson.M{"id": userID}
	if intro.IntroType == "video" {
		update := bson.M{"$set": bson.M{
			"stage_ID":        id,
			"intro_video_uri": intro.URI,
			"intro_type":      intro.IntroType,
		}}

		err := m.Updater(UserCLX, filter, update)
		return err
	} else {
		update := bson.M{"$set": bson.M{
			"stage_ID":        id,
			"intro_audio_uri": intro.URI,
			"intro_type":      intro.IntroType,
		}}

		err := m.Updater(UserCLX, filter, update)
		return err

	}
}

func (m *MongoDB) GetStageFive(id string) (string, string, string, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"intro_video_uri": 1,
		"intro_audio_uri": 1,
		"intro_type":      1,
	}

	creds := new(models.User)

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	err := collection.FindOne(ctx, bson.M{"id": id},
		options.FindOne().SetProjection(projection)).Decode(&creds)

	return creds.IntroVideoUri, creds.IntroAudioUri, creds.IntroType, err
}

func (m *MongoDB) SaveStageSix(id int, images []models.Photo, userID string) error {
	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{
		"stage_ID": id,
		"photos":   images,
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) GetStageSix(id string) ([]models.Photo, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"photos": 1,
	}

	creds := new(models.User)

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	err := collection.FindOne(ctx, bson.M{"id": id},
		options.FindOne().SetProjection(projection)).Decode(&creds)

	return creds.Photos, err
}

func (m *MongoDB) HandleStageCompletion(userID string) error {
	filter := bson.M{"id": userID}
	update := bson.M{"$set": bson.M{
		"is_onboarded": true,
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) GetUserInfo(id string) (*models.User, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"gender":                 1,
		"dob":                    1,
		"relationship_intention": 1,
		"interests":              1,
		"religion":               1,
	}

	creds := new(models.User)

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	err := collection.FindOne(ctx, bson.M{"id": id},
		options.FindOne().SetProjection(projection)).Decode(&creds)

	return creds, err
}
