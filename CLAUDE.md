# Magic Collector — CLAUDE.md

Gerenciador de coleção de cartas Magic: The Gathering. Backend em Go + frontend em React, integrado com a Scryfall API.

## Stack

| Camada    | Tecnologia                                                    |
|-----------|---------------------------------------------------------------|
| Backend   | Go 1.25+, Gin, `database/sql` + `go-sql-driver/mysql`        |
| Frontend  | React 18, Vite, JSX puro (sem TypeScript), `xlsx` para export |
| Banco     | MySQL (externo, via env vars)                                 |
| Deploy    | Docker Compose + Terraform (AWS EC2)                          |
| API MTG   | Scryfall REST API (sem chave, rate-limit gentil)              |

## Estrutura de pastas

```
magic-collector/
├── backend/
│   ├── cmd/
│   │   ├── api/main.go        # entrypoint do servidor Gin
│   │   ├── import/main.go     # importador CSV standalone
│   │   └── testmtg/main.go    # utilitário de teste da Scryfall
│   └── internal/
│       ├── cards/             # handler + service + repository + model
│       ├── decks/             # idem para decks
│       ├── battles/           # idem para batalhas
│       ├── importer/          # importação de pré-cons e deck lists
│       ├── mtgapi/client.go   # cliente Scryfall
│       └── database/mysql.go  # conexão MySQL + auto-migrações CREATE TABLE / ALTER TABLE
├── frontend/
│   └── src/
│       ├── App.jsx            # componente único (formulário + lista + modais)
│       ├── App.css            # tema dark gold inspirado em MTG
│       └── services/api.js    # funções fetch para o backend
├── data/                      # volume Docker (não versionado)
├── docker-compose.yml         # dev
├── docker-compose.prod.yml    # prod
├── .env                       # credenciais do banco (nunca commitar)
└── terraform/                 # infra AWS EC2
```

## Banco de dados

O arquivo `internal/database/mysql.go` cria as tabelas automaticamente na inicialização e adiciona colunas novas via `ALTER TABLE` seguro (ignora erro se já existir). **Nunca use migrações externas** — adicione ao `Open()` em `mysql.go`.

### Tabelas principais

**`cards`** — carta individual na coleção:
- `mtg_id` — UUID Scryfall (cache; evita rebusca)
- `set_code` + `collection_number` — chave de busca na Scryfall
- `foil`, `prerelease`, `commander` — flags booleanas (TINYINT 0/1)
- `deck_id` — FK soft para `decks.id` (0 = sem deck)
- `colors` — JSON array como TEXT (ex: `["W","U"]`)
- `color` — string legível em PT (ex: `"Branco/Azul"`)
- `condition` — `mint | near_mint | played | damaged`
- `rarity` — `L | C | U | R | M | T`

**`decks`** — agrupador de cartas:
- `colors` — códigos separados por vírgula (ex: `"W,U,B"`)
- `set_code` — usado para buscar ícone via Scryfall Sets API
- `icon_uri` — SVG cacheado após primeira busca
- `theme_color` — ID de tema visual (ex: `"sapphire"`)

**`battles`** — histórico de partidas:
- `result` — `win | loss | draw`
- `game_style` — `Commander | Standard | Modern | ...`
- `deck_is_mine` — se o deck usado era do próprio usuário

## Rotas da API

```
GET    /health
GET    /cards               ?q=&page=&page_size=&sort=&order=&deck_id=
GET    /cards/export        retorna todas as cartas (sem paginação)
POST   /cards
GET    /cards/:id           retorna { local: Card, external: ExternalCard? }
PUT    /cards/:id
PATCH  /cards/:id/deck      body: { deck_id: int }
DELETE /cards/:id

GET    /decks
POST   /decks
PUT    /decks/:id
DELETE /decks/:id
PATCH  /decks/:id/icon      busca ícone do set no Scryfall e salva
POST   /decks/import-precon body: { set_code, deck_name, language, ... }
POST   /decks/import-list   body: { deck_list: string, deck_name, ... }

GET    /battles
POST   /battles
DELETE /battles/:id
```

## Cliente Scryfall (`internal/mtgapi/client.go`)

Métodos disponíveis:

| Método              | Uso                                                     |
|---------------------|---------------------------------------------------------|
| `Search`            | Busca por set_code + collection_number (+ idioma)       |
| `SearchByName`      | Busca por nome exato (usado no import-list)             |
| `SearchPreRelease`  | Busca cartas de pré-release por nome (`is:prerelease`)  |
| `GetByMTGID`        | Busca pelo UUID Scryfall (cache)                        |
| `FetchSetCards`     | Busca todas as cartas de um set (paginando)             |
| `GetSetByCode`      | Busca metadados de um set (nome, ícone, data)           |

**Rate limiting:** Scryfall pede máximo ~10 req/s. O cliente usa `time.Sleep(75ms)` entre buscas em lote e `100ms` entre páginas. Não remova esses sleeps.

**Idiomas Scryfall:** PT→`pt`, JP→`ja`, ES→`es`, FR→`fr`, DE→`de`. Veja `toLangCode()`.

## Domínio Magic: The Gathering

