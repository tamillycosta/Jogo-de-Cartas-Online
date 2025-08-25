package uteis


import (
	"bufio"
	"net"
	"sync"
)



type User struct {
	ID       string
	Name     string
	Conn     net.Conn
	ChatRoom *ChatRoom
	Writer   *bufio.Writer
	Reader   *bufio.Reader
}


type ChatRoom struct {
	ID    string
	User1 *User
	User2 *User
}

type Server struct {
	users     map[string]*User
	waitQueue []*User
	chatRooms map[string]*ChatRoom
	userID    string
	roomID    string
	mutex     sync.RWMutex
}