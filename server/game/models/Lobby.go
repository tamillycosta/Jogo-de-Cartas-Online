package models

import (
	request "jogodecartasonline/api/Request"
	response "jogodecartasonline/api/Response"
	"net"
	"sync"

	"gorm.io/gorm"
)

type Lobby struct {
	Mu sync.Mutex
	Players map[string]*Player
    WaitQueue []*Player
    Matchs map[string]*Match
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

    if result.Error == nil { // existe no banco
        if lobby.isLog(username) {
            return resp.MakeErrorResponse(403, "Ação proibida - User já está logado", "403 Forbidden")
        }
   
        player.Conn = conn
        lobby.Players[player.ID] = &player

    } else if result.Error == gorm.ErrRecordNotFound {
        newPlayer := CreateAccount(req, conn)
        lobby.Players[newPlayer.ID] = &newPlayer
        lobby.DB.Create(&newPlayer)

    } else {
        return resp.MakeErrorResponse(500, "Erro ao acessar o banco", "500 Internal Server Error")
    }

    return resp.MakeSuccessResponse("User adicionado com sucesso", map[string]string{
        "id":   player.ID,
        "nome": player.Nome,
    })
}



func (lobby *Lobby) isLog(username string) bool{
    _, ok := lobby.Players[username]
    return ok
}

