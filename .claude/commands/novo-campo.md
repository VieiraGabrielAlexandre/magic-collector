Adicione um novo campo chamado `$ARGUMENTS` ao modelo de carta (`Card`) do Magic Collector, seguindo o padrão do projeto:

1. **`backend/internal/database/mysql.go`** — adicione um `db.Exec("ALTER TABLE cards ADD COLUMN ...")` seguro (sem verificação de erro)
2. **`backend/internal/cards/model.go`** — adicione o campo em `Card`, `CreateCardInput` e `UpdateCardInput` com a tag JSON correta
3. **`backend/internal/cards/repository.go`** — inclua o campo no `INSERT`, `UPDATE` e em todos os `SELECT` (Scan)
4. **`frontend/src/App.jsx`** — adicione o campo no `EMPTY_FORM` e nos JSX dos formulários de cadastro e edição
5. Se o campo precisar ser exibido no modal de detalhes, adicione também na seção do modal em `App.jsx`

Ao final, confirme quais arquivos foram modificados.
