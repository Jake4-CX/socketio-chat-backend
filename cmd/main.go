package main

import (
	"github.com/Jake4-CX/socketio-chat-backend/cmd/initializers"
	globalPkgInitalizer "github.com/Jake4-CX/socketio-chat-backend/pkg/initializers"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func main() {
	globalPkgInitalizer.LoadEnvVariables()
	initializers.InitializeWebsocket()

	router := gin.Default()

	router.Use(GinMiddleware(("*")))

	// Socket.IO
	io := initializers.SocketIO
	router.GET("/socket.io/", gin.WrapH(io.HttpHandler()))

	if os.Getenv("USE_TLS") == "" || os.Getenv("USE_TLS") != "true" {
		log.Fatal(router.Run("0.0.0.0:" + os.Getenv("REST_PORT")))
	} else {
		log.Fatal(autotls.Run(router, "api.load-test.jack.lat"))
	}
}

func GinMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Upgrade, Accept-Encoding, Accept-Language, X-Requested-With")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Request.Header.Del("Origin")

		c.Next()
	}
}
