package main

import (
	"github.com/idena-network/idena-translation/config"
	"github.com/idena-network/idena-translation/core"
	"github.com/idena-network/idena-translation/core/words_mapper"
	"github.com/idena-network/idena-translation/db"
	"github.com/idena-network/idena-translation/db/postgres"
	"github.com/idena-network/idena-translation/node"
	"github.com/idena-network/idena-translation/server"
	log "github.com/inconshreveable/log15"
	"os"
	"runtime"
)

func initLogger(verbosity int) {
	var handler log.Handler
	logLvl := log.Lvl(verbosity)
	if runtime.GOOS == "windows" {
		handler = log.LvlFilterHandler(logLvl, log.StreamHandler(os.Stdout, log.LogfmtFormat()))
	} else {
		handler = log.LvlFilterHandler(logLvl, log.StreamHandler(os.Stderr, log.TerminalFormat()))
	}
	log.Root().SetHandler(handler)
}

func startServer(appConfig *config.Config) {
	initLogger(appConfig.Verbosity)
	log.Info("App is starting...")
	server.NewServer(appConfig.Server.Port, initAuth(appConfig)).Start(appConfig.Swagger)
}

func initAuth(appConfig *config.Config) core.Engine {
	return core.NewEngine(
		initDbAccessor(appConfig),
		initNodeClient(appConfig),
		appConfig.ItemsLimit,
		appConfig.ConfirmedRate,
		words_mapper.NewWordsMapper(appConfig.WordsUrl),
	)
}

func initDbAccessor(appConfig *config.Config) db.Accessor {
	return postgres.NewAccessor(appConfig.Postgres.ConnStr, appConfig.Postgres.ScriptsDir)
}

func initNodeClient(appConfig *config.Config) node.Client {
	return node.NewClient(appConfig.Api.Url)
}
