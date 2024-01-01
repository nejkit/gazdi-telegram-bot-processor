package services

import (
	"context"
	"gazdi-telegram-bot-processor/models"
	"gazdi-telegram-bot-processor/providers"
	"github.com/sirupsen/logrus"
	"strings"
)

type FagService struct {
	questionsProvider providers.QuestionsProvider
	redisProvider     providers.RedisProvider
}

var (
	faqHashName         = "faq:questiontoid"
	fagSetQuestionsName = "faq:questions"
)

func NewFaqService(elasticProvider providers.QuestionsProvider, redisProvider providers.RedisProvider) (*FagService, error) {
	service := &FagService{questionsProvider: elasticProvider, redisProvider: redisProvider}
	if err := service.updateRedisData(context.TODO()); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *FagService) updateRedisData(ctx context.Context) error {
	fagModels, err := s.questionsProvider.GetAllQuestions()

	if err != nil {
		return err
	}

	for _, model := range fagModels {

		if err = s.redisProvider.SetHash(ctx, faqHashName, model.Question, model.Id); err != nil {
			return err
		}
	}
	return nil
}

func (s *FagService) GetQuestionsList(ctx context.Context) ([]models.FAQModel, error) {
	res, err := s.redisProvider.GetAllFromHash(ctx, faqHashName)

	if err != nil {
		return nil, err
	}
	var modelList []models.FAQModel
	for k, v := range res {
		modelList = append(modelList, models.FAQModel{Id: v, Question: k})
	}

	return modelList, nil
}

func (s *FagService) GetFaqById(ctx context.Context, id string) (*models.FAQModel, error) {

	faqModel, err := s.questionsProvider.GetFaqById(id)

	if err != nil {
		return nil, err
	}

	return faqModel, nil
}

func (s *FagService) DeleteFaq(ctx context.Context, model models.FAQModel) {
	go s.redisProvider.DeleteFromHash(ctx, faqHashName, model.Question)
	go s.redisProvider.DeleteFromSet(ctx, fagSetQuestionsName, strings.ReplaceAll(strings.ToLower(model.Question), " ", ""))
	go s.questionsProvider.DeleteById(model.Id)
}

func (s *FagService) AddNewFaq(ctx context.Context, model models.FAQModel) error {
	if err := s.questionsProvider.InsertNewFaq(model); err != nil {
		return err
	}

	if err := s.redisProvider.SetHash(ctx, faqHashName, model.Question, model.Id); err != nil {
		return err
	}
	logrus.Debug(strings.ReplaceAll(strings.ToLower(model.Question), " ", ""))
	if err := s.redisProvider.AddToSet(ctx, fagSetQuestionsName, strings.ReplaceAll(strings.ToLower(model.Question), " ", "")); err != nil {
		return err
	}

	return nil
}

func (s *FagService) GetFagList(ctx context.Context) ([]models.FAQModel, error) {
	return s.questionsProvider.GetAllQuestions()
}

func (s *FagService) CheckUniqueQuestion(ctx context.Context, question string) (bool, error) {
	logrus.Debug(strings.ReplaceAll(strings.ToLower(question), " ", ""))
	return s.redisProvider.CheckMemberOfSet(ctx, fagSetQuestionsName, strings.ReplaceAll(strings.ToLower(question), " ", ""))
}
