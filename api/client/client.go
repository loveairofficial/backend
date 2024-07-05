package client

import (
	"loveair/core/rest"
	"loveair/core/websocket/gorilla"
	"loveair/log"

	"github.com/gorilla/mux"
)

func Route(client *mux.Router, rest *rest.Rest, secret string, socket *gorilla.Socket, serviceLogger log.SLoger) {
	// client.Methods("POST").Path("/auth").HandlerFunc(rest.Authenticate)
	client.Methods("GET").Path("/verify-email").HandlerFunc(rest.VerifyEmail)
	client.Methods("POST").Path("/sign-up").HandlerFunc(rest.SignUp)
	client.Methods("POST").Path("/sign-in").HandlerFunc(rest.SignIn)
	client.Methods("GET").Path("/sign-out").HandlerFunc(rest.SignOut)
	// client.Methods("GET").Path("/refresh").HandlerFunc(rest.Refresh)
	// client.Methods("GET").Path("/get-stream-tkn").HandlerFunc(rest.GenerateStreamToken)

	//Websocket
	soc := client.PathPrefix("/connect").Subrouter()
	soc.Path("/{id}").HandlerFunc(socket.Connect)

	// Onboarding
	client.Methods("GET").Path("/get-stageID").HandlerFunc(rest.GetStage)

	client.Methods("POST").Path("/stage-one").HandlerFunc(rest.HandleStageOne)
	client.Methods("GET").Path("/get-stage-one").HandlerFunc(rest.GetStageOne)

	client.Methods("POST").Path("/stage-two").HandlerFunc(rest.HandleStageTwo)
	client.Methods("GET").Path("/get-stage-two").HandlerFunc(rest.GetStageTwo)

	client.Methods("POST").Path("/stage-three").HandlerFunc(rest.HandleStageThree)
	client.Methods("GET").Path("/get-stage-three").HandlerFunc(rest.GetStageThree)

	client.Methods("POST").Path("/stage-four").HandlerFunc(rest.HandleStageFour)
	client.Methods("GET").Path("/get-stage-four").HandlerFunc(rest.GetStageFour)

	client.Methods("POST").Path("/stage-five").HandlerFunc(rest.HandleStageFive)
	client.Methods("GET").Path("/get-stage-five").HandlerFunc(rest.GetStageFive)

	client.Methods("POST").Path("/stage-six").HandlerFunc(rest.HandleStageSix)
	client.Methods("GET").Path("/get-stage-six").HandlerFunc(rest.GetStageSix)

	client.Methods("GET").Path("/stage-completion").HandlerFunc(rest.HandleStageCompletion)

	// Uploads & signature
	sn := client.PathPrefix("/signature").Subrouter()
	// sn.Use(middleware.Authorization(secret, rest.DB, serviceLogger))
	sn.Methods("GET", "OPTIONS").Path("/get").HandlerFunc(rest.GetSignature)

	// ~ Preference
	preference := client.PathPrefix("/preference").Subrouter()
	// preference.Use(middleware.Authorization(secret, rest.DB, serviceLogger))

	//Query
	preferenceQuery := preference.PathPrefix("/query").Subrouter()
	preferenceQuery.Methods("GET").Path("/").HandlerFunc(rest.GetPreference)

	//Mutate
	preferenceMutate := preference.PathPrefix("/mutate").Subrouter()
	preferenceMutate.Methods("PUT").Path("/").HandlerFunc(rest.UpdatePreference)
	// preferenceMutate.Methods("PUT").Path("/address").HandlerFunc(rest.UpdateAddress)

	// ~ Potential Matches
	potentialMatches := client.PathPrefix("/potential-matches").Subrouter()
	// preference.Use(middleware.Authorization(secret, rest.DB, serviceLogger))

	//Query
	potentialMatchesQuery := potentialMatches.PathPrefix("/query").Subrouter()
	potentialMatchesQuery.Methods("GET").Path("/").HandlerFunc(rest.GetPotentialMatches)

	// ~ Match Call
	// MatcheCall := client.PathPrefix("/match-call").Subrouter()
	// preference.Use(middleware.Authorization(secret, rest.DB, serviceLogger))

	//Mutate
	// MatcheCallMutate := MatcheCall.PathPrefix("/mutate").Subrouter()
	// MatcheCallMutate.Methods("POST").Path("/init").HandlerFunc(rest.InitMatcheCall)

	// ~ Potential Matches
	meetRequests := client.PathPrefix("/meet-requests").Subrouter()
	// preference.Use(middleware.Authorization(secret, rest.DB, serviceLogger))

	//Query
	meetRequestsQuery := meetRequests.PathPrefix("/query").Subrouter()
	meetRequestsQuery.Methods("GET").Path("/").HandlerFunc(rest.GetMeetRequests)

	// ~ Chats
	chats := client.PathPrefix("/chats").Subrouter()
	// preference.Use(middleware.Authorization(secret, rest.DB, serviceLogger))

	//Query
	chatsQuery := chats.PathPrefix("/query").Subrouter()
	chatsQuery.Methods("GET").Path("/").HandlerFunc(rest.GetChats)

	// ~ Profile
	profile := client.PathPrefix("/profile").Subrouter()
	// preference.Use(middleware.Authorization(secret, rest.DB, serviceLogger))

	//Query
	profileQuery := profile.PathPrefix("/query").Subrouter()
	profileQuery.Methods("GET").Path("/").HandlerFunc(rest.GetProfile)
	profileQuery.Methods("GET").Path("/get-account").HandlerFunc(rest.GetAccount)

	//Mutate
	profileMutate := profile.PathPrefix("/mutate").Subrouter()
	profileMutate.Methods("PUT").Path("/").HandlerFunc(rest.UpdateLocation)
	profileMutate.Methods("PUT").Path("/updateProfile").HandlerFunc(rest.UpdateProfile)
	profileMutate.Methods("PUT").Path("/update-account").HandlerFunc(rest.UpdateAccount)
}
