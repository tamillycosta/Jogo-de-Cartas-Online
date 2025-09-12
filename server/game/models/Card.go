package models


import(
"github.com/google/uuid"
 "fmt"
 "encoding/json"
)



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
	InDeck   bool   `gorm:"default:false" json:"inDeck"` 
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


func DecodeCards(data interface{}) ([]*Card, error) {
	cards, ok := data.(string)

	if !ok {
		return nil, fmt.Errorf("dados do player não são uma string válida")
	}

	var newCards []*Card
	err := json.Unmarshal([]byte(cards), &newCards);

	if err != nil {	
		return nil, fmt.Errorf("não foi possivel decodificar as cartas")
	}

	return newCards,nil
}
