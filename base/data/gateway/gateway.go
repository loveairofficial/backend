package gateway

import (
	"loveair/base/data"
	"loveair/base/data/mongo"
)

// Incharge of:
// Creating a DB type
// Knows all available DB
// When called by the main function with the DB type needed and connection string
// It initializes the DB and returns the memory address of the DB initialization
// of type the "DatabaseHandler". The interface that the DB method implements.

type DBTYPE string

const (
	MONGODB DBTYPE = "mongodb"
)

func DBConnect(options DBTYPE, Config map[string]string) (data.Interface, error) {
	switch options {
	case MONGODB:
		return mongo.InitMongoDBInstance(Config)
	}
	return nil, nil
}
