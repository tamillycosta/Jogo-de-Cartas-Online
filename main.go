package main

import (
	"jogodecartasonline/api"
	request "jogodecartasonline/api/Request"
	response "jogodecartasonline/api/Response"
	"jogodecartasonline/server/game/models"

	"encoding/json"
	"fmt"
	"io"
	"jogodecartasonline/server/game/config"
	"net"
)

var Server = api.NewAplication()


func SendResponse(conn net.Conn, resp response.Response) error {
    data, err := json.Marshal(resp)
    if err != nil {
        return err
    }
    
    // Adiciona delimiter no final (quebra de linha)
    message := append(data, '\n')
    
    _, err = conn.Write(message)
    if err != nil {
        fmt.Printf("❌ Erro ao enviar resposta: %v\n", err)
        return err
    }
    
    fmt.Printf("📤 Resposta enviada: %s\n", string(data))
    return nil
}


func handleConnection(conn net.Conn) {
    defer conn.Close()
    
    for {
        buffer := make([]byte, 1024)
        n, err := conn.Read(buffer)
        
        if err != nil {
            if err == io.EOF {
                fmt.Printf("📤 Cliente %s desconectou\n", conn.RemoteAddr())
            } else {
                fmt.Printf("❌ Erro ao ler: %v\n", err)
            }
            return
        }

        req, err := request.Deserialize(buffer[:n])
        if err != nil {
            fmt.Printf("❌ Erro ao desserializar: %v\n", err)
            continue
        }

        fmt.Printf("🔄 Processando: %s\n", req.Method)

        resp := Server.Dispatch(req, conn)
        
        // 🎯 USA A NOVA FUNÇÃO COM DELIMITER
        if err := SendResponse(conn, resp); err != nil {
            fmt.Printf("❌ Erro ao enviar: %v\n", err)
            return
        }
    }
}


func main() {
    db := config.CretaeTable()

    lobby := &models.Lobby{
        Players:   make(map[string]*models.Player),
        Matchs:    make(map[string]*models.Match),
        WaitQueue: []*models.WaitingPlayer{},
        DB:        &db,
    }
    
    go lobby.PrintStats()
    
    // registra as rotas
    Server.AddRoute("addUser", lobby.AddPlayer)
    Server.AddRoute("TryMatch", lobby.TryMatchUsers)

    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        panic(err)
    }
    defer listener.Close()
    
    fmt.Println("Servidor rodando na porta 8080...")

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("Erro ao aceitar conexão: %v\n", err)
            continue
        }
        
        fmt.Printf("Novo cliente conectado: %s\n", conn.RemoteAddr())
        
        // Cada conexão roda em sua própria goroutine
        go handleConnection(conn)
    }
}