Ajude-me com o ambiente de desenvolvimento do Magic Collector.

Comandos úteis:
- Subir tudo: `docker compose up --build -d`
- Só backend (rebuild): `docker compose up --build -d backend`
- Logs ao vivo: `docker compose logs -f backend` ou `docker compose logs -f frontend`
- Parar: `docker compose down`
- Testar Scryfall: `cd backend && go run ./cmd/testmtg/main.go`
- Importar CSV: `cd backend && go run ./cmd/import/main.go [arquivo.csv] [banco.db]`
- Build prod: `docker compose -f docker-compose.prod.yml up --build -d`

URLs locais:
- Frontend: http://localhost:5173
- Backend: http://localhost:8080
- Health: http://localhost:8080/health

Banco MySQL definido via `.env` (copiar de `.env.example`).

O que você precisa fazer? $ARGUMENTS
