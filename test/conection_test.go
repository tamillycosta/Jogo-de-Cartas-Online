package main

import (
   
    "fmt"
  "jogodecartasonline/test/utils"
    "sync"
    "testing"
    "time"

   
)






// TESTA CONXÃO DE N PLAYERS
func TestStressLogin(t *testing.T) {
	totalClients := 10000
	var wg sync.WaitGroup
	wg.Add(totalClients)

	start := time.Now()

	for i := 0; i < totalClients; i++ {
		go func(id int) {
			defer wg.Done()
            name := fmt.Sprintf("stress_player_%d", i)
			client, err := utils.NewFakeClient(t, name)
			
			if err != nil {
				t.Errorf("Cliente %d falhou na conexão: %v", id, err)
				return
			}
			 defer client.Conn.Close()
			
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	t.Logf("Finalizado %d logins em %.4f segundos", totalClients, elapsed.Seconds())
	
}


// TESTA CRIAÇÃO DE PARTIDAS PARA N PLAYERS
func TestStressMatchmaking2(t *testing.T) {
	totalClients := 10000
	var wg sync.WaitGroup
	wg.Add(totalClients)

	start := time.Now()
	var mu sync.Mutex
	failures := 0

	for i := 0; i < totalClients; i++ {
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("stress_player_%d", id)

			client, err := utils.NewFakeClient(t, name)
			if err != nil {
				mu.Lock()
				failures++
				mu.Unlock()
				return
			}
			defer client.Conn.Close()

			if err := client.FindMatch(&testing.T{}); err != nil {
				mu.Lock()
				failures++
				mu.Unlock()
			}

			
			if id%1000 == 0 {
				elapsed := time.Since(start).Seconds()
				t.Logf("[Progresso] %d clientes processados em %.2f segundos", id, elapsed)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start).Seconds()
	media := elapsed / float64(totalClients)

	t.Logf("===== RESULTADOS DO TESTE MATCHMAKING =====")
	t.Logf("Total de clientes: %d", totalClients)
	t.Logf("Tempo total: %.2f segundos", elapsed)
	t.Logf("Tempo médio por cliente: %.6f segundos", media)
	t.Logf("Falhas: %d (%.2f%%)", failures, (float64(failures)/float64(totalClients))*100)
	t.Logf("===========================================")
}


