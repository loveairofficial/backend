package neo4j

import (
	"fmt"
	"loveair/models"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (neo *Neo4j) AddUser(usr models.User) error {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	param := map[string]interface{}{
		"id":         usr.ID,
		"is_active":  usr.IsActive,
		"is_paused":  usr.IsPaused,
		"first_name": usr.FirstName,
	}

	_, err := neo4j.ExecuteQuery(ctx, neo.driver,
		"CREATE (:USER { id: $id, first_name: $first_name, is_active: $is_active, is_paused: $is_paused })",
		param,
		neo4j.EagerResultTransformer)
	if err != nil {
		return err
	}

	// _, err := session.BeginTransaction(func(transaction neo4j.ExplicitTransaction) (interface{}, error) {
	// 	result, err := transaction.Run(
	// 		"CREATE (:User:"+acc.Country+":"+acc.State+"{uid: $uid, is_active: $is_active, first_name: $first_name, last_name: $last_name, email: $email, email_verified: $email_verified, profile_picture_URL: $profile_picture_URL, phone: $phone, phone_verified: $phone_verified, gender: $gender, languages: $languages, state: $state, country: $country, created_at: $created_at, updated_at: $updated_at, identity: $identity, identity_verified: $identity_verified, about: $about, personality: $personality, dob: $dob, interests: $interests, nationality: $nationality, religion: $religion, career: $career})",
	// 		param)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	return nil, result.Err()
	// })

	return err
}

func (neo *Neo4j) UpdateUserInfo(id string, usr *models.User) error {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	// Calculate age from dob
	// now := time.Now()
	// age := now.Year() - usr.DOB.Year()
	// if now.Month() < usr.DOB.Month() || (now.Month() == usr.DOB.Month() && now.Day() < usr.DOB.Day()) {
	// 	age--
	// }

	// Prepare parameters
	param := map[string]interface{}{
		"id":        id,
		"gender":    usr.Gender,
		"dob":       usr.DOB,
		"rel_int":   usr.RelationshipIntention,
		"religion":  usr.Religion,
		"is_paused": false,
	}

	// Build base update query
	updateQuery := `MATCH (n:USER { id: $id}) 
					SET n.gender = $gender, n.dob = $dob, n.rel_int = $rel_int, n.religion = $religion, n.is_paused = $is_paused
					`

	// Build relationship creation queries for interests
	var createInterests []string
	// if len(usr.Interests) > 0 {
	// 	for idx, interest := range usr.Interests {
	// 		// Use unique index to ensure each relationship creation query has a unique identifier
	// 		createInterests = append(createInterests, fmt.Sprintf(`MATCH (i%d:INTEREST {name: $interest%d}) MERGE (n)-[:INTERESTED_IN]->(i%d:INTEREST {name: $interest%d})`, idx, idx, idx, idx))
	// 		param[fmt.Sprintf("interest%d", idx)] = interest
	// 	}
	// }

	if len(usr.Interests) > 0 {
		for idx, interest := range usr.Interests {
			// Use unique index to ensure each relationship creation query has a unique identifier
			createInterests = append(createInterests, fmt.Sprintf(`
				WITH n
				MATCH (i%d:INTEREST {name: $interest%d})
				MERGE (n)-[:INTERESTED_IN]->(i%d)
			`, idx, idx, idx))
			param[fmt.Sprintf("interest%d", idx)] = interest
		}
	}

	// Combine update and relationship creation queries
	finalQuery := updateQuery + "\n" + strings.Join(createInterests, "\n")

	// Execute the final query inside a transaction for batch processing
	_, err := neo4j.ExecuteQuery(ctx, neo.driver, finalQuery, param, neo4j.EagerResultTransformer)
	if err != nil {
		return err
	}

	return nil
}

func (neo *Neo4j) UpdateUserLocation(id string, lat, lon float64) error {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	param := map[string]interface{}{
		"id":        id,
		"lat":       lat,
		"lon":       lon,
		"is_paused": false,
	}

	_, err := neo4j.ExecuteQuery(ctx, neo.driver,
		"MATCH (n:USER {id: $id}) SET n.is_paused = $is_paused, n.geo_loc = point({latitude: $lat, longitude: $lon})",
		param,
		neo4j.EagerResultTransformer)

	return err
}

func (neo *Neo4j) UpdateProfile(id string, usr models.User) error {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	param := map[string]interface{}{
		"id":       id,
		"lat":      usr.Location.Lat,
		"lon":      usr.Location.Lon,
		"gender":   usr.Gender,
		"dob":      usr.DOB,
		"rel_int":  usr.RelationshipIntention,
		"religion": usr.Religion,
	}

	// Build base update query
	updateQuery := `MATCH (n:USER {id: $id}) 
					SET n.gender = $gender, n.dob = $dob, n.rel_int = $rel_int, n.religion = $religion, n.geo_loc = point({latitude: $lat, longitude: $lon})

					WITH n
					OPTIONAL MATCH (n)-[in:INTERESTED_IN]-(i)
					DELETE in
					`

	// Build relationship creation queries for interests
	var createInterests []string
	// if len(usr.Interests) > 0 {
	// 	for idx, interest := range usr.Interests {
	// 		// Use unique index to ensure each relationship creation query has a unique identifier
	// 		createInterests = append(createInterests, fmt.Sprintf(`MATCH (i%d:INTEREST {name: $interest%d}) MERGE (n)-[:INTERESTED_IN]->(i%d:INTEREST {name: $interest%d})`, idx, idx, idx, idx))
	// 		param[fmt.Sprintf("interest%d", idx)] = interest
	// 	}
	// }

	if len(usr.Interests) > 0 {
		for idx, interest := range usr.Interests {
			// Use unique index to ensure each relationship creation query has a unique identifier
			createInterests = append(createInterests, fmt.Sprintf(`
					WITH n
					MATCH (i%d:INTEREST {name: $interest%d})

					MERGE (n)-[:INTERESTED_IN]->(i%d)
				`, idx, idx, idx))
			param[fmt.Sprintf("interest%d", idx)] = interest
		}
	}

	// Combine update and relationship creation queries
	finalQuery := updateQuery + "\n" + strings.Join(createInterests, "\n")

	// Execute the final query inside a transaction for batch processing
	_, err := neo4j.ExecuteQuery(ctx, neo.driver, finalQuery, param, neo4j.EagerResultTransformer)
	if err != nil {
		return err
	}

	return nil
}

func (neo *Neo4j) UpdateAccount(id string, iPaused bool) error {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	param := map[string]interface{}{
		"id":        id,
		"is_paused": iPaused,
	}

	_, err := neo4j.ExecuteQuery(ctx, neo.driver,
		"MATCH (n:USER {id: $id}) SET n.is_paused = $is_paused",
		param,
		neo4j.EagerResultTransformer)

	return err
}

func ConvertInterfaceToStringSlice(input []interface{}) []string {
	var output []string

	for _, element := range input {
		// Type assertion to ensure the element is a string
		if str, ok := element.(string); ok {
			output = append(output, str)
		} else {
			// If the type assertion fails, handle the case (for example, skip the item or log an error)
			fmt.Println("Error: Non-string element found in input")
		}
	}

	return output
}

func (neo *Neo4j) GetPotentialMatches(id string, pref *models.Preference) ([]models.User, error) {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	fmt.Println(id, pref)

	param := map[string]interface{}{
		"id":            id,
		"rel_int":       pref.RelationshipIntention,
		"latitude":      pref.GeoCircle.Lat,
		"longitude":     pref.GeoCircle.Lon,
		"radius":        pref.GeoCircle.Radius,
		"interested_in": pref.InterestedIn,
		"age_min":       pref.AgeRange.Min,
		"age_max":       pref.AgeRange.Max,
		"religion":      pref.Religion,
		"presence":      pref.Presence,
	}

	// Define the query to retrieve the user
	// MATCH (on:USER{is_active: true, is_paused: false, presence: $presence})
	//Todo: presence, boostfactor, (recommendation based on those with like matches as you)
	// query := `
	// MATCH (n:USER{id: $id})
	// MATCH (on:USER{is_active: true, is_paused: false})
	// WHERE on.id <> n.id
	// AND NOT (n)-[:MATCH|PASS|UNMATCH]->(on)
	// AND point.distance(on.geo_loc, point({latitude: $latitude, longitude: $longitude})) <= $radius
	// AND on.rel_int IN $rel_int
	// AND on.gender IN $interested_in
	// AND on.age >= $age_min AND on.age <= $age_max
	// OPTIONAL MATCH (n)-[:INTERESTED_IN]->(i:INTEREST)<-[:INTERESTED_IN]-(on)
	// WITH on, COLLECT(i.name) AS mutualInterests
	// RETURN on.id AS id, on.presence AS presence, mutualInterests
	// LIMIT 5
	// `

	// query := `
	// 	MATCH (n:USER{id: $id})
	// 	MATCH (on:USER{is_active: true, is_paused: false})
	// 	WHERE on.id <> n.id
	// 	AND NOT (n)-[:MATCH|PASS|UNMATCH|REQUESTED_TO_MEET]-(on)
	// 	AND point.distance(on.geo_loc, point({latitude: $latitude, longitude: $longitude})) <= $radius
	// 	AND on.rel_int IN $rel_int
	// 	AND on.gender IN $interested_in
	// 	AND on.age >= $age_min AND on.age <= $age_max
	// 	OPTIONAL MATCH (n)-[:INTERESTED_IN]->(i:INTEREST)<-[:INTERESTED_IN]-(on)
	// 	WITH n, on, COLLECT(i.name) AS mutualInterests
	// 	OPTIONAL MATCH (on)-[:INTERESTED_IN]->(oi:INTEREST)
	// 	WHERE NOT (n)-[:INTERESTED_IN]->(oi)
	// 	WITH on, mutualInterests, COLLECT(oi.name) AS exclusiveInterests
	// 	RETURN on.id AS id, on.presence AS presence, mutualInterests, exclusiveInterests
	// 	LIMIT 5
	// 	`

	baseQuery := `
    WITH date() AS currentDate
    MATCH (n:USER{id: $id})
    MATCH (on:USER{is_active: true, is_paused: false})
    WHERE on.id <> n.id
    AND NOT (n)-[:MATCH|PASS|UNMATCH|REQUESTED_TO_MEET]-(on)
    AND on.rel_int IN $rel_int
    WITH on, currentDate, duration.inDays(on.dob, currentDate).days / 365.25 AS age
    WHERE age >= $age_min AND age <= $age_max`

	// Initialize conditionals
	conditions := []string{}

	// Add location condition if needed
	if !pref.Global {
		locationCondition := `AND point.distance(on.geo_loc, point({latitude: $latitude, longitude: $longitude})) <= $radius`
		conditions = append(conditions, locationCondition)
	}

	// Add gender condition if needed
	if pref.InterestedIn[0] != "Open to all" {
		genderCondition := `AND on.gender IN $interested_in`
		conditions = append(conditions, genderCondition)
	}

	// Add religion condition if needed
	if pref.Religion[0] != "Open to all" {
		religionCondition := `AND on.religion IN $religion`
		conditions = append(conditions, religionCondition)
	}

	// Add presence condition if needed
	if pref.Presence != "Open to all" {
		presenceCondition := `AND on.presence = $presence`
		conditions = append(conditions, presenceCondition)
	}

	// Join all conditions with spaces
	conditionString := strings.Join(conditions, "\n")

	// Combine base query with conditions
	completeQuery := fmt.Sprintf(`
		%s
		%s
		OPTIONAL MATCH (n)-[:INTERESTED_IN]->(i:INTEREST)<-[:INTERESTED_IN]-(on)
		WITH n, on, COLLECT(i.name) AS mutualInterests
		OPTIONAL MATCH (on)-[:INTERESTED_IN]->(oi:INTEREST)
		WHERE NOT (n)-[:INTERESTED_IN]->(oi)
		WITH on, mutualInterests, COLLECT(oi.name) AS exclusiveInterests
		RETURN on.id AS id, on.presence AS presence, mutualInterests, exclusiveInterests
		LIMIT 5`, baseQuery, conditionString)

	// Execute the query
	result, err := neo4j.ExecuteQuery(
		ctx,
		neo.driver,
		completeQuery,
		param,
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	fmt.Println(len(result.Records), result.Keys)

	var potentialMatches []models.User

	// Loop through results and do something with them
	for _, record := range result.Records {
		var potentialMatch models.User

		idI, _ := record.Get("id")
		potentialMatch.ID, _ = idI.(string)
		miI, _ := record.Get("mutualInterests")
		mi, _ := miI.([]interface{})
		potentialMatch.MutualInterest = ConvertInterfaceToStringSlice(mi)
		prI, _ := record.Get("presence")
		potentialMatch.Presence, _ = prI.(string)
		eiI, _ := record.Get("exclusiveInterests")
		ei, _ := eiI.([]interface{})
		potentialMatch.ExclusiveInterest = ConvertInterfaceToStringSlice(ei)

		potentialMatches = append(potentialMatches, potentialMatch)
	}

	return potentialMatches, err
}

func (neo *Neo4j) UpdateUserPresence(id, status string) error {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	param := map[string]interface{}{
		"id":       id,
		"presence": status,
	}

	_, err := neo4j.ExecuteQuery(ctx, neo.driver,
		"MATCH (n:USER {id: $id}) SET n.presence = $presence",
		param,
		neo4j.EagerResultTransformer)

	return err
}

func (neo *Neo4j) AddRequestedToMeetRelationship(mr *models.MeetRequest) error {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	param := map[string]interface{}{
		"id":         mr.ID,
		"timestamp":  mr.Timestamp,
		"rose":       mr.Rose,
		"compliment": mr.Compliment,
		"sid":        mr.SenderID,
		"rid":        mr.RecipientID,
	}

	_, err := neo4j.ExecuteQuery(ctx, neo.driver,
		`MATCH (n:USER{id: $sid})
		MATCH (on:USER{id: $rid})

		OPTIONAL MATCH (n)-[p:MATCH|UNMATCH|PASS|REQUESTED_TO_MEET]-(on)
		DELETE p 

		MERGE (n)-[:REQUESTED_TO_MEET{id: $id, timestamp: $timestamp, compliment: $compliment, rose: $rose}]->(on)`,
		param,
		neo4j.EagerResultTransformer)

	return err
}

func (neo *Neo4j) AddMatchRelationship(ts time.Time, sid, rid string) error {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	param := map[string]interface{}{
		"timestamp": ts,
		"sid":       sid,
		"rid":       rid,
	}

	_, err := neo4j.ExecuteQuery(ctx, neo.driver,
		`
		MATCH (n:USER {id: $sid})
		MATCH (on:USER {id: $rid})

		OPTIONAL MATCH (n)-[p:MATCH|UNMATCH|PASS|REQUESTED_TO_MEET]-(on)
		DELETE p 

		MERGE (n)-[:MATCH {timestamp: $timestamp}]->(on)
		MERGE (on)-[:MATCH {timestamp: $timestamp}]->(n)`,
		param,
		neo4j.EagerResultTransformer)

	return err
}

func (neo *Neo4j) AddPassRelationship(ts time.Time, sid, rid string) error {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	param := map[string]interface{}{
		"timestamp": ts,
		"sid":       sid,
		"rid":       rid,
	}

	_, err := neo4j.ExecuteQuery(ctx, neo.driver,
		`MATCH (n:USER {id: $sid})
		MATCH (on:USER {id: $rid})

		OPTIONAL MATCH (n)-[m:MATCH|UNMATCH|PASS|REQUESTED_TO_MEET]-(on)
		DELETE m

		MERGE (n)-[:PASS {timestamp: $timestamp}]->(on)
		MERGE (on)-[:PASS {timestamp: $timestamp}]->(n)
		
		`,
		param,
		neo4j.EagerResultTransformer)

	return err
}

func (neo *Neo4j) GetMeetRequests(id string) ([]models.MeetRequest, error) {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	param := map[string]interface{}{
		"id": id,
	}

	query := `
		MATCH (n:USER {id: $id})
		MATCH (on)-[r:REQUESTED_TO_MEET]->(n)
		RETURN  r.id AS mrId, r.timestamp AS timestamp, r.compliment AS compliment, r.rose AS rose, on.id AS userID, on.presence AS presence
		`

	// Execute the query
	result, err := neo4j.ExecuteQuery(
		ctx,
		neo.driver,
		query,
		param,
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	fmt.Println(len(result.Records), result.Keys)

	var meetRequests []models.MeetRequest

	// Loop through results and do something with them
	for _, record := range result.Records {
		var meetRequest models.MeetRequest

		idI, _ := record.Get("mrId")
		meetRequest.ID, _ = idI.(string)

		tsI, _ := record.Get("timestamp")
		meetRequest.Timestamp, _ = tsI.(time.Time)

		complimentI, _ := record.Get("compliment")
		meetRequest.Compliment, _ = complimentI.(string)

		roseI, _ := record.Get("exclusiveInterests")
		meetRequest.Rose, _ = roseI.(bool)

		userIDI, _ := record.Get("userID")
		meetRequest.User.ID, _ = userIDI.(string)

		presenceI, _ := record.Get("presence")
		meetRequest.Presence, _ = presenceI.(string)

		meetRequests = append(meetRequests, meetRequest)
	}

	return meetRequests, err
}

func (neo *Neo4j) AddUnmatchRelationship(ts time.Time, sid, rid string) error {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	param := map[string]interface{}{
		"timestamp": ts,
		"sid":       sid,
		"rid":       rid,
	}

	_, err := neo4j.ExecuteQuery(ctx, neo.driver,
		`MATCH (n:USER {id: $sid})
		MATCH (on:USER {id: $rid})

		OPTIONAL MATCH (n)-[m:MATCH|UNMATCH|PASS|REQUESTED_TO_MEET]-(on)
		DELETE m

		MERGE (n)-[:UNMATCH {timestamp: $timestamp}]->(on)
		MERGE (on)-[:UNMATCH {timestamp: $timestamp}]->(n)
		
		`,
		param,
		neo4j.EagerResultTransformer)

	return err
}

func (neo *Neo4j) UpdateUserBoost(id string, boost int) error {
	ctx, cancel := getContext()
	defer cancel()

	session := neo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	param := map[string]interface{}{
		"id":    id,
		"boost": boost,
	}

	_, err := neo4j.ExecuteQuery(ctx, neo.driver,
		"MATCH (n:USER {id: $id}) SET n.boost = $boost",
		param,
		neo4j.EagerResultTransformer)

	return err
}
