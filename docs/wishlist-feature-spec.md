# Wishlist Feature Spec — Magic Collector

## Objetivo

Adicionar uma nova aba chamada **Wishlist** ao Magic Collector.

A Wishlist permite cadastrar cartas desejadas informando:

* sigla da coleção/set;
* número da carta;
* motivo de estar na wishlist;
* se deseja versão foil ou não.

Ao cadastrar, o sistema deve buscar os dados da carta na Scryfall e salvar a entrada da wishlist no banco.

---

## Escopo

Implementar backend, banco e frontend para Wishlist.

Arquivos prováveis:

* `backend/internal/database/mysql.go`
* `backend/internal/wishlist/`
* `backend/internal/mtgapi/client.go`, somente se necessário
* `backend/cmd/api/main.go`
* `frontend/src/App.jsx`
* `frontend/src/App.css`
* `frontend/src/services/api.js`

---

## Banco de dados

Criar tabela `wishlist_cards`.

Campos sugeridos:

```sql
id BIGINT AUTO_INCREMENT PRIMARY KEY,
mtg_id VARCHAR(64),
set_code VARCHAR(10) NOT NULL,
collection_number VARCHAR(20) NOT NULL,
name VARCHAR(255),
printed_name VARCHAR(255),
image_uri TEXT,
artist VARCHAR(255),
rarity VARCHAR(10),
colors TEXT,
color VARCHAR(100),
price_usd DECIMAL(10,2) NULL,
price_usd_foil DECIMAL(10,2) NULL,
foil TINYINT(1) DEFAULT 0,
reason TEXT,
acquired TINYINT(1) DEFAULT 0,
created_at DATETIME,
updated_at DATETIME
```

Seguir o padrão atual do projeto: adicionar criação da tabela e alterações necessárias diretamente em `internal/database/mysql.go`. Não usar migração externa.

---

## API

Criar rotas:

```http
GET    /wishlist
POST   /wishlist
GET    /wishlist/:id
PUT    /wishlist/:id
DELETE /wishlist/:id
POST   /wishlist/:id/acquire
```

### POST /wishlist

Body:

```json
{
  "set_code": "TMC",
  "collection_number": "12",
  "foil": false,
  "reason": "Quero testar no deck Commander das Tartarugas"
}
```

Comportamento:

* Buscar carta na Scryfall usando `set_code` + `collection_number`.
* Salvar dados principais retornados.
* Usar `printed_name` quando disponível.
* Salvar preço, imagem, artista e raridade.
* Normalizar `set_code` para maiúsculo.

### POST /wishlist/:id/acquire

Essa rota deve transformar uma carta da wishlist em uma carta da coleção.

Body sugerido:

```json
{
  "deck_id": 0,
  "condition": "near_mint",
  "commander": false,
  "prerelease": false
}
```

Comportamento:

* Criar uma nova entrada em `cards` usando os dados da wishlist.
* Se `deck_id` for informado, já vincular a carta ao deck.
* Marcar a wishlist como `acquired = 1` ou remover da wishlist, dependendo da implementação mais simples.
* Preferência: marcar como adquirida para preservar histórico.

---

## Frontend

Adicionar nova aba:

```text
Coleção | Decks | Battles | Wishlist
```

A Wishlist deve ter:

* formulário de cadastro;
* lista de cartas desejadas;
* modal de detalhes;
* ação rápida para remover;
* ação rápida para adquirir;
* opção de escolher deck ao adquirir.

---

## Layout da Wishlist

A listagem deve seguir o estilo da listagem de cartas, mas com identidade própria.

Direção visual:

* tema dark;
* destaque em **azul claro** em vez de dourado;
* cards elegantes;
* imagem da carta;
* nome;
* set;
* número;
* raridade;
* artista;
* preço;
* motivo da wishlist;
* badge foil quando aplicável.

Tokens visuais sugeridos:

```css
--wishlist-primary: #7dd3fc;
--wishlist-primary-soft: rgba(125, 211, 252, 0.16);
--wishlist-border: rgba(125, 211, 252, 0.35);
--wishlist-surface: #0f172a;
```

---

## Modal de detalhes da Wishlist

O modal deve usar uma variação visual azul clara.

Deve mostrar:

* imagem grande;
* nome;
* nome impresso, se existir;
* set;
* número;
* raridade;
* artista;
* preço normal;
* preço foil;
* motivo;
* flag foil;
* botão adquirir;
* botão remover;
* botão fechar.

Critérios:

* modal deve caber em 320px;
* deve ter scroll interno se necessário;
* botões não podem ficar sobrepostos;
* visual deve ser diferente do modal padrão dourado.

---

## Ação “Adquirir carta”

Ao clicar em adquirir:

* abrir modal ou painel simples;
* permitir escolher deck;
* permitir condição da carta;
* confirmar cadastro na coleção;
* chamar `POST /wishlist/:id/acquire`;
* atualizar lista da coleção e wishlist;
* exibir feedback visual.

Se não escolher deck, salvar como carta sem deck.

---

## Responsividade

Mobile:

* formulário em uma coluna;
* cards em uma coluna;
* ações empilhadas;
* botões grandes;
* motivo pode aparecer resumido com opção de detalhe no modal.

Tablet:

* cards em grid de 2 colunas;
* formulário pode ter 2 colunas.

Desktop:

* grid com múltiplas colunas;
* cards com largura máxima;
* conteúdo centralizado.

---

## Critérios de aceite

* Existe aba Wishlist.
* É possível cadastrar carta por set + número.
* O sistema busca dados da Scryfall.
* A Wishlist mostra imagem, preço, artista e motivo.
* É possível remover uma carta diretamente pela lista.
* É possível abrir modal de detalhes.
* Modal usa visual azul claro.
* É possível adquirir uma carta com um clique/fluxo simples.
* Ao adquirir, a carta entra na coleção.
* Ao adquirir, pode ser vinculada a um deck.
* Layout funciona bem em mobile, iPad e desktop.
* Não há botão sobreposto.
* Não há scroll horizontal acidental.
