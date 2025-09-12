package models

import (
	request "jogodecartasonline/api/Request"
	"net"
	"gorm.io/gorm"
	"github.com/google/uuid"
	"fmt"
)

type Player struct {
	ID          string `gorm:"type:char(36);primaryKey"`
	Nome        string `gorm:"size:100;not null"`
	Score       int
	Cards       []*Card  `gorm:"foreignKey:PlayerId"`
	MatchID     *string  `gorm:"-"`
	Match       *Match   `gorm:"-"`
	Conn        net.Conn `gorm:"-" json:"-"`
    BattleDeck  []*Card `gorm:"-" json:"battleDeck"`
	CurrentCard *Card    `gorm:"-"`

}

func CreateAccount(req request.Request, conn net.Conn) Player {
	id := uuid.NewString()
	username := req.Params["nome"]
    starterCards := GenerateInicialCards(id)
	score := 0
	player := &Player{ID: id, Nome: username, Score: score, Cards: starterCards, Conn: conn}
	return *player

}

// relaciona 3 cartas iniciais ao player
func (lobby *Lobby) AddCard(player *Player) {
    count := 3 
    for i := 0; i < count; i++ {
        lobby.DB.Create(&player.Cards[i]) 
    }
}

// carrega todas as cartas do banco
func (p *Player) LoadCards(db *gorm.DB) error {
    return db.Where("player_id = ?", p.ID).Find(&p.Cards).Error
}

// Carrega apenas cartas marcadas como InDeck
func (p *Player) LoadBattleDeck(db *gorm.DB) error {
    return db.Where("player_id = ? AND in_deck = true", p.ID).Find(&p.BattleDeck).Error
}


func (p *Player) GetDeckCount(db *gorm.DB) int {
    var count int64
    db.Model(&Card{}).Where("player_id = ? AND in_deck = true", p.ID).Count(&count)
    return int(count)
}

func (p *Player) GetCardByName(cardName string) *Card {
    for _, card := range p.Cards {
        if card.Nome == cardName {
            return card
        }
    }
    return nil
}

func (p *Player) GetDeckCard(cardName string) *Card {
    for _, card := range p.BattleDeck {
        if card.Nome == cardName {
            return card
        }
    }
    return nil
}

// lista as cartas de batalha 


// Escolhe uma carta baseada no índice
func (p *Player) ChooseCardByIndex(index int) *Card {
    if len(p.Cards) == 0 || index < 0 || index >= len(p.Cards) {
        return nil
    }
    p.CurrentCard = p.Cards[index]
    return p.CurrentCard
}



func (lobby *Lobby) ChooseCard(player Player, cardIndex int) *Card {
    // Busca o player atual no mapa
    currentPlayer := lobby.Players[player.Nome]
    if currentPlayer == nil {
        return nil
    }
    
    // Carrega as cartas do banco se não estiverem carregadas
    if len(currentPlayer.Cards) == 0 {
        err := currentPlayer.LoadBattleDeck(lobby.DB)
        if err != nil {
            fmt.Printf("❌ Erro ao carregar cartas: %v\n", err)
            return nil
        }
    }
    
    // Escolhe a carta pelo índice
    return currentPlayer.ChooseCardByIndex(cardIndex)
}



func (player *Player) Atack() int {
	return player.CurrentCard.Power
}

