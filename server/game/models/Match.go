package models

import (
	"fmt"
	request "jogodecartasonline/api/Request"
	"math/rand"
	"net"
	"sync"

	"github.com/google/uuid"
)

const (
	ACTION_CHOOSE_CARD = "chooseCard"
	ACTION_ATTACK      = "attack"
	ACTION_LEAVE_MATCH = "leaveMatch"

	// Estados do jogo
	GAME_STATUS_ACTIVE = "ACTIVE"
	GAME_STATUS_ENDED  = "ENDED"

	// Configurações do jogo
	INITIAL_PLAYER_LIFE = 3
)

// Estrutura para representar um jogador na lista de espera
type WaitingPlayer struct {
	Player *Player
	Conn   net.Conn
}

type Round struct {
	ID     int
	Sender *Player
}

type Match struct {
	ID          string
	Player1     *Player
	Player2     *Player
	Round       *Round
	Player1Life int
	Player2Life int
	Status      string
	Mu          sync.RWMutex
}

// Esturua para representar a ação de um player na partida
type GameActionResult struct {
	Success        bool                   `json:"success"`
	Action         string                 `json:"action"`
	PlayerResult   map[string]interface{} `json:"playerResult"`
	OpponentResult map[string]interface{} `json:"opponentResult"`
	GameState      map[string]interface{} `json:"gameState"`
	GameEnded      bool                   `json:"gameEnded"`
	Winner         *Player                `json:"winner,omitempty"`
	Message        string                 `json:"message"`
}

// Estrutura para representar o dano causado por um ataque na partida
type DamageResult struct {
	DamageDealt           int
	OpponentLifeRemaining int
	OpponentCardHP        int
	GameEnded             bool
	Winner                *Player
}

// retorna uma nova partida
func NewMatch(player1 *Player, player2 *Player) *Match {
	match := &Match{
		ID:          uuid.NewString(),
		Player1:     player1,
		Player2:     player2,
		Round:       &Round{ID: 1},
		Status:      GAME_STATUS_ACTIVE,
		Mu:          sync.RWMutex{},
		Player1Life: INITIAL_PLAYER_LIFE,
		Player2Life: INITIAL_PLAYER_LIFE,
	}
	return match
}



// Processa a escolha de carta de um jogador
func (lobby *Lobby) ProcessChoseCard(match *Match, currentPlayer *Player, req request.Request) GameActionResult {
	match.Mu.Lock()
	defer match.Mu.Unlock()

	if !lobby.IsPlayerTurn(match, currentPlayer) {
		return GameActionResult{
			Success: false,
			Action:  ACTION_CHOOSE_CARD,
			Message: "Não é sua vez de jogar",
		}
	}

	cardIndexStr, exists := req.Params["cardIndex"]
	if !exists {
		cardIndexStr = "0"
	}

	cardIndex := 0
	if _, err := fmt.Sscanf(cardIndexStr, "%d", &cardIndex); err != nil {
		cardIndex = 0
	}

	card := lobby.ChooseCard(*currentPlayer, cardIndex)

	if card == nil {
		fmt.Printf("Erro: carta não encontrada para índice %d\n", cardIndex)
		return GameActionResult{
			Success: false,
			Action:  ACTION_CHOOSE_CARD,
			Message: "Erro ao escolher carta ou carta inválida",
		}
	}

	if card.Health == 0 {
		return GameActionResult{
			Success: false,
			Action:  ACTION_CHOOSE_CARD,
			Message: "Esta carta ja esta sem vida, escolha outra para batalhar",
		}
	}

	return GameActionResult{
		Success: true,
		Action:  ACTION_CHOOSE_CARD,
		Message: fmt.Sprintf("%s escolheu uma carta", currentPlayer.Nome),
		PlayerResult: map[string]interface{}{
			"cardName":   card.Nome,
			"cardPower":  card.Power,
			"cardHealth": card.Health,
			"message":    "Carta escolhida com sucesso!",
		},
		OpponentResult: map[string]interface{}{
			"message":    fmt.Sprintf("%s escolheu uma carta", currentPlayer.Nome),
			"cardName":   card.Nome,
			"cardPower":  card.Power,
			"cardHealth": card.Health,
		},
		GameState: map[string]interface{}{
			"currentTurn": lobby.GetOpponent(match, currentPlayer).Nome,
			"roundId":     match.Round.ID,
		},
	}
}

// Verifica se a ataque de um jogador ira finalizar o jogo
func (lobby *Lobby) applyDamage(match *Match, opponent *Player, attackPower int) DamageResult {
	var opponentLife *int
	var winner *Player

	// Identifica qual vida modificar
	if opponent.ID == match.Player1.ID {
		opponentLife = &match.Player1Life
		winner = match.Player2 // Se Player1 morrer, Player2 ganha
	} else {
		opponentLife = &match.Player2Life
		winner = match.Player1 // Se Player2 morrer, Player1 ganha
	}


	cardHP := opponent.CurrentCard.Health

	// Aplica dano primeiro na carta
	if cardHP > attackPower {
		// Carta absorve todo o dano
		opponent.CurrentCard.Health -= attackPower

	} else {
		// Carta é destruída e o oponete perde 1 ponto
		opponent.CurrentCard.Health = 0
		*opponentLife -= 1
	}

	// Verifica se jogo terminou
	gameEnded := *opponentLife <= 0
	var finalWinner *Player = nil

	if gameEnded {
		match.Status = GAME_STATUS_ENDED
		finalWinner = winner
		match.RestoreCardsHp()
	}

	return DamageResult{

		OpponentLifeRemaining: *opponentLife,
		OpponentCardHP:        opponent.CurrentCard.Health,
		GameEnded:             gameEnded,
		Winner:                finalWinner,
	}
}

