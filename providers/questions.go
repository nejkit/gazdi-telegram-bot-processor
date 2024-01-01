package providers

import (
	"context"
	"gazdi-telegram-bot-processor/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	InsertFaqQuery  = "insert into faq values ($1, $2, $3)"
	GetFaqByIdQuery = "select id, question, answer from faq where id = $1"
)

type QuestionsProvider struct {
	client *pgxpool.Pool
}

func NewQuestionsProvider(url string) (*QuestionsProvider, error) {

	client, err := pgxpool.Connect(context.TODO(), url)
	if err != nil {
		return nil, err
	}
	return &QuestionsProvider{client: client}, nil
}

func (p *QuestionsProvider) InsertNewFaq(model models.FAQModel) error {
	con, err := p.client.Acquire(context.TODO())
	if err != nil {
		return err
	}
	defer con.Release()

	_, err = con.Exec(context.TODO(), InsertFaqQuery, model.Id, model.Question, model.Answer)
	if err != nil {
		return err
	}

	return nil
}

func (p *QuestionsProvider) GetFaqById(id string) (*models.FAQModel, error) {
	con, err := p.client.Acquire(context.TODO())
	if err != nil {
		return nil, err
	}
	defer con.Release()

	var faqModel models.FAQModel

	err = con.QueryRow(context.TODO(), GetFaqByIdQuery, id).Scan(&faqModel.Id, &faqModel.Question, &faqModel.Answer)
	if err != nil {
		return nil, err
	}
	return &faqModel, nil
}

func (p *QuestionsProvider) GetAllQuestions() ([]models.FAQModel, error) {
	con, err := p.client.Acquire(context.TODO())
	if err != nil {
		return nil, err
	}
	defer con.Release()

	var data []models.FAQModel

	rows, err := con.Query(context.TODO(), "select id, question, answer from faq")

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var faqModel models.FAQModel
		if err = rows.Scan(&faqModel.Id, &faqModel.Question, &faqModel.Answer); err != nil {
			return nil, err
		}
		data = append(data, faqModel)
	}

	return data, nil
}

func (p *QuestionsProvider) DeleteById(id string) error {
	con, err := p.client.Acquire(context.TODO())
	if err != nil {
		return err
	}

	defer con.Release()

	_, err = con.Exec(context.TODO(), "delete from faq where id = $1", id)
	if err != nil {
		return err
	}
	return nil
}
