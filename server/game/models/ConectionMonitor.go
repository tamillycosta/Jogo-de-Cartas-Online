package models

import (
    "fmt"
    "encoding/json"
    response "jogodecartasonline/api/Response"
    "sync"
    "time"
)

// Sistema de monitoramento de conexões
type ConnectionMonitor struct {
    Mu              sync.RWMutex
    PlayerHeartbeat map[string]time.Time  
    CheckInterval   time.Duration
    TimeoutDuration time.Duration
    Lobby           *Lobby
}

// Inicializa sistema de monitoramento
func NewConnectionMonitor(lobby *Lobby) *ConnectionMonitor {
    return &ConnectionMonitor{
        PlayerHeartbeat: make(map[string]time.Time),
        CheckInterval:   30 * time.Second,  
        TimeoutDuration: 90 * time.Second,  
        Lobby:           lobby,
    }
}

// Inicia monitoramento em background
func (cm *ConnectionMonitor) Start() {
    go cm.heartbeatMonitor()
    fmt.Printf("🔍 Sistema de monitoramento de conexões iniciado\n")
}

// Monitor principal que roda em background
func (cm *ConnectionMonitor) heartbeatMonitor() {
    ticker := time.NewTicker(cm.CheckInterval)
    defer ticker.Stop()

    for range ticker.C {
        cm.checkAllConnections()
    }
}

// Verifica todas as conexões
func (cm *ConnectionMonitor) checkAllConnections() {
    cm.Lobby.Mu.RLock()
    players := make([]*Player, 0, len(cm.Lobby.Players))
    for _, player := range cm.Lobby.Players {
        players = append(players, player)
    }
    cm.Lobby.Mu.RUnlock()

    for _, player := range players {
        cm.checkPlayerConnection(player)
    }
}

// Verifica conexão de um player específico
func (cm *ConnectionMonitor) checkPlayerConnection(player *Player) {
    if player.Conn == nil {
        return
    }

    // Tenta enviar ping
    if !cm.sendPing(player) {
        fmt.Printf("⚠️ Player %s não respondeu ao ping, removendo...\n", player.Nome)
        cm.handleDisconnectedPlayer(player)
        return
    }

    // Verifica último heartbeat
    cm.Mu.RLock()
    lastHeartbeat, exists := cm.PlayerHeartbeat[player.Nome]
    cm.Mu.RUnlock()

    if exists && time.Since(lastHeartbeat) > cm.TimeoutDuration {
        fmt.Printf("Player %s timeout (último ping: %v atrás)\n", 
            player.Nome, time.Since(lastHeartbeat))
        cm.handleDisconnectedPlayer(player)
    }
}

// Envia ping para verificar se conexão está ativa

func (cm *ConnectionMonitor) sendPing(player *Player) bool {
    resp := response.Response{
        Status:  200,
        Message: "PING", 
        Data:    map[string]string{"type": "PING"},
    }
    
    data, err := json.Marshal(resp)
    if err != nil {
        return false
    }
    
    message := append(data, '\n')
    
    player.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
    _, err = player.Conn.Write(message)
    player.Conn.SetWriteDeadline(time.Time{})
    
    if err != nil {
        fmt.Printf("Ping falhou para %s: %v\n", player.Nome, err)
        return false
    }

    // Atualiza heartbeat
    cm.Mu.Lock()
    cm.PlayerHeartbeat[player.Nome] = time.Now()
    cm.Mu.Unlock()
    
    return true
}

// Processa player desconectado
func (cm *ConnectionMonitor) handleDisconnectedPlayer(player *Player) {
    fmt.Printf("🚨 Processando desconexão de %s\n", player.Nome)

    // Remove do sistema de heartbeat
    cm.Mu.Lock()
    delete(cm.PlayerHeartbeat, player.Nome)
    cm.Mu.Unlock()

    // Se está em partida, notifica oponente e finaliza
    if player.Match != nil {
        cm.handlePlayerInMatchDisconnect(player)
    }

    // Se está na fila de espera, remove
    cm.removeFromWaitQueue(player)

    // Remove do lobby
    cm.Lobby.Mu.Lock()
    delete(cm.Lobby.Players, player.Nome)
    cm.Lobby.Mu.Unlock()

    // Fecha conexão
    if player.Conn != nil {
        player.Conn.Close()
    }

    fmt.Printf("🧹 Player %s removido completamente do servidor\n", player.Nome)
}

// Processa desconexão durante partida
func (cm *ConnectionMonitor) handlePlayerInMatchDisconnect(disconnectedPlayer *Player) {
    match := disconnectedPlayer.Match
    if match == nil {
        return
    }

    opponent := cm.Lobby.GetOpponent(match, disconnectedPlayer)
    if opponent == nil {
        return
    }

    fmt.Printf("⚔️ %s desconectou durante partida contra %s\n", 
        disconnectedPlayer.Nome, opponent.Nome)

   
    match.Status = GAME_STATUS_ENDED

    // Notifica oponente que ganhou por desconexão
    if opponent.Conn != nil {
        cm.notifyOpponentWinByDisconnect(opponent)
    }

    disconnectedPlayer.Match = nil
    opponent.Match = nil
    

   
    cm.Lobby.Mu.Lock()
    delete(cm.Lobby.Matchs, match.ID)
    cm.Lobby.Mu.Unlock()

    // Atualiza score do vencedor
    opponent.Score += 100
    cm.Lobby.DB.Save(opponent)

    fmt.Printf("🏆 %s ganhou por desconexão do oponente\n", opponent.Nome)
}





// Funções auxiliares


// Remove player da fila de espera
func (cm *ConnectionMonitor) removeFromWaitQueue(player *Player) {
    cm.Lobby.Mu.Lock()
    defer cm.Lobby.Mu.Unlock()

    for i, waitingPlayer := range cm.Lobby.WaitQueue {
        if waitingPlayer.Player.Nome == player.Nome {
            cm.Lobby.WaitQueue = append(cm.Lobby.WaitQueue[:i], cm.Lobby.WaitQueue[i+1:]...)
            fmt.Printf("%s removido da fila de espera\n", player.Nome)
            break
        }
    }
}


// Registra player ativo 
func (cm *ConnectionMonitor) RegisterPlayerActivity(playerName string) {
    cm.Mu.Lock()
    cm.PlayerHeartbeat[playerName] = time.Now()
    cm.Mu.Unlock()
}

// Força verificação imediata de um player
func (cm *ConnectionMonitor) CheckPlayerNow(playerName string) {
    cm.Lobby.Mu.RLock()
    player, exists := cm.Lobby.Players[playerName]
    cm.Lobby.Mu.RUnlock()

    if exists {
        cm.checkPlayerConnection(player)
    }
}

// Estatísticas do monitor
func (cm *ConnectionMonitor) GetStats() map[string]interface{} {
    cm.Mu.RLock()
    defer cm.Mu.RUnlock()

    activeConnections := len(cm.PlayerHeartbeat)
    
    // Conta conexões com problemas
    problematicConnections := 0
    now := time.Now()
    
    for _, lastPing := range cm.PlayerHeartbeat {
        if now.Sub(lastPing) > cm.TimeoutDuration/2 {
            problematicConnections++
        }
    }

    return map[string]interface{}{
        "activeConnections":      activeConnections,
        "problematicConnections": problematicConnections,
        "checkInterval":          cm.CheckInterval.String(),
        "timeoutDuration":        cm.TimeoutDuration.String(),
    }
}