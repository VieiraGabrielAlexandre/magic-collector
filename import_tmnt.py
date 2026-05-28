#!/usr/bin/env python3
"""
Importa o deck TMNT "Michelangelo & Leonardo" em inglês.
Uso: python3 import_tmnt.py
     python3 import_tmnt.py http://34.196.63.122/api   # produção
"""
import json
import sys
import urllib.request
import urllib.error

BASE_URL = sys.argv[1].rstrip("/") if len(sys.argv) > 1 else "http://localhost:8080"

DECK_LIST = """\
Commander
1\tMichelangelo, the Heart\t\t$ 1.20
1\tLeonardo, the Balance\t\t$ 0.92
Creatures (29)
1\tApril O'Neil, Live on the Scene\t\t$ 0.35
1\tBebop, Skull & Crossbones\t\t$ 0.33
1\tLita, Little Orphan Amphibian\t\t$ 0.25
1\tRoadkill Rodney\t\t$ 0.33
1\tSplinter, the Mentor\t\t$ 0.69
1\tSteelbane Hydra\t\t$ 0.41
1\tTempestra, Dame of Games\t\t$ 0.35
1\tVoracious Hydra\t\t$ 0.33
1\tDonatello, the Brains\t\t$ 0.98
1\tIrma, Part-Time Mutant\t\t$ 5.12
1\tMona Lisa, Science Geek\t\t$ 0.25
1\tRay Fillet, Wave Warrior\t\t$ 0.40
1\tRocksteady, Mutant Marauder\t\t$ 0.35
1\tTokka & Rahzar, Unsupervised\t\t$ 0.32
1\tBaxter, Fly in the Ointment\t\t$ 0.34
1\tBig Mother Mouser\t\t$ 0.99
1\tCasey Jones, Back Alley Brute\t\t$ 0.40
1\tCorpsejack Menace\t\t$ 0.34
1\tDimension X Pizzasaur\t\t$ 0.35
1\tElectric Seaweed\t\t$ 0.28
1\tRat King, Pale Piper\t\t$ 0.40
1\tAcidic Slime\t\t$ 0.25
1\tBiogenic Ooze\t\t$ 0.33
1\tHeroes in a Half Shell\t\t$ 0.85
1\tKrang, the All-Powerful\t\t$ 0.50
1\tRaphael, the Muscle\t\t$ 0.86
1\tShredder, Shadow Master\t\t$ 1.10
1\tVigor\t\t$ 2.21
1\tLeatherhead, Iron Gator\t\t
Spells (16)
1\tShellshock\t\t$ 0.30
1\tAssassin's Trophy\t\t$ 0.94
1\tContinue?\t\t$ 3.50
1\tSuper Combo\t\t$ 0.40
1\tCultivate\t\t$ 0.40
1\tHere Comes a New Hero!\t\t$ 0.40
1\tSpecial Move\t\t$ 0.32
1\tSwift Demise\t\t$ 0.34
1\tDouble Jump // Flying Kick\t\t$ 0.31
1\tHarmonize\t\t$ 0.25
1\tLessons from Life\t\t$ 0.29
1\tWave Goodbye\t\t$ 0.98
1\tFast Forward\t\t$ 0.40
1\tGame Over\t\t$ 0.40
1\tVanquish the Horde\t\t$ 0.33
1\tBlasphemous Act\t\t$ 0.94
Artifacts (9)
1\tSol Ring\t\t$ 1.48
1\tArcane Signet\t\t$ 0.52
1\tEverything Pizza\t\t$ 0.20
1\tFoot Chopper\t\t$ 0.35
1\tArcade Cabinet\t\t$ 0.59
1\tChromatic Lantern\t\t$ 1.26
1\tExploding Barrel\t\t$ 0.35
1\tCoin of Mastery\t\t$ 1.79
1\tMole Module\t\t$ 0.35
Enchantments (5)
1\tLevel Up\t\t$ 2.30
1\tTogether Forever\t\t$ 0.30
1\tEndless Foot Assault\t\t$ 4.72
1\tHigh Score\t\t$ 1.10
1\tNinja Pizza\t\t$ 3.23
Lands (39)
1\tAsh Barrens\t\t$ 0.25
1\tBig Apple, 3 a.m.\t\t$ 0.47
1\tCinder Glade\t\t$ 0.30
1\tCity of Brass\t\t$ 10.70
1\tCommand Tower\t\t$ 0.33
1\tDragonskull Summit\t\t$ 0.55
1\tEscape Tunnel\t\t$ 0.25
1\tEvolving Wilds\t\t$ 0.21
1\tExotic Orchard\t\t$ 0.30
1\tFabled Passage\t\t$ 1.32
4\tForest\t\t$ 0.00
1\tGrand Coliseum\t\t$ 0.45
1\tHidden Hideout\t\t$ 3.70
1\tHinterland Harbor\t\t$ 0.35
2\tIsland\t\t$ 0.00
2\tMountain\t\t$ 0.00
1\tPath of Ancestry\t\t$ 0.25
2\tPlains\t\t$ 0.00
1\tRain-Slicked Copse\t\t$ 0.31
1\tRootbound Crag\t\t$ 0.39
1\tSmoldering Marsh\t\t$ 0.32
1\tSodden Verdure\t\t$ 0.53
1\tSpire Garden\t\t$ 5.50
1\tSunken Hollow\t\t$ 0.37
2\tSwamp\t\t$ 0.00
1\tThriving Grove\t\t$ 0.20
1\tThriving Isle\t\t$ 0.20
1\tThriving Moor\t\t$ 0.20
1\tTurtle Lair\t\t$ 0.25
1\tUndergrowth Stadium\t\t$ 6.01
1\tVernal Fen\t\t$ 0.50
1\tVibrant Cityscape\t\t$ 0.25
100 Cards Total
"""

