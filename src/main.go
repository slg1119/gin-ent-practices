package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/someJSON", func(c *gin.Context) {
		names := []string{"test", "test1", "test2", "test3"}
		c.SecureJSON(http.StatusOK, names)
	})
	router.Run(":8080")
}
