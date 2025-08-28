package screm

import(
	"fmt"
)

type Screm struct{
	Text string

}

func (s *Screm) ShowInitalMenu(){
	fmt.Print("=========Bem Vindo Ao Magic Card===============\n")
	fmt.Print("1 Entrar no Jogo\n")
	fmt.Print("3 Sair\n")
}