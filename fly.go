package fly

import (
	"flag"
	"net/http"

	"github.com/3vilive/fly/pkg/config"
	"github.com/3vilive/fly/pkg/db"
	"github.com/3vilive/fly/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	configFile string
)

func init() {
	// default configs
	viper.SetDefault("http-server.addr", ":8080")
	viper.SetDefault("gin.mode", "debug")

	// flags
	flag.StringVar(&configFile, "config", "configs/default.yml", "app config")
}

func initComponenets() error {
	// init config
	if err := config.InitConfig(configFile); err != nil {
		return errors.Wrap(err, "init config error")
	}

	// init logger
	if err := log.InitLog(); err != nil {
		return errors.Wrap(err, "init log error")
	}

	// init db
	if err := db.InitDatabase(); err != nil {
		return errors.Wrap(err, "init db error")
	}

	return nil
}

func deinitComponenets() {
	if err := db.DeinitDatabase(); err != nil {
		log.Error("deinit db error", zap.Error(err))
	} else {
		log.Info("deinit database ok")
	}
}

func Bootstrap(run func() error) error {
	// parse flags
	if !flag.Parsed() {
		flag.Parse()
	}

	defer deinitComponenets()
	if err := initComponenets(); err != nil {
		return errors.Wrap(err, "init components error")
	}

	if err := run(); err != nil {
		return err
	}

	return nil
}

func BootstrapHttpServer(initGinEngine func(*gin.Engine), initHttpServer func(*http.Server)) error {
	return Bootstrap(func() error {
		// init gin engine
		gin.SetMode(viper.GetString("gin.mode"))
		r := gin.Default()
		initGinEngine(r)

		// http server
		httpServer := &http.Server{
			Addr:    viper.GetString("http-server.addr"),
			Handler: r,
		}
		initHttpServer(httpServer)

		return RunHttpServer(httpServer)
	})
}
