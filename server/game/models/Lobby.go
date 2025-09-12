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
    ConnectionMonitor *ConnectionMonitor 
}

var GlobalPackSystem = &PackSystem{
    LastPackTime: make(map[string]time.Time),
    PackCooldown: 1 * time.Minute,
}

// Rotas visiveis 

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
        player.LoadBattleDeck(lobby.DB)
        player.Conn = conn
        lobby.Players[player.Nome] = &player

    } else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        // Player novo
        newPlayer := CreateAccount(req, conn)
        lobby.DB.Create(&newPlayer)
        lobby.AddCard(&newPlayer)
        
        // Carrega as cartas recém-criadas
        
        lobby.DB.Preload("Cards").Where("id = ?", newPlayer.ID).First(&newPlayer) 
        
        lobby.Players[newPlayer.Nome] = &newPlayer
        player = newPlayer
    } else {
        return resp.MakeErrorResponse(500, "Erro ao acessar o banco", "500 Internal Server Error")
    }

    return resp.MakeSuccessResponse("Jogador adicionado com sucesso", map[string]string{
        "player": utils.Encode(player),
    })
}

// Metodo Responssavel por desligar um player do server 
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

// Metodo Responssavel por criar uma partida entre dois jogadores 
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

// Metodo Responssavel por verificar status do pacote um jogador
func (lobby *Lobby) CheckPackStatus(req request.Request, conn net.Conn) response.Response {
    resp := response.Response{}
    
    playerName := req.User
    player := lobby.Players[playerName]
    
    if player == nil {
        return resp.MakeErrorResponse(404, "Player não encontrado", "")
    }
    
    canOpen, remaining := GlobalPackSystem.CanOpenPack(player.ID)
    
   fmt.Printf("requisição recebida")
    return resp.MakeSuccessResponse("Status verificado", map[string]string{
        "type":       "PACKAGE_STATUS", 
        "canOpen":    fmt.Sprintf("%t", canOpen),
        "remaining":  remaining.String(),
        "totalCards": fmt.Sprintf("%d", len(player.Cards)),
        "player": utils.Encode(player),
    })
}

// Metodo Responssavel por abir pacote de um joagdor 
func (lobby *Lobby) OpenCardPack(req request.Request, conn net.Conn) response.Response {
    resp := response.Response{}
    username := req.User
    player := lobby.Players[username]
    
    if player == nil {
        return resp.MakeErrorResponse(404, "Player não encontrado", "")
    }
    
    // Verifica se pode abrir
    canOpen, remaining := GlobalPackSystem.CanOpenPack(player.ID)
    if !canOpen {
        return resp.MakeErrorResponse(400, "Pacote em cooldown", remaining.String())
    }
    
    // Abre pacote
    newCards, err := GlobalPackSystem.OpenPack(player.ID)
    if err != nil {
        return resp.MakeErrorResponse(500, "Erro ao abrir pacote", err.Error())
    }

    // Salva novas cartas
    for _, card := range newCards {
        lobby.DB.Create(card)
        player.Cards = append(player.Cards, card)
    }
    lobby.DB.Save(player)
    fmt.Printf("jogador %s abriu um pacote!!", player.Nome)
    return resp.MakeSuccessResponse("Pacote aberto!", map[string]string{
        "type":       "PACKAGE_OPENED",
        "cards":      utils.Encode(newCards),
        "totalCards": fmt.Sprintf("%d", len(player.Cards)),
    })
}

// Metodo Responssavel por apresentar os status das conexões atuais no servidor
func (lobby *Lobby) GetConnectionStats(req request.Request, conn net.Conn) response.Response {
    resp := response.Response{}
    
    if lobby.ConnectionMonitor == nil {
        return resp.MakeErrorResponse(500, "Monitor não inicializado", "")
    }
    
    stats := lobby.ConnectionMonitor.GetStats()
    
    // Adiciona estatísticas do lobby
    lobby.Mu.RLock()
    stats["totalPlayers"] = len(lobby.Players)
    stats["waitingPlayers"] = len(lobby.WaitQueue)
    stats["activeMatches"] = len(lobby.Matchs)
    lobby.Mu.RUnlock()
    
    return resp.MakeSuccessResponse("Estatísticas de conexão", map[string]string{
        "stats": utils.Encode(stats),
    })
}



