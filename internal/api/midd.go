package api

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func MakeToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return
		}
		token := parts[1]
		c.Set("DceToken", token)
		c.Next()
	}
}
