package main


import (
	
   "jogodecartasonline/server/game/models"
    "jogodecartasonline/api"
    request "jogodecartasonline/api/Request"
	 "encoding/json"
	"net"
	"fmt"
    "jogodecartasonline/server/game/config"
)

//

var Server = api.NewAplication()



func handleConnection(conn net.Conn) {


    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    
    if err != nil {
        fmt.Println("Erro ao ler do cliente:", err)
        return
    }


    // desserializa request
    req, err := request.Deserialize(buffer[:n])
    if err != nil {
        fmt.Println("Erro ao desserializar:", err)
        return
    }

    // verifica se existe rota
    resp := Server.Dispatch(req, conn)

    // serializa e envia resposta
    data, _ := json.Marshal(resp)
    conn.Write(data)
}



func main(){

    db := config.CretaeTable()

    lobby := &models.Lobby{
        Players:   make(map[string]*models.Player),
        Matchs:    make(map[string]*models.Match),
        WaitQueue: []*models.Player{},
        DB:        &db,
    }
    
	// registra a rota
    Server.AddRoute("addUser", lobby.AddPlayer)

    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        panic(err)
    }
    defer listener.Close()
    fmt.Println("Servidor rodando na porta 8080...")

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Erro ao aceitar conexão:", err)
            continue
        }
        fmt.Println("Novo cliente conectado:", conn.RemoteAddr()) 
        go handleConnection(conn)
    }
}