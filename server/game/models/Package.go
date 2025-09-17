package models

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/google/uuid"
)

// conta quantas cartas especiais foram distribuidas
var SpecialCardCount = make(map[string]int)
var CardVersions = make(map[string]int)

type CardRarity string

const (
	COMMON    CardRarity = "COMMON"
	UNCOMMON  CardRarity = "UNCOMMON"
	RARE      CardRarity = "RARE"
	EPIC      CardRarity = "EPIC"
	LEGENDARY CardRarity = "LEGENDARY"
)

// Cartas Template (estoque global)

var BaseCards = map[string]Card{

	"starter_mage": {
		TemplateID: "starter_mage",
		Nome:       "Aprendiz Mago",
		Power:      100,
		Health:     100,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_goblin": {
		TemplateID: "starter_goblin",
		Nome:       "Goblin Com Bomba",
		Power:      100,
		Health:     140,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_witch": {
		TemplateID: "starter_witch",
		Nome:       "Bruxa",
		Power:      150,
		Health:     120,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_wolf": {
		TemplateID: "starter_wolf",
		Nome:       "Lobo",
		Power:      100,
		Health:     90,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_fire": {
		TemplateID: "starter_fire",
		Nome:       "Feiticeira de Fogo",
		Power:      100,
		Health:     150,
		Rarity:     string(COMMON),
	},
	"starter_knight": {
		TemplateID: "starter_knight",
		Nome:       "Escudeiro",
		Power:      70,
		Health:     100,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_raven": {
		TemplateID: "starter_raven",
		Nome:       "Corvo Místico",
		Power:      100,
		Health:     95,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_devil": {
		TemplateID: "starter_devil",
		Nome:       "Cavaleiro das Trevas",
		Power:      120,
		Health:     110,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_elf": {
		TemplateID: "starter_elf",
		Nome:       "Elfo Caçador",
		Power:      90,
		Health:     100,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},
	"starter_dragon": {
		TemplateID: "starter_dragon",
		Nome:       "Dragão Comum",
		Power:      50,
		Health:     100,
		Rarity:     string(COMMON),
		MaxCopies:  0,
	},

	// Cartas especiais raras
	"legend_dragon": {
		TemplateID: "legend_dragon",
		Nome:       "Dragão Ancião",
		Power:      350,
		Health:     300,
		Rarity:     string(LEGENDARY),
		IsSpecial:  true,
		MaxCopies:  100,
	},
	"legend_archmage": {
		TemplateID: "legend_archmage",
		Nome:       "Arquimago Supremo",
		Power:      230,
		Health:     280,
		Rarity:     string(LEGENDARY),
		IsSpecial:  true,
		MaxCopies:  100,
	},
	"epic_shadow_witch": {
		TemplateID: "epic_shadow_witch",
		Nome:       "Bruxa das Sombras",
		Power:      200,
		Health:     200,
		Rarity:     string(EPIC),

		IsSpecial: true,
		MaxCopies: 200,
	},

	"epic_phoenix": {
		TemplateID: "epic_phoenix",
		Nome:       "Fênix Dourada",
		Power:      170,
		Health:     200,
		Rarity:     string(EPIC),
		IsSpecial:  true,
		MaxCopies:  200,
	},

	"rare_best": {
		TemplateID: "rare_best",
		Nome:       "Besta Sombria",
		Power:      180,
		Health:     170,
		Rarity:     string(RARE),
		IsSpecial:  true,
		MaxCopies:  200,
	},

	"uncumon_bow": {
		TemplateID: "uncumon_bow",
		Nome:       "Arqueiro Fantasma",
		Power:      150,
		Health:     170,
		Rarity:     string(UNCOMMON),
		IsSpecial:  true,
		MaxCopies:  200,
	},
}
var StarterCardIDs = []string{
	"starter_mage", "starter_goblin", "starter_witch", "starter_wolf",
	"starter_fire", "starter_knight", "starter_raven", "starter_devil",
	"starter_elf", "starter_dragon",
}

type CardParck struct {
	ID    string
	Cards []*Card
}

type CardParckSystem struct {
	MU sync.Mutex
}

func CreatePlayerCard(templateID, playerID string) *Card {

	baseCard, exist := BaseCards[templateID]
	if !exist {
		return nil
	}
	playerCard := baseCard
	playerCard.PlayerId = playerID
	playerCard.ID = uuid.NewString()
	return &playerCard

}

// Geração de carta de pacote
func GeneratePackCard(playerID string) *Card {
	rarity := rollRarity()
	availableCards := getCardsByRarity(rarity)

	if len(availableCards) == 0 {
		availableCards = getCardsByRarity(COMMON)
	}

	templateID := availableCards[rand.Intn(len(availableCards))]
	baseCard := BaseCards[templateID]

	if baseCard.Rarity == string(COMMON) {
		return CreatePlayerCard(templateID, playerID)
	}

	// Para cartas especiais
	// Verifica se ainda tem da versão original
	if IsSpecialCardAvailable(templateID) {

		MarkSpecialCardUsed(templateID)
		return CreatePlayerCard(templateID, playerID)
	}

	// Se esgotou a versão original, cria uma nova versão
	return CreateNextVersion(templateID, playerID)
}

