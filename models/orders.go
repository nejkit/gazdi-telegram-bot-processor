package models

type DeliveryOpt struct {
	PostProvider          string  `json:"post_provider,omitempty"`
	DestinationCity       string  `json:"destination_city,omitempty"`
	DestinationDepartment int32   `json:"destination_department,omitempty"`
	LastName              string  `json:"last_name,omitempty"`
	FirstName             string  `json:"first_name,omitempty"`
	MiddleName            string  `json:"middle_name,omitempty"`
	Phone                 string  `json:"phone,omitempty"`
	Cost                  float64 `json:"cost,omitempty"`
}

type ShortProductInfo struct {
	Id         string  `json:"id,omitempty"`
	ExternalId string  `json:"external_id,omitempty"`
	Name       string  `json:"name,omitempty"`
	Sku        string  `json:"sku,omitempty"`
	Quantity   float64 `json:"quantity,omitempty"`
	Price      float64 `json:"price,omitempty"`
}

type OrderInfo struct {
	Id               string             `json:"id,omitempty"`
	LastName         string             `json:"last_name,omitempty"`
	FirstName        string             `json:"first_name,omitempty"`
	MiddleName       string             `json:"middle_name,omitempty"`
	RequiredCallback bool               `json:"required_callback,omitempty"`
	PhoneForCallback string             `json:"phone_for_callback,omitempty"`
	DateCreated      int64              `json:"date_created,omitempty"`
	Comment          string             `json:"comment,omitempty"`
	FullPrice        float64            `json:"full_price,omitempty"`
	Cart             []ShortProductInfo `json:"cart,omitempty"`
	DeliveryOpt      DeliveryOpt        `json:"delivery_opt"`
}
