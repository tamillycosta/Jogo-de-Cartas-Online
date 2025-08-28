package models


type Package struct{
	Cards []*Card
}


type Card struct {
	ID        string     `gorm:"type:char(36);primaryKey"`
	Nome      string     `gorm:"size:100;not null"`
	Power     int
	Rarity    string     `gorm:"size:50"`
	Players   []*Player  `gorm:"many2many:player_cards;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	
}