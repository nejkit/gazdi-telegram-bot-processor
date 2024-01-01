package api

import (
	"context"
	"errors"
	"gazdi-telegram-bot-processor/models"
	"gazdi-telegram-bot-processor/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type BotApi struct {
	entrypoint  *tgbotapi.BotAPI
	faqService  *services.FagService
	userService *services.UserService
}

func NewBotApi(token string, faqService *services.FagService, userService *services.UserService) (*BotApi, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &BotApi{entrypoint: api, faqService: faqService, userService: userService}, nil
}

func (a *BotApi) handleStartCommand(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	ms := tgbotapi.NewMessage(chatID, "Оберіть команду зі списку: ")
	ms.ReplyMarkup = a.GetMainReplyMenuMarkup(msg.From.ID)
	_, err := a.entrypoint.Send(ms)

	if err != nil {
		logrus.Errorln(err.Error())
	}
}

func (a *BotApi) handlePrivateQuestionsRequest(msg *tgbotapi.Message, updates <-chan tgbotapi.Update) {
	chatID := msg.Chat.ID

	fagList, err := a.faqService.GetFagList(context.TODO())

	if err != nil {
		return
	}

	ms := tgbotapi.NewMessage(chatID, "Список наявних FAQ")
	innerMenu := tgbotapi.NewInlineKeyboardMarkup()

	handlerMap := make(map[string]models.FAQModel)

	for _, faq := range fagList {
		innerMenu.InlineKeyboard = append(innerMenu.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(faq.Question, "view:"+faq.Id)))
		handlerMap[faq.Id] = faq
	}
	ms.ReplyMarkup = innerMenu
	go a.entrypoint.Send(ms)

	ms = tgbotapi.NewMessage(chatID, "Ви знаходитесь у списку FAQ")
	ms.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Повернутися на головну")))

	go a.entrypoint.Send(ms)

	go a.handleFaqAdminMenu(chatID, updates, handlerMap)
}

func (a *BotApi) handleFaqAdminMenu(chatID int64, updates <-chan tgbotapi.Update, handlerMap map[string]models.FAQModel) {
	for {
		select {
		case callback, ok := <-updates:
			if !ok {
				continue
			}
			if callback.Message != nil && callback.Message.Chat.ID == chatID && callback.Message.Text == "Повернутися на головну" {
				ms := tgbotapi.NewMessage(chatID, "Повернення на головну")
				ms.ReplyMarkup = a.GetMainReplyMenuMarkup(callback.Message.From.ID)
				go a.entrypoint.Send(ms)
				return
			}
			if callback.CallbackQuery == nil {
				continue
			}
			if callback.CallbackQuery.Message.Chat.ID != chatID {
				continue
			}

			operationType := strings.Split(callback.CallbackQuery.Data, ":")[0]
			operationID := strings.Split(callback.CallbackQuery.Data, ":")[1]

			info, ok := handlerMap[operationID]
			if !ok {
				callbackResp := tgbotapi.CallbackConfig{CallbackQueryID: callback.CallbackQuery.ID, Text: "FAQ відсутній", ShowAlert: false}
				a.entrypoint.AnswerCallbackQuery(callbackResp)
				continue
			}

			if operationType == "view" {
				callbackResp := tgbotapi.CallbackConfig{CallbackQueryID: callback.CallbackQuery.ID, Text: "Запит оброблюється", ShowAlert: false}
				a.entrypoint.AnswerCallbackQuery(callbackResp)
				ms := tgbotapi.NewMessage(chatID, "Питання: "+info.Question+"\nВідповідь: "+info.Answer)
				ms.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Видалити питання", "delete:"+info.Id)))
				a.entrypoint.Send(ms)
			}
			if operationType == "delete" {
				a.faqService.DeleteFaq(context.TODO(), info)
				callbackResp := tgbotapi.CallbackConfig{CallbackQueryID: callback.CallbackQuery.ID, Text: "FAQ видалено", ShowAlert: false}
				a.entrypoint.AnswerCallbackQuery(callbackResp)
				delete(handlerMap, operationID)
			}
		case <-time.After(time.Minute * 5):
			return

		}
	}
}

