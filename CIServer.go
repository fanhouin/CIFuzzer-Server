package main

import (
	"web-server/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	apiGroup := r.Group("/api")
	routes.RunFuzzer(apiGroup)

	r.Run(":8080")
}
