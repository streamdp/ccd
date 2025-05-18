package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/db"
	"github.com/streamdp/ccd/pkg/sessionrepo"
	"github.com/streamdp/ccd/pkg/symbolsrepo"
	ws "github.com/streamdp/ccd/pkg/wsserver"
	"github.com/streamdp/ccd/server"
)

func main() {
	l := log.New(gin.DefaultWriter, "[CCD] ", log.LstdFlags)

	appCfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}
	gin.SetMode(appCfg.RunMode())

	l.Printf("Run mode:\n")
	l.Printf("\tVersion=%v\n", appCfg.Version())
	l.Printf("\tRun mode=%v\n", appCfg.RunMode())
	l.Printf("\tData provider=%v\n", appCfg.DataProvider)
	l.Printf("\tSession store=%v\n", appCfg.SessionStore)
	l.Printf("\tPort=%v\n", appCfg.Http.Port())

	d, err := db.Connect(appCfg)
	if err != nil {
		l.Fatalln(err)
	}
	database, ok := d.(db.Database)
	if !ok {
		l.Fatalln("database type assertion error")
	}
	defer func() {
		if errClose := database.Close(); errClose != nil {
			l.Printf("failed to close database connection: %v", errClose)
		}
	}()
	go db.Serve(database, l)

	sessionStore, ok := d.(sessionrepo.SessionStore)
	if !ok {
		l.Fatalln("task repo type assertion error")
	}
	sessionRepo, err := newSessionRepo(sessionStore, appCfg)
	if err != nil {
		l.Fatal(err)
	}
	defer func() {
		if errClose := sessionRepo.Close(); errClose != nil {
			l.Printf("failed to close session store: %v", errClose)
		}
	}()

	symbolsStore, ok := d.(symbolsrepo.SymbolsStore)
	if !ok {
		l.Fatalln("symbol repo type assertion error")
	}
	symbolRepo := symbolsrepo.New(symbolsStore)
	if err = symbolRepo.Load(); err != nil {
		l.Fatalln(err)
	}

	restClient, err := initRestClient(appCfg)
	if err != nil {
		l.Fatalln(err)
	}

	ctx := context.Background()

	wsServer := ws.NewServer(ctx, restClient, l, database)

	wsClient, err := initWsClient(ctx, database, wsServer, sessionRepo, l, appCfg)
	if err != nil {
		l.Fatalln(err)
	}
	if err = wsClient.RestoreLastSession(ctx); err != nil {
		l.Printf("error restoring last ws session: %v", err)
	}

	restPuller := clients.NewPuller(restClient, l, sessionRepo, database.DataPipe(), wsServer.DataPipe())
	if err = restPuller.RestoreLastSession(ctx); err != nil {
		l.Printf("error restoring last rest session: %v", err)
	}

	srv := server.NewServer(database, symbolRepo, restClient, wsClient, restPuller, l, appCfg, wsServer)
	if err = srv.InitRouter(ctx); err != nil {
		l.Fatalln(err)
	}

	if err = srv.Run(fmt.Sprintf(":%d", appCfg.Http.Port())); err != nil {
		l.Fatalln(err)
	}

	<-ctx.Done()
}
