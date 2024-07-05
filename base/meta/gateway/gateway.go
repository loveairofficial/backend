package gateway

import (
	"loveair/base/meta"
	"loveair/base/meta/neo4j"
)

// Incharge of:
// Creating a DB type
// Knows all available DB
// When called by the main function with the DB type needed and connection string
// It initializes the DB and returns the memory address of the DB initialization
// of type the "DatabaseHandler". The interface that the DB method implements.

type MBTYPE string

const (
	NEO4J MBTYPE = "neo4j"
)

func ConnectDB(options MBTYPE, mbConfig map[string]string) (meta.Interface, error) {
	switch options {
	case NEO4J:
		return neo4j.NewNeo4jConnection(mbConfig)
	}
	return nil, nil
}
