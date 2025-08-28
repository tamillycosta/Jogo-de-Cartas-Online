package models


import (
	
)


type Round struct{
	ID int
	Sender *Player
}



type Match struct{
	ID string
	Player1 *Player
	Player2 *Player
	Duration int
	Round *Round
	Status map[string]string
}