package main

import (
	"fmt"
	"loveair/api"
	"loveair/email"
	"loveair/log"
	"loveair/push"
	"os"
	"runtime"
	"syscall"
	"time"

	"loveair/base/cache"
	cbasegw "loveair/base/cache/gateway"
	"loveair/base/data"
	dbasegw "loveair/base/data/gateway"
	"loveair/base/meta"
	mbasegw "loveair/base/meta/gateway"

	"loveair/core/websocket/router"
	emailgw "loveair/email/gateway"
	hookgw "loveair/log/hook/gateway"
	pushgw "loveair/push/gateway"
)

var (
	sLogger log.SLoger
	dbaseIf data.Interface
	mbaseIf meta.Interface
	cbaseIf cache.Interface
	emailIf email.Interface
	sRouter *router.Router
	pushIf  push.Interface
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

	redis_remote_uri, redis_remote_user, redis_remote_pass := os.Getenv("REDIS_REMOTE_URI"), os.Getenv("REDIS_REMOTE_USER"), os.Getenv("REDIS_REMOTE_PASS")
	redis_local_uri, redis_local_user, redis_local_pass := os.Getenv("REDIS_LOCAL_URI"), os.Getenv("REDIS_LOCAL_USER"), os.Getenv("REDIS_LOCAL_PASS")

	if redis_remote_uri == "" {
		redis_remote_uri = "localhost:6379"
	}

	if redis_local_uri == "" {
		redis_local_uri = "localhost:6379"
	}

	//~ Init Cachebase (Redis)
	cbaseIf = cbasegw.ConnectCache(cbasegw.CBTYPE("redis"), map[string]string{
		"remote_url":      redis_remote_uri,
		"remote_username": redis_remote_user,
		"remote_password": redis_remote_pass,

		"local_url":      redis_local_uri,
		"local_username": redis_local_user,
		"local_password": redis_local_pass,
	})

	email_api_key := os.Getenv("EMAIL_API_KEY")
	if email_api_key == "" {
		email_api_key = "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"
	}

	// Init Email
	emailIf = emailgw.EConnect(emailgw.ETYPE("sendgrid"), map[string]string{
		"API_KEY": email_api_key},
	)

	// Init Push
	pushIf = pushgw.PConnect(pushgw.PTYPE("expo"))

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

	//~ Increase resources limitations
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	// USE: 512GB Ram = 125,000 Max Goroutines
	// 1GB Ram = 250,000 Max Goroutines
	// 2GB Ram = 500,000 Max Goroutines

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
		emailIf,
		pushIf,
		// mediabaseIf,
		sLogger,
	))
}
