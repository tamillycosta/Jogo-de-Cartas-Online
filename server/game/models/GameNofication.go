package models


import(
	"net"
	
	"encoding/json"
	response "jogodecartasonline/api/Response"
	"fmt"
)



func  NotifyOpponent(opponent *WaitingPlayer, match *Match, requestingPlayer *Player) {
    // Cria uma Response padrão para notificação
    resp := response.Response{}
    notification := resp.MakeSuccessResponse("Partida Encontrada!", map[string]string{
        "matchId":   match.ID,
        "opponent":  requestingPlayer.Nome,
        "type":      "MATCH_FOUND", // Para identificar que é notificação
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
         
    })
}

