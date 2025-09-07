package main

import (
	"encoding/json"
	"fmt"
	response "jogodecartasonline/api/Response"
	"jogodecartasonline/client/model"
	"jogodecartasonline/server/game/models"
	"time"
)

// Estados do sistema de pacotes
type PackageState struct {
	CanOpen       bool
	Remaining     string
	TotalCards    int
	NewCards      []models.Card
	InPackageMenu bool
}

var packageState = &PackageState{}

func ProcessPackageResponse(resp response.Response) {
	fmt.Println("📋 Processando status do pacote...")

	canOpenStr := resp.Data["canOpen"]
	remaining := resp.Data["remaining"]
	totalCards := resp.Data["totalCards"]

	
	canOpen := canOpenStr == "true"

	packageState.CanOpen = canOpen
	packageState.Remaining = remaining
	packageState.TotalCards = parseInt(totalCards)

	if canOpen {
		PackageMenu()
	} else {
		CooldownMenu()
	}
}

func ProcessPackageOpenResponse(resp response.Response) {
	cardsStr := resp.Data["cards"]
	totalCards := resp.Data["totalCards"]

	// Decodifica as cartas recebidas
	var newCards []models.Card
	if err := json.Unmarshal([]byte(cardsStr), &newCards); err != nil {
		fmt.Printf("Erro ao decodificar cartas: %v\n", err)
		return
	}

	// Mostra cartas obtidas
	Menu.ClearScreen()
	fmt.Println("✨ === PACOTE ABERTO! ===")
	fmt.Printf("Total de cartas: %s\n\n", totalCards)
	fmt.Println("Cartas obtidas:")

	for i, card := range newCards {
		fmt.Printf("%d. %s (Poder: %d, Vida: %d, Raridade: %s)\n",
			i+1, card.Nome, card.Power, card.Health, card.Rarity)
	}

	fmt.Println("\nPressione Enter para continuar...")
	fmt.Scanln()

	packageState.InPackageMenu = false
}

func TryOpenPackage(username string, client *model.Client) {
	Menu.ClearScreen()
	packageState.InPackageMenu = true

	fmt.Println("📦 === SISTEMA DE PACOTES ===")
	fmt.Println("Verificando status...")

	// Verifica status do pacote
	err := client.CheckPackStatus(username)
	if err != nil {
		fmt.Printf("Erro ao verificar status: %v\n", err)
		packageState.InPackageMenu = false
		time.Sleep(2 * time.Second)
		return
	}

	// Aguarda resposta do status
	fmt.Println("Aguardando resposta do servidor...")

	// Loop até sair do menu de pacotes
	for packageState.InPackageMenu {
		time.Sleep(100 * time.Millisecond)
	}
}

func openPackage() {
	Menu.ClearScreen()
	fmt.Println("📦 Abrindo pacote...")

	
	// client.OpenPack(username)

	// Simula abertura 
	fmt.Println("✨ PACOTE ABERTO!")
	fmt.Println("\nCartas obtidas:")
	fmt.Println("1. Dragão de Fogo (Raro)")
	fmt.Println("2. Mago Elemental (Comum)")
	fmt.Println("3. Poção de Vida (Comum)")
	fmt.Println("4. Escudo Mágico (Incomum)")
	fmt.Println("5. Fênix Dourada (Épico)")

	fmt.Println("\nPressione Enter para continuar...")
	fmt.Scanln()

	// Volta ao menu principal após abrir
	packageState.InPackageMenu = false
}

func showMyCards() {
	Menu.ClearScreen()
	fmt.Println("🃏 === MINHAS CARTAS ===")
	fmt.Printf("Total: %d cartas\n\n", packageState.TotalCards)

	
	// mostra exemplo
	fmt.Println("1. Dragão Normal (Poder: 50, Vida: 100)")
	fmt.Println("2. Mago Iniciante (Poder: 30, Vida: 80)")
	fmt.Println("3. Guerreiro (Poder: 40, Vida: 90)")

	fmt.Println("\nPressione Enter para voltar...")
	fmt.Scanln()

	PackageMenu()
}

func PackageMenu() {
	Menu.ClearScreen()
	Menu.ShowPackageMenu(packageState.TotalCards)
	var opcao int
	fmt.Print("Escolha: ")
	fmt.Scanln(&opcao)

	switch opcao {
	case 1:
		openPackage()
	case 2:
		showMyCards()
	case 3:
		packageState.InPackageMenu = false
	default:
		fmt.Println("Opção inválida!")
		time.Sleep(1 * time.Second)
		PackageMenu()
	}
}

func CooldownMenu() {
	Menu.ClearScreen()
	Menu.ShowCooldownMessage(packageState.TotalCards, packageState.Remaining)

	var opcao int
	fmt.Print("Escolha: ")
	fmt.Scanln(&opcao)

	switch opcao {
	case 1:
		showMyCards()
	case 2:
		packageState.InPackageMenu = false
	default:
		fmt.Println("Opção inválida!")
		time.Sleep(1 * time.Second)
		CooldownMenu()
	}
}

// Função auxiliar para converter string para int
func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}
