#!/usr/bin/env python3
"""
Atualiza preços de cartas via Scryfall API.
Cartas PT sem preço usam fallback para versão EN automaticamente.

Uso:
  python3 refresh_prices.py                          # todas as cartas (localhost)
  python3 refresh_prices.py --empty-only             # só cartas sem preço
  python3 refresh_prices.py http://34.x.x.x/api     # todas (produção)
  python3 refresh_prices.py http://IP/api --empty-only
"""
import json
import sys
import time
import urllib.request
import urllib.error

args = sys.argv[1:]
BASE_URL = "http://localhost:8080"
EMPTY_ONLY = "--empty-only" in args

for a in args:
    if a.startswith("http"):
        BASE_URL = a.rstrip("/")

url = f"{BASE_URL}/cards/refresh-prices{'?empty_only=1' if EMPTY_ONLY else ''}"

mode = "cartas SEM PREÇO (com fallback EN para cartas PT)" if EMPTY_ONLY else "TODAS as cartas"
print(f"→ POST {url}")
print(f"  Modo: {mode}")
print("  (80–160ms por carta — pode levar alguns minutos)\n")

req = urllib.request.Request(url, data=b"", method="POST",
                              headers={"Content-Type": "application/json"})

start = time.time()
try:
    with urllib.request.urlopen(req, timeout=600) as resp:
        result = json.loads(resp.read())
        elapsed = time.time() - start

        print(f"✓ {result['updated']} preços atualizados")
        print(f"  ⊘ {result['skipped']} cartas sem dados na Scryfall")
        if result.get("failed", 0):
            print(f"  ⚠ {result['failed']} falhas no banco")
        print(f"  Total processado: {result['total']} cartas em {elapsed:.1f}s")
except urllib.error.HTTPError as e:
    body = e.read().decode()
    print(f"✗ HTTP {e.code}: {body}")
except Exception as e:
    print(f"✗ Erro: {e}")
