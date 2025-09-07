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

var Menu screm.Screm

// Estados do matchmaking
type MatchmakingState struct {
	IsSearching bool
	InGame      bool
	CurrentTurn string
	GameState   map[string]interface{}
}

var matchState = &MatchmakingState{}

// Channels para comunicação
var (
	loginResponse     = make(chan response.Response, 1)
	matchFoundChannel = make(chan response.Response, 1)
	gameNotifications = make(chan response.Response, 10)
)

// Configuração para debug
const DEBUG_MODE = false // Mude para true se quiser ver debug

// Estrutura para dados da ação decodificados
type GameAction struct {
	Success        bool                   `json:"success"`
	Action         string                 `json:"action"`
	PlayerResult   map[string]interface{} `json:"playerResult"`
	OpponentResult map[string]interface{} `json:"opponentResult"`
	GameState      map[string]interface{} `json:"gameState"`
	GameEnded      bool                   `json:"gameEnded"`
	Message        string                 `json:"message"`
}

func resetGameState() {
	matchState.GameState = nil
	matchState.CurrentTurn = ""

}

// Decodifica os responses do servidor
func handleServerMessages(client *model.Client) {
	isFirstMessage := true

	for {
		resp, err := client.ReceiveResponse()
		if err != nil {
			fmt.Printf("❌ Erro ao receber: %v\n", err)
			return
		}

		// Verifica se é ping do servidor
		if resp.Data["type"] == "PING" {
			continue
		}

		if isFirstMessage {
			loginResponse <- resp
			isFirstMessage = false
			continue
		}

		messageType := resp.Data["type"]

		switch messageType {
		case "MATCH_FOUND":
			fmt.Println("\n🎉 PARTIDA ENCONTRADA!")
			matchState.IsSearching = false
			matchState.InGame = true
			matchFoundChannel <- resp

		case "GAME_ENDED":

			fmt.Println("\n🏆 JOGO FINALIZADO!")

			result := resp.Data["result"]
			winner := resp.Data["winner"]
			reason := resp.Data["reason"]

			if result == "WIN" {
				fmt.Printf("🎉 VOCÊ VENCEU! 🎉\n")
				if reason == "leaveMatch" {
					fmt.Println("Oponente desistiu da partida")
				} else {
					fmt.Println("Parabéns pela vitória!")
				}
			} else {
				fmt.Printf("💀 VOCÊ FOI DERROTADO!\n")
				fmt.Printf("Vencedor: %s\n", winner)
				if reason == "attack" {
					fmt.Println("Você foi derrotado em combate")
				}
			}

			matchState.InGame = false

			// Pequena pausa para ler resultado
			time.Sleep(3 * time.Second)

		case "OPPONENT_ACTION":
			gameNotifications <- resp

		case "GAME_OVER":
			fmt.Println("\n🏆 JOGO FINALIZADO!")
			gameNotifications <- resp

		case "CHECKPACKAGE": 
			fmt.Println("📦 Status do pacote recebido")
			ProcessPackageResponse(resp)

		case "PACKAGE_OPENED":
			fmt.Println("📦 Pacote aberto!")
			ProcessPackageOpenResponse(resp)

		default:
			if resp.Message == "Procurando partida..." {
				fmt.Printf("⏳ Procurando oponente... (Posição: %s)\n", resp.Data["posicao"])

			} else if resp.Data["matchId"] != "" {
				fmt.Println("\n🎉 PARTIDA ENCONTRADA!")
				matchState.IsSearching = false
				matchState.InGame = true
				matchFoundChannel <- resp

			} else if resp.Message == "Ação processada" {
				ProcessPlayerActionResponse(resp)
			} else {
				fmt.Printf("📩 %s\n", resp.Message)
			}
		}
	}
}

func ProcessGameNotifications() {
	for notification := range gameNotifications {
		switch notification.Data["type"] {
		case "OPPONENT_ACTION":
			ProcessOpponentAction(notification)

		case "GAME_OVER":
			fmt.Printf("\n🏆 %s\n", notification.Message)
			matchState.InGame = false

		default:
			fmt.Printf("\n📢 %s\n", notification.Message)
		}
	}
}

// Processa atulizaçaõ do estado do jogador

func main() {
	Menu.ClearScreen()
	Menu.ShowInitalMenu()

	var opcao int
	fmt.Scanln(&opcao)

	if opcao == 1 {
		conn, err := net.Dial("tcp", "localhost:8080")
		if err != nil {
			fmt.Println("Erro ao conectar no servidor:", err)
			return
		}
		defer conn.Close()

		Menu.ClearScreen()
		fmt.Print("Informe seu username: ")
		var nome string
		fmt.Scanln(&nome)

		client := model.Client{
			Conn:   conn,
			Reader: bufio.NewReader(conn),
			Nome:   nome,
		}

		go handleServerMessages(&client)
		go ProcessGameNotifications()

		//Adicionar verificação caso o user ja esteja logado
		fmt.Println("Fazendo login...")
		client.LoginServer(nome)

		var player *models.Player
		select {
		case loginResp := <-loginResponse:
			if loginResp.Status == 200 {
				fmt.Printf("✅ %s\n", loginResp.Message)
				player, err = model.DecodePlayer(loginResp.Data["player"])
				if err != nil {
					fmt.Printf("Erro ao decodificar player: %v\n", err)
					return
				}
				time.Sleep(1 * time.Second)
			} else {
				fmt.Printf("Erro no login: %s\n", loginResp.Message)
				return
			}

		case <-time.After(10 * time.Second):
			fmt.Println("Timeout no login")
			return
		}

		// LOBBY PRINCIPAL
		for {
			if matchState.IsSearching || matchState.InGame {
				time.Sleep(1 * time.Second)
				continue
			}

			Menu.ClearScreen()
			fmt.Printf("Bem-vindo, %s!\n\n", player.Nome)
			Menu.ShowLobbyMenu()
			fmt.Scanln(&opcao)

			switch opcao {
			case 1:
				WaitForMatch(&client, player)

			case 2:
				Menu.ClearScreen()
				TryOpenPackage(player.Nome, &client)

			case 3:
				Menu.ClearScreen()
				fmt.Println("Saindo do jogo...")
				client.LeaveServer(player.Nome)
				return

			default:
				fmt.Println("⚠️ Opção inválida")
				time.Sleep(1 * time.Second)
			}
		}
	}
	if opcao == 2 {
		fmt.Println("Saindo do server...")

		return
	}
}
