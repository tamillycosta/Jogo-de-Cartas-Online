package main

import (
	"encoding/json"
	"fmt"
	response "jogodecartasonline/api/Response"
	model "jogodecartasonline/client/routes"
	"jogodecartasonline/server/game/models"
	"strings"
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


// Envia requisição para trocar carta do deck
func ChangeDeckCard(client *model.Client, oldCardIndex, newCardIndex int) {
	SetWaitingForDeck(true)

	err := client.ChangeDeckCard(oldCardIndex, newCardIndex)
	if err != nil {
		fmt.Printf("❌ Erro ao trocar carta: %v\n", err)
		SetWaitingForDeck(false)
		return
	}

	fmt.Println("⏳ Processando troca...")
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
	deckCards, err1 := models.DecodeCards(resp.Data["deckCards"])
	otherCards, err2 := models.DecodeCards(resp.Data["otherCards"])

	if err1 != nil || err2 != nil {
		fmt.Printf("❌ Erro ao decodificar cartas: deck=%v, other=%v\n", err1, err2)
		return
	}

	// Verifica se está no contexto de gerenciamento de deck
	if IsDeckMenuActive() {
		ProcessListCardsForDeck(resp, client)
		return
	}

	// Contexto normal de visualização
	Menu.ShowListCards(deckCards, otherCards)
	fmt.Println("\nPressione ENTER para voltar...")
	inputManager.ReadString()
	SetPackageMenuActive(false)
}


// Processa a requisição de troca de carta do deck
func ProcessNewDeck(resp response.Response, client *model.Client) {
	SetWaitingForDeck(false)

	Menu.ClearScreen()
	Menu.ShowNewCard(resp.Data)

	fmt.Println("\nPressione ENTER para continuar...")
	inputManager.ReadString()
	
	SetDeckMenuActive(false)
	SetPackageMenuActive(false)
}

// Processa a  requisição de listagem de cartas para gerenciamento de deck
func ProcessListCardsForDeck(resp response.Response, client *model.Client) {
	deckCards, err1 := models.DecodeCards(resp.Data["deckCards"])
	otherCards, err2 := models.DecodeCards(resp.Data["otherCards"])

	if err1 != nil || err2 != nil {
		fmt.Printf("❌ Erro ao decodificar cartas: deck=%v, other=%v\n", err1, err2)
		SetDeckMenuActive(false)
		return
	}

	ShowDeckManagement(client, deckCards, otherCards)
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


func ManageDeck(client *model.Client, player *models.Player) {
	fmt.Println("🔧 Carregando gerenciamento de deck...")
	SetDeckMenuActive(true)

	err := client.ListCards(player)
	if err != nil {
		fmt.Printf("❌ Erro ao carregar cartas: %v\n", err)
		time.Sleep(2 * time.Second)
		SetDeckMenuActive(false)
		return
	}
}


// Seleciona carta para remover e adiconar no deck 
func HandleCardSwapWithDebug(client *model.Client, deckCards, otherCards []*models.Card) bool {

	fmt.Println("\n🔄 ===== TROCAR CARTA =====")
	fmt.Println("Qual carta deseja REMOVER do deck?")
	Menu.ShowListCard(deckCards)
	

	fmt.Printf("\nEscolha (1-%d) ou 0 para cancelar: ", len(deckCards))
	oldChoice, err := inputManager.ReadInt()
	if err != nil || oldChoice < 0 || oldChoice > len(deckCards) {
		fmt.Println("⚠️ Entrada inválida!")
		time.Sleep(1 * time.Second)
		return false
	}

	if oldChoice == 0 {
		return false // Cancelar
	}

	oldCardIndex := oldChoice - 1
	selectedOldCard := deckCards[oldCardIndex]
	
	
	// Seleciona carta para adicionar ao deck
	fmt.Println("\nQual carta deseja ADICIONAR ao deck?")
	Menu.ShowListCard(otherCards)
	
	fmt.Printf("\nEscolha (1-%d) ou 0 para cancelar: ", len(otherCards))
	newChoice, err := inputManager.ReadInt()
	if err != nil || newChoice < 0 || newChoice > len(otherCards) {
		fmt.Println("⚠️ Entrada inválida!")
		time.Sleep(1 * time.Second)
		return false
	}

	if newChoice == 0 {
		return false // Cancelar
	}

	newCardIndex := newChoice - 1
	selectedNewCard := otherCards[newCardIndex]
	
	Menu.ShowConfirmChange(*selectedNewCard,*selectedOldCard,oldCardIndex, newCardIndex)
	
	confirm := inputManager.ReadString()
	if strings.ToLower(strings.TrimSpace(confirm)) != "s" {
		fmt.Println("❌ Troca cancelada!")
		time.Sleep(1 * time.Second)
		return false
	}

	// Executa a troca
	ChangeDeckCard(client, oldCardIndex, newCardIndex)
	return true
}




// --------------------- Menus

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
			ManageDeck(client,player)
			return

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
			ManageDeck(client,player)
			return

		case 3:
			SetPackageMenuActive(false)

		default:
			fmt.Println("⚠️ Opção inválida!")
			time.Sleep(1 * time.Second)
		}
	}
}

// Interface de gerenciamento de deck
func ShowDeckManagement(client *model.Client, deckCards, otherCards []*models.Card) {
	for IsDeckMenuActive() {
		if IsWaitingForDeck() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		Menu.ClearScreen()
		
		fmt.Println("\n════════════════════════════════════════════════")
		fmt.Println("⚔️ GERENCIAR DECK")
		fmt.Println("════════════════════════════════════════════════")

		Menu.ShowListCards(deckCards, otherCards)
	
		Menu.ShowDeckManagementMenu()
		opcao, err := inputManager.ReadInt()
		if err != nil {
			fmt.Println("⚠️ Entrada inválida!")
			time.Sleep(1 * time.Second)
			continue
		}

		switch opcao {
		case 1:
			if len(deckCards) == 0 {
				fmt.Println("❌ Não há cartas no deck para trocar!")
				time.Sleep(2 * time.Second)
				continue
			}
			if len(otherCards) == 0 {
				fmt.Println("❌ Não há cartas disponíveis para adicionar!")
				time.Sleep(2 * time.Second)
				continue
			}

			if HandleCardSwapWithDebug(client, deckCards, otherCards) {
				return 
			}

		case 2:
			SetDeckMenuActive(false)
			SetPackageMenuActive(false)
			return

		default:
			fmt.Println("⚠️ Opção inválida!")
			time.Sleep(1 * time.Second)
		}
	}
}

