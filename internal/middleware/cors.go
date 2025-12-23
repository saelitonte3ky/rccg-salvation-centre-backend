package middleware

import "github.com/gin-gonic/gin"

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		origin := c.Request.Header.Get("Origin")

		allowedOrigins := map[string]struct{}{
			"http://localhost:3000":                    {},
			"https://rccg-salvation-centre.vercel.app": {},
			"https://www.rccgsalvationcentre.org":      {},
		}

		if _, ok := allowedOrigins[origin]; ok {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set(
			"Access-Control-Allow-Headers",
			"Origin, Content-Type, Authorization",
		)
		c.Writer.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, PATCH, DELETE, OPTIONS",
		)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
