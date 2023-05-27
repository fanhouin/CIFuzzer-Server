package main

import (
	"log"
	"os"
	"web-server/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	apiGroup := r.Group("/api")
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return
	}
	routes.RunFuzzer(apiGroup, workDir)

	r.Run(":8080")
}
