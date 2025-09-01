package models

import (
	"fmt"
	request "jogodecartasonline/api/Request"
	response "jogodecartasonline/api/Response"
	"jogodecartasonline/utils"
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
func NewMatch(player1 Player, player2 Player) *Match {
	match := &Match{
		ID:          uuid.NewString(),
		Player1:     &player1,
		Player2:     &player2,
		Round:       &Round{ID: 1},
		Status:      GAME_STATUS_ACTIVE,
		Mu:          sync.RWMutex{},
		Player1Life: INITIAL_PLAYER_LIFE,
		Player2Life: INITIAL_PLAYER_LIFE,
	}
	return match
}

// retorna o jogador a começar
func (match *Match) ChoseStartPlayer(player1 Player, player2 Player) {
	if rand.Intn(2) == 0 {
		match.Round.Sender = &player1
	} else {
		match.Round.Sender = &player2
	}
}

// Processa as jogadas de um player na partida
func (lobby *Lobby) ProcessGameAction(req request.Request, conn net.Conn) response.Response {
	resp := response.Response{}

	lobby.Mu.RLock()
	player := lobby.Players[req.User]
	lobby.Mu.RUnlock()

	if player == nil {
		return resp.MakeErrorResponse(404, "Player não encontrado", "")
	}

	if player.Match == nil {
		return resp.MakeErrorResponse(400, "Player não está em uma partida", "")
	}

	match := player.Match

	if match.Status != GAME_STATUS_ACTIVE {
		return resp.MakeErrorResponse(400, "Partida já finalizada", "")
	}

	var actionResult GameActionResult
	switch req.Method {
	case ACTION_CHOOSE_CARD:
		actionResult = lobby.ProcessChoseCard(match, player, req)
	case ACTION_ATTACK:
		actionResult = lobby.ProcessAttack(match, player, req)
	case ACTION_LEAVE_MATCH:
		actionResult = lobby.ProcessLeaveMatch(match, player)
	default:
		return resp.MakeErrorResponse(400, "Ação não reconhecida", "")
	}

	if actionResult.Success && !actionResult.GameEnded {
		opponent := lobby.GetOpponent(match, player)
		if opponent != nil {
			if err := NotifyOpponentAction(player, actionResult); err != nil {
				fmt.Printf("⚠️ Erro ao notificar oponente %s: %v\n", opponent.Nome, err)
			}
		}
	}

	if actionResult.Success && !actionResult.GameEnded {
		lobby.SwitchTurn(match)
	}

	return resp.MakeSuccessResponse("Ação processada", map[string]string{
		"result": utils.Encode(actionResult),
	})

}

// Processa a escolha de carta de um jogador
func (lobby *Lobby) ProcessChoseCard(match *Match, currentPlayer *Player, req request.Request) GameActionResult {
	match.Mu.Lock()
	defer match.Mu.Unlock()

	// verifica se é a vez do jogador
	if !lobby.IsPlayerTurn(match, currentPlayer) {
		return GameActionResult{
			Success: false,
			Action:  ACTION_CHOOSE_CARD,
			Message: "Não é sua vez de jogar",
		}
	}

	card := currentPlayer.ChoseCard()

	return GameActionResult{
		Success: true,
		Action:  ACTION_CHOOSE_CARD,
		Message: fmt.Sprintf("%s escolheu uma carta", currentPlayer.Nome),
		PlayerResult: map[string]interface{}{
			"card":    utils.Encode(card),
			"message": "Carta escolhida com sucesso!",
		},
		OpponentResult: map[string]interface{}{
			"message": fmt.Sprintf("%s escolheu uma carta", currentPlayer.Nome),
		},
		GameState: map[string]interface{}{
			"currentTurn": lobby.GetOpponent(match, currentPlayer).Nome,
			"round":       match.Round,
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
	if cardHP >= attackPower {
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
		lobby.EndMatch(match, finalWinner)
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

	return GameActionResult{
		Success:   true,
		Action:    ACTION_ATTACK,
		Message:   fmt.Sprintf("%s atacou com poder %d", currentPlayer.Nome, attackPower),
		GameEnded: damageResult.GameEnded,
		Winner:    damageResult.Winner,
		PlayerResult: map[string]interface{}{
			"attackPower":    attackPower,
			"damageDealt":    damageResult.DamageDealt,
			"opponentLife":   damageResult.OpponentLifeRemaining,
			"opponentCardHP": damageResult.OpponentCardHP,
			"message":        "Ataque realizado!",
		},
		OpponentResult: map[string]interface{}{
			"damageTaken":     damageResult.DamageDealt,
			"lifeRemaining":   damageResult.OpponentLifeRemaining,
			"cardHPRemaining": damageResult.OpponentCardHP,
			"message":         fmt.Sprintf("Você recebeu %d de dano", damageResult.DamageDealt),
		},
		GameState: lobby.getGameStateMap(match),
	}
}

// Processa saída do match
func (lobby *Lobby) ProcessLeaveMatch(match *Match, leavingPlayer *Player) GameActionResult {
	match.Mu.Lock()
	defer match.Mu.Unlock()

	opponent := lobby.GetOpponent(match, leavingPlayer)
	match.Status = GAME_STATUS_ENDED

	// Remove match do lobby
	lobby.Mu.Lock()
	delete(lobby.Matchs, match.ID)
	delete(lobby.Players, leavingPlayer.Nome)
	lobby.Mu.Unlock()

	// Fecha conexão
	leavingPlayer.LeaveMatch(*leavingPlayer)

	return GameActionResult{
		Success:   true,
		Action:    ACTION_LEAVE_MATCH,
		Message:   fmt.Sprintf("%s saiu da partida", leavingPlayer.Nome),
		GameEnded: true,
		Winner:    opponent, // Oponente ganha por W.O.
		PlayerResult: map[string]interface{}{
			"message": "Você saiu da partida",
		},
		OpponentResult: map[string]interface{}{
			"message": fmt.Sprintf("%s saiu da partida. Você ganhou!", leavingPlayer.Nome),
		},
	}
}








// -------------------- Funções auxiliares -----------------------------------

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

// Finaliza o match
func (lobby *Lobby) EndMatch(match *Match, winner *Player) {
	// Remove match do lobby
	lobby.Mu.Lock()
	delete(lobby.Matchs, match.ID)
	lobby.Mu.Unlock()

	// Atualiza scores no banco
	if winner != nil {
		winner.Score++
		lobby.DB.Save(winner)
	}

	fmt.Printf("🏆 Match %s finalizado. Vencedor: %s\n", match.ID, winner.Nome)
}

func (lobby *Lobby) getGameStateMap(match *Match) map[string]interface{} {
	var currentTurn string
	if match.Round != nil && match.Round.Sender != nil {
		currentTurn = match.Round.Sender.Nome
	} else {
		currentTurn = "UNKNOWN"
	}

	return map[string]interface{}{
		"matchId":     match.ID,
		"currentTurn": currentTurn,
		"roundNumber": match.Round.ID,
		"player1Life": match.Player1Life,
		"player2Life": match.Player2Life,
		"status":      match.Status,
		"player1Name": match.Player1.Nome,
		"player2Name": match.Player2.Nome,
	}
}
