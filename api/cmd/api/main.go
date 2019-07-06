package main

import (
	"embly/api/pkg/build"
	"embly/api/pkg/config"
	"embly/api/pkg/dbutil"
	"embly/api/pkg/routing"
	"embly/api/pkg/rustcompile"

	"github.com/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetReportCaller(true)
	config.Register(
		"DB_HOST",
		"DB_DATABASE",
		"DB_PASSWORD",
		"DB_PORT",
		"DB_USERNAME",
		"REDIS_HOST",
	)

	r := redis.NewClient(&redis.Options{
		Addr: config.Get("REDIS_HOST"),
	})
	if _, err := r.Ping().Result(); err != nil {
		logrus.Fatal(err)
	}

	dbclient, err := dbutil.Connect()
	if err != nil {
		logrus.Fatal(err)
	}
	rcc, err := rustcompile.NewRustCompileClient("rustcompile:9000")
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "error creating rustcompile client"))
	}
	rc := &routing.Context{
		DB:       dbclient,
		RCClient: rcc,
	}

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"msg": "OK"})
	})
	build.ApplyRoutes(rc, router.Group("/api/"))

	router.Run(":3000")
}
