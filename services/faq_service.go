package services

import (
	"context"
	"encoding/json"
	"gazdi-telegram-bot-processor/models"
	"gazdi-telegram-bot-processor/providers"
)

type FagService struct {
	elasticProvider providers.ElasticProvider
	redisProvider   providers.RedisProvider
}

var (
	faqHashName  = "faq:questions"
	fagIndexName = "faq-notes"
)

func NewFaqService(elasticProvider providers.ElasticProvider, redisProvider providers.RedisProvider) (*FagService, error) {
	service := &FagService{elasticProvider: elasticProvider, redisProvider: redisProvider}
	if err := service.updateRedisData(context.TODO()); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *FagService) updateRedisData(ctx context.Context) error {
	dataFromEs, err := s.elasticProvider.GetAllFromIndex(fagIndexName)

	if err != nil {
		return err
	}

	for _, data := range dataFromEs {
		fagModel := models.FAQModel{}
		if err = json.Unmarshal([]byte(data), &fagModel); err != nil {
			continue
		}
		if err = s.redisProvider.SetHash(ctx, faqHashName, fagModel.Question, fagModel.Id); err != nil {
			return err
		}
	}
	return nil
}

func (s *FagService) GetQuestionsList(ctx context.Context) ([]string, error) {
	res, err := s.redisProvider.GetAllKeysFromHash(ctx, faqHashName)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *FagService) GetAnswerByQuestion(ctx context.Context, question string) (*string, error) {
	res, err := s.redisProvider.GetOneFromHash(ctx, faqHashName, question)

	if err != nil {
		return nil, err
	}

	dataEs, err := s.elasticProvider.GetDataById(fagIndexName, *res)

	if err != nil {
		return nil, err
	}

	faqModel := models.FAQModel{}

	if err = json.Unmarshal([]byte(*dataEs), &faqModel); err != nil {
		return nil, err
	}

	return &faqModel.Answer, nil
}

func (s *FagService) AddNewFaq(ctx context.Context, model models.FAQModel) error {
	if err := s.redisProvider.SetHash(ctx, faqHashName, model.Question, model.Id); err != nil {
		return err
	}

	data, err := json.Marshal(model)

	if err != nil {
		return err
	}

	if err = s.elasticProvider.InsertData(fagIndexName, model.Id, string(data)); err != nil {
		go s.redisProvider.DeleteFromHash(ctx, faqHashName, model.Question)
		return err
	}
	return nil

}
