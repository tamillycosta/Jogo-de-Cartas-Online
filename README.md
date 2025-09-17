# 🌟 MagiCards 🌟

Bem-vindo ao **MagiCards**, um jogo de cartas multiplayer onde a batalha e a fantasia se misturam!  

⚔️ **Regras básicas do jogo**  
- Cada jogador começa com **3 vidas**.  
- A cada carta perdida, o jogador perde **1 vida**.  
- Não existe sistema de ranks, apenas sobrevivência.  
- O último jogador com vidas vence a partida.  

---

## 🚀 Como rodar o projeto

### ✅ Pré-requisitos
- [Go](https://go.dev/dl/)
- [Docker](https://www.docker.com/) 

---

### 🖥️ Rodando o **Servidor** 

1. Clone o repositório:
   ```bash
   https://github.com/tamillycosta/Jogo-de-Cartas-Online.git
   cd Jogo-de-Cartas-Online/
2. Suba o container do serivor:
   ```bash
     docker-compose up --build

### 🎮 Rodando o Cliente (sem Docker)
1. Entre no diretório do cliente:
   ```bash
   cd Jogo-de-Cartas-Online/client
2. Rode o projeto:
   ```bash
     go run .

### 🎮 Rodando o Cliente (Com Docker)
1. Na raiz do projeto rode:
   ```bash
   docker build -t magicards-client -f client/Dockerfile .
2. Rode o projeto:
   ```bash
   docker run -it --rm magicards-client
