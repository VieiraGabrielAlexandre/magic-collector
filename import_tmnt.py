#!/usr/bin/env python3
"""
Importa o deck TMNT "Michelangelo & Leonardo" em inglês.
Uso:
  python3 import_tmnt.py                        # novo import completo (localhost)
  python3 import_tmnt.py --retry 11             # importa só as que falharam para o deck 11
  python3 import_tmnt.py http://IP/api          # novo import completo (produção)
  python3 import_tmnt.py http://IP/api --retry 11
"""
import json
import sys
import urllib.request
import urllib.error

args = sys.argv[1:]
BASE_URL = "http://localhost:8080"
RETRY_DECK_ID = None

for i, a in enumerate(args):
    if a == "--retry" and i + 1 < len(args):
        RETRY_DECK_ID = int(args[i + 1])
    elif a.startswith("http"):
        BASE_URL = a.rstrip("/")

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

def do_request(url, payload, timeout=600):
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/json"}, method="POST")
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.loads(resp.read())

def print_result(result, deck_id=None):
    print(f"✓ {result['imported']} cartas importadas!")
    failed = result.get("failed_cards", [])
    if failed:
        print(f"  ⚠ {result['failed']} ainda falharam:")
        for name in failed:
            print(f"    ✗ {name}")
    print(f"  Total processado: {result['total_from_api']}")
    if deck_id:
        print(f"  Deck ID: {deck_id}")

# Modo retry: importa só cartas específicas para um deck existente
FAILED_CARDS = [
    "Dimension X Pizzasaur",
    "Electric Seaweed",
    "Rat King, Pale Piper",
    "Acidic Slime",
    "Biogenic Ooze",
    "Heroes in a Half Shell",
    "Krang, the All-Powerful",
    "Raphael, the Muscle",
    "Shredder, Shadow Master",
    "Vigor",
    "Leatherhead, Iron Gator",
    "Shellshock",
    "Assassin's Trophy",
    "Continue?",
    "Super Combo",
    "Cultivate",
    "Here Comes a New Hero!",
    "Special Move",
    "Swift Demise",
    "Double Jump // Flying Kick",
    "Harmonize",
    "Lessons from Life",
    "Wave Goodbye",
    "Fast Forward",
    "Game Over",
    "Vanquish the Horde",
    "Blasphemous Act",
    "Sol Ring",
    "Arcane Signet",
    "Everything Pizza",
    "Foot Chopper",
    "Arcade Cabinet",
    "Chromatic Lantern",
    "Exploding Barrel",
    "Coin of Mastery",
    "Mole Module",
    "Level Up",
    "Together Forever",
    "Endless Foot Assault",
    "High Score",
    "Ninja Pizza",
    "Ash Barrens",
    "Big Apple, 3 a.m.",
    "Cinder Glade",
    "City of Brass",
    "Command Tower",
    "Dragonskull Summit",
    "Escape Tunnel",
    "Evolving Wilds",
    "Exotic Orchard",
    "Fabled Passage",
    "Forest",
    "Grand Coliseum",
]

try:
    if RETRY_DECK_ID:
        retry_list = "\n".join(f"1\t{name}" for name in FAILED_CARDS)
        retry_url = f"{BASE_URL}/decks/{RETRY_DECK_ID}/import-cards"
        print(f"→ POST {retry_url}")
        print(f"  {len(FAILED_CARDS)} cartas para retry no deck {RETRY_DECK_ID}")
        print("  Aguarde (~60s)…\n")
        result = do_request(retry_url, {"deck_list": retry_list, "set_code": "", "language": "EN"})
        print_result(result, RETRY_DECK_ID)
    else:
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
        card_lines = [l for l in DECK_LIST.splitlines() if l.strip() and l.strip()[0].isdigit()]
        url = f"{BASE_URL}/decks/import-list"
        print(f"→ POST {url}")
        print(f"  {len(card_lines)} entradas na lista, idioma EN")
        print("  Aguarde (~60s)…\n")
        result = do_request(url, payload)
        deck_id = result['deck_id']
        print(f"  Deck ID: {deck_id} — {result.get('deck_name', '')}")
        print_result(result, deck_id)

except urllib.error.HTTPError as e:
    body = e.read().decode()
    print(f"✗ HTTP {e.code}: {body}")
except Exception as e:
    print(f"✗ Erro: {e}")
