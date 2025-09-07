package models


import(
	"fmt"
	"time"
	"sync"
)


type PackSystem struct {
    LastPackTime map[string]time.Time // guarda ultima vez que um jogador abriu um pacote 
    PackCooldown time.Duration
    Mu           sync.RWMutex
}



// Verifica se um jogador pode abrir um pacote 
func (s *PackSystem) CanOpenPack(playerID string) (bool, time.Duration) {
    s.Mu.RLock()
    lastTime, exists := s.LastPackTime[playerID]
    s.Mu.RUnlock()
    
    if !exists {
        return true, 0
    }
    
	
    timeSince := time.Since(lastTime)
    if timeSince >= s.PackCooldown {
        return true, 0
    }
    
	// retorna o tempo que falta 
    remaining := s.PackCooldown - timeSince
    return false, remaining
}


func (s *PackSystem) OpenPack(playerID string) ([]*Card, error) {
    canOpen, remaining := s.CanOpenPack(playerID)
    if !canOpen {
        return nil, fmt.Errorf("aguarde %v para abrir outro pacote", remaining.Round(time.Second))
    }
    
    cards := make([]*Card, 0, 5)
    
    for i := 0; i < 5; i++ {
        // Cria carta para o player
        card := GeneratePackCard(playerID)        
        cards = append(cards, card)
    }
    
    // Atualiza último tempo de acessa do jogador
    s.Mu.Lock()
    s.LastPackTime[playerID] = time.Now()
    s.Mu.Unlock()
    
    return cards, nil
}