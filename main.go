package main

import (
	"github.com/everstake/nebulas-tg-bot/config"
	"github.com/everstake/nebulas-tg-bot/dao"
	"github.com/everstake/nebulas-tg-bot/services/bot"
	"github.com/everstake/nebulas-tg-bot/services/modules"
	"log"
	"os"
	"os/signal"
)

func main() {
	err := os.Setenv("TZ", "UTC")
	if err != nil {
		log.Fatal("os.Setenv (TZ): %s", err.Error())
	}

	cfg := config.GetConfig()
	d, err := dao.NewDAO(cfg)
	if err != nil {
		log.Fatal("dao.NewDAO: %s", err.Error())
	}

	b := bot.NewBot(d, cfg)

	g := modules.NewGroup(b)
	g.Run()

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill)

	<-interrupt
	g.Stop()

	os.Exit(0)
}
