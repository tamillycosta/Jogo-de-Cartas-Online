package models

import (
	"jogodecartasonline/api/Request"
	
	
	"net"

	"github.com/google/uuid"
)


type Player struct {
	ID        string           `gorm:"type:char(36);primaryKey"`                 // UUID string
	Nome      string           `gorm:"size:100;not null"`
	Score     int
	Cards     []*Card          `gorm:"many2many:player_cards;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	MatchID   *string          `gorm:"-"`                      // partida atual (nullable)
	Match     *Match           `gorm:"-"` 
	Conn      net.Conn         `gorm:"-" json:"-"`                                    // ignorado pelo GORM
	CurrentCard *Card           `gorm:"-"` 
}



func CreateAccount(req request.Request, conn net.Conn) Player {
	id := uuid.NewString()
    username := req.Params["nome"]
	score := 0
	var cards []*Card 
	card := NewCard()
	cards = append(cards, &card )
	player := &Player{ID:id, Nome: username, Score: score, Cards: cards, Conn: conn}
    return *player
    
}



func (player *Player) ChoseCard () *Card{	
	return  player.Cards[0]
}


func  (player *Player) Atack() int{
	
	return player.Cards[0].Power
}


func(p *Player) LeaveMatch  (player Player){
	p.Conn.Close()
}