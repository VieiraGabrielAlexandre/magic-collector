#!/usr/bin/env python3
"""
Importa o deck "Perrie, the Pulverizer" (NCC) em português.
Uso: python3 import_perrie.py
     python3 import_perrie.py http://34.196.63.122/api   # produção
"""
import json
import sys
import urllib.request
import urllib.error

BASE_URL = sys.argv[1].rstrip("/") if len(sys.argv) > 1 else "http://localhost:8080"

DECK_LIST = """\
Commander
1\tPerrie, the Pulverizer\t\t
Creatures (30)
1\tAven Courier\t\t$ 2.25
1\tDevoted Druid\t\t$ 1.37
1\tGrateful Apparition\t\t$ 0.38
1\tIncubation Druid\t\t$ 2.28
1\tLuminarch Aspirant\t\t$ 0.55
1\tScavenging Ooze\t\t$ 0.35
1\tSkyship Plunderer\t\t$ 0.25
1\tSteelbane Hydra\t\t$ 6.88
1\tThrummingbird\t\t$ 0.37
1\tWall of Roots\t\t$ 0.35
1\tAngelic Sleuth\t\t$ 0.61
1\tAven Mimeomancer\t\t$ 0.40
1\tCrystalline Giant\t\t$ 0.54
1\tEvolution Sage\t\t$ 3.65
1\tJenara, Asura of War\t\t$ 0.45
1\tPark Heights Maverick\t\t$ 0.50
1\tRishkar, Peema Renegade\t\t$ 0.44
1\tVorel of the Hull Clade\t\t$ 0.94
1\tWingspan Mentor\t\t$ 0.24
1\tDenry Klin, Editor in Chief\t\t$ 1.73
1\tFathom Mage\t\t$ 0.49
1\tForgotten Ancient\t\t$ 0.59
1\tKros, Defense Contractor\t\t$ 0.50
1\tSlippery Bogbonder\t\t$ 4.60
1\tWickerbough Elder\t\t$ 0.25
1\tAvenging Huntbonder\t\t$ 0.32
1\tRoalesk, Apex Hybrid\t\t$ 0.50
1\tShield Broker\t\t$ 0.30
1\tSkyboon Evangelist\t\t$ 0.50
1\tBribe Taker\t\t$ 0.30
Planeswalkers (1)
1\tAjani Unyielding\t\t$ 0.75
Spells (13)
1\tDeclaration in Stone\t\t$ 0.35
1\tBant Charm\t\t$ 0.28
1\tBrokers Charm\t\t$ 0.29
1\tContractual Safeguard\t\t$ 1.22
1\tExotic Pets\t\t$ 0.22
1\tGenerous Gift\t\t$ 1.34
1\tStorm of Forms\t\t$ 0.31
1\tTezzeret's Gambit\t\t$ 0.40
1\tBrokers Confluence\t\t$ 1.26
1\tDamning Verdict\t\t$ 9.89
1\tPlanar Outburst\t\t$ 0.33
1\tUrban Evolution\t\t$ 0.25
1\tRishkar's Expertise\t\t$ 3.14
Artifacts (12)
1\tEverflowing Chalice\t\t$ 0.44
1\tSol Ring\t\t$ 1.58
1\tArcane Signet\t\t$ 0.65
1\tFellwar Stone\t\t$ 1.59
1\tGavel of the Righteous\t\t$ 0.52
1\tPower Conduit\t\t$ 8.94
1\tSwiftfoot Boots\t\t$ 2.30
1\tAgent's Toolkit\t\t$ 6.15
1\tCommander's Sphere\t\t$ 0.34
1\tMidnight Clock\t\t$ 0.45
1\tOblivion Stone\t\t$ 0.40
1\tOracle's Vault\t\t$ 0.30
Enchantments (5)
1\tHoofprints of the Stag\t\t$ 0.30
1\tTogether Forever\t\t$ 0.35
1\tFamily's Favor\t\t$ 1.77
1\tPrimal Empathy\t\t$ 0.33
1\tResourceful Defense\t\t$ 2.06
Lands (38)
1\tAsh Barrens\t\t$ 0.25
1\tBant Panorama\t\t$ 0.40
1\tBrokers Hideout\t\t$ 0.79
1\tCanopy Vista\t\t$ 0.41
1\tCommand Tower\t\t$ 0.40
1\tExotic Orchard\t\t$ 0.39
1\tFlooded Grove\t\t$ 0.55
2\tForest\t\t$ 0.40
2\tForest\t\t$ 0.48
1\tForest\t\t$ 0.37
1\tFortified Village\t\t$ 0.35
1\tGavony Township\t\t$ 5.90
2\tIsland\t\t$ 0.38
1\tIsland\t\t$ 0.20
1\tIsland\t\t$ 0.34
1\tKarn's Bastion\t\t$ 3.50
1\tLittjara Mirrorlake\t\t$ 0.25
1\tLlanowar Reborn\t\t$ 0.25
1\tMyriad Landscape\t\t$ 0.40
1\tNesting Grounds\t\t$ 0.47
1\tPath of Ancestry\t\t$ 0.32
2\tPlains\t\t$ 0.50
2\tPlains\t\t$ 0.38
1\tPlains\t\t$ 0.36
1\tPort Town\t\t$ 0.35
1\tPrairie Stream\t\t$ 0.40
1\tSeaside Citadel\t\t$ 0.40
1\tSkycloud Expanse\t\t$ 0.31
1\tSungrass Prairie\t\t$ 0.32
1\tTemple of Mystery\t\t$ 0.33
1\tVivid Creek\t\t$ 0.35
1\tVivid Grove\t\t$ 0.37
1\tVivid Meadow\t\t$ 0.40
"""

payload = {
    "deck_name":   "Perrie, the Pulverizer",
    "set_code":    "ncc",
    "language":    "PT",
    "commander":   True,
    "colors":      "W,U,G",
    "theme_color": "teal",
    "description": "Pré-con New Capenna Commander — Brokers Ascendancy",
    "deck_list":   DECK_LIST,
}

url = f"{BASE_URL}/decks/import-list"
print(f"→ POST {url}")
print(f"  {len([l for l in DECK_LIST.splitlines() if l.strip() and l.strip()[0].isdigit()])} entradas na lista, idioma PT")
print("  Aguarde (~15s)…\n")

data = json.dumps(payload).encode("utf-8")
req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/json"}, method="POST")

try:
    with urllib.request.urlopen(req, timeout=120) as resp:
        result = json.loads(resp.read())
        print(f"✓ {result['imported']} cartas importadas!")
        if result.get("failed", 0):
            print(f"  ⚠ {result['failed']} falhas")
        print(f"  Total na lista: {result['total_from_api']}")
        print(f"  Deck ID: {result['deck_id']} — {result['deck_name']}")
except urllib.error.HTTPError as e:
    body = e.read().decode()
    print(f"✗ HTTP {e.code}: {body}")
except Exception as e:
    print(f"✗ Erro: {e}")
