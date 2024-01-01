package main

import (
	"context"
	"gazdi-telegram-bot-processor/api"
	"gazdi-telegram-bot-processor/providers"
	"gazdi-telegram-bot-processor/services"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	ctx := context.Background()
	logrus.SetLevel(logrus.DebugLevel)

	elasticProvider, err := providers.NewQuestionsProvider("postgres://qazdi:admin@localhost:5432/questions")
	if err != nil {
		logrus.Errorln(err.Error(), "haha")
		return
	}

	redisProvider := providers.NewRedisProvider("localhost:6379")
	faqService, _ := services.NewFaqService(*elasticProvider, *redisProvider)
	usersService := services.NewUserService(*redisProvider)

	api, err := api.NewBotApi("6453546139:AAHTQrR9gSpmwgRHf9ZOjDyIXl1PJ1CNAak", faqService, &usersService)
	if err != nil {
		return
	}

	go api.ListenUpdates(ctx)
	for {
		time.Sleep(time.Millisecond)
	}
}
