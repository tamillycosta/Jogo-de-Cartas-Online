package models

import "github.com/google/uuid"

type Package struct {
	Cards []*Card
}

type Card struct {
	ID       string `gorm:"type:char(36);primaryKey"`
	Nome     string `gorm:"size:100;not null"`
	Power    int
	Rarity   string `gorm:"size:50"`
	PlayerId string // <- FK para Player.ID
	Health   int
	TemplateID string `gorm:"size:100;not null"`
	CurrentCopies int 
	MaxCopies int 
	IsSpecial bool
}



func NewCard(player *Player) Card {
	id := uuid.NewString()
	Nome := "Dragão Normal"
	Power := 50
	Rarity := "Normal"

	Health := 100
	card := &Card{ID: id, Nome: Nome, Power: Power, Rarity: Rarity, PlayerId: player.ID, Health: Health, CurrentCopies: 0, MaxCopies: 0}
	return *card
}
