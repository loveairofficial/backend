package client

import (
	"loveair/api/client/middleware"
	"loveair/core/rest"
	"loveair/core/websocket/gorilla"
	"loveair/log"

	"github.com/gorilla/mux"
)

func Route(client *mux.Router, rest *rest.Rest, secret string, socket *gorilla.Socket, serviceLogger log.SLoger) {
	// client.Methods("POST").Path("/auth").HandlerFunc(rest.Authenticate)
	client.Methods("GET").Path("/verify-email").HandlerFunc(rest.VerifyEmail)
	client.Methods("GET").Path("/verify-email-verification-pin").HandlerFunc(rest.VerifyEmailVerificationPin)
	client.Methods("GET").Path("/send-password-reset-pin").HandlerFunc(rest.HandleSendPasswordResetPin)
	client.Methods("GET").Path("/verify-password-reset-pin").HandlerFunc(rest.HandleVerifyPasswordResetPin)
	client.Methods("PUT").Path("/reset-password").HandlerFunc(rest.HandlePasswordReset)
	client.Methods("POST").Path("/sign-up").HandlerFunc(rest.SignUp)
	client.Methods("POST").Path("/sign-in").HandlerFunc(rest.SignIn)
	client.Methods("POST").Path("/reactivate-account").HandlerFunc(rest.ReactivateAccount)
	client.Methods("GET").Path("/sign-out").HandlerFunc(rest.SignOut)
	client.Methods("PUT").Path("/refresh").HandlerFunc(rest.Refresh)

	//~ Websocket
	soc := client.PathPrefix("/connect").Subrouter()
	soc.Use(middleware.Authorization(secret, serviceLogger))
	soc.Path("/{id}").HandlerFunc(socket.Connect)

	// ~ Onboarding
	onboarding := client.PathPrefix("/onboarding").Subrouter()
	onboarding.Use(middleware.Authorization(secret, serviceLogger))

	//Query
	onboardingQuery := onboarding.PathPrefix("/query").Subrouter()
	onboardingQuery.Methods("GET").Path("/get-stageID").HandlerFunc(rest.GetStage)
	onboardingQuery.Methods("GET").Path("/get-stage-one").HandlerFunc(rest.GetStageOne)
	onboardingQuery.Methods("GET").Path("/get-stage-two").HandlerFunc(rest.GetStageTwo)
	onboardingQuery.Methods("GET").Path("/get-stage-three").HandlerFunc(rest.GetStageThree)
	onboardingQuery.Methods("GET").Path("/get-stage-four").HandlerFunc(rest.GetStageFour)
	onboardingQuery.Methods("GET").Path("/get-stage-five").HandlerFunc(rest.GetStageFive)
	onboardingQuery.Methods("GET").Path("/get-stage-six").HandlerFunc(rest.GetStageSix)

	//Mutate
	onboardingMutate := onboarding.PathPrefix("/mutate").Subrouter()
	onboardingMutate.Methods("POST").Path("/stage-one").HandlerFunc(rest.HandleStageOne)
	onboardingMutate.Methods("POST").Path("/stage-two").HandlerFunc(rest.HandleStageTwo)
	onboardingMutate.Methods("POST").Path("/stage-three").HandlerFunc(rest.HandleStageThree)
	onboardingMutate.Methods("POST").Path("/stage-four").HandlerFunc(rest.HandleStageFour)
	onboardingMutate.Methods("POST").Path("/stage-five").HandlerFunc(rest.HandleStageFive)
	onboardingMutate.Methods("POST").Path("/stage-six").HandlerFunc(rest.HandleStageSix)
	onboardingMutate.Methods("POST").Path("/stage-completion").HandlerFunc(rest.HandleStageCompletion)

	//~ Uploads & signature
	sn := client.PathPrefix("/signature").Subrouter()
	sn.Use(middleware.Authorization(secret, serviceLogger))

	//Query
	sn.Methods("GET", "OPTIONS").Path("/get").HandlerFunc(rest.GetSignature)

	//~ Preference
	preference := client.PathPrefix("/preference").Subrouter()
	preference.Use(middleware.Authorization(secret, serviceLogger))

	//Query
	preferenceQuery := preference.PathPrefix("/query").Subrouter()
	preferenceQuery.Methods("GET").Path("/").HandlerFunc(rest.GetPreference)

	//Mutate
	preferenceMutate := preference.PathPrefix("/mutate").Subrouter()
	preferenceMutate.Methods("PUT").Path("/").HandlerFunc(rest.UpdatePreference)
	// preferenceMutate.Methods("PUT").Path("/address").HandlerFunc(rest.UpdateAddress)

	//~ Potential Matches
	potentialMatches := client.PathPrefix("/potential-matches").Subrouter()
	preference.Use(middleware.Authorization(secret, serviceLogger))

	//Query
	potentialMatchesQuery := potentialMatches.PathPrefix("/query").Subrouter()
	potentialMatchesQuery.Methods("GET").Path("/").HandlerFunc(rest.GetPotentialMatches)

	//~ Meet Requests
	meetRequests := client.PathPrefix("/meet-requests").Subrouter()
	meetRequests.Use(middleware.Authorization(secret, serviceLogger))

	//Query
	meetRequestsQuery := meetRequests.PathPrefix("/query").Subrouter()
	meetRequestsQuery.Methods("GET").Path("/").HandlerFunc(rest.GetMeetRequests)
	meetRequestsQuery.Methods("GET").Path("/check-free-trial-availability").HandlerFunc(rest.CheckFreeTrialAvailability)

	//~ Chats
	chats := client.PathPrefix("/chats").Subrouter()
	chats.Use(middleware.Authorization(secret, serviceLogger))

	//Query
	chatsQuery := chats.PathPrefix("/query").Subrouter()
	chatsQuery.Methods("GET").Path("/").HandlerFunc(rest.GetChats)

	//~ Profile
	profile := client.PathPrefix("/profile").Subrouter()
	profile.Use(middleware.Authorization(secret, serviceLogger))

	//Query
	profileQuery := profile.PathPrefix("/query").Subrouter()
	profileQuery.Methods("GET").Path("/").HandlerFunc(rest.GetProfile)
	profileQuery.Methods("GET").Path("/get-account").HandlerFunc(rest.GetAccount)

	//Mutate
	profileMutate := profile.PathPrefix("/mutate").Subrouter()
	profileMutate.Methods("PUT").Path("/").HandlerFunc(rest.UpdateLocation)
	profileMutate.Methods("PUT").Path("/notification").HandlerFunc(rest.UpdateNotification)
	profileMutate.Methods("PUT").Path("/updateProfile").HandlerFunc(rest.UpdateProfile)
	profileMutate.Methods("PUT").Path("/update-account").HandlerFunc(rest.UpdateAccount)
	profileMutate.Methods("PUT").Path("/update-password").HandlerFunc(rest.UpdatePassword)
	profileMutate.Methods("PUT").Path("/deactivate").HandlerFunc(rest.DeactivateAccount)

	//~ Subscription
	subscription := client.PathPrefix("/subscription").Subrouter()

	//Mutate
	subscriptionMutate := subscription.PathPrefix("/mutate").Subrouter()
	subscriptionMutate.Methods("POST").Path("/glassfy-webhook").HandlerFunc(rest.HandleGlassfyWebhook)

	//~ Config
	config := client.PathPrefix("/config").Subrouter()

	//Query
	configQuery := config.PathPrefix("/query").Subrouter()
	configQuery.Methods("GET").Path("/latest-stable-build").HandlerFunc(rest.GetLatestStableBuildNumber)
	configQuery.Methods("GET").Path("/terms").HandlerFunc(rest.GetTerms)

	configQuery.Methods("GET").Path("/privacy-policy").HandlerFunc(rest.GetPrivacyPolicy)
	configQuery.Methods("GET").Path("/how-loveair-works").HandlerFunc(rest.GetHowLoveairWorks)
}
