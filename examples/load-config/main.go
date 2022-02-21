package main

import (
	"net/http"

	"github.com/3vilive/fly"
	"github.com/3vilive/fly/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	err := fly.BootstrapHttpServer(
		func(e *gin.Engine) {
			e.GET("/ping", func(c *gin.Context) {
				log.Info("http-server.addr", zap.String("http-server.addr", viper.GetString("http-server.addr")))
				c.String(http.StatusOK, "pong")
			})

			e.GET("/config", func(c *gin.Context) {
				key := c.Query("key")
				typ := c.Query("type")

				switch typ {
				case "duration":
					dur := viper.GetDuration(key)
					log.Info("get duration by key", zap.String("key", key), zap.Duration("dur", dur))
				}

				c.String(200, "ok")
			})
		},
		func(s *http.Server) {

		},
	)
	if err != nil {
		panic(err)
	}
}
