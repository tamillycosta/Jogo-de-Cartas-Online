package models

import (
	"math/rand"
	"net"

	"github.com/google/uuid"
)

type WaitingPlayer struct {
	Player  *Player
	Conn    net.Conn
	MatchCh chan MatchResult // Channel para receber resultado do match
}

type Round struct {
	ID     int
	Sender *Player
}

type Match struct {
	ID       string
	Player1  *Player
	Player2  *Player
	Duration int
	Round    *Round
	Status   map[string]string
}

type MatchResult struct {
	Success bool
	Match   *Match
	Player1 *Player
	Player2 *Player
	Error   string
}

func NewMatch(player1 Player, player2 Player) Match {
	match := &Match{
		ID:       uuid.NewString(),
		Player1:  &player1,
		Player2:  &player2,
		Round:    &Round{},
		Duration: 0,
		Status:   map[string]string{},
	}
	return *match
}

func (match *Match) ChoseStartPlayer(player1 Player, player2 Player) {
	if rand.Intn(2) == 0 {
		match.Round.Sender = &player1
	} else {
		match.Round.Sender = &player2
	}
}
