package mongo

import (
	"loveair/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Users
func (m *MongoDB) GetAdminCredential(email string) (*models.Admin, error) {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"name":            1,
		"profile_picture": 1,
		"password":        1,
		"is_onboarded":    1,
		"role":            1,
	}

	creds := new(models.Admin)

	database := m.client.Database(LADB)
	collection := database.Collection(AdminCLX)

	err := collection.FindOne(ctx, bson.M{"email": email},
		options.FindOne().SetProjection(projection)).Decode(&creds)

	return creds, err
}

func (m *MongoDB) CheckAdminCredential(email string) error {
	ctx, cancel := getContext()
	defer cancel()

	projection := bson.M{
		"_id": 1,
	}

	creds := new(models.Admin)

	database := m.client.Database(LADB)
	collection := database.Collection(AdminCLX)
	err := collection.FindOne(ctx, bson.M{"email": email}, options.FindOne().SetProjection(projection)).Decode(&creds)

	return err
}

func (m *MongoDB) GetUsers(count, offset int64) (*[]models.User, int64, error) {
	ctx, cancel := getContext()
	defer cancel()

	users := make([]models.User, 0)

	database := m.client.Database(LADB)
	collection := database.Collection(UserCLX)

	// Define options for sorting
	opts := options.Find().SetSort(map[string]int{"joined_at": -1}) // Sort by joined field in descending order
	opts.SetLimit(count)
	opts.SetSkip(offset)

	// Count the documents in the collection
	usersCount, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return &users, 0, err
	}

	cur, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return &users, 0, err
	}

	for cur.Next(ctx) {
		user := models.User{}
		if err = cur.Decode(&user); err != nil {
			return &users, 0, err
		}
		users = append(users, user)
	}

	return &users, usersCount, err
}

func (m *MongoDB) SuppressAccount(id string) error {
	filter := bson.M{"id": id}
	update := bson.M{"$set": bson.M{
		"is_suppressed": true,
		"status":        "declined",
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

func (m *MongoDB) UnSuppressAccount(id string) error {
	filter := bson.M{"id": id}
	update := bson.M{"$set": bson.M{
		"is_suppressed": false,
		"status":        "approved",
	}}

	err := m.Updater(UserCLX, filter, update)
	return err
}

// Roles
func (m *MongoDB) GetRoles() (*[]models.Role, error) {
	ctx, cancel := getContext()
	defer cancel()

	roles := make([]models.Role, 0)

	database := m.client.Database(LADB)
	collection := database.Collection(RoleCLX)

	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return &roles, err
	}

	for cur.Next(ctx) {
		role := models.Role{}
		if err = cur.Decode(&role); err != nil {
			return &roles, err
		}
		roles = append(roles, role)
	}

	return &roles, err
}

// Admins
func (m *MongoDB) AddAdmin(a *models.Admin) error {
	ctx, cancel := getContext()
	defer cancel()

	id := primitive.NewObjectID()

	data := primitive.M{
		"_id":             id,
		"is_active":       a.IsActive,
		"name":            a.Name,
		"phone":           a.Phone,
		"email":           a.Email,
		"password":        a.Password,
		"address":         a.Address,
		"profile_picture": a.ProfilePicture,

		"joined":     a.Joined,
		"role":       a.Role,
		"activities": a.Activities,
	}

	database := m.client.Database(LADB)
	collection := database.Collection(AdminCLX)
	_, err := collection.InsertOne(ctx, data)
	return err
}

func (m *MongoDB) GetAdmins() (*[]models.Admin, error) {
	ctx, cancel := getContext()
	defer cancel()

	admins := make([]models.Admin, 0)

	database := m.client.Database(LADB)
	collection := database.Collection(AdminCLX)

	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return &admins, err
	}

	for cur.Next(ctx) {
		admin := models.Admin{}
		if err = cur.Decode(&admin); err != nil {
			return &admins, err
		}
		admins = append(admins, admin)
	}

	return &admins, err
}
