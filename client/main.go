package main

import (
	"bufio"
	"fmt"
	response "jogodecartasonline/api/Response"
	"jogodecartasonline/client/model"
	"jogodecartasonline/client/screm"
	"jogodecartasonline/server/game/models"

	"net"
	"time"
)

var menu screm.Screm

// chanels de comunicação
var (
	loginResponse    = make(chan response.Response, 1)
	matchResponse    = make(chan response.Response, 1)
	gameNotification = make(chan response.Response, 10) // Buffer maior para notificações
)

// Decodifica as requisições do player
func handleServerMessages(client *model.Client) {
	messageCount := 0

	for {
		resp, err := client.ReceiveResponse()
		if err != nil {
			fmt.Printf("❌ Erro ao receber: %v\n", err)
			return
		}

		messageCount++

		//  ROTEIA BASEADO NO NÚMERO DA MENSAGEM E CONTEÚDO
		if messageCount == 1 {
			// Primeira mensagem sempre é login
			loginResponse <- resp

		} else if resp.Data["matchId"] != "" || resp.Message == "Procurando partida..." {
			// Mensagens relacionadas a match
			matchResponse <- resp

		} else {
			// Outras notificações (jogo, oponente, etc)
			gameNotification <- resp
		}
	}
}

func processGameNotifications() {
	for {
		select {
		case notification := <-gameNotification:
			fmt.Printf("🔔 %s\n", notification.Message)

			if notification.Data["type"] == "GAME_OVER" {
				fmt.Printf("🏆 Jogo terminou!\n")
			}

		case <-time.After(100 * time.Millisecond):
			// Não bloqueia se não há notificações
			continue
		}
	}
}

func gameLoop(client *model.Client, player *models.Player) {
	for {
		menu.ShowGameLoop()

		var opcao int
		fmt.Scanln(&opcao)

		switch opcao {
		case 1:
			fmt.Print("Escolha uma carta (0-4): ")
			var cardIndex int
			fmt.Scanln(&cardIndex)
			client.ChooseCard(player, cardIndex)

		case 2:
			client.Attack(player)

		case 3:
			client.PassTurn(player)

		case 4:
			client.LeaveMatch(player)
			return
		}

		// Aguarda resposta
		time.Sleep(1 * time.Second)
	}
}

func main() {

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
			Nome:   nome,
		}

		// Inicia goroutines
		go handleServerMessages(&client)
		go processGameNotifications()

		// LOGIN com resposta sincronizada
		client.LoginServer(nome)
		var player *models.Player
		select {
		case loginResp := <-loginResponse:
			fmt.Println("✅", loginResp.Message)
			player, _ = model.DecodePlayer(loginResp.Data["player"])

		case <-time.After(10 * time.Second):
			fmt.Println("timeout no login")
			return
		}

		// LOBBY LOOP
		for {
			menu.ShowLobbyMenu()
			fmt.Scanln(&opcao)

			if opcao == 1 {
				// Busca partida
				client.FoundMatch(player)

				// Espera resposta do match
				select {
				case matchResp := <-matchResponse:
					if matchResp.Data["matchId"] != "" {
					
						fmt.Println("PARTIDA ENCONTRADA!")
						gameLoop(&client, player)
					} else {
						// ⏳ AINDA PROCURANDO
						fmt.Printf("⏳ %s\n", matchResp.Message)
					}

				case <-time.After(60 * time.Second):
					fmt.Println("Timeout procurando partida")
				}
			}
		}
	}
}
