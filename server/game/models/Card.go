package models

import "github.com/google/uuid"


type Package struct{
	Cards []*Card
}


type Card struct {
	ID        string     `gorm:"type:char(36);primaryKey"`
	Nome      string     `gorm:"size:100;not null"`
	Power     int
	Rarity    string     `gorm:"size:50"`
	Players   []*Player  `gorm:"many2many:player_cards;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Health    int
}


func NewCard() Card{
	id := uuid.NewString()
	Nome := "Dragão Normal"
	Power := 50
	Rarity := "Normal"
	Players := []*Player{}
	Health := 100
	card := &Card{ID: id, Nome: Nome, Power: Power, Rarity: Rarity, Players: Players, Health: Health}
	return *card
}