# Magic Collector

Gerenciador de coleção de cartas Magic: The Gathering com backend em Go e frontend em React.

---

## Sumário

- [Pré-requisitos](#pré-requisitos)
- [Executar o projeto](#executar-o-projeto)
- [Importar cartas via CSV](#importar-cartas-via-csv)
- [Estrutura do CSV](#estrutura-do-csv)
- [Usando o frontend](#usando-o-frontend)
- [Lógica de busca na API externa](#lógica-de-busca-na-api-externa)
- [Estrutura do projeto](#estrutura-do-projeto)
- [API REST](#api-rest)

---

## Pré-requisitos

- [Docker](https://docs.docker.com/get-docker/) e Docker Compose
- Go 1.25+ (apenas para rodar o script de importação ou testes locais)

---

## Executar o projeto

```bash
# Clonar o repositório
git clone <url-do-repositorio>
cd magic-collector

# Garantir que a pasta de dados existe e tem permissão de escrita
mkdir -p data
chmod 775 data

# Subir os serviços
docker compose up --build -d
```

| Serviço   | URL                        |
|-----------|----------------------------|
| Frontend  | http://localhost:5173       |
| Backend   | http://localhost:8080       |
| Health    | http://localhost:8080/health|

Para parar:
```bash
docker compose down
```

Para ver os logs:
```bash
docker compose logs -f backend
docker compose logs -f frontend
```

---

## Importar cartas via CSV

O script de importação lê um arquivo CSV, agrupa linhas duplicadas somando a quantidade, e insere tudo no banco de dados diretamente — sem passar pela API externa.

```bash
cd backend

# Padrão: lê data/cards.csv e grava em data/collection.db
go run ./cmd/import/main.go

# Caminhos personalizados
go run ./cmd/import/main.go /caminho/para/cartas.csv /caminho/para/banco.db
```

> O banco é criado automaticamente se não existir. Recomenda-se parar os containers antes de importar para evitar conflitos de acesso ao arquivo SQLite, ou rodar com os containers no ar (o volume é compartilhado via `./data`).

---

## Estrutura do CSV

O arquivo deve ter exatamente **12 colunas** com cabeçalho na primeira linha:

| Coluna            | Tipo     | Exemplo              | Obrigatório |
|-------------------|----------|----------------------|-------------|
| Nome              | texto    | Artesão Braçoluz     | Sim         |
| Cor               | texto    | Branco               | Não         |
| Tipo              | texto    | Criatura             | Não         |
| Subtitulo         | texto    | Anão Artesão         | Não         |
| Numero na coleção | número   | 17                   | Não*        |
| Raridade          | letra    | C / U / R / M / L / T | Não       |
| Sigla             | texto    | KLD                  | Não*        |
| Lingua            | código   | PT / EN / ES / JP    | Não         |
| Ano               | número   | 2016                 | Não         |
| Artista           | texto    | Ryan Pancoast        | Não         |
| Empresa           | texto    | Wizards of the Coast | Não         |
| Foil              | booleano | Sim / Não / Yes / No | Não         |

> *Sigla + Número na coleção são necessários para a busca automática de dados externos via Scryfall.

**Exemplo de cabeçalho:**
```
Nome,Cor,Tipo,Subtitulo,Numero na coleção,Raridade,Sigla,Lingua,Ano,Artista,Empresa,Foil
```

**Linhas duplicadas** (mesmo Nome + Sigla + Número + Idioma + Foil) são agrupadas automaticamente e a quantidade é somada.

**Foil** aceita: `sim`, `yes`, `true` (case-insensitive) → verdadeiro. Qualquer outro valor → falso.

---

## Usando o frontend

### Cadastrar uma carta manualmente

Preencha o formulário à esquerda e clique em **Cadastrar**. O backend tenta buscar dados complementares na Scryfall automaticamente ao salvar.

### Pesquisar e ordenar

- Use a barra de busca para filtrar por nome, coleção, cor, tipo ou artista
- Clique nos botões de ordenação (Nome, Coleção, Cor, Raridade, Ano, Nº) para ordenar; clique novamente para inverter
- A paginação é automática (20 cartas por página)

### Ver detalhes de uma carta

Clique em **Ver** em qualquer carta da lista. O modal exibe:

- Imagem de alta qualidade (via Scryfall)
- Nome local (ex: `Aeroesquife`) e nome em inglês (`Sky Skiff`)
- Tipo, custo de mana, poder/resistência
- Nome completo da coleção
- Texto da habilidade no idioma da carta
- Texto de sabor (*flavor text*)
- Preços atuais em USD e EUR (via Scryfall)
- Link direto para a página da carta no Scryfall
- Dados locais: idioma, condição, quantidade, artista, notas

### Remover uma carta

Clique em **✕** na lista ou em **Remover carta** dentro do modal de detalhes.

---

## Lógica de busca na API externa

Ao clicar em **Ver**, o backend consulta a [Scryfall API](https://scryfall.com/docs/api) para enriquecer os dados locais. O fluxo é:

```
GET /cards/:id
    │
    ├─ 1. Há mtg_id em cache (UUID Scryfall)?
    │       └─ SIM → GET https://api.scryfall.com/cards/{uuid}
    │                     └─ OK → retorna dados (rápido, sem nova busca)
    │
    └─ 2. mtg_id ausente ou inválido → busca por Coleção + Número
              │
              ├─ Idioma ≠ EN?
              │     └─ GET /cards/{sigla}/{número}/{idioma}  (ex: /cards/kld/233/pt)
              │           ├─ OK + artista bate → retorna carta localizada
              │           │     → salva UUID Scryfall no banco para cache futuro
              │           └─ 404 ou artista diferente → continua
              │
              └─ GET /cards/{sigla}/{número}  (EN, padrão)
                    ├─ OK → retorna carta em inglês
                    │     → salva UUID Scryfall no banco para cache futuro
                    └─ 404 → nenhum dado externo retornado
```

**Por que Coleção + Número e não o nome?**
Cartas em português têm nomes diferentes dos originais em inglês. A busca por set code + número do colecionador é independente de idioma e retorna sempre a carta correta.

**Validação pelo artista**
Na busca localizada (etapa 2a), o artista retornado pela Scryfall é comparado com o artista salvo localmente para garantir que o número no CSV está correto. A comparação é case-insensitive e aceita correspondência parcial — `"Dan Scott"` bate com `"Dan Murayama Scott"`.

**Cache por UUID**
Após a primeira busca bem-sucedida, o UUID Scryfall é gravado no campo `mtg_id`. Nas próximas consultas ao mesmo card, o backend usa diretamente `GET /cards/{uuid}`, evitando a busca dupla.

**Campos localizados**
Quando a carta é buscada com idioma PT (ou outro não-EN), a Scryfall retorna campos adicionais:
- `printed_name` — nome no idioma da carta
- `printed_text` — texto da habilidade traduzido
- `printed_type_line` — linha de tipo traduzida

---

## Estrutura do projeto

```
magic-collector/
├── backend/
│   ├── cmd/
│   │   ├── api/          # Entrypoint do servidor HTTP
│   │   ├── import/       # Script de importação CSV
│   │   └── testmtg/      # Utilitário de teste da integração Scryfall
│   ├── internal/
│   │   ├── cards/        # Handler, Service, Repository, Model
│   │   ├── database/     # Conexão SQLite + migrações
│   │   └── mtgapi/       # Cliente HTTP para a Scryfall API
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── App.jsx       # Componente principal (formulário + lista + modal)
│   │   ├── App.css
│   │   └── services/
│   │       └── api.js    # Funções de acesso ao backend
│   ├── Dockerfile
│   └── vite.config.js    # Proxy /api → backend:8080
├── data/                 # Volume persistente (não versionado)
│   ├── collection.db     # Banco SQLite
│   └── cards.csv         # CSV de importação (opcional)
└── docker-compose.yml
```

---

## API REST

| Método | Rota         | Descrição                                    |
|--------|--------------|----------------------------------------------|
| GET    | /health      | Health check                                 |
| GET    | /cards       | Lista cartas com paginação, busca e ordenação |
| POST   | /cards       | Cadastra uma carta                           |
| GET    | /cards/:id   | Detalha uma carta + dados externos Scryfall  |
| DELETE | /cards/:id   | Remove uma carta                             |

### Parâmetros de listagem (`GET /cards`)

| Parâmetro  | Padrão | Descrição                                      |
|------------|--------|------------------------------------------------|
| `q`        | —      | Busca livre (nome, set, cor, tipo, artista)    |
| `page`     | 1      | Página atual                                   |
| `page_size`| 20     | Itens por página (máx. 100)                    |
| `sort`     | name   | Campo: `name`, `set_code`, `rarity`, `color`, `year`, `collection_number` |
| `order`    | asc    | `asc` ou `desc`                                |

### Resposta de `GET /cards/:id`

```json
{
  "local": { /* dados do banco local */ },
  "external": {
    "id": "c58fb256-...",
    "name": "Sky Skiff",
    "printed_name": "Aeroesquife",
    "set": "KLD",
    "set_name": "Kaladesh",
    "rarity": "common",
    "type": "Artifact — Vehicle",
    "printed_type": "Artefato — Veículo",
    "mana_cost": "{2}",
    "text": "Flying\nCrew 1 ...",
    "printed_text": "Voar\nTripular 1 ...",
    "flavor_text": "...",
    "artist": "Richard Wright",
    "image_url": "https://cards.scryfall.io/normal/...",
    "power": "2",
    "toughness": "3",
    "prices": { "usd": "0.05", "eur": "0.03", ... },
    "scryfall_uri": "https://scryfall.com/card/kld/233/pt/..."
  }
}
```
