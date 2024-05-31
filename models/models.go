package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Account struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	Balance  float64            `json:"balance" bson:"balance"`
	Currency string             `json:"currency" bson:"currency"`
	Security *Secret      		`json:"security" bson:"security"`
}

type Secret struct{
	NickName string `json:"nickname" bson:"nickname"`
	Password string `json:"password" bson:"password"`
}
