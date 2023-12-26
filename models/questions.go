package models

type FAQModel struct {
	Id       string `json:"id,omitempty"`
	Question string `json:"question,omitempty"`
	Answer   string `json:"answer,omitempty"`
}
