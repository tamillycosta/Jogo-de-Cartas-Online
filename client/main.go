package main

import (
	"bufio"
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
            Conn:   conn,
            Reader: bufio.NewReader(conn),
            Nome: nome,
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

		fmt.Print(resp.Message)
		player, err := model.DecodePlayer(resp.Data["player"])

		if err != nil {
			fmt.Println("Erro:", err)
			return
		}

		// goritine para escutar o servidor

		go func() {
			for {
				resp, err := client.ReceiveResponse()
				if err != nil {
					fmt.Println("Erro ao receber resposta:", err)
					return
				}

				fmt.Printf("📩 Nova resposta do servidor: %+v\n", resp.Message)
			}
		}()

		for {

			menu.ShowLobbyMenu()
			fmt.Scanln(&opcao)

			if opcao == 1 {
				client.FoundMatch(player)

			}

			if opcao == 2 {
				//client.Exit()

				return
			}
		}

	}
}
