package main

import (
   
    "fmt"
    "jogodecartasonline/client/model"
    "jogodecartasonline/client/screm"
    "net"
)



func main() {
    menu := &screm.Screm{}
    menu.ShowInitalMenu()

    var opcao int
    fmt.Scanln(&opcao)

    if opcao == 1 {
        conn, err := net.Dial("tcp", "localhost:8080")
        if err != nil {
            fmt.Println("Erro ao conectar no servidor:", err)
            return
        }
        defer conn.Close()

        fmt.Print("Informe seu username: ")
        var nome string
        fmt.Scanln(&nome)

        client := model.Client{
            Nome: nome,
            Conn: conn,
        }

        // envia request
        err = client.LoginServer(nome)
        if err != nil {
            fmt.Println("Erro ao enviar request:", err)
            return
        }

        // recebe resposta
        resp, err := client.ReceiveResponse()
        if err != nil {
            fmt.Println("Erro ao receber resposta:", err)
            return
        }

        fmt.Println("Resposta do servidor:", resp.Message, resp.Data)
    }
}
