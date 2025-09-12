package main

import (
	"encoding/json"
	"fmt"
	response "jogodecartasonline/api/Response"
	model "jogodecartasonline/client/routes"
	"jogodecartasonline/server/game/models"

	"time"
)

// Envia a requisição de status do pacote
func EnterPackageSystem(client *model.Client, player *models.Player) {
	fmt.Println("📦 Verificando status dos pacotes...")
	SetPackageMenuActive(true)
	SetWaitingForPackage(false)

	err := client.CheckPackStatus(player.Nome)
	if err != nil {
		fmt.Printf("❌ Erro ao verificar pacotes: %v\n", err)
		time.Sleep(2 * time.Second)
		SetPackageMenuActive(false)
		return
	}

}

// Envia a requisição para abertura de pacotes
func openPackage(client *model.Client) {

	// Define como aguardando antes de enviar a requisição
	SetWaitingForPackage(true)

	err := client.OpenPack(client.Nome)

	if err != nil {
		fmt.Printf("❌ Erro ao abrir pacote: %v\n", err)
		SetWaitingForPackage(false) 

		return
	}

	fmt.Println("⏳ Aguardando resposta do servidor...")

}

// Envia a requisição para abertura de pacotes
func ListCards(client *model.Client, player *models.Player) {

	if player == nil {
		fmt.Println("❌ Erro: Player não encontrado")
		return
	}

	err := client.ListCards(player)
	if err != nil {
		fmt.Printf("❌ Erro listar as cartas %v\n", err)
		return
	}

}

// Processa a requisição para status do pacote
func ProcessPackageStatus(resp response.Response, client *model.Client) {
	canOpen := resp.Data["canOpen"] == "true"
	remaining := resp.Data["remaining"]
	totalCards := resp.Data["totalCards"]

	var player *models.Player
	var err error

	if playerData, exists := resp.Data["player"]; exists && playerData != "" {
		player, err = model.DecodePlayer(playerData)
		if err != nil {
			fmt.Printf("❌ Erro ao decodificar player: %v\n", err)
			SetPackageMenuActive(false)
			return
		}
	} else {
		fmt.Println("❌ Dados do player não encontrados na resposta")
		SetPackageMenuActive(false)
		return
	}

	Menu.ClearScreen()
	if canOpen {
		PackageMenu(client, totalCards, player)
	} else {
		CooldownMenu(client, totalCards, remaining, player)
	}
}

// Processa a requisição da abertura de pacotes
func ProcessPackageOpened(resp response.Response, client *model.Client) {
	cardsStr := resp.Data["cards"]
	totalCards := resp.Data["totalCards"]

	var newCards []models.Card
	if err := json.Unmarshal([]byte(cardsStr), &newCards); err != nil {
		fmt.Printf("Erro ao decodificar cartas: %v\n", err)
		SetWaitingForPackage(false)
		return
	}

	SetWaitingForPackage(false)

	showOpenedPackage(newCards, totalCards)
}

// Processa a requisição de listagem das cartas de um jogador
func ProcessListCards(resp response.Response, client *model.Client) {
	DeckCards, err1 := models.DecodeCards(resp.Data["deckCards"])
	OtherCards, err2 := models.DecodeCards(resp.Data["otherCards"])

	if err1 != nil || err2 != nil {
		fmt.Printf("❌ Erro ao decodificar cartas: deck=%v, other=%v\n", err1, err2)
		return
	}

	Menu.ShowListCards(DeckCards, OtherCards)

	fmt.Println("\nPressione ENTER para voltar...")
	inputManager.ReadString()
	SetPackageMenuActive(false)
}

// --------------------- Auxliliares

func showOpenedPackage(newCards []models.Card, totalCards string) {
	Menu.ClearScreen()
	Menu.ShowOpenPackResult(totalCards)

	for i, card := range newCards {
		rarity := Menu.GetRarityEmoji(card.Rarity)
		fmt.Printf("  %d. %s %s (⚔️%d 💚%d)\n",
			i+1, rarity, card.Nome, card.Power, card.Health)
	}

	fmt.Println("Aperte qualquer tecla para voltar ao lobby")

	inputManager.ReadString()
	SetPackageMenuActive(false) // Sai do sistema de pacotes

}

func PackageMenu(client *model.Client, totalCards string, player *models.Player) {

	for IsPackageMenuActive() {
	
		if IsWaitingForPackage() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		Menu.ClearScreen()
		Menu.ShowPackageMenu(totalCards)

		fmt.Print("\nEscolha: ")
		opcao, err := inputManager.ReadInt()
		if err != nil {
			fmt.Println("⚠️ Entrada inválida!")
			time.Sleep(1 * time.Second)
			continue
		}

		switch opcao {
		case 1:
			openPackage(client)
			return
		case 2:
			fmt.Println("📋 Listando cartas... ")
			ListCards(client, player)
			return

		case 3:
			fmt.Println("🔧 Gerenciando deck... ")
			time.Sleep(2 * time.Second)
			SetPackageMenuActive(false)

		case 4:
			SetPackageMenuActive(false)

		default:
			fmt.Println("⚠️ Opção inválida!")
			time.Sleep(1 * time.Second)
		}
	}
}

func CooldownMenu(client *model.Client, totalCards, remaining string, player *models.Player) {

	for IsPackageMenuActive() {
		Menu.ClearScreen()
		Menu.ShowCooldownMessage(totalCards, remaining)

		fmt.Print("\nEscolha: ")
		opcao, err := inputManager.ReadInt()
		if err != nil {
			fmt.Println("⚠️ Entrada inválida!")
			time.Sleep(1 * time.Second)
			continue
		}

		switch opcao {
		case 1:
			fmt.Println("📋 Listando cartas... ")
			ListCards(client, player)
			return

		case 2:
			fmt.Println("🔧 Gerenciando deck... ")
			time.Sleep(2 * time.Second)
			SetPackageMenuActive(false)

		case 3:
			SetPackageMenuActive(false)

		default:
			fmt.Println("⚠️ Opção inválida!")
			time.Sleep(1 * time.Second)
		}
	}
}
