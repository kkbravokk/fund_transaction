package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"funds_transaction/config"
	"funds_transaction/internal/router"
	"funds_transaction/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	conf := config.LoadConfig("./default.yaml")
	initComponents(conf)

	ServerRun(conf)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-ch
	switch sig {
	case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
		os.Exit(0)
	default:
		logrus.Infof("main receive signal: %v", sig)
	}
}

func initComponents(conf *config.Config) {
	initLog(conf)
	initDB(conf)
}

func initLog(conf *config.Config) {
	logrus.SetLevel(logrus.Level(conf.Log.LogLevel))

	output, err := os.OpenFile(conf.Log.Path, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	logrus.SetOutput(output)

	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func initDB(conf *config.Config) {
	db, err := gorm.Open(sqlite.Open(conf.DB.Name), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	database.DB = db
}

func ServerRun(conf *config.Config) {
	r := gin.New()
	router.InitRouter(r)

	if conf.HTTP.Pprof {
		runtime.SetBlockProfileRate(int(time.Second))
		runtime.SetMutexProfileFraction(1)
		go func() {
			err := http.ListenAndServe(":6060", nil)
			if err != nil {
				logrus.Errorf("start pprof err: %v", err)
			}
		}()

	}

	_ = r.Run(fmt.Sprintf("%s:%s", conf.HTTP.Host, conf.HTTP.Port))
}
