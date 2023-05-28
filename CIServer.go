package main

import (
	"log"
	"os"
	"web-server/routes"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	apiGroup := r.Group("/api")
	routes.RunFuzzer(apiGroup, workDir)

	return r
}

func main() {
	r := SetupRouter()
	r.Run(":8080")
}
