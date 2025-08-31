package models


import(
	"net"
	"jogodecartasonline/utils"
	"encoding/json"
	response "jogodecartasonline/api/Response"
	"fmt"
)


// Envia Notificação do encontro de uma partida
func MatchFound(player1Waiting WaitingPlayer, player2Waiting WaitingPlayer, matchResult MatchResult  ) {
	// Notifica player 1
	select {
	case player1Waiting.MatchCh <- matchResult:
	default: // Se channel estiver fechado, ignora
	}
	
	// Notifica player 2  
	select {
	case player2Waiting.MatchCh <- matchResult:
	default: // Se channel estiver fechado, ignora
	}
}


func NotifyMatchFound(conn net.Conn, match *Match, player1, player2 *Player) {
    resp := response.Response{}
    notification := resp.MakeSuccessResponse("Partida Encontrada!", map[string]string{
        "type":     "MATCH_FOUND",
        "matchId":  match.ID,
        "player1":  utils.Encode(player1),
        "player2":  utils.Encode(player2),
        
    })

    data, err := json.Marshal(notification)
    if err != nil {
        fmt.Printf("❌ Erro ao serializar notificação: %v\n", err)
        return
    }

    _, err = conn.Write(data)
    if err != nil {
        fmt.Printf("❌ Erro ao enviar notificação para %s: %v\n", conn.RemoteAddr(), err)
        return
    }

    fmt.Printf("🎮 Notificação de match enviada para %s\n", conn.RemoteAddr())
}


// Envia notificação de timeout
func (lobby *Lobby) SendTimeoutNotification(conn net.Conn) {
    resp := response.Response{}
    notification := resp.MakeErrorResponse(408, "Timeout", "Nenhuma partida encontrada em 60 segundos")

    data, _ := json.Marshal(notification)
    conn.Write(data)
}
