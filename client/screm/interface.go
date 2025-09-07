package screm

import (
	"fmt"
	response "jogodecartasonline/api/Response"
	"os"
	"os/exec"
	"runtime"
)

type Screm struct {
	Text string
}

func (*Screm) ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func (s *Screm) ShowInitalMenu() {
	fmt.Print("=============Bem Vindo Ao MagiCards===============\n")
	fmt.Print("1 Entrar no Jogo\n")
	fmt.Print("2 Sair\n")
}

func (s *Screm) ShowLobbyMenu() {
	fmt.Print("=============LOBBY===============\n")
	fmt.Print("1 Buscar Partida\n")
	fmt.Print("2 Abrir Pacote\n")
	fmt.Print("3 Sair Do Jogo\n")
}

func (s *Screm) ShowGameLoop() {
	fmt.Println("\n===========PARTIDA===============")
	fmt.Println("1. Escolher carta")
	fmt.Println("2. Atacar")
	fmt.Println("3. Passar vez")
	fmt.Println("4. Sair da partida")
}

func (s *Screm) ShowPlayerResultCard(playerResult map[string]interface{}) {

	fmt.Println("\n🃏 Oponente escolheu uma carta")
	fmt.Printf("\n✅ Carta escolhida: %s (Poder: %.0f, Vida: %.0f)\n",
		playerResult["cardName"],
		playerResult["cardPower"],
		playerResult["cardHealth"])
}

func (s *Screm) ShowPlayerResultAtack(playerResult map[string]interface{}) {
	attackPower := playerResult["attackPower"]
	opponentLife := playerResult["opponentLife"]
	opponentCardHP := playerResult["opponentCardHP"]
	fmt.Printf("\n⚔️ Ataque realizado! Poder: %.0f\n", attackPower)
	fmt.Printf("   Vida do oponente: %.0f | Vida da carta: %.0f\n",
		opponentLife, opponentCardHP)
}

func (s *Screm) ShowOpponentResultCard(opponenteResult map[string]interface{}) {
	fmt.Println("\n🃏 Oponente escolheu uma carta")
	fmt.Printf("\n✅ Carta escolhida: %s (Poder: %.0f, Vida: %.0f)\n",
		opponenteResult["cardName"],
		opponenteResult["cardPower"],
		opponenteResult["cardHealth"])
}

func (s *Screm) ShowOpponentResultAtack(opponentResult map[string]interface{}) {
	fmt.Printf("\n💥 Oponente te atacou! Dano recebido: %.0f\n", opponentResult["damageTaken"])
	fmt.Printf("   Sua vida: %.0f | Vida da sua carta: %.0f\n",
		opponentResult["lifeRemaining"],
		opponentResult["cardHPRemaining"])
}

func (s *Screm) ShowFoundMatchMake(response response.Response) {
	fmt.Println("🎉 === PARTIDA ENCONTRADA! ===")
	fmt.Printf("🆚 Oponente: %s\n", response.Data["opponent"])
	fmt.Printf("🎯 Match ID: %s\n", response.Data["matchId"])
}

func (s *Screm) ShowOpponentGameEnd(opponentResult map[string]interface{}) {
	fmt.Println("🏆 ==== VOCÊ GANHOU! ====")
	fmt.Print(opponentResult["message"], "\n")
	fmt.Printf("Seu score : %d\n", opponentResult["score"])
}

func (s *Screm) ShowCooldownMessage(totalCards int, remaining string) {
	fmt.Println("⏰ === PACOTE EM COOLDOWN ===")
	fmt.Printf("Total de cartas: %d\n", totalCards)
	fmt.Printf("Próximo pacote em: %s\n", remaining)
	fmt.Println("\n1. Ver Minhas Cartas")
	fmt.Println("2. Voltar ao Lobby")
}

func (s *Screm) ShowPackageMenu(totalCards int) {
	fmt.Println("📦 === PACOTES DISPONÍVEIS ===")
	fmt.Printf("Total de cartas: %d\n", totalCards)
	fmt.Println("\n1. Abrir Pacote (5 cartas)")
	fmt.Println("2. Ver Minhas Cartas")
	fmt.Println("3. Voltar ao Lobby")
}
