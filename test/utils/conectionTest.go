package utils

import (
    "bufio"
    "encoding/json"
    "fmt"
    "net"
 
    "testing"
    

    request "jogodecartasonline/api/Request"
    response "jogodecartasonline/api/Response"
)


type FakeClient struct {
    Conn   net.Conn
    Reader *bufio.Reader
    Name   string
}

// Função para enviar request e aguardar response
func (fc *FakeClient) SendRequest(req request.Request) (response.Response, error) {
    data, err := json.Marshal(req)
    if err != nil {
        return response.Response{}, err
    }
    
    // Envia com delimitador
    message := append(data, '\n')
    _, err = fc.Conn.Write(message)
    if err != nil {
        return response.Response{}, err
    }
    
    // Lê resposta
    line, err := fc.Reader.ReadBytes('\n')
    if err != nil {
        return response.Response{}, err
    }
    
    // Remove delimitador
    line = line[:len(line)-1]
    
    var resp response.Response
    err = json.Unmarshal(line, &resp)
    return resp, err
}

// Cria cliente fake e faz login
func NewFakeClient(t *testing.T, name string) (*FakeClient, error) {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        return nil, fmt.Errorf("erro ao conectar %s: %v", name, err)
    }

    client := &FakeClient{
        Conn:   conn,
        Reader: bufio.NewReader(conn),
        Name:   name,
    }

    // Faz login
    loginReq := request.Request{
        Method: "addUser",
        Params: map[string]string{"nome": name},
    }
    
    loginResp, err := client.SendRequest(loginReq)
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("erro no login de %s: %v", name, err)
    }
    
    if loginResp.Status != 200 {
        conn.Close()
        return nil, fmt.Errorf("login falhou para %s: %s", name, loginResp.Message)
    }
    
    fmt.Printf("✅ %s logado com sucesso\n", name)
    return client, nil
}



func (fc *FakeClient) FindMatch(t *testing.T) error {
    // Cria dados do player para a requisição
    playerData := fmt.Sprintf(`{"nome":"%s"}`, fc.Name)
    
    matchReq := request.Request{
        Method: "TryMatch",
        User:   fc.Name,
        Params: map[string]string{
            "player": playerData,
        },
    }
    
    matchResp, err := fc.SendRequest(matchReq)
    if err != nil {
        return fmt.Errorf("erro ao buscar partida para %s: %v", fc.Name, err)
    }
    
    if matchResp.Status == 200 {
        if matchResp.Data["matchId"] != "" {
            fmt.Printf("🎉 %s encontrou partida! ID: %s\n", fc.Name, matchResp.Data["matchId"])
        } else {
            fmt.Printf("⏳ %s na fila (posição: %s)\n", fc.Name, matchResp.Data["posicao"])
        }
    } else {
        return fmt.Errorf("erro na busca para %s: %s", fc.Name, matchResp.Message)
    }
    
    return nil
}