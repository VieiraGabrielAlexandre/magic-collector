---
name: run
description: Sobe o ambiente de desenvolvimento do Magic Collector. Use quando precisar iniciar o app, rodar o projeto, subir os containers, ver se está rodando, ou acessar o frontend/backend localmente.
---

## Visão geral

O projeto roda via Docker Compose. Backend Go na porta 8080, frontend React/Vite na porta 5173. Banco é MySQL externo configurado via `.env`.

## Passo 1 — Verificar se já está no ar

```bash
docker compose ps
curl -s http://localhost:8080/health
```

Se retornar `{"status":"ok"}`, o app já está rodando. Acesse http://localhost:5173.

## Passo 2 — Garantir que o `.env` existe

```bash
ls .env 2>/dev/null || cp .env.example .env
```

Se copiou, edite `.env` com as credenciais MySQL antes de continuar.

## Passo 3 — Subir tudo

```bash
docker compose up --build -d
```

Aguarde ~5s para o backend conectar ao MySQL. Verifique os logs se o backend não subir:

```bash
docker compose logs backend --tail=30
```

## Passo 4 — Rebuild seletivo (quando necessário)

```bash
# Backend mudou (Go requer rebuild obrigatório)
docker compose up --build -d backend

# Frontend: hot reload via volume mount — geralmente não precisa rebuild
# Só rebuild se mudou package.json, vite.config ou Dockerfile
docker compose up --build -d frontend
```

## URLs

| Serviço  | URL                         |
|----------|-----------------------------|
| Frontend | http://localhost:5173        |
| Backend  | http://localhost:8080        |
| Health   | http://localhost:8080/health |

## Parar

```bash
docker compose down
```

## Checklist

- [ ] `.env` existe com credenciais MySQL válidas
- [ ] `docker compose ps` mostra backend e frontend como `running`
- [ ] `curl /health` retorna 200
- [ ] http://localhost:5173 carrega a página
