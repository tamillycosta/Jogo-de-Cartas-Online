package main

import (
	"encoding/json"
	"fmt"
	response "jogodecartasonline/api/Response"
	"jogodecartasonline/client/routes"
	"jogodecartasonline/server/game/models"
	"time"
	
)





func ProcessPlayerActionResponse(resp response.Response) {
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
		matchState.mu.Lock()
		matchState.GameState = gameAction.GameState
		matchState.mu.Unlock()
	}


	// Processa resultado da própria ação
	switch gameAction.Action {
	case "chooseCard":
		if playerResult := gameAction.PlayerResult; playerResult != nil {
			Menu.ShowPlayerResultCard(playerResult)
		}

	case "attack":
		if playerResult := gameAction.PlayerResult; playerResult != nil {
			Menu.ShowPlayerResultAtack(playerResult)
			time.Sleep(3 * time.Second)
			if result, ok := playerResult["result"].(string); ok {
				if result == "WIN" {
					Menu.ShowplayerGameEnd(playerResult)
					
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
func ProcessOpponentAction(notification response.Response) {
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
		matchState.mu.Lock()
		matchState.GameState = gameAction.GameState
		matchState.mu.Unlock()
	}


	// Processa ações que continuam o jogo
	switch gameAction.Action {
	case "chooseCard":
		if opponentResult := gameAction.OpponentResult; opponentResult != nil {
			Menu.ShowOpponentResultCard(opponentResult)
		}

	case "attack":
		if opponentResult := gameAction.OpponentResult; opponentResult != nil {
		Menu.ShowOpponentResultAtack(opponentResult)
		}
	}

	if gameAction.GameEnded {
		matchState.InGame = false
		time.Sleep(3 * time.Second)
	}
}


// Apresenta ao jogadores o estado dos rouds
func ShowGameStatus(myName string) {
	matchState.mu.RLock()
	defer matchState.mu.RUnlock()
	
	if matchState.InGame && matchState.GameState != nil {
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


// Loop da partidas
func gameLoop(client *model.Client, player *models.Player, myName string) {
	Menu.ClearScreen()
	fmt.Println("🎮 === INICIANDO JOGO ===")

	resetGameState()

	for {
		// Verifica se ainda está no jogo
		matchState.mu.RLock()
		inGame := matchState.InGame
		matchState.mu.RUnlock()
		
		if !inGame {
			break
		}

		ShowGameStatus(myName)
		Menu.ShowGameLoop()

		opcao, err := inputManager.ReadInt()
		if err != nil {
			
			time.Sleep(1 * time.Second)
			continue
		}

		switch opcao {
		case 1:
			fmt.Print("Escolha uma carta (0-2): ")
			cardIndex, err := inputManager.ReadInt()
			if err != nil {
				fmt.Println("⚠️ Índice inválido!")
				time.Sleep(1 * time.Second)
				continue
			}
			fmt.Printf("⏳ Escolhendo carta %d...\n", cardIndex)
			client.ChooseCard(player, cardIndex)

		case 2:
			fmt.Printf("⏳ Atacando...\n")
			client.Attack(player)

		case 3:
			fmt.Printf("👋 Saindo da partida...\n")
			client.LeaveMatch(player)
			matchState.SetInGame(false)
		
		default:
			fmt.Println("⚠️ Opção inválida!")
		}

		time.Sleep(1 * time.Second)
		Menu.ClearScreen()
	}

	fmt.Println("\n🔙 Voltando ao lobby...")
	resetGameState()
	time.Sleep(2 * time.Second)
}

func WaitForMatch(client *model.Client, player *models.Player) {
	Menu.ClearScreen()
	fmt.Println("Entrando na fila de matchmaking...")

	err := client.FoundMatch(player)
	if err != nil {
		fmt.Printf("Erro ao entrar na fila: %v\n", err)
		matchState.SetSearching(false)
		return
	}

	matchState.SetSearching(true)
	fmt.Println("⏳ Aguardando oponente...")

	// Contador visual
	dots := ""
	for {
		matchState.mu.RLock()
		searching := matchState.IsSearching
		matchState.mu.RUnlock()
		
		if !searching {
			break
		}

		select {
		case matchResp := <-matchFoundChannel:
			Menu.ClearScreen()
			Menu.ShowFoundMatchMake(matchResp)

			if matchResp.Data["yourTurn"] == "true" {
				fmt.Println("▶️ Você começa!")
			} else {
				fmt.Println("⏸️ Aguarde sua vez")
			}

			time.Sleep(2 * time.Second)
			gameLoop(client, player, player.Nome)
			return

		case <-time.After(1 * time.Second):
			matchState.mu.RLock()
			stillSearching := matchState.IsSearching
			matchState.mu.RUnlock()
			
			if stillSearching {
				dots += "."
				if len(dots) > 3 {
					dots = ""
				}
				fmt.Printf("\r⏳ Procurando oponente%s   ", dots)
			}
		}
	}
}
