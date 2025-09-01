package models

import (
	"net"

	"encoding/json"
	"fmt"
	"jogodecartasonline/utils"
	response "jogodecartasonline/api/Response"
)



func  NotifyOpponent(opponent *WaitingPlayer, match *Match, requestingPlayer *Player) {
    // Cria uma Response padrão para notificação
    resp := response.Response{}
    notification := resp.MakeSuccessResponse("Partida Encontrada!", map[string]string{
        "matchId":   match.ID,
        "opponent":  requestingPlayer.Nome,
        "type":      "MATCH_FOUND", // Para identificar que é notificação
		"yourTurn": fmt.Sprintf("%t", match.Round.Sender.Nome != opponent.Player.Nome),
    })

    // Envia via conexão TCP usando JSON padrão
    data, err := json.Marshal(notification)
    if err != nil {
        fmt.Printf("Erro ao serializar notificação: %v\n", err)
        return
    }

    message := append(data, '\n')
    _, err = opponent.Conn.Write(message)
    if err != nil {
        fmt.Printf("Erro ao notificar %s: %v\n", opponent.Player.Nome, err)
        return
    }

    fmt.Printf("✅ %s notificado sobre match\n", opponent.Player.Nome)
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









// Envia notificação de timeout
func (lobby *Lobby) SendTimeoutNotification(conn net.Conn) {
    resp := response.Response{}
    notification := resp.MakeErrorResponse(408, "Timeout", "Nenhuma partida encontrada em 60 segundos")

    data, _ := json.Marshal(notification)
    conn.Write(data)
}



// Cria resposta padrão para quem fez a requisição
func  MakeMatchResponse(match *Match, opponent *Player) response.Response {
    resp := response.Response{}
    return resp.MakeSuccessResponse("Partida Encontrada!", map[string]string{
        "matchId":  match.ID,
        "opponent": opponent.Nome,
        "yourTurn": fmt.Sprintf("%t", match.Round.Sender.Nome == opponent.Nome),
        "matchStatus" :match.Status,
    })
}

