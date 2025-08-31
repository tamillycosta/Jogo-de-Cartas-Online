package models

import (
	"encoding/json"
	"errors"
	"fmt"
	request "jogodecartasonline/api/Request"
	response "jogodecartasonline/api/Response"
	

	"jogodecartasonline/utils"
	"log"

	"net"
	"sync"
	"time"

	"gorm.io/gorm"
)

type Lobby struct {
	Mu        sync.RWMutex
	Players   map[string]*Player
	WaitQueue []*WaitingPlayer
	Matchs    map[string]*Match
	DB        *gorm.DB
}

// Metodo Responssavel por adicionar um jogador ao Lobby
func (lobby *Lobby) AddPlayer(req request.Request, conn net.Conn) response.Response {
	lobby.Mu.Lock()
	defer lobby.Mu.Unlock()
	resp := response.Response{}

	username := req.Params["nome"]

	var player Player
	result := lobby.DB.Where("nome = ?", username).First(&player)

	if result.Error == nil {
		// existe no banco
		if lobby.isLog(username) {
			return resp.MakeErrorResponse(403, "Ação proibida - User já está logado", "403 Forbidden")
		}
		fmt.Print("entrou aqui")
		player.Conn = conn
		lobby.Players[player.Nome] = &player

	} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		newPlayer := CreateAccount(req, conn)
		lobby.DB.Create(&newPlayer)
		lobby.Players[newPlayer.Nome] = &newPlayer
		player = newPlayer
	} else {
		return resp.MakeErrorResponse(500, "Erro ao acessar o banco", "500 Internal Server Error")
	}

	return resp.MakeSuccessResponse("Jogador adicionado com sucesso", map[string]string{
		"player": utils.Encode(player),
	})
}

// Metodo Responssavel por adicionar jogador a lista de espera para encontrar uma partida
// Remove toda a complexidade de channels para matchmaking
func (lobby *Lobby) AddToWaitQueue(req request.Request, conn net.Conn) *WaitingPlayer {
    playerJson := req.Params["player"]

    var player Player
    if err := json.Unmarshal([]byte(playerJson), &player); err != nil {
        return nil
    }

    // Estrutura simples, sem channels desnecessários
    waitingPlayer := &WaitingPlayer{
        Player: &player,
        Conn:   conn,
        // Remove MatchCh - não precisa para matchmaking
    }

    lobby.Mu.Lock()
    lobby.WaitQueue = append(lobby.WaitQueue, waitingPlayer)
    lobby.Mu.Unlock()

    return waitingPlayer
}

// Tenta combinar dois players em uma partida
func (lobby *Lobby) TryMatchUsers(req request.Request, conn net.Conn) response.Response {
    resp := response.Response{}

    // Adiciona à fila
    waitingPlayer := lobby.AddToWaitQueue(req, conn)
    if waitingPlayer == nil {
        return resp.MakeErrorResponse(402, "Erro ao Adicionar Jogador a Fila de espera", "")
    }

    lobby.Mu.Lock()
    queueLength := len(lobby.WaitQueue)
    
    if queueLength >= 2 {
        // MATCH ENCONTRADO
        player1 := lobby.WaitQueue[0] 
        player2 := lobby.WaitQueue[1]
        lobby.WaitQueue = lobby.WaitQueue[2:]
        lobby.Mu.Unlock()

        // Cria match
        match := NewMatch(*player1.Player, *player2.Player)
        match.ChoseStartPlayer(*player1.Player, *player2.Player)

        lobby.Mu.Lock()
        lobby.Matchs[match.ID] = &match
        player1.Player.Match = &match
        player2.Player.Match = &match
        lobby.Mu.Unlock()

        // 🎯 IDENTIFICA QUEM FEZ A REQUISIÇÃO
        if player1.Player.Nome == req.User {
            // Player1 fez a requisição
            lobby.NotifyOpponent(player2, &match, player1.Player)
            return lobby.MakeMatchResponse(&match, player2.Player)
        } else {
            // Player2 fez a requisição  
            lobby.NotifyOpponent(player1, &match, player2.Player)
            return lobby.MakeMatchResponse(&match, player1.Player)
        }
       
	

    } else {
        lobby.Mu.Unlock()
        return resp.MakeSuccessResponse("Procurando partida...", map[string]string{
            "posicao": fmt.Sprintf("%d", queueLength),
        })
    }
}


func (lobby *Lobby) NotifyOpponent(opponent *WaitingPlayer, match *Match, requestingPlayer *Player) {
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
        fmt.Printf("❌ Erro ao serializar notificação: %v\n", err)
        return
    }

    message := append(data, '\n')
    _, err = opponent.Conn.Write(message)
    if err != nil {
        fmt.Printf("❌ Erro ao notificar %s: %v\n", opponent.Player.Nome, err)
        return
    }

    fmt.Printf("✅ %s notificado sobre match\n", opponent.Player.Nome)
}

// Cria resposta padrão para quem fez a requisição
func (lobby *Lobby) MakeMatchResponse(match *Match, opponent *Player) response.Response {
    resp := response.Response{}
    return resp.MakeSuccessResponse("Partida Encontrada!", map[string]string{
        "matchId":  match.ID,
        "opponent": opponent.Nome,
         
    })
}
















// Remove player da fila (em caso de timeout ou desconexão)
func (lobby *Lobby) RemoveFromQueue(targetPlayer *WaitingPlayer) {
	lobby.Mu.Lock()
	defer lobby.Mu.Unlock()

	for i, waitingPlayer := range lobby.WaitQueue {
		if waitingPlayer == targetPlayer {
			lobby.WaitQueue = append(lobby.WaitQueue[:i], lobby.WaitQueue[i+1:]...)
			break
		}
	}
}

// Verifica se o User esta logado
func (lobby *Lobby) isLog(username string) bool {
	_, ok := lobby.Players[username]
	return ok
}

// status do siistema
func (l *Lobby) PrintStats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		l.Mu.RLock()
		totalUsers := len(l.Players)
		waitingUsers := len(l.WaitQueue)
		activeChats := len(l.Matchs)
		l.Mu.RUnlock()

		log.Printf("Stats: %d usuários conectados, %d na fila, %d chats ativos",
			totalUsers, waitingUsers, activeChats)
	}
}
