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
    Round *Round
    mu    sync.Mutex
}




type Server struct {
	users     map[string]*User
	waitQueue []*User
	chatRooms map[string]*ChatRoom
	
	mutex     sync.RWMutex
}

type Round struct{
	ID 		int
	Sender  *User
	
}