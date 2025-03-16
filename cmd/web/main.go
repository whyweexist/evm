package main

import (
	"github.com/whyweexist/evm/common"
	"github.com/whyweexist/evm/config"
	"github.com/whyweexist/evm/docs"
	"github.com/whyweexist/evm/internal"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/spf13/viper"
	"log"
)



func main() {
	// set config file
	viper.SetConfigFile(".env")
	// LOAD APP ENV
	config.LoadEnv()
	// INIT LANDLORD/MAIN DATABASE FOR DATA STORE
	config.Instance.InitPostgresConn()
	// INIT REDIS CONNECTION FOR DATA CACHE
	config.Instance.InitRedisConn()
	// INIT MAILER CONNECTION FOR EMAIL NOTIFICATION
	config.Instance.InitMailerConn()
	// INIT GIN ENGINE
	config.Instance.InitGinEngine()
	// INIT SENTRY
	if !config.Instance.AppDebug {
		config.Instance.InitSentryConn()
		config.Engine.Use(sentrygin.New(sentrygin.Options{}))
	}
	// INIT SWAGGER DOCS FEW CONFIG
	docs.SwaggerInfo.BasePath = config.Engine.BasePath()
	docs.SwaggerInfo.Host = config.Instance.AppURL
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	// INIT GOOGLE FORM SERVICE
	config.Instance.InitGoogleFormConn()

	// PRINT RUNNING LOG
	log.Printf("Run %s(%s)",
		config.Instance.AppName,
		common.Version)

	// BOOTSTRAP APP
	internal.RunApp(
		internal.WithEngine(config.Engine),
		internal.WithPostgreDatabase(config.Postgre),
		internal.WithRedisCache(config.Redis),
		internal.WithMailer(config.Mailer),
		internal.WithGoogleFormService(config.GoogleForm))

	// RUN SERVER
	log.Fatalln(config.Engine.Run(config.Instance.AppURL))
}