func (lobby *Lobby) SelectMatchDeck(req request.Request, conn net.Conn) response.Response {
    resp := response.Response{}
    
    username := req.User
    oldCardName := req.Params["oldCardName"] // Carta a remover do deck
    newCardName := req.Params["newCardName"] // Carta a adicionar ao deck
    
   
    var player Player
    if err := lobby.DB.Preload("Cards").Where("nome = ?", username).First(&player).Error; err != nil {
        return resp.MakeErrorResponse(404, "Player não encontrado no banco", "")
    }
    
    // Encontra as cartas
    var oldCard *Card = nil
    var newCard *Card = nil
    
    for _, card := range player.Cards {
        if card.Nome == oldCardName && card.InDeck {
            oldCard = card
        }
        if card.Nome == newCardName && !card.InDeck {
            newCard = card
        }
    }
    
    if oldCard == nil {
        return resp.MakeErrorResponse(400, "Carta antiga não encontrada no deck", "")
    }
    
    if newCard == nil {
        return resp.MakeErrorResponse(400, "Carta nova não encontrada ou já está no deck", "")
    }
    
    // Atualiza status no banco
    oldCard.InDeck = false
    newCard.InDeck = true
    
    lobby.DB.Save(oldCard)
    lobby.DB.Save(newCard)
    
    // Atualiza player logado se existir
    if loggedPlayer := lobby.Players[username]; loggedPlayer != nil {
        loggedPlayer.LoadBattleDeck(lobby.DB)
    }
    
    return resp.MakeSuccessResponse("Deck atualizado com sucesso!", map[string]string{
        "removed": oldCardName,
        "added":   newCardName,
    })
}



func (lobby *Lobby) ListCards(req request.Request, conn net.Conn) response.Response {
    resp := response.Response{}
    playerID := req.Params["ID"]
    
    var player Player
    if err := lobby.DB.Preload("Cards").Where("id = ?", playerID).First(&player).Error; err != nil {
        return resp.MakeErrorResponse(404, "Player não encontrado", "")
    }
    
    // Separa cartas do deck das outras
    var deckCards []*Card
    var otherCards []*Card
    
    for _, card := range player.Cards {
        if card.InDeck {
            deckCards = append(deckCards, card)
        } else {
            otherCards = append(otherCards, card)
        }
    }
    
    fmt.Printf("jogador %s listou suas cartas", playerID)
    return resp.MakeSuccessResponse("Cartas listadas com sucesso!", map[string]string{
        "type": "LIST_CARDS",
        "deckCards":  utils.Encode(deckCards),
        "otherCards": utils.Encode(otherCards),
        "deckCount":  fmt.Sprintf("%d", len(deckCards)),
        "totalCards": fmt.Sprintf("%d", len(player.Cards)),
    })
}


//--------------------------------- Auxliares

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



// status do sistema
func (l *Lobby) PrintStats() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        l.Mu.RLock()
        totalUsers := len(l.Players)
        waitingUsers := len(l.WaitQueue)
        activeMatches := len(l.Matchs)
        l.Mu.RUnlock()

        log.Printf("Stats: %d usuários conectados, %d na fila, %d partidas ativas",
            totalUsers, waitingUsers, activeMatches)
            
        // Estatísticas de conexão
        if l.ConnectionMonitor != nil {
            stats := l.ConnectionMonitor.GetStats()
            if problematic, ok := stats["problematicConnections"].(int); ok && problematic > 0 {
                log.Printf("⚠️ Conexões com problemas: %d", problematic)
            }
        }
    }
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