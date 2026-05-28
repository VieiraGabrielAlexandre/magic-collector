---
name: verify
description: Verifica se uma mudança no Magic Collector funciona corretamente. Use quando precisar confirmar que uma feature funciona, testar uma correção, validar integração com Scryfall, ou checar regressões após mudanças.
---

## Visão geral

Verificação é sempre: health check → logs → teste da rota/funcionalidade alterada → smoke test nos fluxos vizinhos.

## Passo 1 — Confirmar que o backend está saudável

```bash
curl -s http://localhost:8080/health
docker compose logs backend --tail=20
```

Se não responder, rebuild primeiro: `docker compose up --build -d backend`.

## Passo 2 — Testar a rota ou funcionalidade alterada

### Cartas
```bash
# Listar (confirma paginação, busca, ordenação)
curl -s "http://localhost:8080/cards?page=1&sort=name&order=asc" | python3 -m json.tool | head -50

# Criar uma carta de teste
curl -s -X POST http://localhost:8080/cards \
  -H "Content-Type: application/json" \
  -d '{"name":"Lightning Bolt","set_code":"m11","collection_number":"149","language":"EN","quantity":1}' \
  | python3 -m json.tool

# Detalhar (inclui dados Scryfall — confirma integração externa)
curl -s http://localhost:8080/cards/{id} | python3 -m json.tool
```

### Decks
```bash
curl -s http://localhost:8080/decks | python3 -m json.tool
```

### Batalhas
```bash
curl -s http://localhost:8080/battles | python3 -m json.tool
```

## Passo 3 — Testar integração Scryfall (se tocou mtgapi)

```bash
cd backend && go run ./cmd/testmtg/main.go
```

## Passo 4 — Smoke test manual no frontend

Abrir http://localhost:5173 e verificar:

**Aba Coleção:**
- Cadastrar carta com `set_code` + número → modal deve mostrar imagem e dados do Scryfall
- Buscar por nome/cor/tipo
- Editar e salvar uma carta
- Remover uma carta

**Aba Decks:**
- Criar deck → atribuir cartas → verificar contagem
- Ícone do set carrega ao gerenciar deck

**Aba Batalhas:**
- Registrar resultado → aparece na lista

## Passo 5 — Checklist de regressão

- [ ] Health endpoint retorna 200
- [ ] Nenhum erro vermelho no console do browser
- [ ] Listagem de cartas carrega com paginação
- [ ] Modal de detalhes abre com dados Scryfall
- [ ] Funcionalidade modificada funciona no caminho feliz
- [ ] Funcionalidades vizinhas não quebraram
