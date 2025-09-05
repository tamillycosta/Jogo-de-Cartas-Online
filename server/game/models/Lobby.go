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
    result := lobby.DB.Preload("Cards").Where("nome = ?", username).First(&player) 

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
        lobby.AddCard(&newPlayer)
        
        // Carrega as cartas recém-criadas
        lobby.DB.Preload("Cards").Where("id = ?", newPlayer.ID).First(&newPlayer) // MUDANÇA AQUI
        
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
func (lobby *Lobby) AddToWaitQueue(req request.Request, conn net.Conn) *WaitingPlayer {
    playerJson := req.Params["player"]

    var player Player
    if err := json.Unmarshal([]byte(playerJson), &player); err != nil {
        return nil
    }

    waitingPlayer := &WaitingPlayer{
        Player: &player,
        Conn:   conn,
      
    }

    lobby.Mu.Lock()
    lobby.WaitQueue = append(lobby.WaitQueue, waitingPlayer)
    lobby.Mu.Unlock()

    return waitingPlayer
}

// Tenta combinar dois players em uma partida
func (lobby *Lobby) TryMatchUsers(req request.Request, conn net.Conn) response.Response {
	resp := response.Response{}
	
	playerName := req.User
	
	// Verifica se o jogador ja esta na fila 
	lobby.Mu.Lock()
	for _, waitingPlayer := range lobby.WaitQueue {
		if waitingPlayer.Player.Nome == playerName {
			lobby.Mu.Unlock()
			return resp.MakeErrorResponse(400, "Você já está na fila de espera", "")
		}
	}
	lobby.Mu.Unlock()
	
	

	// Adiciona à fila
	waitingPlayer := lobby.AddToWaitQueue(req, conn)
	if waitingPlayer == nil {
		return resp.MakeErrorResponse(402, "Erro ao Adicionar Jogador a Fila de espera", "")
	}

	lobby.Mu.Lock()
	queueLength := len(lobby.WaitQueue)
	
	// Se tiverem dois jogadores na lista 
	if queueLength >= 2 {
		
		waiting1 := lobby.WaitQueue[0]
        waiting2 := lobby.WaitQueue[1]
        lobby.WaitQueue = lobby.WaitQueue[2:]
        lobby.Mu.Unlock()


		player1 := lobby.Players[waiting1.Player.Nome]
        player2 := lobby.Players[waiting2.Player.Nome]

		if player1 == nil || player2 == nil {
            return resp.MakeErrorResponse(500, "Erro: Players não encontrados no lobby", "")
        }
		

		// Cria match
		match := NewMatch(player1, player2)
		match.ChoseStartPlayer(*player1, *player2)

		lobby.Mu.Lock()
		lobby.Matchs[match.ID] = match
		player1.Match = match
		player2.Match = match
		lobby.Mu.Unlock()

		
		if player1.Nome == playerName {
			// Player1 fez a requisição - notifica Player2
			NotifyMatchFound(waiting2, match, player1)
			return MakeMatchFoundResponse(match, player2)
		} else {
			// Player2 fez a requisição - notifica Player1  
			NotifyMatchFound(waiting1, match, player2)
			return MakeMatchFoundResponse(match, player1)
		}
		
	} else {
		lobby.Mu.Unlock()
		
		
		return resp.MakeSuccessResponse("Procurando partida...", map[string]string{
			"type":    "SEARCHING",
			"posicao": fmt.Sprintf("%d", queueLength),
		})
	}
}




func (lobby *Lobby) DeletePlayer(req request.Request, conn net.Conn) response.Response{
	lobby.Mu.Lock()
    defer lobby.Mu.Unlock()

	resp := response.Response{}
    username := req.Params["nome"]

	if !lobby.isLog(username){
		return resp.MakeErrorResponse(403, "Ação proibida - User Não Esta Conectado", "403 Forbidden")
	}

	player := lobby.Players[username]
	delete(lobby.Players, username)
	

	return resp.MakeSuccessResponse("Jogador removido do server com sucesso", map[string]string{
        "player": utils.Encode(player),
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
