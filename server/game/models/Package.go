package models

import (
	"sync"
    "math/rand"
    
	"github.com/google/uuid"
)



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
        Nome:        "Poção de Cura",
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
        MaxCopies:   10,
    },
    "legend_archmage": {
        TemplateID:  "legend_archmage",
        Nome:        "Arquimago Supremo",
        Power:       230,
        Health:      280,
        Rarity: string(LEGENDARY),
        IsSpecial:   true,
        MaxCopies:   15,
    },
    "epic_shadow_witch": {
        TemplateID:  "epic_shadow_witch",
        Nome:        "Bruxa das Sombras",
        Power:       200,
        Health:      200,
        Rarity:      string(EPIC),
      
        IsSpecial:   true,
        MaxCopies:   50,
    },

    "epic_phoenix": {
        TemplateID:  "epic_phoenix",
        Nome:        "Fênix Dourada",
        Power:       170,
        Health:      200,
        Rarity:      string(EPIC),
        IsSpecial:   true,
        MaxCopies:   30,
    },

    "rare_best": {
        TemplateID:  "epic_phoenix",
        Nome:        "Besta Sombria",
        Power:       180,
        Health:      170,
        Rarity:      string(RARE),
        IsSpecial:   true,
        MaxCopies:   50,
    },

    "uncumon_bow": {
        TemplateID:  "epic_phoenix",
        Nome:        "Arqueiro Fantasma",
        Power:       150,
        Health:      170,
        Rarity:      string(UNCOMMON),
        IsSpecial:   true,
        MaxCopies:   100,
    },


}
var StarterCardIDs = []string{
    "starter_mage", "starter_goblin", "starter_witch", "starter_wolf",
    "starter_fire", "starter_knight", "starter_raven", "starter_devil", 
    "starter_elf", "starter_dragon",
}

var SpecialCardCount = make(map[string]int)



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
          cards[i] = card
      }
      return cards
  }


func IsSpecialCardAvailable(templateID string) bool {
    baseCard, exists := BaseCards[templateID]
    if !exists || !baseCard.IsSpecial {
        return true // Cartas normais sempre disponíveis
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
    
    // Marca como usada se especial
    if BaseCards[templateID].IsSpecial {
        MarkSpecialCardUsed(templateID)
    }
    
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



