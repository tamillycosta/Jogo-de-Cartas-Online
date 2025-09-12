package models

import (
	"sync"
    "math/rand"
    "fmt"
	"github.com/google/uuid"
)


var SpecialCardCount = make(map[string]int)
var CardVersions = make(map[string]int) 
type CardRarity string

const (
	COMMON    CardRarity = "COMMON"    // 60% chance
	UNCOMMON  CardRarity = "UNCOMMON"  // 25% chance
	RARE      CardRarity = "RARE"      // 12% chance
	EPIC      CardRarity = "EPIC"      // 2.5% chance
	LEGENDARY CardRarity = "LEGENDARY" // 0.5% chance

)

// Cartas Template

var BaseCards = map[string]Card {

	"starter_mage": {
        TemplateID:  "starter_mage",
        Nome:        "Aprendiz Mago",
        Power:       100,
        Health:      100,
        Rarity:      string(COMMON),
        MaxCopies:   0,
    },
    "starter_goblin": {
        TemplateID:  "starter_goblin",
        Nome:        "Goblin Com Bomba",
        Power:       100,
        Health:      140,
        Rarity:      string(COMMON),
        MaxCopies:   0,
    },
    "starter_witch": {
        TemplateID:  "starter_witch",
        Nome:        "Bruxa",
        Power:       150,
        Health:      120,
        Rarity:      string(COMMON),
        MaxCopies:   0,
    },
    "starter_wolf": {
        TemplateID:  "starter_wolf",
        Nome:        "Lobo",
        Power:       100,
        Health:      90,
        Rarity:      string(COMMON),
        MaxCopies:   0,
    },
    "starter_fire": {
        TemplateID:  "starter_fire",
        Nome:        "Feiticeira de Fogo",
        Power:       100,
        Health:      150,
        Rarity:      string(COMMON),
    },
    "starter_knight": {
        TemplateID:  "starter_knight",
        Nome:        "Escudeiro",
        Power:       70,
        Health:      100,
        Rarity:      string(COMMON),
        MaxCopies:   0,
    },
    "starter_raven": {
        TemplateID:  "starter_raven",
        Nome:        "Corvo Místico",
        Power:       100,
        Health:      95,
        Rarity:      string(COMMON),
        MaxCopies:   0,
    },
    "starter_devil": {
        TemplateID:  "starter_devil",
        Nome:        "Cavaleiro das Trevas",
        Power:       120,
        Health:      110,
        Rarity:      string(COMMON),
        MaxCopies:   0,
    },
    "starter_elf": {
        TemplateID:  "starter_elf",
        Nome:        "Elfo Caçador",
        Power:       90,
        Health:      100,
        Rarity:      string(COMMON),
        MaxCopies:   0,
    },
    "starter_dragon": {
        TemplateID:  "starter_dragon",
        Nome:        "Dragão Comum",
        Power:       50,
        Health:      100,
        Rarity:      string(COMMON),
        MaxCopies:   0,
    },

    // Cartas especiais raras
    "legend_dragon": {
        TemplateID:  "legend_dragon",
        Nome:        "Dragão Ancião",
        Power:       350,
        Health:      300,
        Rarity: string(LEGENDARY),
        IsSpecial:   true,
        MaxCopies:   200,
    },
    "legend_archmage": {
        TemplateID:  "legend_archmage",
        Nome:        "Arquimago Supremo",
        Power:       230,
        Health:      280,
        Rarity: string(LEGENDARY),
        IsSpecial:   true,
        MaxCopies:   200,
    },
    "epic_shadow_witch": {
        TemplateID:  "epic_shadow_witch",
        Nome:        "Bruxa das Sombras",
        Power:       200,
        Health:      200,
        Rarity:      string(EPIC),
      
        IsSpecial:   true,
        MaxCopies:   200,
    },

    "epic_phoenix": {
        TemplateID:  "epic_phoenix",
        Nome:        "Fênix Dourada",
        Power:       170,
        Health:      200,
        Rarity:      string(EPIC),
        IsSpecial:   true,
        MaxCopies:   200,
    },

    "rare_best": {
        TemplateID:  "rare_best",
        Nome:        "Besta Sombria",
        Power:       180,
        Health:      170,
        Rarity:      string(RARE),
        IsSpecial:   true,
        MaxCopies:   200,
    },

    "uncumon_bow": {
        TemplateID:  "uncumon_bow",
        Nome:        "Arqueiro Fantasma",
        Power:       150,
        Health:      170,
        Rarity:      string(UNCOMMON),
        IsSpecial:   true,
        MaxCopies:   200,
    },


}
var StarterCardIDs = []string{
    "starter_mage", "starter_goblin", "starter_witch", "starter_wolf",
    "starter_fire", "starter_knight", "starter_raven", "starter_devil", 
    "starter_elf", "starter_dragon",
}


