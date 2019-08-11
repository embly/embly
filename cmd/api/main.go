package main

import (
	"embly/pkg/build"
	"embly/pkg/routing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetReportCaller(true)
	roc, err := routing.NewContext()
	if err != nil {
		logrus.Fatal(err)
	}

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"msg": "OK"})
	})
	build.ApplyRoutes(roc, router.Group("/api/"))

	router.Run(":3000")
}
