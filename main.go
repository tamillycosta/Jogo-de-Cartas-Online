package main

import (
	"jogodecartasonline/api"
	request "jogodecartasonline/api/Request"
	response "jogodecartasonline/api/Response"
	"jogodecartasonline/server/game/models"
    "time"
    "bufio"
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
    
  
    message := append(data, '\n')
    
    _, err = conn.Write(message)
    if err != nil {
        fmt.Printf("Erro ao enviar resposta: %v\n", err)
        return err
    }
    
    fmt.Printf("📤 Resposta enviada: %s\n", string( data))
    return nil
}


func handleConnection(conn net.Conn, monitor *models.ConnectionMonitor) {
    defer func() {
        conn.Close()
        fmt.Printf("Conexão fechada: %s\n", conn.RemoteAddr())
    }()
    
    reader := bufio.NewReader(conn)
    var currentPlayer string

    for {
        line, err := reader.ReadBytes('\n')
        if err != nil {
            if err == io.EOF {
                fmt.Printf("Cliente %s desconectou (EOF)\n", conn.RemoteAddr())
            } else {
                fmt.Printf("Erro ao ler de %s: %v\n", conn.RemoteAddr(), err)
            }
            
            // Se temos player identificado, processa desconexão
            if currentPlayer != "" {
                monitor.CheckPlayerNow(currentPlayer)
            }
            return
        }

        // Remove delimitador
        line = line[:len(line)-1]

        // Ignora mensagens de ping do monitor
        if string(line) == "PONG" {
            if currentPlayer != "" {
                monitor.RegisterPlayerActivity(currentPlayer)
            }
            continue
        }

        req, err := request.Deserialize(line)
        if err != nil {
            fmt.Printf("Erro ao deserializar de %s: %v\n", conn.RemoteAddr(), err)
            continue
        }

        // Registra atividade do player
        if req.User != "" {
            currentPlayer = req.User
            monitor.RegisterPlayerActivity(req.User)
        }

        // Processa requisição
        resp := Server.Dispatch(req, conn)
        
        if resp.Status != 0 || resp.Message != "" {
            data, _ := json.Marshal(resp)
            message := append(data, '\n')
            
            // Tenta enviar resposta com timeout
            conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            _, err = conn.Write(message)
            conn.SetWriteDeadline(time.Time{})
            
            if err != nil {
                fmt.Printf("Erro ao enviar resposta para %s: %v\n", conn.RemoteAddr(), err)
                if currentPlayer != "" {
                    monitor.CheckPlayerNow(currentPlayer)
                }
                return
            }
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
    

    // Inicializa sistema de monitoramento
    connectionMonitor := models.NewConnectionMonitor(lobby)
    connectionMonitor.Start()

    go lobby.PrintStats()
    
    // registra as rotas
    Server.AddRoute("addUser", lobby.AddPlayer)
    Server.AddRoute("TryMatch", lobby.TryMatchUsers)
    Server.AddRoute("ProcessGameAction", lobby.ProcessGameAction)
    Server.AddRoute("DeletePlayer", lobby.DeletePlayer)
    Server.AddRoute("ConnectionStats", lobby.GetConnectionStats)
    Server.AddRoute("OpenPack", lobby.OpenCardPack)       
    Server.AddRoute("PackStatus", lobby.CheckPackStatus)   
    Server.AddRoute("ListCards", lobby.ListCards)
    Server.AddRoute("SelectMatchDeck", lobby.SelectMatchDeck)
    Server.AddRoute("SendUserPing", lobby.SendUserPing)

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
        go handleConnection(conn, connectionMonitor)
    }
}

