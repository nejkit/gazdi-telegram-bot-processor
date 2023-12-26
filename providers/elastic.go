package providers

import (
	"context"
	"github.com/olivere/elastic/v7"
)

var (
	FaqIndexName = "faq"
)

type ElasticProvider struct {
	client *elastic.Client
}

func NewElasticSearchProvider(url string) (*ElasticProvider, error) {

	client, err := elastic.NewClient(elastic.SetURL(url), elastic.SetSniff(false))
	if err != nil {
		return nil, err
	}
	return &ElasticProvider{client: client}, nil
}

func (p *ElasticProvider) InsertData(indexName string, docId string, data string) error {
	_, err := p.client.Index().
		Index(indexName).
		Id(docId).
		BodyJson(data).
		Refresh(indexName).
		Do(context.TODO())

	if err != nil {
		return err
	}

	return nil
}

func (p *ElasticProvider) GetDataById(indexName string, id string) (*string, error) {
	query := elastic.NewMatchPhraseQuery("id", id)
	searchResult, err := p.client.Search().
		Index(indexName).
		Query(query).
		Do(context.TODO())

	if err != nil {
		return nil, err
	}

	var data string

	for _, hit := range searchResult.Hits.Hits {
		data = string(hit.Source)
	}

	return &data, err
}

func (p *ElasticProvider) GetAllFromIndex(indexName string) ([]string, error) {
	searchResult, err := p.client.Search().Index(indexName).Size(10000).Do(context.TODO())

	if err != nil {
		return nil, err
	}

	var data []string

	for _, hit := range searchResult.Hits.Hits {
		data = append(data, string(hit.Source))
	}

	return data, nil
}
