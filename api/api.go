package api

import (
	"context"
	"gazdi-telegram-bot-processor/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
	"time"
)

type BotApi struct {
	entrypoint *tgbotapi.BotAPI
	faqService *services.FagService
}

func NewBotApi(token string, faqService *services.FagService) (*BotApi, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &BotApi{entrypoint: api, faqService: faqService}, nil
}

func (a *BotApi) handleStartCommand(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	ms := tgbotapi.NewMessage(chatID, "Оберіть команду зі списку: ")
	ms.ReplyMarkup = GetReplyMenuMarkup()
	_, err := a.entrypoint.Send(ms)

	if err != nil {
		logrus.Errorln(err.Error())
	}
}

func (a *BotApi) handleFaq(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	innerMenu := tgbotapi.NewInlineKeyboardMarkup()
	questions, err := a.faqService.GetQuestionsList(context.TODO())

	for _, question := range questions {
		innerMenu.InlineKeyboard = append(innerMenu.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(question, question)))
	}

	if err != nil {
		logrus.Errorln(err.Error())
	}

	ms := tgbotapi.NewMessage(chatID, "Список питань: ")
	ms.ReplyMarkup = innerMenu
	_, err = a.entrypoint.Send(ms)

	if err != nil {
		logrus.Errorln(err.Error())
	}
}

func (a *BotApi) handleFaqQuestion(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	resp := tgbotapi.CallbackConfig{
		CallbackQueryID: query.ID,
		Text:            "Запит оброблюється",
		ShowAlert:       false,
	}
	a.entrypoint.AnswerCallbackQuery(resp)

	question := query.Data
	logrus.Infoln(question)
	answer, err := a.faqService.GetAnswerByQuestion(context.TODO(), question)

	if err != nil {
		logrus.Errorln(err.Error())
		return
	}

	ms := tgbotapi.NewMessage(chatID, "Питання: "+question+"\nВідповідь: "+*answer)
	_, err = a.entrypoint.Send(ms)
	if err != nil {
		logrus.Errorln(err.Error())
		return
	}
}

func GetReplyMenuMarkup() tgbotapi.ReplyKeyboardMarkup {
	btn1 := tgbotapi.NewKeyboardButton("Найчастіші питання")
	return tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(btn1))
}

func (a *BotApi) ListenUpdates(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := a.entrypoint.GetUpdatesChan(u)

	if err != nil {
		return
	}
	for {
		select {
		case update := <-updates:
			logrus.Debug("Receive new message", update)
			if update.Message != nil && update.Message.IsCommand() && update.Message.Command() == "start" {
				go a.handleStartCommand(update.Message)
				continue
			}
			if update.Message != nil && update.Message.Text == "Найчастіші питання" {
				go a.handleFaq(update.Message)
				continue
			}

			if update.CallbackQuery != nil && update.CallbackQuery.Message.Text == "Список питань:" {
				logrus.Debug("Try answer...")
				go a.handleFaqQuestion(update.CallbackQuery)
				continue
			}
		default:
			time.Sleep(time.Millisecond * 20)
		}
	}
}