payload = {
    "deck_name":   "TMNT — Michelangelo & Leonardo",
    "set_code":    "",
    "language":    "EN",
    "commander":   True,
    "colors":      "W,U,B,R,G",
    "theme_color": "forest",
    "description": "Teenage Mutant Ninja Turtles Commander — 5 cores",
    "deck_list":   DECK_LIST,
}

url = f"{BASE_URL}/decks/import-list"
card_lines = [l for l in DECK_LIST.splitlines() if l.strip() and l.strip()[0].isdigit()]
print(f"→ POST {url}")
print(f"  {len(card_lines)} entradas na lista, idioma EN")
print("  Aguarde (~30s)…\n")

data = json.dumps(payload).encode("utf-8")
req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/json"}, method="POST")

def do_request(url, payload, timeout=180):
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/json"}, method="POST")
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.loads(resp.read())

try:
    result = do_request(url, payload)
    print(f"✓ {result['imported']} cartas importadas!")
    failed = result.get("failed_cards", [])
    if failed:
        print(f"  ⚠ {result['failed']} falhas:")
        for name in failed:
            print(f"    - {name}")

        # Retry automático: importa só as que falharam para o mesmo deck
        retry_names = [n for n in failed if not n.endswith(")")]  # exclui erros com sufixo "(db: ...)"
        retry_names += [n[:n.rfind(" (db:")] for n in failed if " (db:" in n]
        if retry_names:
            retry_list = "\n".join(f"1\t{name}" for name in retry_names)
            retry_payload = {
                "deck_list": retry_list,
                "set_code": payload["set_code"],
                "language": payload["language"],
            }
            retry_url = f"{BASE_URL}/decks/{result['deck_id']}/import-cards"
            print(f"\n  → Retry de {len(retry_names)} cartas em {retry_url}…")
            try:
                r2 = do_request(retry_url, retry_payload, timeout=180)
                print(f"  ✓ Retry: {r2['imported']} importadas, {r2['failed']} ainda falharam")
                if r2.get("failed_cards"):
                    for name in r2["failed_cards"]:
                        print(f"    ✗ {name}")
            except Exception as e2:
                print(f"  ✗ Retry falhou: {e2}")

    print(f"\n  Total na lista: {result['total_from_api']}")
    print(f"  Deck ID: {result['deck_id']} — {result.get('deck_name', '')}")
except urllib.error.HTTPError as e:
    body = e.read().decode()
    print(f"✗ HTTP {e.code}: {body}")
except Exception as e:
    print(f"✗ Erro: {e}")
