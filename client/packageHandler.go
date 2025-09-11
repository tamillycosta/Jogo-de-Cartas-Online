package main

import (
	"encoding/json"
	"fmt"
	response "jogodecartasonline/api/Response"
	"jogodecartasonline/client/model"
	"jogodecartasonline/server/game/models"

	"time"
)


// Envia a requisição de status do pacote
func EnterPackageSystem(client *model.Client, username string) {
	fmt.Println("📦 Verificando status dos pacotes...")
	SetPackageMenuActive(true)
	SetWaitingForPackage(false)


	err := client.CheckPackStatus(username)
	if err != nil {
		fmt.Printf("❌ Erro ao verificar pacotes: %v\n", err)
		time.Sleep(2 * time.Second)
		SetPackageMenuActive(false)
		return
	}

	
}


// Envia a requisição para abertura de pacotes
func openPackage(client *model.Client) {
	fmt.Println("📦 Enviando requisição...")

	// Define como aguardando ANTES de enviar a requisição
	SetWaitingForPackage(true)

	err := client.OpenPack(client.Nome)

	if err != nil {
		fmt.Printf("❌ Erro ao abrir pacote: %v\n", err)
		SetWaitingForPackage(false) // Libera em caso de erro
		// Não sai do sistema - permite tentar novamente
		return
	}

	fmt.Println("⏳ Aguardando resposta do servidor...")
	// Não precisa definir WaitingForPackage aqui novamente
}


// Processa a requisição para status do pacote
func ProcessPackageStatus(resp response.Response, client *model.Client) {
	canOpen := resp.Data["canOpen"] == "true"
	remaining := resp.Data["remaining"]
	totalCards := resp.Data["totalCards"]

	Menu.ClearScreen()
	if canOpen {
		PackageMenu(client, totalCards)
	} else {
		CooldownMenu(client, totalCards, remaining)
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



func PackageMenu(client *model.Client, totalCards string) {

	for IsPackageMenuActive() {
		// Se está aguardando resposta do servidor, mostra status e continua aguardando
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
			time.Sleep(2 * time.Second)
			SetPackageMenuActive(false)

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

func CooldownMenu(client *model.Client, totalCards, remaining string) {

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
			time.Sleep(2 * time.Second)
			SetPackageMenuActive(false) 

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
