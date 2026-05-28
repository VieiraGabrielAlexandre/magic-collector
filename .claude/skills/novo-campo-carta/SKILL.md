---
name: novo-campo-carta
description: Adiciona um novo campo ao modelo de carta do Magic Collector. Use quando precisar adicionar um atributo novo às cartas, como um campo de grading, preço manual, condição extra, ou qualquer dado novo que precisa ser salvo por carta.
---

## Visão geral

Adicionar um campo em `cards` toca obrigatoriamente 4 arquivos no backend e 1 no frontend. Esquecer qualquer um causa bug silencioso (campo não salva ou não aparece).

## Passo 1 — Migração do banco (`internal/database/mysql.go`)

Adicionar no final do bloco de `ALTER TABLE` seguro:

```go
db.Exec(`ALTER TABLE cards ADD COLUMN <nome> <TIPO> NOT NULL DEFAULT <default>`)
```

Exemplos de tipos: `VARCHAR(100)`, `INT`, `TINYINT` (bool), `TEXT`, `DECIMAL(10,2)`.
O `db.Exec` sem verificar erro é intencional — ignora se a coluna já existe.

## Passo 2 — Model (`internal/cards/model.go`)

Adicionar o campo em **três structs**:

```go
// Em Card:
NomeCampo string `json:"nome_campo"`

// Em CreateCardInput:
NomeCampo string `json:"nome_campo"`

// Em UpdateCardInput:
NomeCampo string `json:"nome_campo"`
```

## Passo 3 — Repository (`internal/cards/repository.go`)

Três lugares para tocar:

**INSERT** — adicionar `nome_campo` na lista de colunas e `?` nos valores, e `card.NomeCampo` nos args.

**UPDATE** — adicionar `nome_campo = ?` no SET e `card.NomeCampo` nos args (antes do `WHERE id = ?`).

**SELECT / Scan** — adicionar `nome_campo` no `SELECT` e `&card.NomeCampo` no `rows.Scan(...)` na mesma posição.

## Passo 4 — Frontend: estado inicial (`frontend/src/App.jsx`)

Adicionar no objeto `EMPTY_FORM`:

```js
const EMPTY_FORM = {
  // ... campos existentes ...
  nome_campo: "",   // ou false para bool, ou 0 para número
};
```

## Passo 5 — Frontend: formulários (`frontend/src/App.jsx`)

**Formulário de cadastro** (buscar pela seção `<form className="card form">`):
Adicionar o campo usando o helper `field()` para inputs simples, ou JSX manual para selects/checkboxes.

**Formulário de edição** (buscar pela seção `editMode && (`):
Repetir o mesmo campo dentro do `<div className="edit-grid">`.

**Modal de detalhes** (opcional — se faz sentido exibir):
Adicionar na grade de `<div className="modal-grid">`:
```jsx
<div><span>Label</span>{selectedCard.local.nome_campo || "—"}</div>
```

## Checklist

- [ ] `ALTER TABLE` adicionado em `mysql.go`
- [ ] Campo em `Card`, `CreateCardInput` e `UpdateCardInput` em `model.go`
- [ ] Campo no `INSERT`, `UPDATE` e `SELECT`/`Scan` em `repository.go`
- [ ] Campo em `EMPTY_FORM` em `App.jsx`
- [ ] Campo nos formulários de cadastro e edição em `App.jsx`
- [ ] Rebuild do backend: `docker compose up --build -d backend`
- [ ] Testar criação e edição de carta com o novo campo
