package main

import (
	
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
    "strings"
	"jogodecartasonline/test/utils"
)




// TESTE DE CONCORRÊNCIA PARA ABERTURA DOS PACOTES
// N CLIENTES TENTAM ABRIR PACOTES AO MESMO TEMPO 
// VERIFICA SUCESSO DE ABERTURA , DUPLICIDADE DAS CARTAS E DISTRIBUIÇÃO DE RARIDADE
func TestOpenPackagesConcurrency(t *testing.T) {
    numClients := 100      // ou 10000
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

    fmt.Println("\n📊 Checando estatísticas no servidor...")
   
    
    // Dá um tempo para o servidor processar e imprimir
    time.Sleep(500 * time.Millisecond)
    
    t.Log("✅ Teste concluído! Verifique os logs do servidor para estatísticas detalhadas.")
}



// Teste DE ESTOQUE GLOABAL
func TestCardStatistics(t *testing.T) {
	

	numClients := 100
	packagesPerClient := 1
	var wg sync.WaitGroup
	
	

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientIndex int) {
			defer wg.Done()
			
			playerName := fmt.Sprintf("test_player_%d", clientIndex)
			client, err := utils.NewPackageTestClient(t, playerName)
			if err != nil {
				return
			}
			defer client.Conn.Close()
			
			
			for j := 0; j < packagesPerClient; j++ {
				client.OpenPackage()
				
			}
			
		
		}(i)
	}
	
	wg.Wait()
	
	
	
	// Agora busca as estatísticas
	statsClient, err := utils.NewPackageTestClient(t, "stats_collector")
	if err != nil {
		t.Fatalf("Erro ao criar cliente de stats: %v", err)
	}
	defer statsClient.Conn.Close()
	
	resp, err := statsClient.CheckServerStats(t)
	if err != nil {
		t.Fatalf("Erro ao buscar estatísticas: %v", err)
	}
	
	if resp.Status != 200 {
		t.Fatalf("Erro na resposta: %s", resp.Message)
	}
	
	stats := resp.Data
	
	fmt.Println("\n📈 RESULTADO FINAL - CARTAS DISTRIBUÍDAS:")
	
	rarities := []string{"COMMON","UNCOMMON", "RARE", "EPIC", "LEGENDARY"}
	totalSpecial := 0
	exceededLimit := false
	
	for _, rarity := range rarities {
		if countStr, exists := stats[strings.ToLower(rarity)]; exists {
			count, _ := strconv.Atoi(countStr)
			status := "✅"
			if count > 5 && rarity != "COMMON" {
				status = "⚠️ "
				exceededLimit = true
			}
			fmt.Printf("  %s %s: %d\n", status, strings.ToUpper(rarity), count)
			if rarity != "COMMON" {
				totalSpecial += count
			}
		} else {
			fmt.Printf("  ✅ %s: 0\n", strings.ToUpper(rarity))
		}
	}
	
	fmt.Printf("\n🔥 TOTAL DE CARTAS ESPECIAIS: %d\n", totalSpecial)
	
	// Verifica versões criadas
	totalVersions := 0
	if totalVersionsStr, exists := stats["totalVersions"]; exists {
		totalVersions, _ = strconv.Atoi(totalVersionsStr)
	}
	
	
	// ANÁLISE FINAL
	fmt.Println("\n🎯 === ANÁLISE DO SISTEMA DE VERSÕES ===")
	
	if !exceededLimit {
		fmt.Println("ℹ️  Nenhuma raridade passou do limite de 200")
	
	} else {
		fmt.Println("✅ Algumas raridades passaram do limite de 200!")
		
		if totalVersions == 0 {
			t.Error("❌ FALHA: Cartas passaram do limite mas NENHUMA versão foi criada!")
			fmt.Println("🐛 Possível problema no sistema CreateNextVersion")
		} else {
			t.Log("✅ SUCESSO: Sistema de versões está funcionando!")
			fmt.Printf("🎉 %d versões foram criadas automaticamente\n", totalVersions)
		}
	}
	
	fmt.Printf("\n📊 Resumo: %d pacotes geraram %d cartas especiais \n", 
		numClients*packagesPerClient, totalSpecial)
	
	fmt.Println("=== TESTE CONCLUÍDO ===")
}