func GenerateInicialCards(playerId string) []*Card {

	// embaralha os ids das cartas basicas
	shuffled := make([]string, len(StarterCardIDs))
	copy(shuffled, StarterCardIDs)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	cards := make([]*Card, 3)
	for i := 0; i < 3; i++ {
		card := CreatePlayerCard(shuffled[i], playerId)
		if card == nil {
			return nil
		}
		card.InDeck = true
		cards[i] = card
	}
	return cards
}

func IsSpecialCardAvailable(templateID string) bool {
	baseCard := BaseCards[templateID]

	if baseCard.Rarity == string(COMMON) {
		return true
	}

	distributed := SpecialCardCount[templateID]

	if distributed >= baseCard.MaxCopies {
		fmt.Printf("⚠️  A carta %s está esgotada (%d/%d)\n",
			baseCard.Nome, distributed, baseCard.MaxCopies)
		return false
	}

	return true
}

func MarkSpecialCardUsed(templateID string) {
	SpecialCardCount[templateID]++
}

// cria uma nova versão de uma carta especial caso a original tenha esgotado
func CreateNextVersion(originalTemplateID string, playerID string) *Card {
	baseCard := BaseCards[originalTemplateID]

	CardVersions[originalTemplateID]++
	version := CardVersions[originalTemplateID]

	fmt.Printf("🔧 DEBUG: Criando nova versão para %s - Versão: %d\n", baseCard.Nome, version)

	newPower, newHealth := generateRandomStats(baseCard.Rarity)

	newCard := Card{
		ID:         uuid.NewString(),
		TemplateID: fmt.Sprintf("%s_v%d", originalTemplateID, version),
		Nome:       fmt.Sprintf("%s V%d", baseCard.Nome, version),
		Power:      newPower,
		Health:     newHealth,
		Rarity:     baseCard.Rarity,
		PlayerId:   playerID,
		IsSpecial:  true,
		MaxCopies:  baseCard.MaxCopies,
		InDeck:     false,
	}

	BaseCards[newCard.TemplateID] = newCard

	SpecialCardCount[newCard.TemplateID] = 1

	fmt.Printf("✅ Nova versão criada: %s (Power: %d, Health: %d)\n",
		newCard.Nome, newCard.Power, newCard.Health)

	return &newCard
}

// ------------------------------------------ Funções auxiliares
func rollRarity() CardRarity {
	roll := rand.Float64() * 100

	switch {
	case roll < 2.0:
		return LEGENDARY
	case roll < 10.0:
		return EPIC
	case roll < 20.0:
		return RARE
	case roll < 30.0:
		return UNCOMMON
	default:
		return COMMON
	}
}

func getCardsByRarity(rarity CardRarity) []string {
	var available []string

	for templateID, card := range BaseCards {

		if card.Rarity == string(rarity) {
			available = append(available, templateID)
		}
	}

	return available
}

func generateRandomStats(rarity string) (int, int) {
	var minPower, maxPower, minHealth, maxHealth int

	switch rarity {
	case string(UNCOMMON):
		minPower, maxPower = 120, 180
		minHealth, maxHealth = 120, 180
	case string(RARE):
		minPower, maxPower = 160, 220
		minHealth, maxHealth = 160, 220
	case string(EPIC):
		minPower, maxPower = 200, 280
		minHealth, maxHealth = 200, 280
	case string(LEGENDARY):
		minPower, maxPower = 280, 350
		minHealth, maxHealth = 280, 350
	default: // COMMON
		minPower, maxPower = 80, 120
		minHealth, maxHealth = 80, 120
	}

	power := rand.Intn(maxPower-minPower+1) + minPower
	health := rand.Intn(maxHealth-minHealth+1) + minHealth

	return power, health
}

func InitializeCardCounts() {
	fmt.Println("🔧 Inicializando contadores de cartas especiais...")

	for templateID, card := range BaseCards {
		if card.Rarity != string(COMMON) {
			if _, exists := SpecialCardCount[templateID]; !exists {
				SpecialCardCount[templateID] = 0
				fmt.Printf("   %s: 0/%d\n", card.Nome, card.MaxCopies)
			}
		}
	}

	for templateID := range BaseCards {
		if _, exists := CardVersions[templateID]; !exists {
			CardVersions[templateID] = 0
		}
	}

	fmt.Println("✅ Contadores inicializados!")
}

// Função de debug para mostrar estado atual

func PrintCardStats() {
	fmt.Println("\n📊 === ESTATÍSTICAS DE CARTAS ===")

	// Conta por raridade
	rarityCount := make(map[string]int)

	for templateID, count := range SpecialCardCount {
		if card, exists := BaseCards[templateID]; exists {
			rarityCount[card.Rarity] += count
		}
	}

	// Imprime por raridade
	fmt.Println("Cartas distribuídas por raridade:")
	for rarity, count := range rarityCount {
		fmt.Printf("  %s: %d\n", rarity, count)
	}

	// Total
	total := 0
	for _, count := range rarityCount {
		total += count
	}
	fmt.Printf("  TOTAL ESPECIAIS: %d\n", total)

	// Versões criadas
	versionsCreated := 0
	for _, version := range CardVersions {
		versionsCreated += version
	}

	if versionsCreated > 0 {
		fmt.Printf("🆕 Versões criadas: %d\n", versionsCreated)
		fmt.Println("Detalhes das versões:")
		for templateID, version := range CardVersions {
			if version > 0 {
				if card, exists := BaseCards[templateID]; exists {
					fmt.Printf("  %s: %d versões\n", card.Nome, version)
				}
			}
		}
	}

	fmt.Println("=================================")
}
