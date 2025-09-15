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
func TestOpenPackages(t *testing.T) {
    numClients := 100          
    packagesPerClient := 1   
    var wg sync.WaitGroup

    stats := &utils.PackageStats{
        CardsByRarity: make(map[string]int),
    }

    var cardIDs sync.Map // usado para verificar duplicação de IDs

    for i := 0; i < numClients; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()

            playerName := fmt.Sprintf("player_%d", i)
            client, err := utils.NewPackageTestClient(t, playerName)
            if err != nil {
                t.Errorf("falha ao criar client %s: %v", playerName, err)
                stats.AddError()
                return
            }
            defer client.Conn.Close()

            stats.AddPlayer()

            for j := 0; j < packagesPerClient; j++ {
                resp, err := client.OpenPackage()
                if err != nil {
                    t.Errorf("[%s] erro ao abrir pacote: %v", playerName, err)
                    stats.AddError()
                    continue
                }

                if resp.Message == "Pacote em cooldown" {
                    fmt.Printf("seu [%s] Pacote ainda não esta pronto", playerName)
                    time.Sleep(1 * time.Minute)
                    continue
                }

                stats.AddPack()

                // decodifica o JSON das cartas
                cardsJSON, ok := resp.Data["cards"]
                if !ok {
                    t.Errorf("[%s] pacote retornou cards inválidos", playerName)
                    stats.AddError()
                    continue
                }

                var cards []map[string]interface{}
                if err := json.Unmarshal([]byte(cardsJSON), &cards); err != nil {
                    t.Errorf("[%s] erro ao decodificar cartas: %v", playerName, err)
                    stats.AddError()
                    continue
                }

                for _, card := range cards {
                    id := fmt.Sprintf("%v", card["ID"])
                    rarity := fmt.Sprintf("%v", card["Rarity"])

                    // verifica duplicação de ID
                    if _, loaded := cardIDs.LoadOrStore(id, true); loaded {
                        t.Errorf("ID duplicado detectado: %s", id)
                    }

                    stats.AddCard(rarity)
                }

                // log do progresso
                t.Logf("[%s] abriu pacote %d -> %d cartas recebidas", playerName, j+1, len(cards))
            }
        }(i)
    }

    // espera todas as goroutines terminarem
    wg.Wait()

    totalPacks, rarityCounts, totalCards, players, errors := stats.GetStats()

    t.Logf("\n===== RESULTADOS DO TESTE DE PACOTES =====")
    t.Logf("Jogadores simulados: %d", players)
    t.Logf("Total de pacotes abertos: %d", totalPacks)
    t.Logf("Total de cartas geradas: %d", totalCards)
    t.Logf("Erros encontrados: %d", errors)
    t.Logf("Distribuição por raridade:")
    for rarity, count := range rarityCounts {
        t.Logf("  %s -> %d", rarity, count)
    }



}