### Cores das cartas (WUBRG)
| Código | Nome (EN) | Nome (PT) | Ícone no frontend |
|--------|-----------|-----------|-------------------|
| W      | White     | Branco    | `white.svg`       |
| U      | Blue      | Azul      | `blue.svg`        |
| B      | Black     | Preto     | `black.svg`       |
| R      | Red       | Vermelho  | `red.svg`         |
| G      | Green     | Verde     | `green.svg`       |
| C      | Colorless | Incolor   | `incolour.svg`    |

Cartas multicoloridas (ouro/dourado) têm múltiplos códigos. Cartas incolores (ex: artefatos genéricos) têm `C`.

### Raridades
| Letra | Raridade | Cor no CSS  |
|-------|----------|-------------|
| L     | Land     | `--r-l` verde claro |
| C     | Common   | `--r-c` cinza       |
| U     | Uncommon | `--r-u` azul/prata  |
| R     | Rare     | `--r-r` dourado     |
| M     | Mythic   | `--r-m` laranja     |
| T     | Token    | `--r-t` roxo        |

### Códigos de set comuns
`KLD` Kaladesh · `GRN` Guilds of Ravnica · `DMU` Dominaria United · `BRO` The Brothers' War · `MKM` Murders at Karlov Manor · `OTJ` Outlaws of Thunder Junction · `DSK` Duskmourn · `FDN` Foundations · `NCC` New Capenna Commander

O código do set vai sempre para `set_code` no banco e é 2–5 letras maiúsculas. A Scryfall aceita minúsculas na URL.

### Formatos de jogo
- **Commander** — 100 cartas singleton, 4 jogadores, carta comandante define identidade de cor
- **Standard** — apenas os últimas ~2 anos de sets
- **Modern / Pioneer / Legacy** — pool de cartas mais amplo

### Cartas especiais
- **Foil** — tratamento de impressão com brilho; geralmente mais valioso
- **Pré-release** — distribuídas em eventos antes do lançamento oficial; têm carimbo especial e set code diferente (ex: `pgrn` em vez de `grn`)
- **Token** — não é uma carta de jogo; entra no banco com rarity `T`
- **Double-faced cards (DFC)** — cartas com dois lados; a Scryfall retorna `card_faces` em vez de `image_uris` no topo

### Scryfall: campos localizados
Quando buscada com idioma não-EN, a Scryfall pode retornar:
- `printed_name` — nome traduzido
- `printed_text` — texto de habilidade traduzido
- `printed_type_line` — linha de tipo traduzida

O frontend exibe `printed_name` quando disponível, caindo de volta para `name`.

## Padrões do projeto

### Adicionar um campo novo a `cards`
1. `internal/database/mysql.go` → adicionar `db.Exec("ALTER TABLE cards ADD COLUMN ...")`
2. `internal/cards/model.go` → adicionar campo em `Card`, `CreateCardInput`, `UpdateCardInput`
3. `internal/cards/repository.go` → incluir o campo em `INSERT`, `UPDATE` e `SELECT`
4. `internal/cards/handler.go` → nenhuma mudança normalmente (faz bind automático)
5. `frontend/src/App.jsx` → adicionar campo em `EMPTY_FORM` e nos JSX de formulário/edição
6. `frontend/src/services/api.js` → incluir campo no payload, se necessário

### Padrão Go: handler → service → repository
Cada domínio (`cards`, `decks`, `battles`) segue o mesmo padrão:
- **Handler** — parse de request, chama service, escreve JSON
- **Service** — regras de negócio, chama repository e mtgapi
- **Repository** — SQL puro com `database/sql`, sem ORM
- **Model** — structs Go com tags `json`

### Frontend: estado único em App.jsx
Toda a UI é um único componente React. Estado gerenciado com `useState`/`useEffect`. Não há roteamento nem Redux. Para mudar de "tela" usa-se `activeTab` (collection / decks / battles) e modais controlados por estado booleano.

## Comandos de desenvolvimento

```bash
# Subir o ambiente completo
docker compose up --build -d

# Ver logs do backend
docker compose logs -f backend

# Importar CSV para o banco
cd backend && go run ./cmd/import/main.go

# Testar integração Scryfall
cd backend && go run ./cmd/testmtg/main.go

# Build de produção
docker compose -f docker-compose.prod.yml up --build -d
```

## Variáveis de ambiente

```
DB_HOST      # host MySQL
DB_PORT      # default 3306
DB_USER
DB_PASSWORD
DB_NAME      # default magic_collector
```

Definidas em `.env` (não versionado). Ver `.env.example`.

## Observações importantes

- O banco é MySQL externo (não SQLite — o README original está desatualizado neste ponto).
- O frontend não tem TypeScript. JSX puro com ESLint básico.
- Não há testes automatizados. Validação é manual via `cmd/testmtg`.
- A Scryfall API não exige chave mas pede `User-Agent` no header — já está configurado como `magic-collector/1.0`.
- `data/collection.db` no `.gitignore` é resquício do SQLite; o banco real é MySQL.
- Ao importar pré-con em PT, cada carta gera 2 requests à Scryfall (EN + PT). Para um set de 100 cartas, leva ~30s.
- 
Para alterações visuais/responsivas, leia obrigatoriamente:

- `docs/frontend-responsive-spec.md`
- `docs/frontend-responsive-plan.md`

A spec define o resultado esperado.
O plan define a ordem de execução.