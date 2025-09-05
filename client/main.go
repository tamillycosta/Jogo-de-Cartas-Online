package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	response "jogodecartasonline/api/Response"
	"jogodecartasonline/client/model"
	"jogodecartasonline/client/screm"
	"jogodecartasonline/server/game/models"
	"net"

	"time"
)

var menu screm.Screm

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

//Decodifica os responses do servidor
func handleServerMessages(client *model.Client) {
	isFirstMessage := true

	for {
		resp, err := client.ReceiveResponse()
		if err != nil {
			fmt.Printf("❌ Erro ao receber: %v\n", err)
			return
		}

		if DEBUG_MODE {
			fmt.Printf("DEBUG - Recebido: Status=%d, Message='%s', Type='%s'\n",
				resp.Status, resp.Message, resp.Data["type"])
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

		default:
			if resp.Message == "Procurando partida..." {
				fmt.Printf("⏳ Procurando oponente... (Posição: %s)\n", resp.Data["posicao"])

			} else if resp.Data["matchId"] != "" {
				fmt.Println("\n🎉 PARTIDA ENCONTRADA!")
				matchState.IsSearching = false
				matchState.InGame = true
				matchFoundChannel <- resp

			} else if resp.Message == "Ação processada" {
				processPlayerActionResponse(resp)
			} else {
				fmt.Printf("📩 %s\n", resp.Message)
			}
		}
	}
}

// Processa atulizaçaõ do estado do jogador
func processPlayerActionResponse(resp response.Response) {
	resultStr, ok := resp.Data["result"]
	if !ok {
		return
	}

	var gameAction GameAction
	if err := json.Unmarshal([]byte(resultStr), &gameAction); err != nil {
		return
	}

	if !gameAction.Success {
		fmt.Printf("\n❌ %s\n", gameAction.Message)
		return
	}

	// Atualiza estado do jogo
	if gameAction.GameState != nil {
		matchState.GameState = gameAction.GameState
	}

	// Processa resultado da própria ação
	switch gameAction.Action {
	case "chooseCard":
		if playerResult := gameAction.PlayerResult; playerResult != nil {
			menu.ShowPlayerResultCard(playerResult)
		}

	case "attack":
		if playerResult := gameAction.PlayerResult; playerResult != nil {
			menu.ShowPlayerResultAtack(playerResult)
			if result, ok := playerResult["result"].(string); ok {
				if result == "WIN" {
					fmt.Println("\n🏆 VOCÊ VENCEU! Parabéns!")
				} else {
					fmt.Println("   Aguarde a vez do oponente...")
				}
			}
		}

	case "leaveMatch":
		if playerResult := gameAction.PlayerResult; playerResult != nil {
			fmt.Println("\n👋 Você saiu da partida")
		}
	}

	if gameAction.GameEnded {
		matchState.InGame = false
		time.Sleep(3 * time.Second)
	}
}

// Processa atulizaçaõ do estado do jogador receptor
func processOpponentAction(notification response.Response) {
	actionResultStr, ok := notification.Data["actionResult"]
	if !ok {
		return
	}

	var gameAction GameAction
	if err := json.Unmarshal([]byte(actionResultStr), &gameAction); err != nil {
		return
	}

	if gameAction.GameEnded {
		return 
	}

	// Atualiza estado do jogo
	if gameAction.GameState != nil {
		matchState.GameState = gameAction.GameState
	}

	// Processa ações que continuam o jogo
	switch gameAction.Action {
	case "chooseCard":
		if opponentResult := gameAction.OpponentResult; opponentResult != nil {
			menu.ShowOpponentResultCard(opponentResult)
		}

	case "attack":
		if opponentResult := gameAction.OpponentResult; opponentResult != nil {
			menu.ShowOpponentResultAtack(opponentResult)
		}
	}

	if gameAction.GameEnded {
		matchState.InGame = false
		time.Sleep(3 * time.Second)
	}
}
// Apresenta ao jogadores o estado dos rouds 
func showGameStatus(myName string) {
	if  matchState.InGame && matchState.GameState != nil {
		fmt.Println("\n" + "===============================")
		fmt.Printf("🎮 ESTADO DO JOGO\n")
		if currentTurn, ok := matchState.GameState["currentTurn"].(string); ok {
			if currentTurn == myName {
				fmt.Printf("▶️  SUA VEZ\n")
			} else {
				fmt.Printf("⏸️  Vez do oponente\n")
			}
		}
		if roundId, ok := matchState.GameState["roundId"].(float64); ok {
			fmt.Printf("🔄 Round: %.0f\n", roundId)
		}
		fmt.Println("===============================")
	}
}

