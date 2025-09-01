package screm

import(
	"fmt"
)

type Screm struct{
	Text string

}

func (s *Screm) ShowInitalMenu(){
	fmt.Print("=============Bem Vindo Ao Magic Card===============\n")
	fmt.Print("1 Entrar no Jogo\n")
	fmt.Print("2 Sair\n")
}


func (s *Screm) ShowLobbyMenu(){
	fmt.Print("=============LOBBY===============\n")
	fmt.Print("1 Buscar Partida\n")
	fmt.Print("2 Sair Do Jogo\n")
}

func (s *Screm) ShowGameLoop(){
	fmt.Println("\n=== SEU TURNO ===")
        fmt.Println("1. Escolher carta")
        fmt.Println("2. Atacar")
        fmt.Println("3. Passar vez")
        fmt.Println("4. Sair da partida")
}