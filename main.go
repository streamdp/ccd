package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/db"
	"github.com/streamdp/ccd/repos"
	"github.com/streamdp/ccd/router"
)

func main() {
	l := log.New(gin.DefaultWriter, "[CCD] ", log.LstdFlags)

	appCfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}
	gin.SetMode(appCfg.RunMode)

	l.Printf("Run mode:\n")
	l.Printf("\tVersion=%v\n", appCfg.Version)
	l.Printf("\tRun mode=%v\n", appCfg.RunMode)
	l.Printf("\tData provider=%v\n", appCfg.DataProvider)
	l.Printf("\tSession store=%v\n", appCfg.SessionStore)

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

	taskRepo, ok := d.(repos.TaskStore)
	if !ok {
		l.Fatalln("task repo type assertion error")
	}
	s, err := newSessionStore(taskRepo, appCfg)
	if err != nil {
		l.Fatal(err)
	}
	defer func() {
		if errClose := s.Close(); errClose != nil {
			l.Printf("failed to close session store: %v", errClose)
		}
	}()

	symbolRepo, ok := d.(repos.SymbolsStore)
	if !ok {
		l.Fatalln("symbol repo type assertion error")
	}
	sr := repos.NewSymbolRepository(symbolRepo)
	if err = sr.Load(); err != nil {
		l.Fatalln(err)
	}

	r, err := initRestClient(appCfg)
	if err != nil {
		l.Fatalln(err)
	}

	ctx := context.Background()

	w, err := initWsClient(ctx, database, l, appCfg)
	if err != nil {
		l.Fatalln(err)
	}

	p := clients.NewPuller(r, l, s, database.DataPipe())
	if err = p.RestoreLastSession(); err != nil {
		l.Printf("error restoring last session: %v", err)
	}

	e := gin.Default()
	if err = router.InitRouter(ctx, e, database, l, sr, r, w, p); err != nil {
		l.Fatalln(err)
	}

	if err = e.Run(fmt.Sprintf(":%d", appCfg.Http.Port())); err != nil {
		l.Fatalln(err)
	}

	<-ctx.Done()
}
