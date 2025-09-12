package main

import(	"sync")

// Status do sistema de pacotes 
var (
	PackageMenuActive bool
	WaitingForPackage bool 
	packageMutex      sync.RWMutex
)

func SetPackageMenuActive(active bool) {
	packageMutex.Lock()
	defer packageMutex.Unlock()
	PackageMenuActive = active
}

func IsPackageMenuActive() bool {
	packageMutex.RLock()
	defer packageMutex.RUnlock()
	return PackageMenuActive
}

func SetWaitingForPackage(waiting bool) {
	packageMutex.Lock()
	defer packageMutex.Unlock()
	WaitingForPackage = waiting
}

func IsWaitingForPackage() bool {
	packageMutex.RLock()
	defer packageMutex.RUnlock()
	return WaitingForPackage
}


// Status da partida 
type MatchmakingState struct {
	IsSearching bool
	InGame      bool
	CurrentTurn string
	GameState   map[string]interface{}
	mu          sync.RWMutex
}

func (ms *MatchmakingState) SetSearching(searching bool) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.IsSearching = searching
}

func (ms *MatchmakingState) SetInGame(inGame bool) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.InGame = inGame
}

func (ms *MatchmakingState) IsInAnyState() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.IsSearching || ms.InGame || PackageMenuActive
}