func processGameNotifications() {
	for notification := range gameNotifications {
		switch notification.Data["type"] {
		case "OPPONENT_ACTION":
			processOpponentAction(notification)

		case "GAME_OVER":
			fmt.Printf("\n🏆 %s\n", notification.Message)
			matchState.InGame = false

		default:
			fmt.Printf("\n📢 %s\n", notification.Message)
		}
	}
}

// Loop da partidas
func gameLoop(client *model.Client, player *models.Player, myName string) {
	menu.ClearScreen()
	fmt.Println("🎮 === INICIANDO JOGO ===")

	resetGameState()

	for matchState.InGame {
		showGameStatus(myName)
		menu.ShowGameLoop()

		var opcao int
		fmt.Scanln(&opcao)

		switch opcao {
		case 1:
			fmt.Print("Escolha uma carta (0-2): ")
			var cardIndex int
			fmt.Scanln(&cardIndex)
			fmt.Printf("⏳ Escolhendo carta %d...\n", cardIndex)
			client.ChooseCard(player, cardIndex)

		case 2:
			fmt.Printf("⏳ Atacando...\n")
			client.Attack(player)

		case 3:
			fmt.Printf("👋 Saindo da partida...\n")
			client.LeaveMatch(player)
			matchState.InGame = false

		}

		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n🔙 Voltando ao lobby...")
	resetGameState()
	time.Sleep(2 * time.Second)
}


func waitForMatch(client *model.Client, player *models.Player) {
	menu.ClearScreen()
	fmt.Println("Entrando na fila de matchmaking...")

	err := client.FoundMatch(player)
	if err != nil {
		fmt.Printf("Erro ao entrar na fila: %v\n", err)
		matchState.IsSearching = false
		return
	}

	matchState.IsSearching = true
	fmt.Println("⏳ Aguardando oponente...")

	// Contador visual
	dots := ""
	for matchState.IsSearching {
		select {
		case matchResp := <-matchFoundChannel:
			menu.ClearScreen()
			menu.ShowFoundMatchMake(matchResp)

			if matchResp.Data["yourTurn"] == "true" {
				fmt.Println("▶️ Você começa!")
			} else {
				fmt.Println("⏸️ Aguarde sua vez")
			}

			time.Sleep(2 * time.Second)
			gameLoop(client, player, player.Nome)
			return

		case <-time.After(1 * time.Second):
			if matchState.IsSearching {
				dots += "."
				if len(dots) > 3 {
					dots = ""
				}
				fmt.Printf("\r⏳ Procurando oponente%s   ", dots)
			}
		}
	}
}

func main() {
	menu.ClearScreen()
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

		menu.ClearScreen()
		fmt.Print("Informe seu username: ")
		var nome string
		fmt.Scanln(&nome)

		client := model.Client{
			Conn:   conn,
			Reader: bufio.NewReader(conn),
			Nome:   nome,
		}

		go handleServerMessages(&client)
		go processGameNotifications()

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

			menu.ClearScreen()
			fmt.Printf("Bem-vindo, %s!\n\n", player.Nome)
			menu.ShowLobbyMenu()
			fmt.Scanln(&opcao)

			switch opcao {
			case 1:
				waitForMatch(&client, player)

			case 2:
				menu.ClearScreen()
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
