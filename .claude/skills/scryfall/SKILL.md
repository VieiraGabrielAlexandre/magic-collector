---
name: scryfall
description: Trabalha com a Scryfall API no Magic Collector. Use quando precisar adicionar uma nova busca de carta, consultar dados de set, entender como a integração funciona, depurar retornos da API, ou implementar um novo método no cliente Scryfall.
---

## Visão geral

O cliente Scryfall está em `backend/internal/mtgapi/client.go`. Não requer chave de API, mas exige header `User-Agent: magic-collector/1.0`. Retorna `*ExternalCard, error` — retorne `nil, nil` em 404 (não é erro, carta não encontrada).

## Endpoints principais

| Operação | Endpoint |
|---|---|
| Por set + número | `GET /cards/{set}/{number}` |
| Por set + número + idioma | `GET /cards/{set}/{number}/{lang}` |
| Por UUID cacheado | `GET /cards/{uuid}` |
| Por nome exato | `GET /cards/named?exact={name}` |
| Por nome + set | `GET /cards/named?exact={name}&set={set}` |
| Pré-releases por nome | `GET /cards/search?q=is:prerelease name:"{name}"` |
| Todas as cartas de um set | `GET /cards/search?q=e:{set}&order=set&unique=prints` |
| Metadados de set | `GET /sets/{set_code}` |

Base URL: `https://api.scryfall.com`

## Códigos de idioma

| App | Scryfall |
|-----|----------|
| PT  | pt       |
| JP  | ja       |
| ES  | es       |
| FR  | fr       |
| DE  | de       |
| EN  | en       |

Ver `toLangCode()` em `client.go` para a lista completa.

## Campos localizados (idioma ≠ EN)

Quando buscada com idioma não-EN, a Scryfall pode retornar:
- `printed_name` — nome traduzido (ex: `"Aeroesquife"` para `"Sky Skiff"`)
- `printed_text` — texto de habilidade traduzido
- `printed_type_line` — linha de tipo traduzida

O frontend exibe `printed_name` quando disponível, caindo para `name`.

## Campos especiais de cartas double-faced (DFC)

Cartas com dois lados (ex: Werewolves, MDFCs) retornam `card_faces` em vez de `image_uris` no topo. O `toExternal()` já lida com isso — ao adicionar novos campos, verifique se precisam buscar em `card_faces[0]`.

## Rate limiting

Scryfall pede máximo ~10 req/s. **Não remover os sleeps existentes:**
- `75ms` entre buscas individuais em lote
- `100ms` entre páginas de listagem

## Passo a passo: adicionar novo método ao cliente

```go
// 1. Defina a assinatura (sempre *ExternalCard ou tipo específico, error)
func (c *Client) NovaBusca(param string) (*ExternalCard, error) {

    // 2. Construa a URL
    endpoint := fmt.Sprintf("https://api.scryfall.com/cards/...", param)

    // 3. Use c.fetch() para endpoints de carta única
    return c.fetch(endpoint)

    // 4. Para listas, use c.fetchPage() e itere paginando
}
```

O método `fetch()` já seta o User-Agent, decodifica JSON e retorna `nil, nil` em 404.

## Testar sem subir o servidor

```bash
cd backend && go run ./cmd/testmtg/main.go
```

Edite `cmd/testmtg/main.go` para testar o método novo antes de integrar ao serviço.

## Campos que vêm da Scryfall e são salvos em cache

Ao criar/importar uma carta, o service salva automaticamente em `cards`:
- `mtg_id` — UUID Scryfall (evita rebusca futura)
- `rarity` — raridade oficial
- `type` — tipo em inglês
- `mana_cost` — custo de mana
- `colors` — array JSON de cores (ex: `["W","U"]`)

## Checklist ao adicionar nova busca

- [ ] Método segue padrão `*ExternalCard, error`
- [ ] Header `User-Agent: magic-collector/1.0` setado (via `c.fetch()`)
- [ ] Retorna `nil, nil` em 404 (não erro)
- [ ] Sleep adicionado se for chamado em lote
- [ ] Testado via `cmd/testmtg/main.go`
- [ ] Integrado ao service/handler correto
