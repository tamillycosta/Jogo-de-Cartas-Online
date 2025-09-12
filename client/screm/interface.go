package screm

import (
	
	"fmt"
	response "jogodecartasonline/api/Response"
	"jogodecartasonline/server/game/models"
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
	fmt.Println("════════════════════════════════════════════════")
	fmt.Println("✨ Bem-vindo ao 🌟 MagiCards 🌟")
	fmt.Println("════════════════════════════════════════════════")
	fmt.Println("1️⃣  Entrar no Jogo")
	fmt.Println("2️⃣  Sair")
	fmt.Println("════════════════════════════════════════════════")
}

func (s *Screm) ShowLobbyMenu() {
	fmt.Println("════════════════════════════════════════════════")
	fmt.Println("🏰 LOBBY")
	fmt.Println("════════════════════════════════════════════════")
	fmt.Println("1️⃣  Buscar Partida")
	fmt.Println("2️⃣  Menu de Cartas")
	fmt.Println("3️⃣  Sair do Jogo")
	fmt.Println("════════════════════════════════════════════════")
}

func (s *Screm) ShowGameLoop() {
	fmt.Println("\n════════════════════════════════════════════════")
	fmt.Println("⚔️ PARTIDA")
	fmt.Println("════════════════════════════════════════════════")
	fmt.Println("1️⃣  Escolher carta")
	fmt.Println("2️⃣  Atacar")
	fmt.Println("3️⃣  Sair da partida")
	fmt.Println("════════════════════════════════════════════════")
}

func (s *Screm) ShowPlayerResultCard(playerResult map[string]interface{}) {
	fmt.Println("\n🃏 Você escolheu uma carta!")
	fmt.Printf("✅ %s | ⚔️ %.0f | 💚 %.0f\n",
		playerResult["cardName"],
		playerResult["cardPower"],
		playerResult["cardHealth"])
}

func (s *Screm) ShowPlayerResultAtack(playerResult map[string]interface{}) {
	fmt.Println("\n⚔️ Seu ataque foi lançado!")
	fmt.Printf("   Poder do ataque: %.0f\n", playerResult["attackPower"])
	fmt.Printf("   Vida do oponente: 💔 %.0f | HP da carta inimiga: 💚 %.0f\n",
		playerResult["opponentLife"],
		playerResult["opponentCardHP"])
}

func (s *Screm) ShowOpponentResultCard(opponentResult map[string]interface{}) {
	fmt.Println("\n🃏 Oponente escolheu uma carta!")
	fmt.Printf("❌ %s | ⚔️ %.0f | 💚 %.0f\n",
		opponentResult["cardName"],
		opponentResult["cardPower"],
		opponentResult["cardHealth"])
}

func (s *Screm) ShowOpponentResultAtack(opponentResult map[string]interface{}) {
	fmt.Println("\n💥 Oponente te atacou!")
	fmt.Printf("   Dano recebido: %.0f\n", opponentResult["damageTaken"])
	fmt.Printf("   Sua vida: 💔 %.0f | HP da sua carta: 💚 %.0f\n",
		opponentResult["lifeRemaining"],
		opponentResult["cardHPRemaining"])
}



func (s *Screm) ShowFoundMatchMake(response response.Response) {
	fmt.Println("\n🎉 ================================")
	fmt.Println("       🚨  PARTIDA ENCONTRADA! 🚨")
	fmt.Println("===================================")
	fmt.Printf("🆚 Oponente : %s\n", response.Data["opponent"])
	fmt.Printf("🎯 Match ID : %s\n", response.Data["matchId"])
	fmt.Println("===================================")
}


func (s *Screm) ShowOpponentGameEnd(opponentResult map[string]interface{}) {
	fmt.Println("\n🏆 =================================")
	fmt.Println("           🎖️  RESULTADO 🎖️")
	fmt.Println("====================================")
	fmt.Println(opponentResult["message"])
	fmt.Printf("📊 Seu Score : %d\n", opponentResult["score"])
	fmt.Println("====================================")
}



func (s *Screm) ShowCooldownMessage(totalCards string, remaining string) {
		fmt.Println("📦 === SISTEMA DE PACOTES ===")
		fmt.Printf("📊 Total de cartas: %s\n", totalCards)
		fmt.Printf("⏰ Próximo pacote em: %s\n\n", remaining)
		fmt.Println("1. 🃏 Ver minhas cartas")
		fmt.Println("2. 🔄 Gerenciar deck")
		fmt.Println("3. ⬅️ Voltar ao lobby")
}


func (s *Screm) ShowPackageMenu(totalCards string) {
		fmt.Println("📦 === SISTEMA DE PACOTES ===")
		fmt.Printf("📊 Total de cartas: %s\n\n", totalCards)
		fmt.Println("1. 📦 Abrir pacote")
		fmt.Println("2. 🃏 Ver minhas cartas")
		fmt.Println("3. 🔄 Gerenciar deck")
		fmt.Println("4. ⬅️ Voltar ao lobby")
}


func (s *Screm) ShowOpenPackResult(totalCards  string){
	fmt.Println("✨ === PACOTE ABERTO! ===")
	fmt.Printf("Total de cartas: %s\n\n", totalCards)
	fmt.Println("🎉 Cartas obtidas:")
}


func (s *Screm) GetRarityEmoji(rarity string) string {
	switch rarity {
	case "Common":
		return "⚪"
	case "Rare":
		return "🔵"
	case "Epic":
		return "🟣"
	case "Legendary":
		return "🟡"
	default:
		return "⚪"
	}
}

func (s *Screm) ShowListCards(DeckCards []*models.Card, OtherCards []*models.Card) {
    fmt.Println("══════════════════════════════")
    fmt.Println("🃏 CARTAS DO DECK DE BATALHA")
    fmt.Println("══════════════════════════════")

    if len(DeckCards) == 0 {
        fmt.Println("⚠️ Nenhuma carta no deck!")
    } else {
        for i, card := range DeckCards {
            fmt.Printf("%d) %s | ⚔️ %d  💚 %d\n",
                i+1, card.Nome, card.Power, card.Health)
        }
    }

    fmt.Println("\n══════════════════════════════")
    fmt.Println("📦 CARTAS EM ESTOQUE")
    fmt.Println("══════════════════════════════")

    if len(OtherCards) == 0 {
        fmt.Println("⚠️ Nenhuma carta em estoque!")
    } else {
        for i, card := range OtherCards {
            fmt.Printf("%d) %s | ⚔️ %d  💚 %d\n",
                i+1, card.Nome, card.Power, card.Health)
        }
    }

    fmt.Println("══════════════════════════════")
}
