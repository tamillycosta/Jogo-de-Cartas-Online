package main

import (
	
	"encoding/json"
	"fmt"
	
	"sync"
	"testing"
	"time"
    
	"jogodecartasonline/test/utils"
)

// Estrutura para estatísticas dos pacotes



// TESTE UNITARIO PARA ABERTURA DE PACOTES
func TestCooldownSystem(t *testing.T) {
    client, err := utils.NewPackageTestClient(t, "cooldown_test_player")
    if err != nil {
        t.Fatalf("Erro ao criar cliente: %v", err)
    }
    defer client.Conn.Close()
    
    fmt.Printf("Testando sistema de cooldown...\n")
    
    // Verifica status inicial
    statusResp, err := client.CheckPackageStatus()
    if err != nil {
        t.Fatalf("Erro ao verificar status: %v", err)
    }
    
    canOpen := statusResp.Data["canOpen"]
    remaining := statusResp.Data["remaining"]
    totalCards := statusResp.Data["totalCards"]
    
    fmt.Printf("Status inicial:\n")
    fmt.Printf("  - Pode abrir: %s\n", canOpen)
    fmt.Printf("  - Tempo restante: %s\n", remaining)
    fmt.Printf("  - Total de cartas: %s\n", totalCards)
    
    if canOpen == "true" {
        fmt.Printf("Player pode abrir pacote imediatamente!\n")
        
        // Abre o pacote
        packResp, err := client.OpenPackage()
        if err != nil {
            t.Errorf("Erro ao abrir pacote: %v", err)
            return
        }
        
        if packResp.Status == 200 {
            fmt.Printf("Pacote aberto com sucesso!\n")
            
            // Verifica status após abertura
            statusResp2, err := client.CheckPackageStatus()
            if err != nil {
                t.Errorf("Erro ao verificar status pós-abertura: %v", err)
                return
            }
            
            fmt.Printf("Status após abertura:\n")
            fmt.Printf("  - Pode abrir: %s\n", statusResp2.Data["canOpen"])
            fmt.Printf("  - Tempo restante: %s\n", statusResp2.Data["remaining"])
            fmt.Printf("  - Total de cartas: %s\n", statusResp2.Data["totalCards"])
            
        } else {
            fmt.Printf("Falha ao abrir pacote: %s\n", packResp.Message)
        }
    } else {
        fmt.Printf("Player está em cooldown\n")
        if remaining == "" {
            fmt.Printf("ATENÇÃO: Campo 'remaining' está vazio!\n")
        }
    }
}



// TESTE DE CONCORRÊNCIA PARA ABERTURA DOS PACOTES
// N CLIENTES TENTAM ABRIR PACOTES AO MESMO TEMPO 
// VERIFICA SUCESSO DE ABERTURA , DUPLICIDADE DAS CARTAS E DISTRIBUIÇÃO DE RARIDADE
func TestOpenPackagesSafe(t *testing.T) {
    numClients := 3000
    packagesPerClient := 1

    stats := &utils.PackageStats{
        CardsByRarity: make(map[string]int),
    }
    var cardIDs sync.Map
    done := make(chan bool, numClients)
    start := time.Now()

    // Canal para logs temporários (progresso)
    logChan := make(chan string, numClients*packagesPerClient)

    for i := 0; i < numClients; i++ {
        go func(i int) {
            playerName := fmt.Sprintf("player_%d", i)
            client, err := utils.NewPackageTestClient(t, playerName)
            if err != nil {
                stats.AddError()
                logChan <- fmt.Sprintf("Falha ao criar client %s: %v", playerName, err)
                done <- false
                return
            }
            defer client.Conn.Close()

            stats.AddPlayer()

            for j := 0; j < packagesPerClient; j++ {
                resp, err := client.OpenPackage()
                if err != nil {
                    stats.AddError()
                    logChan <- fmt.Sprintf("[%s] erro ao abrir pacote: %v", playerName, err)
                    continue
                }

                // Ignorando cooldowns no teste de stress
                if resp.Message == "Pacote em cooldown" {
                    continue
                }

                stats.AddPack()

                cardsJSON, ok := resp.Data["cards"]
                if !ok {
                    stats.AddError()
                    logChan <- fmt.Sprintf("[%s] pacote retornou cards inválidos", playerName)
                    continue
                }

                var cards []map[string]interface{}
                if err := json.Unmarshal([]byte(cardsJSON), &cards); err != nil {
                    stats.AddError()
                    logChan <- fmt.Sprintf("[%s] erro ao decodificar cartas: %v", playerName, err)
                    continue
                }

                for _, card := range cards {
                    id := fmt.Sprintf("%v", card["ID"])
                    rarity := fmt.Sprintf("%v", card["Rarity"])
                    if _, loaded := cardIDs.LoadOrStore(id, true); loaded {
                        logChan <- fmt.Sprintf("ID duplicado detectado: %s", id)
                    }
                    stats.AddCard(rarity)
                }

                // Envia log resumido pro canal
                if j%packagesPerClient == 0 {
                    logChan <- fmt.Sprintf("[%s] abriu pacote %d -> %d cartas", playerName, j+1, len(cards))
                }
            }

            done <- true
        }(i)
    }

    // Espera todos os clientes terminarem
    for i := 0; i < numClients; i++ {
        <-done
    }
    close(logChan)

    // Imprime logs resumidos
    for l := range logChan {
        t.Log(l)
    }

    // Resumo final
    elapsed := time.Since(start).Seconds()
    totalPacks, rarityCounts, totalCards, players, errors := stats.GetStats()

    t.Logf("\n===== RESULTADOS DO TESTE DE PACOTES =====")
    t.Logf("Jogadores simulados: %d", players)
    t.Logf("Total de pacotes abertos: %d", totalPacks)
    t.Logf("Total de cartas geradas: %d", totalCards)
    t.Logf("Erros encontrados: %d", errors)
    t.Logf("Tempo total de execução: %.2f segundos", elapsed)
    t.Logf("Distribuição por raridade:")
    for rarity, count := range rarityCounts {
        t.Logf("  %s -> %d", rarity, count)
    }
    t.Logf("===========================================")
}

