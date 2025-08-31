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





// porta d entrada do usuario no server
func (c *Client) LoginServer(nome string) error {
	req := request.Request{
		User:   nome,
		Method: "addUser",
		Params: map[string]string{
			"nome": nome,
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
    
    fmt.Printf("📩 JSON recebido: %s\n", string(line))
    
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
