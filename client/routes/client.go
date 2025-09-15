package model

import (
	"bufio"
	"encoding/json"
	request "jogodecartasonline/api/Request"
	response "jogodecartasonline/api/Response"
	"jogodecartasonline/server/game/models"
	"strconv"
	"fmt"
	"net"
	"time"
	"sync"
)


var (
	pingStartTime time.Time
	waitingPong   bool = false
	pingMutex     sync.Mutex
)

type Client struct {
	Nome   string
	Conn   net.Conn
	Reader *bufio.Reader
}

// --------------------- Requisições Basicas do lobby

// requisição para login do usuario no server
func (c *Client) LoginServer(name string) error {
	req := request.Request{
		User:   name,
		Method: "addUser",
		Params: map[string]string{
			"nome": name,
		},
	}
	return c.SendRequest(req)
}

// requisição para sair do servidor
func (c *Client) LeaveServer(username string) error {
	req := request.Request{
		User:   username,
		Method: "DeletePlayer",
		Params: map[string]string{
			"nome": username,
		},
	}
	return c.SendRequest(req)
}

// requisição para achar uma partida
func (c *Client) FoundMatch(player *models.Player) error {
	playerJson, _ := json.Marshal(player)

	req := request.Request{
		User:   player.Nome,
		Method: "TryMatch",
		Params: map[string]string{
			"player": string(playerJson),
		},
	}
	return c.SendRequest(req)
}

func (c *Client) CheckResponseTime() error {
    pingMutex.Lock()
    defer pingMutex.Unlock()
    
    req := request.Request{
        User:   c.Nome,
        Method: "SendUserPing",
        Params: map[string]string{
            "timestamp": fmt.Sprintf("%d", time.Now().UnixNano()),
        },
    }
    
    fmt.Println("📡 Enviando ping para o servidor...")
    
    pingStartTime = time.Now()
    waitingPong = true
    
    err := c.SendRequest(req)
    if err != nil {
        fmt.Printf("❌ Erro ao enviar ping: %v\n", err)
        waitingPong = false
        return err
    }
    
   
    return nil
}


// --------------------- Requisições para os pacotes

// requisições para verificar status do pacote
func (c *Client) CheckPackStatus(username string) error {
	req := request.Request{
		User:   username,
		Method: "PackStatus",
		Params: map[string]string{},
	}
	return c.SendRequest(req)

}


// requisições de abrir pacote
func (c *Client) OpenPack(username string) error {
	req := request.Request{
		User:   username,
		Method: "OpenPack",
		Params: map[string]string{},
	}
	return c.SendRequest(req)
}

// requisição para listar as cartas 
func (c *Client) ListCards(player *models.Player) error{
	req := request.Request{
		User:   player.Nome,
		Method: "ListCards",
		Params: map[string]string{
			"ID": player.ID,
		},
	}
	return c.SendRequest(req)
}

// requisição para selecionar cartas do deck de batalha
func (c *Client) ChangeDeckCard(oldCardIndex, newCardIndex int) error {
	req := request.Request{
		Method: "SelectMatchDeck",
		User:   c.Nome,
		Params: map[string]string{
			"oldCardIndex": strconv.Itoa(oldCardIndex),
			"newCardIndex": strconv.Itoa(newCardIndex),
		},
	}
	return c.SendRequest(req)
	
}




//----------------- Requisições de um match

// requisição para escolher carta
func (c *Client) ChooseCard(player *models.Player, cardIndex int) error {
	req := request.Request{
		User:   player.Nome,
		Method: "ProcessGameAction",
		Params: map[string]string{
			"action":    "chooseCard",
			"cardIndex": fmt.Sprintf("%d", cardIndex),
		},
	}
	return c.SendRequest(req)
}

// requisição para atacar
func (c *Client) Attack(player *models.Player) error {
	req := request.Request{
		User:   player.Nome,
		Method: "ProcessGameAction",
		Params: map[string]string{
			"action": "attack",
		},
	}
	return c.SendRequest(req)
}

// requisição para sair da partida
func (c *Client) LeaveMatch(player *models.Player) error {
	req := request.Request{
		User:   player.Nome,
		Method: "ProcessGameAction",
		Params: map[string]string{
			"action": "leaveMatch",
		},
	}
	return c.SendRequest(req)
}

// -----  METODOS AUXILIARES ------
func (c *Client) SendRequest(req request.Request) error {
	dados, err := json.Marshal(req)
	if err != nil {
		return err
	}

	message := append(dados, '\n')

	_, err = c.Conn.Write(message)
	return err
}

// Recebe resposta do servidor
func (c *Client) ReceiveResponse() (response.Response, error) {

	line, err := c.Reader.ReadBytes('\n')
	if err != nil {
		return response.Response{}, err
	}

	line = line[:len(line)-1]

	var resp response.Response
	err = json.Unmarshal(line, &resp)
	if err != nil {
		fmt.Printf(" Erro ao unmarshall: %v\n", err)
		return response.Response{}, err
	}

	return resp, nil
}

func DecodePlayer(data interface{}) (*models.Player, error) {
	playerJSON, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("dados do player não são uma string válida")
	}

	var player models.Player
	err := json.Unmarshal([]byte(playerJSON), &player)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar player: %w", err)
	}

	return &player, nil
}

func (c *Client) HandlePongSimple(resp response.Response) {
    pingMutex.Lock()
    defer pingMutex.Unlock()
    
    if waitingPong {
        pingDuration := time.Since(pingStartTime)
        fmt.Printf("\n📡 ✅ Tempo de resposta: %.2f ms\n", 
            float64(pingDuration.Nanoseconds())/1000000.0)
        waitingPong = false
    } else {
        fmt.Println("\n📡 ⚠️ Pong recebido sem ping correspondente")
    }
}
