# Life Counter Feature Spec — Magic Collector

## Objetivo

Adicionar uma nova aba chamada **Pontuação** ou **Life Counter** ao Magic Collector.

Essa feature deve funcionar como um marcador de jogo de Magic: The Gathering para partidas casuais e Commander.

Ela deve permitir:

* criar sessão de jogo;
* adicionar jogadores;
* controlar vida;
* controlar dano de comandante;
* controlar marcadores de tóxico;
* identificar jogadores eliminados;
* salvar sessões no banco;
* restaurar sessões em andamento;
* encerrar sessões;
* listar sessões anteriores.

---

## Escopo

Implementar backend, banco e frontend para sessões de pontuação.

Arquivos prováveis:

* `backend/internal/database/mysql.go`
* `backend/internal/game_sessions/`
* `backend/cmd/api/main.go`
* `frontend/src/App.jsx`
* `frontend/src/App.css`
* `frontend/src/services/api.js`

---

## Regras de Magic consideradas

### Vida

* Cada jogador possui vida total.
* Valor inicial padrão:

  * Commander: 40
  * Casual/Outros: 20
* Se a vida chegar a 0 ou menos, o jogador perde.

### Dano de comandante

* No Commander, se um jogador receber 21 ou mais pontos de dano de comandante do mesmo comandante/jogador, ele perde.
* O sistema deve mostrar claramente quando um jogador foi eliminado por dano de comandante.
* Ao atingir 21 de dano de comandante, exibir uma caveira no jogador eliminado.

### Tóxico / Poison

* Se um jogador atingir 10 marcadores de poison/toxic, ele perde.
* Exibir estado visual de eliminado.

### Jogadores

* Mínimo: 2 jogadores.
* Máximo: 8 jogadores.
* Cada jogador deve ter:

  * nome;
  * sigla de até 3 caracteres;
  * vida;
  * dano de comandante recebido;
  * marcadores tóxicos;
  * status eliminado ou ativo.

---

## Banco de dados

Criar tabelas para sessões.

### `game_sessions`

Campos sugeridos:

```sql
id BIGINT AUTO_INCREMENT PRIMARY KEY,
name VARCHAR(255) NOT NULL,
format VARCHAR(50) DEFAULT 'Commander',
status VARCHAR(20) DEFAULT 'active',
starting_life INT DEFAULT 40,
created_at DATETIME,
updated_at DATETIME,
ended_at DATETIME NULL
```

Status possíveis:

```text
active
finished
```

Jogos com status `finished` não podem ser alterados.

---

### `game_session_players`

Campos sugeridos:

```sql
id BIGINT AUTO_INCREMENT PRIMARY KEY,
session_id BIGINT NOT NULL,
name VARCHAR(255) NOT NULL,
short_code VARCHAR(3) NOT NULL,
life INT NOT NULL,
poison INT DEFAULT 0,
commander_damage_received INT DEFAULT 0,
is_eliminated TINYINT(1) DEFAULT 0,
eliminated_reason VARCHAR(50) NULL,
created_at DATETIME,
updated_at DATETIME
```

`eliminated_reason` pode ser:

```text
life
commander_damage
poison
manual
```

---

## API

Criar rotas:

```http
GET    /game-sessions
POST   /game-sessions
GET    /game-sessions/:id
PUT    /game-sessions/:id
DELETE /game-sessions/:id

POST   /game-sessions/:id/players
PATCH  /game-sessions/:id/players/:player_id
DELETE /game-sessions/:id/players/:player_id

POST   /game-sessions/:id/reset
POST   /game-sessions/:id/finish
POST   /game-sessions/:id/restore
```

### POST /game-sessions

Body:

```json
{
  "name": "Commander sexta-feira",
  "format": "Commander",
  "starting_life": 40,
  "players": [
    { "name": "Gabriel", "short_code": "GAB" },
    { "name": "João", "short_code": "JOA" }
  ]
}
```

Validações:

* mínimo 2 jogadores;
* máximo 8 jogadores;
* `short_code` máximo 3 caracteres;
* vida inicial obrigatória;
* se formato for Commander, vida padrão 40;
* se formato for Casual, vida padrão 20.

---

## Atualização de pontuação

A pontuação deve ser fácil de alterar.

Cada jogador deve ter botões rápidos:

### Vida

* `+1`
* `-1`
* `+5`
* `-5`

### Dano de comandante

* `+1`
* `-1`
* `+5`
* `-5`

### Tóxico

* `+1`
* `-1`

Os principais campos sempre visíveis:

* vida total;
* dano de comandante.

