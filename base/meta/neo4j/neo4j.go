package neo4j

import (
	"context"
	"loveair/base/meta"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4j struct {
	driver neo4j.DriverWithContext
}

func NewNeo4jConnection(dbConfig map[string]string) (meta.Interface, error) {
	driver, err := neo4j.NewDriverWithContext(dbConfig["url"], neo4j.BasicAuth(dbConfig["user"], dbConfig["pass"], ""))
	if err != nil {
		panic(err)
	}

	return &Neo4j{
		driver,
	}, err
}

func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// interfaceToSliceConverter converts []interface to []string or []int with the help of golang generic type system.
// func interfaceToSliceConverter[gen string | int](d []interface{}) []gen {
// 	data := make([]gen, len(d))
// 	for i := range d {
// 		// !try to understand this!
// 		// fmt.Println(i, 88)
// 		data[i] = d[i].(gen)
// 	}
// 	return data
// }
