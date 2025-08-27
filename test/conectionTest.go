package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"
)

func simulateClient(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Erro ao conectar:", err)
		return
	}
	defer conn.Close()

	// Enviar login/entrada
	fmt.Fprintf(conn, "USER%d entrou na sala\n", id)

	// Simula algumas jogadas
	for i := 0; i < 3; i++ {
		fmt.Fprintf(conn, "USER%d jogou carta %d\n", id, i)
		time.Sleep(500 * time.Millisecond)
	}

	// Ler resposta do servidor (se tiver)
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		fmt.Println("Servidor respondeu:", scanner.Text())
	}

	// Sair
	fmt.Fprintf(conn, "/quit\n")

}

func main() {
	var wg sync.WaitGroup
	numClients := 50 // número de usuários simulados

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go simulateClient(i, &wg)
		time.Sleep(100 * time.Millisecond) // escalonar entrada
	}

	wg.Wait()
	fmt.Println("Teste de carga finalizado.")
}
