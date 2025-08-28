
package main

import (
	"fmt"
	"net"
    "jogodecartasonline/server/chat"
)



func main() {
    conn, err := net.Dial("tcp", "172.16.201.11:8080")
    if err != nil {
        fmt.Println("Não foi possível conectar ao server:", err)
        return
    }
    defer conn.Close()

    fmt.Print("Informe seu nome: ")
    var nome string
    fmt.Scanln(&nome)

    client := &chat.CLient{Nome: nome, Conn: conn}

    // envia nome para o servidor
    conn.Write([]byte(nome))
	
    // rodar envio e recebimento em paralelo
    go client.SendMessage()
    go client.ReciveMessage()

    // mantem o programa rodando
    select {} // bloqueia main indefinidamente
}

