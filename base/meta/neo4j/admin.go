package neo4j

import "github.com/neo4j/neo4j-go-driver/v5/neo4j"

func (neo *Neo4j) SuppressAccount(id string) error {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	param := map[string]interface{}{
		"id":            id,
		"is_suppressed": true,
	}

	_, err := neo4j.ExecuteQuery(ctx, neo.driver,
		`
		MATCH (n:USER {id: $id}) 
		SET n.is_suppressed = $is_suppressed
		`,
		param,
		neo4j.EagerResultTransformer)

	return err
}

func (neo *Neo4j) UnSuppressAccount(id string) error {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	param := map[string]interface{}{
		"id":            id,
		"is_suppressed": false,
	}

	_, err := neo4j.ExecuteQuery(ctx, neo.driver,
		`
		MATCH (n:USER {id: $id}) 
		SET n.is_suppressed = $is_suppressed
		`,
		param,
		neo4j.EagerResultTransformer)

	return err
}
