package utils


import (
    "bufio"
    "encoding/json"
    "fmt"
    "net"
   
    "testing"
  
	"sync"
    request "jogodecartasonline/api/Request"
    response "jogodecartasonline/api/Response"
)



// Estrutura para estatísticas dos pacotes 


type PackageStats struct {
    TotalPacksOpened int
    CardsByRarity    map[string]int
    TotalCards       int
    Players          int
    Errors           int
    mu               sync.Mutex
}

func (ps *PackageStats) AddCard(rarity string) {
    ps.mu.Lock()
    defer ps.mu.Unlock()
    ps.CardsByRarity[rarity]++
    ps.TotalCards++
}

func (ps *PackageStats) AddPack() {
    ps.mu.Lock()
    defer ps.mu.Unlock()
    ps.TotalPacksOpened++
}

func (ps *PackageStats) AddPlayer() {
    ps.mu.Lock()
    defer ps.mu.Unlock()
    ps.Players++
}

func (ps *PackageStats) AddError() {
    ps.mu.Lock()
    defer ps.mu.Unlock()
    ps.Errors++
}

func (ps *PackageStats) GetStats() (int, map[string]int, int, int, int) {
    ps.mu.Lock()
    defer ps.mu.Unlock()
    return ps.TotalPacksOpened, ps.CardsByRarity, ps.TotalCards, ps.Players, ps.Errors
}



type PackageTestClient struct {
    Conn   net.Conn
    Reader *bufio.Reader
    Name   string
}

func (ptc *PackageTestClient) SendRequest(req request.Request) (response.Response, error) {
    data, err := json.Marshal(req)
    if err != nil {
        return response.Response{}, err
    }
    
    message := append(data, '\n')
    _, err = ptc.Conn.Write(message)
    if err != nil {
        return response.Response{}, err
    }
    
    line, err := ptc.Reader.ReadBytes('\n')
    if err != nil {
        return response.Response{}, err
    }
    
    line = line[:len(line)-1]
    
    var resp response.Response
    err = json.Unmarshal(line, &resp)
    return resp, err
}

func NewPackageTestClient(t *testing.T, name string) (*PackageTestClient, error) {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        return nil, fmt.Errorf("erro ao conectar %s: %v", name, err)
    }

    client := &PackageTestClient{
        Conn:   conn,
        Reader: bufio.NewReader(conn),
        Name:   name,
    }

    // Login
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
    
    return client, nil
}

func (ptc *PackageTestClient) CheckPackageStatus() (response.Response, error) {
    req := request.Request{
        Method: "PackStatus",
        User:   ptc.Name,
        Params: map[string]string{},
    }
    
    return ptc.SendRequest(req)
}

func (ptc *PackageTestClient) OpenPackage() (response.Response, error) {
    req := request.Request{
        Method: "OpenPack",
        User:   ptc.Name,
        Params: map[string]string{},
    }
    
    return ptc.SendRequest(req)
}
