package models

import (
	

	"encoding/json"
	"fmt"
	"jogodecartasonline/utils"
	response "jogodecartasonline/api/Response"
)


func NotifyMatchFound(waitingPlayer *WaitingPlayer, match *Match, opponent *Player) {
	resp := response.Response{}
	notification := resp.MakeSuccessResponse("Partida Encontrada!", map[string]string{
		"type":     "MATCH_FOUND",
		"matchId":  match.ID,
		"opponent": opponent.Nome,
		"yourTurn": fmt.Sprintf("%t", match.Round.Sender.Nome == waitingPlayer.Player.Nome),
	})

	data, err := json.Marshal(notification)
	if err != nil {
		fmt.Printf("❌ Erro ao serializar notificação: %v\n", err)
		return
	}

	message := append(data, '\n')
	_, err = waitingPlayer.Conn.Write(message)
	if err != nil {

		return
	}

	fmt.Printf("✅ %s notificado sobre partida encontrada\n", waitingPlayer.Player.Nome)
}


func NotifyGameEnd(player *Player, gameResult GameActionResult, isWinner bool) {
    resp := response.Response{}
    
    resultType := "LOSS"
    message := "Você foi derrotado!"
    
    if isWinner {
        resultType = "WIN"
        message = "Você venceu!"
    }
    
    notification := resp.MakeSuccessResponse(message, map[string]string{
        "type":       "GAME_ENDED",
        "result":     resultType,
        "winner":     gameResult.Winner.Nome,
        "reason":     gameResult.Action, // attack ou leaveMatch
        "message":    message,
    })

    data, err := json.Marshal(notification)
    if err != nil {
        fmt.Printf("❌ Erro ao serializar notificação de fim de jogo: %v\n", err)
        return
    }

    message_bytes := append(data, '\n')
    _, err = player.Conn.Write(message_bytes)
    if err != nil {
        
        return
    }

    
}

func NotifyOpponentAction(opponent *Player, actionResult GameActionResult) error {
    resp := response.Response{}
    notification := resp.MakeSuccessResponse("Ação do oponente", map[string]string{
        "type":           "OPPONENT_ACTION",
        "action":         actionResult.Action,
        "actionResult":   utils.Encode(actionResult),
        "opponentResult": utils.Encode(actionResult.OpponentResult),
        "gameState":      utils.Encode(actionResult.GameState),
    })

    data, err := json.Marshal(notification)
    if err != nil {
        return fmt.Errorf("erro ao serializar: %v", err)
    }

    message := append(data, '\n')
    _, err = opponent.Conn.Write(message)
    if err != nil {
        return fmt.Errorf("erro ao enviar: %v", err)
    }

    fmt.Printf("🔔 %s notificado sobre ação\n", opponent.Nome)
    return nil
}

// Resposta para quem fez a requisição e encontrou match
func MakeMatchFoundResponse(match *Match, opponent *Player) response.Response {
	resp := response.Response{}
	return resp.MakeSuccessResponse("Partida Encontrada!", map[string]string{
		"type":     "MATCH_FOUND",
		"matchId":  match.ID,
		"opponent": opponent.Nome,
		"yourTurn": fmt.Sprintf("%t", match.Round.Sender.Nome != opponent.Nome),
	})
}

func (cm *ConnectionMonitor) notifyOpponentWinByDisconnect(opponent *Player) {
    resp := response.Response{}
    notification := resp.MakeSuccessResponse("O outro jogador desconectou", map[string]string{
        "type":       "GAME_ENDED",
        "result":     "WIN",
        "winner":     opponent.Nome,
        "reason":     "leaveMatch", 
    })

    data, err := json.Marshal(notification)
    if err != nil {
        fmt.Printf("❌ Erro ao serializar notificação de fim de jogo: %v\n", err)
        return
    }

    message_bytes := append(data, '\n')
    _, err = opponent.Conn.Write(message_bytes)
    if err != nil {
     
        return
    }

}



func (lobby *Lobby) processAttackStatus(match *Match, currentPlayer *Player, attackPower int, damageResult DamageResult) GameActionResult {

	return GameActionResult{
		Success:   true,
		Action:    ACTION_ATTACK,
		Message:   fmt.Sprintf("%s atacou com poder %d", currentPlayer.Nome, attackPower),
		GameEnded: damageResult.GameEnded,
		Winner:    damageResult.Winner,
		PlayerResult: map[string]interface{}{
			"attackPower":    attackPower,
			"opponentLife":   damageResult.OpponentLifeRemaining,
			"opponentCardHP": damageResult.OpponentCardHP,
			"score": damageResult.Winner.Score,
			"result": func() string {
				if damageResult.GameEnded {
					
					return "WIN"
				}
				return "ATTACK_SUCCESS"
			}(),
		},
		OpponentResult: map[string]interface{}{
			"damageTaken":     attackPower,
			"lifeRemaining":   damageResult.OpponentLifeRemaining,
			"cardHPRemaining": damageResult.OpponentCardHP,

			"result": func() string {
				if damageResult.GameEnded {
					return "LOSS"
				}
				return "DAMAGE_TAKEN"
			}(),
		},
		GameState: map[string]interface{}{
			"currentTurn": func() string {
				if damageResult.GameEnded {
					return "GAME_OVER"
				}
				return lobby.GetOpponent(match, currentPlayer).Nome
			}(),
			"roundId": match.Round.ID,
		},
	}
}