func (a *BotApi) handlePublicQuestionsRequest(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	innerMenu := tgbotapi.NewInlineKeyboardMarkup()
	questions, err := a.faqService.GetQuestionsList(context.TODO())

	for _, question := range questions {
		innerMenu.InlineKeyboard = append(innerMenu.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(question.Question, "answer:"+question.Id)))
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

func (a *BotApi) handleCreationQuestionRequest(msg *tgbotapi.Message, updates <-chan tgbotapi.Update) {
	ChatID := msg.Chat.ID

	ms := tgbotapi.NewMessage(ChatID, "Введіть питання")
	ms.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Відміна")))

	_, err := a.entrypoint.Send(ms)
	if err != nil {
		return
	}
	var faqModel models.FAQModel
	err = a.handleEnterCreationFAQ(updates, ChatID, msg.From.ID, &faqModel, func(faqModel *models.FAQModel, text string) {
		faqModel.Question = text
	})
	if err != nil {
		return
	}

	check, err := a.faqService.CheckUniqueQuestion(context.TODO(), faqModel.Question)

	if err != nil {
		ms := tgbotapi.NewMessage(ChatID, "Під час створення питання виникла помилка")
		ms.ReplyMarkup = a.GetMainReplyMenuMarkup(msg.From.ID)
		go a.entrypoint.Send(ms)
		return
	}

	if check {
		ms := tgbotapi.NewMessage(ChatID, "Таке питання вже існує, введіть унікальне питання")
		go a.entrypoint.Send(ms)
		go a.handleCreationQuestionRequest(msg, updates)
		return
	}

	ms = tgbotapi.NewMessage(ChatID, "Введіть відповідь")
	ms.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Відміна")))

	_, err = a.entrypoint.Send(ms)

	err = a.handleEnterCreationFAQ(updates, ChatID, msg.From.ID, &faqModel, func(faqModel *models.FAQModel, text string) {
		faqModel.Answer = text
	})
	if err != nil {
		return
	}
	faqModel.Id = uuid.NewString()

	err = a.faqService.AddNewFaq(context.TODO(), faqModel)

	if err != nil {
		ms = tgbotapi.NewMessage(ChatID, "Виникла помилка при створенні питання! спробуйте трохи пізніше...")
		ms.ReplyMarkup = a.GetMainReplyMenuMarkup(msg.From.ID)
		a.entrypoint.Send(ms)
		return
	}

	ms = tgbotapi.NewMessage(ChatID, "FAQ було успішно додано!")
	ms.ReplyMarkup = a.GetMainReplyMenuMarkup(msg.From.ID)
	a.entrypoint.Send(ms)
}

func (a *BotApi) handleEnterCreationFAQ(updates <-chan tgbotapi.Update, ChatID int64, userID int, model *models.FAQModel, setterFunc func(faqModel *models.FAQModel, text string)) error {
	for {
		select {
		case questMsg, ok := <-updates:
			if !ok {
				continue
			}
			if questMsg.Message.Chat.ID != ChatID {
				continue
			}
			if questMsg.Message.Text == "Відміна" {
				ms := tgbotapi.NewMessage(ChatID, "Створення скасовано")
				ms.ReplyMarkup = a.GetMainReplyMenuMarkup(questMsg.Message.From.ID)
				go a.entrypoint.Send(ms)
				return errors.New("Cancelled")
			}
			setterFunc(model, questMsg.Message.Text)
			break
		case <-time.After(time.Minute):
			ms := tgbotapi.NewMessage(ChatID, "Час на створення питання витрачено")
			ms.ReplyMarkup = a.GetMainReplyMenuMarkup(userID)
			go a.entrypoint.Send(ms)
			return errors.New("Cancelled")
		}
		break
	}
	return nil
}

func (a *BotApi) handleGetAnswer(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	resp := tgbotapi.CallbackConfig{
		CallbackQueryID: query.ID,
		Text:            "Запит оброблюється",
		ShowAlert:       false,
	}
	a.entrypoint.AnswerCallbackQuery(resp)

	questionId := strings.Split(query.Data, ":")[1]
	answer, err := a.faqService.GetFaqById(context.TODO(), questionId)

	if err != nil {
		logrus.Errorln(err.Error())
		return
	}

	ms := tgbotapi.NewMessage(chatID, "Питання: "+answer.Question+"\nВідповідь: "+answer.Answer)
	_, err = a.entrypoint.Send(ms)
	if err != nil {
		logrus.Errorln(err.Error())
		return
	}
}

func (a *BotApi) GetMainReplyMenuMarkup(userId int) tgbotapi.ReplyKeyboardMarkup {
	btn1 := tgbotapi.NewKeyboardButton("Найчастіші питання")
	root, err := a.userService.CheckUserAdminRole(userId)
	if err != nil {
		logrus.Errorln(err.Error())
		return tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(btn1))
	}
	if root {
		btn2 := tgbotapi.NewKeyboardButton("Додати FAQ")
		btn3 := tgbotapi.NewKeyboardButton("Список FAQ")
		return tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(btn1), tgbotapi.NewKeyboardButtonRow(btn2, btn3))
	}

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
				go a.handlePublicQuestionsRequest(update.Message)
				continue
			}

			if update.Message != nil && update.Message.Text == "Додати FAQ" {
				go a.handleCreationQuestionRequest(update.Message, updates)
				continue
			}

			if update.Message != nil && update.Message.Text == "Список FAQ" {
				go a.handlePrivateQuestionsRequest(update.Message, updates)
				continue
			}

			if update.CallbackQuery != nil && update.CallbackQuery.Message.Text == "Список питань:" {
				logrus.Debug("Try answer...")
				go a.handleGetAnswer(update.CallbackQuery)
				continue
			}
		default:
			time.Sleep(time.Millisecond * 20)
		}
	}
}
