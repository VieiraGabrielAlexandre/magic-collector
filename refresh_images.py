#!/usr/bin/env python3
"""
Atualiza a imagem (image_url) de todas as cartas da coleção via Scryfall API.
Cartas com mtg_id são buscadas diretamente; as demais usam set_code + número.

Uso:
  python3 refresh_images.py                        # localhost
  python3 refresh_images.py http://34.x.x.x/api   # produção
"""
import json
import sys
import time
import urllib.request
import urllib.error

BASE_URL = sys.argv[1].rstrip("/") if len(sys.argv) > 1 else "http://localhost:8080"
url = f"{BASE_URL}/cards/refresh-images"

print(f"→ POST {url}")
print("  Atualizando imagens via Scryfall para todas as cartas…")
print("  (80ms por carta — pode levar alguns minutos para coleções grandes)\n")

req = urllib.request.Request(url, data=b"", method="POST",
                              headers={"Content-Type": "application/json"})

start = time.time()
try:
    with urllib.request.urlopen(req, timeout=600) as resp:
        result = json.loads(resp.read())
        elapsed = time.time() - start

        print(f"✓ {result['updated']} imagens atualizadas")
        print(f"  ⊘ {result['skipped']} cartas sem imagem na Scryfall")
        if result.get("failed", 0):
            print(f"  ⚠ {result['failed']} falhas")
        print(f"  Total processado: {result['total']} cartas em {elapsed:.1f}s")
except urllib.error.HTTPError as e:
    body = e.read().decode()
    print(f"✗ HTTP {e.code}: {body}")
except Exception as e:
    print(f"✗ Erro: {e}")
