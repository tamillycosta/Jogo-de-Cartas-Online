package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func menu(){
	fmt.Println("Conectado ao servidor de chat!")
	fmt.Println("Comandos disponíveis:")
	fmt.Println("  /quit - Sair do chat atual")
	fmt.Println("  /status - Ver estatísticas do servidor")
	fmt.Println("  /exit - Sair do programa")
	fmt.Println()
}


func main() {
	// Conecta ao servidor
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Erro ao conectar ao servidor:", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	inputReader := bufio.NewReader(os.Stdin)

	fmt.Print("informe seu nome\n")
	var nome string 
	fmt.Scanln(nome)

	
	menu()
	

	writer.WriteString(nome)
	// Goroutine para receber mensagens do servidor
	go func() {
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Conexão perdida com o servidor")
				os.Exit(1)
			}
			fmt.Print(message)
		}
	}()

	// Loop principal para enviar mensagens
	for {
		input, err := inputReader.ReadString('\n')
		if err != nil {
			fmt.Println("Erro ao ler input:", err)
			break
		}

		input = strings.TrimSpace(input)

		if input == "/exit" {
			fmt.Println("Desconectando...")
			break
		}

		// Envia mensagem para o servidor
		writer.WriteString(input + "\n")
		writer.Flush()
	}
}