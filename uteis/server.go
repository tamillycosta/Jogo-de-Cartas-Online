package uteis

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
	"github.com/google/uuid"
)




func NewServer() *Server {
	return &Server{
		users:     make(map[string]*User),
		waitQueue: make([]*User, 0),
		chatRooms: make(map[string]*ChatRoom),
		userID:   "",
		roomID:    "",
	}
}



func (s *Server) PrintStats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		s.mutex.RLock()
		totalUsers := len(s.users)
		waitingUsers := len(s.waitQueue)
		activeChats := len(s.chatRooms)
		s.mutex.RUnlock()
		
		log.Printf("Stats: %d usuários conectados, %d na fila, %d chats ativos", 
			totalUsers, waitingUsers, activeChats)
	}
}


// Adiciona um user ao server - 
// Cria uma instância de um novo usuário e o insere na lista de users
func (server *Server) addUser(conn net.Conn) *User {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	user := &User{
		ID:     uuid.NewString(),
		Conn:   conn,
		Writer: bufio.NewWriter(conn),
		Reader: bufio.NewReader(conn),
	}
	
	server.users[user.ID] = user
	
	return user
}



// Remove um user do server -
// Caso o user esteja em um chat o segundo user será  notificado e retirado do chat
// e o chat será removido
func (server *Server) removeUser(user *User) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	
	// Se estiver em um chat, notifica o outro usuário antes de retirar
	if user.ChatRoom != nil {
		room := user.ChatRoom
		var otherUser *User
		
		if room.User1.ID == user.ID {
			otherUser = room.User2
		} else {
			otherUser = room.User1
		}
		
		// notifica o outro user do chat
		if otherUser != nil {
			server.sendMessage(otherUser, "SYSTEM: Seu parceiro de chat desconectou. Você foi colocado de volta na fila de espera.\n")
			otherUser.ChatRoom = nil
			server.waitQueue = append(server.waitQueue, otherUser)
		}
		
		delete(server.chatRooms, room.ID)
	}

	
	delete(server.users, user.ID)
	user.Conn.Close()
	server.tryMatchUsers()
}



// Adiciona um Usuario a lista de espera - 
func (server *Server) addToWaitQueue(user *User) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	
	server.waitQueue = append(server.waitQueue, user)
	server.sendMessage(user, fmt.Sprintf("Você está na fila de espera. Posição: %d\n", len(server.waitQueue)))
	
	server.tryMatchUsers()
}



// Tenta gerar a combinação de 2 usuarios (chat) - 
// Verifica se a lista de espera tem ao menos 2 users para formar um chat
func (server *Server) tryMatchUsers() {
	if len(server.waitQueue) >= 2 {
		user1 := server.waitQueue[0]
		user2 := server.waitQueue[1]
		
		// 🔹 Evita chat com o mesmo usuário (bug de duplicação)
		if user1.ID == user2.ID {
			log.Println("Tentativa de criar chat com o mesmo usuário, ignorado.")
			return
		}


		server.waitQueue = server.waitQueue[2:]
		
		// Cria uma nova sala de chat
		room := &ChatRoom{
			ID:    uuid.NewString(),
			User1: user1,
			User2: user2,
		}

		server.chatRooms[room.ID] = room
		
		// Associa os usuários à sala
		user1.ChatRoom = room
		user2.ChatRoom = room
		
		server.sendMessage(user1, fmt.Sprintf("SYSTEM: Você foi conectado com %s! Digite /quit para sair do chat.\n", user2.Name))
		server.sendMessage(user2, fmt.Sprintf("SYSTEM: Você foi conectado com %s! Digite /quit para sair do chat.\n", user1.Name))
		log.Printf("Chat criado: %s e %s na sala %s", user1.Name, user2.Name, room.ID)
	}
}


// Escreve uma menssagem no Buffer de envio
func (s *Server) sendMessage(user *User, message string) {
	user.Writer.WriteString(message)
	user.Writer.Flush()
}



// Cria a troca de menssagem entre dois usuarios no chat - 
// servidor verifica quem é o usuario emissor e a menssagem 
// envia a menssagem para o receptor
func (s *Server) broadcastToRoom(sender *User, message string) {
	if sender.ChatRoom == nil {
		return
	}
	
	// pega o chat no qual o usuario esta
	room := sender.ChatRoom
	var receiver *User
	
	// Verifica quem é o receptor 
	if room.User1.ID == sender.ID {
		receiver = room.User2
	} else {
		receiver = room.User1
	}
	
	if receiver != nil {
		formattedMessage := fmt.Sprintf("%s: %s", sender.Name, message)
		s.sendMessage(receiver, formattedMessage)
	}
}


func (server *Server) handleConnetion(user *User) {
	
	
	name, err := user.Reader.ReadString('\n')

	if err != nil {
		log.Printf("Erro ao ler nome do usuário %s: %v", user.ID, err)
		return
	}
	
	user.Name = strings.TrimSpace(name)
	log.Printf("Usuário %s (ID: %s) conectado", user.Name, user.ID)
	
	server.sendMessage(user, fmt.Sprintf("Bem-vindo, %s! Você será conectado com outro usuário em breve...\n", user.Name))
	
	// Adiciona à fila de espera
	server.addToWaitQueue(user)
	
	// Loop de mensagens
	for {
		message, err := user.Reader.ReadString('\n')
		if err != nil {
			log.Printf("Usuário %s desconectado: %v", user.Name, err)
			break
		}
		
		message = strings.TrimSpace(message)
		
		// Comando para sair do chat atual
		if message == "/quit" {
			server.removeUser(user)
		}
		
		// Se estiver em um chat, envia a mensagem
		if user.ChatRoom != nil {
			server.broadcastToRoom(user, message+"\n")
		} else {
			server.sendMessage(user, "SYSTEM: Você está na fila de espera. Aguarde ser conectado com alguém.\n")
		}
	}
}


func (server *Server) Start(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("Erro ao iniciar servidor:", err)
	}
	defer listener.Close()
	
	log.Printf("Servidor iniciado na porta %s", port)
	log.Println("Comandos disponíveis:")
	log.Println("  /quit - Sair do chat atual")
	log.Println("  /status - Ver estatísticas do servidor")
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Erro ao aceitar conexão: %v", err)
			continue
		}
		
		user := server.addUser(conn)
		go server.handleConnetion(user)
	}
}