Tóxico também deve estar acessível, mas pode ocupar menos destaque visual.

---

## Regras automáticas de eliminação

Sempre que atualizar pontuação:

* se `life <= 0`, marcar eliminado por vida;
* se `commander_damage_received >= 21`, marcar eliminado por dano de comandante;
* se `poison >= 10`, marcar eliminado por poison.

Quando eliminado:

* mostrar caveira;
* destacar visualmente o jogador;
* bloquear ou reduzir destaque dos controles;
* manter dados visíveis.

Importante:

* jogador eliminado por dano de comandante deve mostrar caveira.
* jogador com vida 0 ou menos também deve mostrar caveira.
* jogador eliminado não deve sumir da sessão.

---

## Encerramento de sessão

Ao encerrar jogo:

* alterar `status` para `finished`;
* preencher `ended_at`;
* bloquear alterações futuras;
* manter sessão visível na listagem;
* exibir como encerrada.

Jogos encerrados não podem ser editados.

---

## Reset de sessão

A opção resetar deve:

* voltar todos os jogadores para a vida inicial;
* zerar dano de comandante;
* zerar tóxico;
* remover status eliminado;
* manter os mesmos jogadores.

Não permitir reset se sessão estiver encerrada.

---

## Frontend

Adicionar nova aba:

```text
Coleção | Decks | Battles | Wishlist | Pontuação
```

A tela deve conter:

* criação de nova sessão;
* listagem de sessões;
* abertura/restauração de sessão ativa;
* painel de pontuação da sessão atual;
* botão resetar;
* botão encerrar jogo;
* opção de adicionar jogador;
* opção de remover jogador antes ou durante sessão ativa, respeitando mínimo 2.

---

## Layout visual

A tela de pontuação deve ter visual de mesa de jogo.

Direção visual:

* cards grandes por jogador;
* vida em destaque;
* dano de comandante sempre visível;
* caveira para eliminado;
* botões grandes e fáceis de tocar;
* boa leitura em celular;
* layout bonito para iPad em mesa;
* desktop com grid organizado.

Sugestão visual:

* fundo dark;
* cards com bordas por jogador;
* jogador ativo em destaque;
* eliminado com opacidade reduzida;
* caveira grande no card;
* badges para Commander Damage e Toxic.

---

## Card de jogador

Cada card deve exibir:

* nome;
* sigla;
* vida atual;
* dano de comandante;
* tóxico;
* status;
* motivo da eliminação, se houver;
* controles de incremento/decremento.

Exemplo visual conceitual:

```text
[GAB] Gabriel        ☠ se eliminado

Vida
40
[-5] [-1] [+1] [+5]

Comandante
0 / 21
[-5] [-1] [+1] [+5]

Tóxico
0 / 10
[-1] [+1]
```

---

## Responsividade

Mobile:

* um jogador por linha;
* botões grandes;
* controles empilhados;
* sessão atual sempre clara;
* evitar excesso de informação escondida.

Tablet/iPad:

* grid de 2 jogadores por linha;
* ótimo para usar na mesa;
* botões grandes para toque.

Desktop:

* grid de 3 ou 4 jogadores por linha;
* painel de sessão no topo;
* listagem de sessões em área separada.

---

## Listagem de sessões

Cada sessão deve mostrar:

* nome;
* formato;
* status;
* jogadores;
* data de criação;
* data de encerramento, se houver;
* botão restaurar/abrir;
* botão encerrar se ativa;
* botão excluir, se existir.

Sessões encerradas:

* devem ficar visíveis;
* não podem ser alteradas;
* podem ser abertas em modo leitura.

---

## Critérios de aceite

* Existe aba Pontuação/Life Counter.
* É possível criar sessão com 2 a 8 jogadores.
* Cada jogador tem sigla de até 3 caracteres.
* Vida inicial respeita formato.
* Vida total e dano de comandante ficam sempre visíveis.
* É fácil adicionar e remover vida.
* É fácil adicionar dano de comandante.
* É possível controlar tóxico.
* Ao chegar a 0 de vida, jogador é eliminado.
* Ao chegar a 21 de dano de comandante, jogador é eliminado.
* Ao chegar a 10 de tóxico, jogador é eliminado.
* Jogador eliminado mostra caveira.
* É possível resetar sessão ativa.
* É possível encerrar sessão.
* Sessão encerrada não pode ser alterada.
* Sessões são salvas no banco.
* É possível restaurar sessão em andamento.
* Existe listagem de sessões.
* Layout funciona bem em mobile, iPad e desktop.
* Nenhum botão fica sobreposto.
* Não há scroll horizontal acidental.
