package admin

import (
	"loveair/api/admin/middleware"
	// mw "loveair/api/middleware"
	"loveair/core/rest"
	"loveair/core/websocket/gorilla"
	"loveair/log"

	"github.com/gorilla/mux"
)

//! Solution for websocket & RBAC - use http for all request so that RBAC can work then when a user request transaction check if they have access then while thy on that page upgrade the connection and use websocket.

func Route(ar *mux.Router, rest *rest.Rest, websocket *gorilla.Socket, secret string, serviceLogger log.SLoger) {
	ar.Methods("POST").Path("/login").HandlerFunc(rest.AdminLogin)

	// ~ Users
	users := ar.PathPrefix("/users").Subrouter()
	//Query
	usersQuery := users.PathPrefix("/query").Subrouter()
	usersQuery.Use(middleware.RBAC(secret, "users", "query", serviceLogger))
	usersQuery.Methods("GET", "OPTIONS").Path("/").HandlerFunc(rest.GetUsers)
	//Mutate
	usersMutate := users.PathPrefix("/mutate").Subrouter()
	usersMutate.Use(middleware.RBAC(secret, "users", "mutate", serviceLogger))
	usersMutate.Methods("PUT", "OPTIONS").Path("/suppress").HandlerFunc(rest.SuppressAccount)
	usersMutate.Methods("PUT", "OPTIONS").Path("/unsuppress").HandlerFunc(rest.UnSuppressAccount)

	// ~ Admins
	admins := ar.PathPrefix("/admins").Subrouter()
	//Query
	adminsQuery := admins.PathPrefix("/query").Subrouter()
	adminsQuery.Use(middleware.RBAC(secret, "admins", "query", serviceLogger))
	adminsQuery.Methods("GET", "OPTIONS").Path("/").HandlerFunc(rest.GetAdmins)

	//Mutate
	adminsMutate := admins.PathPrefix("/mutate").Subrouter()
	adminsMutate.Use(middleware.RBAC(secret, "admins", "mutate", serviceLogger))
	adminsMutate.Methods("POST", "OPTIONS").Path("/add").HandlerFunc(rest.AddAdmin)
	// adminsMutate.Methods("PUT", "OPTIONS").Path("/update").HandlerFunc(rest.UpdateAdmin)
	// rolesMutate.Methods("POST").Path("/delete").HandlerFunc(rest.DeleteRole)

	// ~ Roles
	roles := ar.PathPrefix("/roles").Subrouter()
	//Query
	rolesQuery := roles.PathPrefix("/query").Subrouter()
	rolesQuery.Use(middleware.RBAC(secret, "roles", "query", serviceLogger))
	rolesQuery.Methods("GET", "OPTIONS").Path("/").HandlerFunc(rest.GetRoles)

	//Mutate
	rolesMutate := roles.PathPrefix("/mutate").Subrouter()
	rolesMutate.Use(middleware.RBAC(secret, "roles", "mutate", serviceLogger))
	// rolesMutate.Methods("POST", "OPTIONS").Path("/add").HandlerFunc(rest.AddRole)
	// rolesMutate.Methods("POST", "OPTIONS").Path("/delete").HandlerFunc(rest.DeleteRole)
}

/**

- Scope: Dashboard
- Permissions: Query (View) & Mutations (Edit).

- Scope: Customers (KYC)
- Permissions: Query (View) & Mutations (Edit).

- Scope: Transactions
- Permissions: Query (View) & Mutations (Edit).

- Scope: Currencies
- Permissions: Query (View) & Mutations (Edit).

- Scope: Admins
- Permissions: Query (View) & Mutations (Edit).

- Scope: Roles & Permissions
- Permissions: Query (View) & Mutations (Edit).

- Scope: Referral
- Permissions: Query (View) & Mutations (Edit).

- Scope: Settings
- Permissions: - Permissions: Query (View) & Mutations (Edit).
**/
