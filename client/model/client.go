package model

import (
	response "jogodecartasonline/api/Response"
	request "jogodecartasonline/api/Request"
	"net"
	 "encoding/json"
)

type Client struct {
	Nome string
	Conn net.Conn
}


func (c *Client) LoginServer(nome string) error{
	req := request.Request{
		Method: "addUser",
		Params: map[string]string{
			"nome": nome,
		},
	}
	return c.SendRequest(req)
}




func (c *Client) SendRequest ( req request.Request) error {
	dados, err := json.Marshal(req)

	if err != nil{
		return err
	}
	_, err = c.Conn.Write(dados)
	return  err
}



// Recebe resposta do servidor
func (c *Client) ReceiveResponse() (response.Response, error) {
    buffer := make([]byte, 1024)
    n, err := c.Conn.Read(buffer)
    if err != nil {
        return response.Response{}, err
    }

    var resp response.Response
    err = json.Unmarshal(buffer[:n], &resp)
    return resp, err
}
