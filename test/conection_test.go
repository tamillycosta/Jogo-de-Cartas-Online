package main

import (
   
    "fmt"
  "jogodecartasonline/test/utils"
    "sync"
    "testing"
    "time"

   
)



// TESTA CONEXÃO DE UM PLAYER
func TestBasicConnection(t *testing.T) {
    client, err := utils.NewFakeClient(t, "test_player")
    if err != nil {
        t.Fatalf("Erro ao criar cliente: %v", err)
    }
    defer client.Conn.Close()
    
    err = client.FindMatch(t)
    if err != nil {
        t.Fatalf("Erro na busca por partida: %v", err)
    }
    
    fmt.Printf("✅ Teste básico passou\n")
}

// Teste de stress com múltiplas conexões e criação de partidas 
func TestStressConnections(t *testing.T) {
    const numClients = 550 

    var wg sync.WaitGroup
    var mu sync.Mutex
    errors := make([]error, 0)
    successes := 0

    start := make(chan struct{})

    wg.Add(numClients)

    for i := 0; i < numClients; i++ {
        go func(i int) {
            defer wg.Done()

            name := fmt.Sprintf("stress_player_%d", i)

          
            <-start

            client, err := utils.NewFakeClient(t, name)
            if err != nil {
                mu.Lock()
                errors = append(errors, fmt.Errorf("[%s] erro ao conectar: %v", name, err))
                mu.Unlock()
                return
            }
            defer client.Conn.Close()

            // Buscar partida
            err = client.FindMatch(t)
            if err != nil {
                mu.Lock()
                errors = append(errors, fmt.Errorf("[%s] erro ao buscar partida: %v", name, err))
                mu.Unlock()
                return
            }

            // Simula jogando por um tempo aleatório
            time.Sleep(time.Duration(1+i%3) * time.Second)

            mu.Lock()
            successes++
            mu.Unlock()
        }(i)
    }

    // Dispara todos os goroutines ao mesmo tempo
    close(start)

    wg.Wait()

    // 📊 Resultados
    fmt.Printf("\n📊 RESULTADOS DO TESTE DE CONCORRÊNCIA:\n")
    fmt.Printf("✅ Sucessos: %d/%d\n", successes, numClients)
    fmt.Printf("❌ Erros: %d/%d\n", len(errors), numClients)

    if len(errors) > 0 {
        fmt.Printf("\n🔍 DETALHES DOS ERROS:\n")
        for i, err := range errors {
            fmt.Printf("%d. %v\n", i+1, err)
        }
    }

   
    successRate := float64(successes) / float64(numClients)
    if successRate < 1.0 {
        t.Errorf("Taxa de sucesso muito baixa: %.2f%% (esperado: == 100%%)", successRate*100)
    }
}


