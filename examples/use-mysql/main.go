package main

import (
	"net/http"

	"github.com/3vilive/fly"
	"github.com/3vilive/fly/pkg/log"
	"github.com/3vilive/fly/pkg/storage"
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

			e.GET("/data", func(c *gin.Context) {
				db, err := storage.GetDatabase("example")
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"msg": err.Error(),
					})
					return
				}

				type Stat struct {
					NumberOfUser int64 `gorm:"column:number_of_user" json:"number_of_user"`
					MaxUserAge   int64 `gorm:"column:max_user_age" json:"max_user_age"`
				}
				var stat Stat
				if err := db.Raw("select count(*) number_of_user, max(age) max_user_age from users").Scan(&stat).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"msg": err.Error(),
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"data": stat,
				})
			})
		},
		func(s *http.Server) {

		},
	)
	if err != nil {
		panic(err)
	}
}
