package main

import (
	"context"
	"gazdi-telegram-bot-processor/api"
	"gazdi-telegram-bot-processor/providers"
	"gazdi-telegram-bot-processor/services"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	elasticProvider, err := providers.NewElasticSearchProvider("http://localhost:9200")
	if err != nil {
		logrus.Errorln(err.Error(), "haha")
		return
	}

	redisProvider := providers.NewRedisProvider("localhost:6379")

	service, _ := services.NewFaqService(*elasticProvider, *redisProvider)

	api, err := api.NewBotApi("6453546139:AAHTQrR9gSpmwgRHf9ZOjDyIXl1PJ1CNAak", service)
	if err != nil {
		return
	}
	ctx := context.Background()

	go api.ListenUpdates(ctx)
	for {

	}
}