// Processa ataque de um player
func (lobby *Lobby) ProcessAttack(match *Match, currentPlayer *Player, req request.Request) GameActionResult {
	match.Mu.Lock()
	defer match.Mu.Unlock()

	if !lobby.IsPlayerTurn(match, currentPlayer) {
		return GameActionResult{
			Success: false,
			Action:  ACTION_ATTACK,
			Message: "Não é sua vez de atacar",
		}
	}

	//  Verifica se player tem carta atual
	if currentPlayer.CurrentCard == nil {
		return GameActionResult{
			Success: false,
			Action:  ACTION_ATTACK,
			Message: "Você precisa escolher uma carta antes de atacar",
		}
	}

	if currentPlayer.CurrentCard.Health == 0 {
		return GameActionResult{
			Success: false,
			Action:  ACTION_ATTACK,
			Message: "Sua carta ja esta sem vida, escolha outra",
		}
	}

	// Procura oponente
	opponent := lobby.GetOpponent(match, currentPlayer)
	if opponent == nil {
		return GameActionResult{
			Success: false,
			Action:  ACTION_ATTACK,
			Message: "Oponente não encontrado",
		}
	}

	//  Verifica se oponente tem carta
	if opponent.CurrentCard == nil {
		return GameActionResult{
			Success: false,
			Action:  ACTION_ATTACK,
			Message: "Oponente ainda não escolheu carta",
		}
	}

	attackPower := currentPlayer.Atack()
	damageResult := lobby.applyDamage(match, opponent, attackPower)

	return lobby.processAttackStatus(match, currentPlayer, attackPower, damageResult)
}


// Processa saída de um jogador da partida
func (lobby *Lobby) ProcessLeaveMatch(match *Match, leavingPlayer *Player) GameActionResult {
	match.Mu.Lock()
	defer match.Mu.Unlock()

	opponent := lobby.GetOpponent(match, leavingPlayer)

	match.Status = GAME_STATUS_ENDED

	fmt.Printf("🚪 %s saiu da partida %s\n", leavingPlayer.Nome, match.ID)

	return GameActionResult{
		Success:   true,
		Action:    ACTION_LEAVE_MATCH,
		Message:   fmt.Sprintf("%s saiu da partida", leavingPlayer.Nome),
		GameEnded: true,
		Winner:    opponent,
		PlayerResult: map[string]interface{}{
			"message": "Você saiu da partida",
			"result":  "LEFT",
		},
		OpponentResult: map[string]interface{}{
			"message": fmt.Sprintf("%s saiu da partida. Você ganhou!", leavingPlayer.Nome),
			"result":  "WIN_BY_FORFEIT",
			"score":   opponent.Score + 100,
		},
	}
}




// -------------------- Funções auxiliares -----------------------------------


func (match *Match) RestoreCardsHp(){

	for i := 0; i < 2; i++ {
		card := match.Player1.Cards[i]
		templateCard := BaseCards[card.TemplateID]
		match.Player1.Cards[i].Health = templateCard.Health
	}

	for i := 0; i < 2; i++ {
		card := match.Player2.Cards[i]
		templateCard := BaseCards[card.TemplateID]
		match.Player2.Cards[i].Health = templateCard.Health
	}
	
}


func (lobby *Lobby) GetOpponent(match *Match, currentPlayer *Player) *Player {
	if match == nil || currentPlayer == nil {
		return nil
	}

	if match.Player1 == nil || match.Player2 == nil {
		return nil
	}

	if match.Player1.ID == currentPlayer.ID {
		return match.Player2
	} else if match.Player2.ID == currentPlayer.ID {
		return match.Player1
	}

	return nil // Player não pertence ao match
}

// Verifica se é a vez do player
func (lobby *Lobby) IsPlayerTurn(match *Match, player *Player) bool {
	return match.Round.Sender.ID == player.ID
}

// Troca o turno
func (lobby *Lobby) SwitchTurn(match *Match) {
	if match.Round == nil || match.Round.Sender == nil {
		return
	}

	if match.Round.Sender.ID == match.Player1.ID {
		match.Round.Sender = match.Player2
	} else {
		match.Round.Sender = match.Player1
	}

	match.Round.ID++
}

// Retira a partida finalizada do lobby
func (lobby *Lobby) CleanupFinishedMatch(match *Match, winner *Player) {
	fmt.Printf("🧹 Limpando match finalizado: %s\n", match.ID)

	lobby.Mu.Lock()
	delete(lobby.Matchs, match.ID)

	// Limpa referências dos players
	if match.Player1 != nil {
		match.Player1.Match = nil

	}
	if match.Player2 != nil {
		match.Player2.Match = nil
	}
	lobby.Mu.Unlock()

	// Atualiza score do vencedor
	if winner != nil {
		winner.Score += 100
		lobby.DB.Save(winner)
		fmt.Printf("🏆 %s ganhou 100 pontos! Score total: %d\n", winner.Nome, winner.Score)
	}
}

func (match *Match) ChoseStartPlayer(player1 Player, player2 Player) {
	if rand.Intn(2) == 0 {
		match.Round.Sender = &player1
	} else {
		match.Round.Sender = &player2
	}
}

