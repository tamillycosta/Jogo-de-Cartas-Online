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



// Menus ----------------------------------------------------------

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
	fmt.Println("3️⃣  Ping")
	fmt.Println("4️⃣  Regras do Jogo")
	fmt.Println("5️⃣  Sair do Jogo")
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


func (s *Screm) ShowDeckManagementMenu(){

	fmt.Println("\n════════════════════════════════════════════════")
	fmt.Println("⚔️ GERENCIAR DECK")
	fmt.Println("════════════════════════════════════════════════")
	fmt.Println("1️⃣ Trocar carta do deck")
	fmt.Println("2️⃣  Voltar ao lobby")
	fmt.Println("════════════════════════════════════════════════")

}


func (s *Screm) ShowCooldownMessage(totalCards string, remaining string) {
	fmt.Println("\n════════════════════════════════════════════════")
	fmt.Println("🃏 SISTEMA DE PACOTES")
	fmt.Println("════════════════════════════════════════════════")
	fmt.Printf("📊 Total de cartas: %s\n", totalCards)
	fmt.Printf("⏰   Próximo pacote em: %s\n\n", remaining)
	fmt.Println("1️⃣  Ver minhas cartas")
	fmt.Println("2️⃣  Gerenciar deck")
	fmt.Println("3️⃣. Voltar ao lobby")
}


func (s *Screm) ShowPackageMenu(totalCards string) {
	fmt.Println("\n════════════════════════════════════════════════")
	fmt.Println("🃏 SISTEMA DE PACOTES")
	fmt.Println("════════════════════════════════════════════════")
	fmt.Printf("📊 Total de cartas: %s\n\n", totalCards)
	fmt.Println("1️⃣  Abrir pacote")
	fmt.Println("2️⃣  Ver minhas cartas")
	fmt.Println("3️⃣  Gerenciar deck")
	fmt.Println("4️⃣. Voltar ao lobby")
}



func (s *Screm) ShowDeckManagementInitial(){

	fmt.Println("\n════════════════════════════════════════════════")
	fmt.Println("⚔️ GERENCIAR DECK")
	fmt.Println("════════════════════════════════════════════════")
}
// Staus da Partida ------------------------------------------------

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
	fmt.Printf(" %s | ⚔️ %.0f | 💚 %.0f\n",
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
	fmt.Println("       🚨 PARTIDA ENCONTRADA! 🚨")
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

func (s *Screm) ShowplayerGameEnd(playerResult map[string]interface{} ){
	fmt.Println("\n══════════════════════════════")
fmt.Println(" 🏆 	  VOCÊ VENCEU!!! 🎉 ")
fmt.Println("══════════════════════════════")
fmt.Println("\n══════════════════════════════")
fmt.Println(" 🏆 	   SEU SCORE!!!  🏆 ")
fmt.Println("══════════════════════════════")
fmt.Printf("score %d", playerResult["score"] )
}

// Sistema de pacotes --------------------------------------------------


func (s *Screm) ShowOpenPackResult(totalCards  string){
	fmt.Println("✨ === PACOTE ABERTO! ===")
	fmt.Printf("Total de cartas: %s\n\n", totalCards)
	fmt.Println("🎉 Cartas obtidas:")
}
 

func (s *Screm) GetRarityEmoji(rarity string) string {
	switch rarity {
	case "UNCOMON":
		return "🟢"
	case "RARE":
		return "🔵"
	case "EPIC":
		return "🟣"
	case "LEGENDARY":
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
			rarity := s.GetRarityEmoji(card.Rarity)
            fmt.Printf("%d) %s | ⚔️ %d  💚 %d  %s\n",
                i+1, card.Nome, card.Power, card.Health, rarity)
        }
    }

    fmt.Println("\n══════════════════════════════")
    fmt.Println("📦 CARTAS EM ESTOQUE")
    fmt.Println("══════════════════════════════")

    if len(OtherCards) == 0 {
        fmt.Println("⚠️ Nenhuma carta em estoque!")
    } else {
        for i, card := range OtherCards {
			rarity := s.GetRarityEmoji(card.Rarity)
            fmt.Printf("%d) %s | ⚔️ %d  💚 %d  %s\n",
                i+1, card.Nome, card.Power, card.Health, rarity)
        }
    }



    fmt.Println("══════════════════════════════")
}


func (s *Screm) ShowNewCard(data map[string]string) {

	fmt.Println("\n✨ ==== DECK ATUALIZADO! ====")
	fmt.Printf("🔄 Carta removida: %s\n", data["removed"])
	fmt.Printf("🆕 Nova carta adicionada: %s\n", data["added"])
	fmt.Printf("   Poder: %s | Vida: %s | Raridade: %s\n",
		data["newCard-Power"],
		data["newCard-Life"],
		data["newCard-rarity"])
	fmt.Println("✅ Deck atualizado com sucesso!")
}



func (s *Screm) ShowConfirmChange(selectedOldCard models.Card , selectedNewCard models.Card, oldCardIndex int , newCardIndex int ){
		// Confirmação
		fmt.Printf("\n🔄 CONFIRMAR TROCA:\n")
		fmt.Printf("   ➖ Remover: %s (Índice: %d)\n", selectedOldCard.Nome, oldCardIndex)
		fmt.Printf("   ➕ Adicionar: %s (Índice: %d)\n", selectedNewCard.Nome, newCardIndex)
		fmt.Print("\nConfirmar? (s/N): ")
}



func (s *Screm) ShowListCard(cards []*models.Card){
	for i, card := range cards {
		rarity := s.GetRarityEmoji(card.Rarity)
		fmt.Printf("   %d. %s %s (⚔️%d 💚%d) \n",
			i+1, rarity, card.Nome, card.Power, card.Health,)
	}
}

func (s *Screm) ShowGameIntro() {
    fmt.Println("══════════════════════════════════════════════")
    fmt.Println("🌟 Bem-vindo ao 🌟")
    fmt.Println("🎴 M A G I C A R D S 🎴")
    fmt.Println("══════════════════════════════════════════════")
    fmt.Println("\n📖 Era uma vez...")
    fmt.Println("Em um mundo onde a batalha e a fantasia se misturam,")
    fmt.Println("os maiores mestres duelam em busca da glória suprema!")
    fmt.Println("\n⚔️ Suas cartas são suas armas, sua estratégia é seu poder.")
    fmt.Println("Colecione, monte seu deck e prove seu valor em combate.")
    fmt.Println("\n📜 REGRAS DO JOGO:")
    fmt.Println("1️⃣ Cada jogador começa com 3 vidas ❤️❤️❤️")
    fmt.Println("2️⃣ Cada carta perdida → custa 1 vida")
    fmt.Println("3️⃣ Ao perder todas as vidas, o jogador é derrotado ☠️")
    fmt.Println("4️⃣ Não existe sistema de pontuação, apenas vitória ou derrota")
    fmt.Println("\n✨ Prepare-se, duelista... sua jornada começa agora!")
    fmt.Println("══════════════════════════════════════════════")
}