type CardParck struct{
    ID string
    Cards []*Card
  }
  
  type CardParckSystem struct{
      MU sync.Mutex
  
  }
  
func CreatePlayerCard(templateID ,playerID string)(*Card){
      
      baseCard, exist := BaseCards[templateID]
      if(!exist){
          return nil
      }
      playerCard := baseCard
      playerCard.PlayerId = playerID
      playerCard.ID = uuid.NewString()
      return &playerCard
  
  }
  
func GenerateInicialCards(playerId string)([]*Card){
  
      // embaralha os ids das cartas basicas 
      shuffled := make([]string, len(StarterCardIDs))
      copy(shuffled, StarterCardIDs)
      
      for i := len(shuffled) - 1; i > 0; i-- {
          j := rand.Intn(i + 1)
          shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
      }
  
      cards := make([]*Card, 3)
      for i := 0 ; i < 3; i++{
          card := CreatePlayerCard(shuffled[i], playerId)
          if(card == nil){
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
    return distributed < baseCard.MaxCopies
}

func MarkSpecialCardUsed(templateID string) {
    SpecialCardCount[templateID]++
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
    
    // Se é carta COMMON, cria normalmente 
    if baseCard.Rarity == string(COMMON) {
        return CreatePlayerCard(templateID, playerID)
    }
    
    // Se é carta especial, verifica limite
    if !IsSpecialCardAvailable(templateID) {
        // Acabou o limite, cria versão nova
        return CreateNextVersion(templateID, playerID)
    }
    
    // Ainda tem da versão original
    MarkSpecialCardUsed(templateID)
    return CreatePlayerCard(templateID, playerID)
}

// Funções auxiliares
func rollRarity() CardRarity {
    roll := rand.Float64() * 100
    
    switch {
    case roll < 0.5:
        return LEGENDARY
    case roll < 3.0:
        return EPIC
    case roll < 15.0:
        return RARE
    case roll < 40.0:
        return UNCOMMON
    default:
        return COMMON
    }
}

func getCardsByRarity(rarity CardRarity) []string {
    var available []string
    
    for templateID, card := range BaseCards {
        if card.Rarity == string(rarity) && IsSpecialCardAvailable(templateID) {
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

func CreateNextVersion(originalTemplateID string, playerID string) *Card {
    baseCard := BaseCards[originalTemplateID]
    
    // Incrementa versão
    CardVersions[originalTemplateID]++
    version := CardVersions[originalTemplateID] + 1 // +1 porque começou da versão 1
    
    // Gera stats aleatórios baseados na raridade
    newPower, newHealth := generateRandomStats(baseCard.Rarity)
    
    // Cria nova carta
    newCard := Card{
        ID:         uuid.NewString(),
        TemplateID: fmt.Sprintf("%s_v%d", originalTemplateID, version),
        Nome:       fmt.Sprintf("%s V%d", baseCard.Nome, version),
        Power:      newPower,
        Health:     newHealth,
        Rarity:     baseCard.Rarity,
        PlayerId:   playerID,
        IsSpecial:  true,
        MaxCopies:  baseCard.MaxCopies, // Mesmo limite da original
        InDeck:     false,
    }
    
    // Adiciona nova versão ao mapa de cartas base para futuras criações
    BaseCards[newCard.TemplateID] = newCard
    
    // Reset contador para nova versão
    SpecialCardCount[newCard.TemplateID] = 1
    
    fmt.Printf("Nova versão criada: %s (Power: %d, Health: %d)\n", 
        newCard.Nome, newCard.Power, newCard.Health)
    
    return &newCard
}