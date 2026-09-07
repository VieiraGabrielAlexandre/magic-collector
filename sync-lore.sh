#!/usr/bin/env bash
# Sincroniza arquivos de lore e rebuilda o backend.
#
# Os .md editáveis ficam em doc/lore/ (se existir) ou diretamente
# em backend/internal/lore/data/. Os arquivos em data/ são
# embutidos no binário via //go:embed na compilação.
#
# Uso:
#   ./sync-lore.sh          # copia doc/lore/*.md → backend/internal/lore/data/ e rebuilda
#   ./sync-lore.sh --only-rebuild   # só rebuilda (arquivos já editados em data/)

set -e

DEST="backend/internal/lore/data"
SRC="doc/lore"

if [[ "$1" != "--only-rebuild" ]] && [ -d "$SRC" ]; then
  echo "→ Copiando $SRC/*.md → $DEST/"
  cp "$SRC"/*.md "$DEST"/
  echo "  OK"
else
  echo "→ Usando arquivos já em $DEST/ (edição direta)"
fi

echo "→ Rebuild do backend..."
docker compose up --build -d backend
echo "✓ Lore atualizada."