func TestOpenPackagesHighConcurrency(t *testing.T) {
    numClients := 5000       // ou 10000
    packagesPerClient := 1

    stats := &utils.PackageStats{
        CardsByRarity: make(map[string]int),
    }
    var cardIDs sync.Map
    done := make(chan bool, numClients)
    start := time.Now()

    for i := 0; i < numClients; i++ {
        go func(i int) {
            playerName := fmt.Sprintf("player_%d", i)
            client, err := utils.NewPackageTestClient(t, playerName)
            if err != nil {
                stats.AddError()
                done <- false
                return
            }
            defer client.Conn.Close()

            // Timeout de leitura para evitar goroutines travadas
            client.Conn.SetReadDeadline(time.Now().Add(5 * time.Second))

            stats.AddPlayer()

            for j := 0; j < packagesPerClient; j++ {
                resp, err := client.OpenPackage()
                if err != nil {
                    stats.AddError()
                    continue
                }

                // Ignora cooldown para teste de stress
                if resp.Message == "Pacote em cooldown" {
                    continue
                }

                stats.AddPack()

                cardsJSON, ok := resp.Data["cards"]
                if !ok {
                    stats.AddError()
                    continue
                }

                var cards []map[string]interface{}
                if err := json.Unmarshal([]byte(cardsJSON), &cards); err != nil {
                    stats.AddError()
                    continue
                }

                for _, card := range cards {
                    id := fmt.Sprintf("%v", card["ID"])
                    rarity := fmt.Sprintf("%v", card["Rarity"])
                    if _, loaded := cardIDs.LoadOrStore(id, true); loaded {
                        t.Logf("ID duplicado detectado: %s", id)
                    }
                    stats.AddCard(rarity)
                }
            }

            done <- true
        }(i)
    }

    // Espera todos terminarem
    for i := 0; i < numClients; i++ {
        <-done
    }

    // Resumo final
    elapsed := time.Since(start).Seconds()
    totalPacks, rarityCounts, totalCards, players, errors := stats.GetStats()

    t.Logf("\n===== RESULTADOS DO TESTE DE PACOTES =====")
    t.Logf("Jogadores simulados: %d", players)
    t.Logf("Total de pacotes abertos: %d", totalPacks)
    t.Logf("Total de cartas geradas: %d", totalCards)
    t.Logf("Erros encontrados: %d", errors)
    t.Logf("Tempo total de execução: %.2f segundos", elapsed)
    t.Logf("Distribuição por raridade:")
    for rarity, count := range rarityCounts {
        t.Logf("  %s -> %d", rarity, count)
    }
    t.Logf("===========================================")
}
