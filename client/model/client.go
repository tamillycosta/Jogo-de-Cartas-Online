package model

import (
	request "jogodecartasonline/api/Request"
	response "jogodecartasonline/api/Response"
	"jogodecartasonline/server/game/models"
	"bufio"
	"encoding/json"

	"fmt"
	"net"
)

type Client struct {
	Nome string
	Conn net.Conn
	Reader *bufio.Reader
}



// requisição de login do usuario no server
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

func (c *Client) LeaveServer(username string) error{
    req := request.Request{
        User: username,
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

// Player escolhe carta
func (c *Client) ChooseCard(player *models.Player, cardIndex int) error {
    req := request.Request{
        User:   player.Nome,
        Method: "ProcessGameAction",
        Params: map[string]string{
            "action" : "chooseCard",
            "cardIndex": fmt.Sprintf("%d", cardIndex),
            
        },
    }
    return c.SendRequest(req)
}

// Player ataca
func (c *Client) Attack(player *models.Player) error {
    req := request.Request{
        User:   player.Nome,
        Method: "ProcessGameAction",
        Params: map[string]string{
            "action" : "attack",
        },
    }
    return c.SendRequest(req)
}

// Player passa a vez
func (c *Client) PassTurn(player *models.Player) error {
    req := request.Request{
        User:   player.Nome,
        Method: "ProcessGameAction",
        Params: map[string]string{},
    }
    return c.SendRequest(req)
}

// Player sai da partida
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
