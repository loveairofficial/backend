package main

import (
	"fmt"
	"loveair/api"
	"loveair/log"
	"os"
	"runtime"
	"time"

	"loveair/base/cache"
	cbasegw "loveair/base/cache/gateway"
	"loveair/base/data"
	dbasegw "loveair/base/data/gateway"
	"loveair/base/meta"
	mbasegw "loveair/base/meta/gateway"

	"loveair/core/websocket/router"
	hookgw "loveair/log/hook/gateway"
)

var (
	sLogger log.SLoger
	dbaseIf data.Interface
	mbaseIf meta.Interface
	cbaseIf cache.Interface
	sRouter *router.Router
)

func init() {
	// Init Service Log
	sLogger = log.InitServiceLoger("info")
	hk, err := hookgw.ConnectHook(
		hookgw.HOOKTYPE("amazonKinesis"), map[string]string{
			"accesskey": "",
			"secretkey": "",
			"region":    "",
		})

	if err != nil {
		sLogger.Log.Errorln(err)
	} else {
		hook := hk.GetHookOrigin()
		sLogger.Log.Hooks.Add(hook)
	}

	mongo_uri := os.Getenv("MONGO_URI")
	// If SECRET is not set, assign an empty string
	if mongo_uri == "" {
		mongo_uri = "mongodb://127.0.0.1:27017"
	}

	// Init Database
	dbaseIf, err = dbasegw.DBConnect(dbasegw.DBTYPE("mongodb"), map[string]string{
		"url":      mongo_uri,
		"username": "",
		"password": "",
	})
	if err != nil {
		sLogger.Log.Errorln(err)
	}

	neo4j_uri, neo4j_user, neo4j_pass := os.Getenv("NEO4J_URI"), os.Getenv("NEO4J_USER"), os.Getenv("NEO4J_PASS")
	// If SECRET is not set, assign an empty string
	if neo4j_uri == "" {
		neo4j_uri = "bolt://localhost:7687"
	}

	if neo4j_user == "" {
		neo4j_user = "neo4j"
	}

	if neo4j_pass == "" {
		neo4j_pass = "12345678"
	}

	//~ Init Metabase (Neo4J)
	mbaseIf, err = mbasegw.ConnectDB(mbasegw.MBTYPE("neo4j"), map[string]string{
		"url":  neo4j_uri,
		"user": neo4j_user,
		"pass": neo4j_pass,
	})

	if err != nil {
		sLogger.Log.Errorln(err)
	}

	redis_uri, redis_user, redis_pass := os.Getenv("REDIS_URI"), os.Getenv("REDIS_USER"), os.Getenv("REDIS_PASS")
	// If SECRET is not set, assign an empty string
	if redis_uri == "" {
		neo4j_uri = "localhost:6379"
	}

	//~ Init Cachebase (Redis)
	cbaseIf = cbasegw.ConnectCache(cbasegw.CBTYPE("redis"), map[string]string{
		"url":      redis_uri,
		"username": redis_user,
		"password": redis_pass,
	})

	//! turn on when ready.
	// Init Router
	// sRouter, err = router.NewRouter("ws://127.0.0.1:8090/connect/", sLogger)
	// if err != nil {
	// 	sLogger.Log.Errorln(err)
	// 	go sRouter.Daemon()
	// 	go sRouter.KeepAlive()
	// } else {
	// 	// Start Service Router Reader & Writer
	// 	go sRouter.Daemon()
	// 	go sRouter.StartIO()
	// }

}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("rfx panic: ", r)
			time.Sleep(1 * time.Second)
		}
	}()

	// TODO:
	// USE: 512GB Ram = 125,000 Max Goroutines
	// 1GB Ram = 250,000 Max Goroutines
	// 2GB Ram = 500,000 Max Goroutines

	//!make a cron job that runs daily and updates users age or fine a way to use raw time to query age preference.
	//!solve online status like this immediately you focus on a page start showing the user realtime update of that person alone, sonly who is in focus do you get more update about.
	//! when users update their data, remove the data from the cache before making the update.

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			if runtime.NumGoroutine() < 100 {
				sLogger.Log.Infoln(runtime.NumGoroutine())
				continue
			} else {
				sLogger.Log.Warningln(runtime.NumGoroutine())
			}
		}
	}()

	secret := os.Getenv("SECRET")
	// If SECRET is not set, assign an empty string
	if secret == "" {
		secret = "rfxv0.1"
	}

	port := os.Getenv("PORT")
	// If SECRET is not set, assign an empty string
	if port == "" {
		port = "9090"
	}

	//RESTful API start
	sLogger.Log.Panic(api.Start(
		secret,
		port,
		dbaseIf,
		mbaseIf,
		cbaseIf,
		sRouter,
		// emailIf,
		// mediabaseIf,
		sLogger,
	))
}
