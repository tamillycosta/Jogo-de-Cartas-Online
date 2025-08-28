package main

import (
	"fmt"
	"log"
	"net"
	"jogodecartasonline/server/chat"
)


func handleConnection(cht *chat.Chat ,conn net.Conn){
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
	
		data, err := conn.Read(buffer) // preenche o buffer com os dados do client 

		if(err != nil){
			log.Printf("Conexão encerrada")
			return
		}

		nome := string(buffer[:data])
		client := &chat.CLient{Nome: nome, Conn: conn}
		cht.AcceptClients(client)
		
		conn.Write([]byte("Bem-vindo ao chat, " + nome + "!\n"))
	}

}


func main() {
	chat := &chat.Chat{} 
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
        go handleConnection(chat,conn)
    }
}