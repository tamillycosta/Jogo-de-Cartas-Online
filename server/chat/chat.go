package chat


import(
	"sync"
	"fmt"
	"net"

)

type Chat struct{
	mu sync.Mutex
	clients []*CLient
} 


type CLient struct {
	Nome string
	Conn net.Conn
}

func ( chat *Chat) AcceptClients (client *CLient){
	chat.mu.Lock()
	defer chat.mu.Unlock()
	chat.clients = append(chat.clients, client)
}


func (c *CLient) SendMessage() {
    for {
        var msg string
        fmt.Print("Digite algo: ")
        fmt.Scanln(&msg) // lê uma linha do terminal
        _, err := c.Conn.Write([]byte(msg))
        if err != nil {
            fmt.Println("Erro ao enviar mensagem:", err)
            return
        }
    }
}


func (c *CLient) ReciveMessage(){
	buffer := make([]byte, 1024)
	for{
		mensagem, err :=  c.Conn.Read(buffer)
		if(err != nil){
			fmt.Printf("Erro ao ler a messagem")
			return
		}
		fmt.Printf("O usuario %s enviou a menssagem %s",c.Nome, string(buffer[:mensagem]))
	}
}
