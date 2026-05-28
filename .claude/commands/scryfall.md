Consulte a Scryfall API para a seguinte operação: $ARGUMENTS

Contexto do projeto:
- O cliente Scryfall está em `backend/internal/mtgapi/client.go`
- A URL base é `https://api.scryfall.com/cards`
- Idiomas: PT→`pt`, JP→`ja`, ES→`es`, FR→`fr`, DE→`de`
- Busca por set+número: `GET /cards/{set}/{number}` ou `GET /cards/{set}/{number}/{lang}`
- Busca por UUID cacheado: `GET /cards/{uuid}`
- Busca por nome: `GET /cards/named?exact={name}`
- Busca por set completo: `GET /cards/search?q=e:{set_code}&order=set&unique=prints`
- Metadados de set (ícone, nome): `GET /sets/{set_code}`
- Para cartas pré-release: `GET /cards/search?q=is:prerelease name:"{name}"`

Se precisar adicionar um novo método ao cliente, siga o padrão de `fetch()` em `client.go`: set User-Agent `magic-collector/1.0`, return `*ExternalCard, error`, return `nil, nil` em 404.

Rate limit: máximo ~10 req/s. Use `time.Sleep(75*time.Millisecond)` entre buscas em lote.
