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
	Match     *Match          `gorm:"-"` 
	Conn      net.Conn         `gorm:"-"`                                       // ignorado pelo GORM
	
}



func CreateAccount(req request.Request, conn net.Conn) Player {
	id := uuid.NewString()
    username := req.Params["nome"]
	score := 0
	var card []*Card
	player := &Player{ID:id, Nome: username, Score: score, Cards: card, Conn: conn}

    return *player
    
}


