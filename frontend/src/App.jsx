import { useEffect, useRef, useState } from "react";
import * as XLSX from "xlsx";
import { assignCardToDeck, createBattle, createCard, createDeck, deleteCard, deleteBattle, deleteDeck, evaluateDeck, exportCards, fetchDeckIcon, getCard, importDeckList, importPrecon, listBattles, listCards, listDecks, suggestDecks, updateCard, updateDeck } from "./services/api";
import "./App.css";

const EMPTY_FORM = {
  name: "", color: "", type: "", subtitle: "", collection_number: "",
  rarity: "", set_code: "", language: "PT", year: "", artist: "",
  company: "Wizards of the Coast", foil: false, prerelease: false, commander: false, deck_id: 0, quantity: 1,
  condition: "played", notes: "",
};

const SORT_OPTIONS = [
  { value: "name", label: "Nome" },
  { value: "set_code", label: "Coleção" },
  { value: "color", label: "Cor" },
  { value: "rarity", label: "Raridade" },
  { value: "year", label: "Ano" },
  { value: "collection_number", label: "Nº" },
];

const MTG_COLORS = [
  { code: "W", cls: "w", label: "W" },
  { code: "U", cls: "u", label: "U" },
  { code: "B", cls: "b", label: "B" },
  { code: "R", cls: "r", label: "R" },
  { code: "G", cls: "g", label: "G" },
];

function toggleColor(colors, code) {
  const set = new Set((colors || "").split(",").filter(Boolean));
  if (set.has(code)) set.delete(code); else set.add(code);
  return [...set].join(",");
}

function ColorPips({ colors }) {
  if (!colors) return null;
  const codes = colors.split(",").filter(Boolean);
  if (!codes.length) return null;
  return (
    <span className="color-pips">
      {codes.map((c) => <span key={c} className={`cp cp-${c.toLowerCase()}`}>{c}</span>)}
    </span>
  );
}

// ── Mana / card colors ──────────────────────────────────────────────────
const CARD_COLORS = [
  { icon: "white",    pt: "Branco",   code: "W" },
  { icon: "blue",     pt: "Azul",     code: "U" },
  { icon: "black",    pt: "Preto",    code: "B" },
  { icon: "red",      pt: "Vermelho", code: "R" },
  { icon: "green",    pt: "Verde",    code: "G" },
  { icon: "incolour", pt: "Incolor",  code: "C" },
];

const PTBR_TO_ICON = {
  branco: "white", azul: "blue", preto: "black", preta: "black",
  verde: "green", vermelha: "red", vermelho: "red", incolor: "incolour", incolour: "incolour",
};
const ICON_TO_PTBR = {
  white: "Branco", blue: "Azul", black: "Preto", red: "Vermelho", green: "Verde", incolour: "Incolor",
};
const MTG_CODE_TO_ICON = { W: "white", U: "blue", B: "black", R: "red", G: "green", C: "incolour" };

function colorStrToIconSet(colorStr) {
  const parts = (colorStr || "").split(/[,/]+/).map((s) => s.trim().toLowerCase());
  const set = new Set();
  for (const p of parts) {
    const icon = PTBR_TO_ICON[p];
    if (icon) set.add(icon);
  }
  return set;
}

function iconSetToColorStr(iconSet) {
  return [...iconSet].map((icon) => ICON_TO_PTBR[icon]).filter(Boolean).join("/");
}

function parseCardColorIcons(card) {
  if (card.colors && card.colors !== "[]" && card.colors !== "") {
    try {
      const arr = JSON.parse(card.colors);
      if (Array.isArray(arr) && arr.length > 0) {
        return [...new Set(arr.map((c) => MTG_CODE_TO_ICON[c.toUpperCase()]).filter(Boolean))];
      }
    } catch {}
  }
  if (!card.color) return [];
  return [...colorStrToIconSet(card.color)];
}

function CardColorIcons({ card }) {
  const icons = parseCardColorIcons(card);
  if (!icons.length) return null;
  return (
    <span className="card-color-icons">
      {icons.map((icon) => (
        <img key={icon} src={`/mana-icons/${icon}.svg`} className="mana-icon" alt={icon} />
      ))}
    </span>
  );
}

function ManaColorPicker({ value, onChange }) {
  const iconSet = colorStrToIconSet(value);
  return (
    <div className="mana-color-picker">
      <span className="mana-color-picker-label">Cor</span>
      <div className="mana-color-picker-row">
        {CARD_COLORS.map(({ icon, pt, code }) => {
          const active = iconSet.has(icon);
          return (
            <button key={code} type="button" title={pt}
              className={`mana-picker-btn${active ? " active" : ""}`}
              onClick={() => {
                const s = new Set(iconSet);
                if (active) s.delete(icon); else s.add(icon);
                onChange(iconSetToColorStr(s));
              }}
            >
              <img src={`/mana-icons/${icon}.svg`} alt={pt} />
            </button>
          );
        })}
      </div>
    </div>
  );
}

// ── AI Evaluation markdown renderer ─────────────────────────────────────
function renderInlineBold(text) {
  const parts = text.split(/\*\*(.+?)\*\*/g);
  return parts.map((part, i) => (i % 2 === 1 ? <strong key={i}>{part}</strong> : part));
}

function renderEvalMarkdown(text) {
  if (!text) return null;
  const lines = text.split("\n");
  const elements = [];
  let listItems = [];

  const flushList = (key) => {
    if (listItems.length > 0) {
      elements.push(<ul key={`ul-${key}`} className="eval-list">{listItems}</ul>);
      listItems = [];
    }
  };

  lines.forEach((line, i) => {
    const t = line.trim();
    if (t.startsWith("## ")) {
      flushList(i);
      elements.push(<h3 key={i} className="eval-heading">{t.slice(3)}</h3>);
    } else if (t.startsWith("- ")) {
      listItems.push(<li key={i}>{renderInlineBold(t.slice(2))}</li>);
    } else if (t === "") {
      flushList(i);
    } else {
      flushList(i);
      elements.push(<p key={i} className="eval-para">{renderInlineBold(t)}</p>);
    }
  });
  flushList("end");
  return elements;
}

// ── Deck theme colors ───────────────────────────────────────────────────
const DECK_THEME_COLORS = [
  { id: "",          label: "Nenhuma"                                                   },
  { id: "crimson",   label: "Vermelho",  bg: "#3a0c0c", border: "#7a2020", text: "#ff9090" },
  { id: "sapphire",  label: "Azul",      bg: "#081428", border: "#1a4080", text: "#80c0ff" },
  { id: "emerald",   label: "Verde",     bg: "#071a0c", border: "#1a5828", text: "#80d890" },
  { id: "violet",    label: "Roxo",      bg: "#150a28", border: "#4a2890", text: "#c090f0" },
  { id: "gold",      label: "Dourado",   bg: "#221404", border: "#805010", text: "#e8c060" },
  { id: "teal",      label: "Turquesa",  bg: "#041820", border: "#0c6878", text: "#60d0c8" },
  { id: "ember",     label: "Laranja",   bg: "#280e04", border: "#904010", text: "#f09050" },
  { id: "silver",    label: "Prata",     bg: "#0e1218", border: "#384858", text: "#a0b8c8" },
  { id: "rose",      label: "Rosa",      bg: "#280a18", border: "#8a1e50", text: "#f080b0" },
  { id: "bone",      label: "Marfim",    bg: "#1c1608", border: "#6a5828", text: "#e0d8b0" },
];

function getDeckTheme(themeId) {
  return DECK_THEME_COLORS.find((c) => c.id === themeId) || DECK_THEME_COLORS[0];
}

function getDeckItemStyle(themeId) {
  const t = getDeckTheme(themeId);
  if (!t?.bg) return {};
  return { background: t.bg, borderColor: t.border, borderLeftColor: t.border };
}

function getDeckBadgeStyle(themeId) {
  const t = getDeckTheme(themeId);
  if (!t?.bg) return {};
  return { background: t.bg, borderColor: t.border, color: t.text };
}

function DeckColorSelect({ value, onChange }) {
  const theme = getDeckTheme(value);
  return (
    <label>
      Cor do deck
      <div className="deck-color-select-row">
        {theme?.bg && <span className="deck-color-swatch" style={{ background: theme.bg, borderColor: theme.border }} />}
        <select
          value={value}
          onChange={(e) => onChange(e.target.value)}
          style={theme?.bg ? { background: theme.bg, borderColor: theme.border, color: theme.text } : {}}
        >
          {DECK_THEME_COLORS.map((c) => (
            <option key={c.id} value={c.id}>{c.label}</option>
          ))}
        </select>
      </div>
    </label>
  );
}

export default function App() {
  const [form, setForm] = useState(EMPTY_FORM);

  const [cards, setCards] = useState([]);
  const [total, setTotal] = useState(0);
  const [totalQuantity, setTotalQuantity] = useState(0);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  const [search, setSearch] = useState("");
  const [sort, setSort] = useState("name");
  const [order, setOrder] = useState("asc");
  const [filterFoil, setFilterFoil] = useState("");
  const [filterRarity, setFilterRarity] = useState("");
  const [filterDeck, setFilterDeck] = useState("");
  const [deckBuilderModal, setDeckBuilderModal] = useState(false);
  const [deckBuilderLoading, setDeckBuilderLoading] = useState(false);
  const [deckBuilderResult, setDeckBuilderResult] = useState(null);

  const [selectedCard, setSelectedCard] = useState(null);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [editForm, setEditForm] = useState({});
  const [propagate, setPropagate] = useState(true);

  const [activeTab, setActiveTab] = useState("collection");
  const [decks, setDecks] = useState([]);
  const [deckForm, setDeckForm] = useState({ name: "", description: "", commander: false, colors: "", set_code: "", theme_color: "" });
  const [editDeckModal, setEditDeckModal] = useState(null);

  const [managingDeck, setManagingDeck] = useState(null);
  const [deckCards, setDeckCards] = useState([]);
  const [unassignedCards, setUnassignedCards] = useState([]);
  const [unassignedPage, setUnassignedPage] = useState(1);
  const [unassignedTotalPages, setUnassignedTotalPages] = useState(1);
  const [unassignedTotal, setUnassignedTotal] = useState(0);
  const [deckSearch, setDeckSearch] = useState("");
  const [deckInnerTab, setDeckInnerTab] = useState("cards");
  const [evaluating, setEvaluating] = useState(false);

  const EMPTY_IMPORT_FORM = { set_code: "", deck_name: "", language: "PT", colors: "", commander: false, theme_color: "", description: "" };
  const [importModal, setImportModal] = useState(false);
  const [importForm, setImportForm] = useState(EMPTY_IMPORT_FORM);
  const [importLoading, setImportLoading] = useState(false);
  const [importResult, setImportResult] = useState(null);
  const [importError, setImportError] = useState("");

  const [battles, setBattles] = useState([]);
  const EMPTY_BATTLE_FORM = { result: "win", opponents: ["", "", ""], player_count: 4, game_style: "Commander", deck_id: 0, deck_name: "", deck_is_mine: true, notes: "" };
  const [battleForm, setBattleForm] = useState(EMPTY_BATTLE_FORM);

  const EMPTY_LIST_FORM = { deck_name: "", set_code: "", language: "PT", colors: "", commander: false, theme_color: "", description: "", deck_list: "" };
  const [listModal, setListModal] = useState(false);
  const [listForm, setListForm] = useState(EMPTY_LIST_FORM);
  const [listLoading, setListLoading] = useState(false);
  const [listResult, setListResult] = useState(null);
  const [listError, setListError] = useState("");

  const searchTimer = useRef(null);
  const deckSearchTimer = useRef(null);

  async function loadCards(opts = {}) {
    const deckId = (opts.filterDeck ?? filterDeck) === "" ? undefined : parseInt(opts.filterDeck ?? filterDeck);
    const result = await listCards({
      q: opts.q ?? search,
      page: opts.page ?? page,
      sort: opts.sort ?? sort,
      order: opts.order ?? order,
      pageSize: 20,
      foil: opts.filterFoil ?? filterFoil,
      rarity: opts.filterRarity ?? filterRarity,
      deckId,
    });
    setCards(result.data ?? []);
    setTotal(result.total ?? 0);
    setTotalQuantity(result.total_quantity ?? 0);
    setPage(result.page ?? 1);
    setTotalPages(result.total_pages ?? 1);
  }

  async function loadDecks() {
    const data = await listDecks();
    setDecks(data ?? []);
  }

  async function loadBattles() {
    const data = await listBattles();
    setBattles(data ?? []);
  }

  async function handleBattleSubmit(e) {
    e.preventDefault();
    await createBattle(battleForm);
    setBattleForm(EMPTY_BATTLE_FORM);
    await loadBattles();
  }

  async function handleBattleDelete(id) {
    await deleteBattle(id);
    await loadBattles();
  }

  async function loadDeckCards(deckId) {
    const dc = await listCards({ deckId, pageSize: 500 });
    setDeckCards(dc.data ?? []);
  }

  async function loadUnassigned(pg = 1, q = "") {
    const result = await listCards({ deckId: 0, pageSize: 20, page: pg, q });
    setUnassignedCards(result.data ?? []);
    setUnassignedPage(result.page ?? 1);
    setUnassignedTotalPages(result.total_pages ?? 1);
    setUnassignedTotal(result.total ?? 0);
  }

  useEffect(() => { loadCards(); }, [sort, order, filterFoil, filterRarity, filterDeck]);
  useEffect(() => { loadDecks(); }, []);
  useEffect(() => { loadBattles(); }, []);

  function handleSearchChange(e) {
    const q = e.target.value;
    setSearch(q);
    clearTimeout(searchTimer.current);
    searchTimer.current = setTimeout(() => {
      setPage(1);
      loadCards({ q, page: 1 });
    }, 350);
  }

  function handleSort(field) {
    if (field === sort) {
      setOrder((o) => (o === "asc" ? "desc" : "asc"));
    } else {
      setSort(field);
      setOrder("asc");
    }
    setPage(1);
  }

  function handlePage(p) {
    setPage(p);
    loadCards({ page: p });
  }

  async function handleSubmit(e) {
    e.preventDefault();
    await createCard({ ...form, quantity: Number(form.quantity), year: Number(form.year) || 0 });
    setForm(EMPTY_FORM);
    setPage(1);
    loadCards({ page: 1 });
  }

  async function handleDetails(id) {
    setLoadingDetail(true);
    setSelectedCard(null);
    try {
      const data = await getCard(id);
      setSelectedCard(data);
    } finally {
      setLoadingDetail(false);
    }
  }

  async function handleDelete(id) {
    await deleteCard(id);
    setSelectedCard(null);
    loadCards();
  }

  function handleEditStart() {
    const c = selectedCard.local;
    setEditForm({
      name: c.name, color: c.color, type: c.type, subtitle: c.subtitle,
      collection_number: c.collection_number, rarity: c.rarity, set_code: c.set_code,
      language: c.language, year: c.year, artist: c.artist, company: c.company,
      foil: c.foil, prerelease: c.prerelease, commander: c.commander, deck_id: c.deck_id ?? 0, quantity: c.quantity, condition: c.condition, notes: c.notes,
    });
    setEditMode(true);
  }

  async function handleEditSave() {
    await updateCard(selectedCard.local.id, { ...editForm, year: Number(editForm.year) || 0, quantity: Number(editForm.quantity) || 1, propagate });
    setEditMode(false);
    loadCards();
    handleDetails(selectedCard.local.id);
  }

  const EXPORT_HEADERS = ["name","color","type","subtitle","collection_number","rarity","set_code","language","year","artist","company","foil","quantity","condition","notes","commander","deck"];

  function cardToRow(c) {
    const deckName = c.deck_id > 0 ? (decks.find(d => d.id === c.deck_id)?.name ?? "") : "";
    return [c.name, c.color, c.type, c.subtitle, c.collection_number, c.rarity,
            c.set_code, c.language, c.year, c.artist, c.company,
            c.foil ? "sim" : "nao", c.quantity, c.condition, c.notes,
            c.commander ? "sim" : "nao", deckName];
  }

  async function handleDeckCreate(e) {
    e.preventDefault();
    await createDeck(deckForm);
    setDeckForm({ name: "", description: "", commander: false, colors: "", set_code: "", theme_color: "" });
    loadDecks();
  }

  async function handleDeckDelete(id) {
    await deleteDeck(id);
    loadDecks();
  }

  async function handleDeckUpdate(e) {
    e.preventDefault();
    await updateDeck(editDeckModal.id, {
      name: editDeckModal.name,
      description: editDeckModal.description || "",
      commander: editDeckModal.commander || false,
      colors: editDeckModal.colors || "",
      set_code: editDeckModal.set_code || "",
      theme_color: editDeckModal.theme_color || "",
    });
    setEditDeckModal(null);
    loadDecks();
  }

  async function handleManageDeck(deck) {
    setManagingDeck(deck);
    setDeckSearch("");
    setDeckInnerTab("cards");
    setUnassignedPage(1);
    await Promise.all([loadDeckCards(deck.id), loadUnassigned(1, "")]);
    if (deck.set_code && !deck.icon_uri) {
      const result = await fetchDeckIcon(deck.id);
      if (result?.icon_uri) {
        setManagingDeck((prev) => ({ ...prev, icon_uri: result.icon_uri }));
        loadDecks();
      }
    }
  }

  async function handleSuggestDecks() {
    setDeckBuilderModal(true);
    setDeckBuilderLoading(true);
    setDeckBuilderResult(null);
    try {
      const result = await suggestDecks();
      setDeckBuilderResult(result);
    } catch (e) {
      setDeckBuilderResult({ error: e.message });
    } finally {
      setDeckBuilderLoading(false);
    }
  }

  async function handleEvaluate() {
    setEvaluating(true);
    try {
      const result = await evaluateDeck(managingDeck.id);
      setManagingDeck((prev) => ({ ...prev, evaluation: result.evaluation, evaluated_at: result.evaluated_at }));
      setDecks((prev) => prev.map((d) => d.id === managingDeck.id ? { ...d, evaluation: result.evaluation, evaluated_at: result.evaluated_at } : d));
    } catch (e) {
      alert("Erro ao gerar avaliação: " + e.message);
    } finally {
      setEvaluating(false);
    }
  }

  async function handleAssignCard(cardId) {
    await assignCardToDeck(cardId, managingDeck.id);
    await Promise.all([loadDeckCards(managingDeck.id), loadUnassigned(unassignedPage, deckSearch)]);
    loadDecks();
  }

  async function handleUnassignCard(cardId) {
    await assignCardToDeck(cardId, 0);
    await Promise.all([loadDeckCards(managingDeck.id), loadUnassigned(unassignedPage, deckSearch)]);
    loadDecks();
  }

  async function handleExportCSV() {
    const data = await exportCards();
    const escape = (v) => `"${String(v ?? "").replace(/"/g, '""')}"`;
    const lines = [EXPORT_HEADERS.join(","), ...data.map((c) => cardToRow(c).map(escape).join(","))];
    const blob = new Blob([lines.join("\n")], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url; a.download = "colecao.csv"; a.click();
    URL.revokeObjectURL(url);
  }

  async function handleExportXLSX() {
    const data = await exportCards();
    const rows = data.map(cardToRow);
    const ws = XLSX.utils.aoa_to_sheet([EXPORT_HEADERS, ...rows]);
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, "Coleção");
    XLSX.writeFile(wb, "colecao.xlsx");
  }

  async function handleImportList(e) {
    e.preventDefault();
    setListLoading(true);
    setListError("");
    setListResult(null);
    try {
      const result = await importDeckList(listForm);
      setListResult(result);
      await loadDecks();
    } catch (err) {
      setListError(err.message);
    } finally {
      setListLoading(false);
    }
  }

  async function handleImportPrecon(e) {
    e.preventDefault();
    setImportLoading(true);
    setImportError("");
    setImportResult(null);
    try {
      const result = await importPrecon(importForm);
      setImportResult(result);
      await loadDecks();
    } catch (err) {
      setImportError(err.message);
    } finally {
      setImportLoading(false);
    }
  }

  const field = (label, key, extra = {}) => (
    <label>
      {label}
      <input value={form[key]} onChange={(e) => setForm({ ...form, [key]: e.target.value })} {...extra} />
    </label>
  );

  const orderIcon = (f) => {
    if (sort !== f) return null;
    return order === "asc" ? " ↑" : " ↓";
  };

  const deckTotalQuantity = deckCards.reduce((sum, c) => sum + (c.quantity || 1), 0);

  return (
    <main className="app">
      <section className="hero">
        <h1>Magic Collector</h1>
        <p>Cadastre, organize e consulte sua coleção de cartas Magic: The Gathering</p>
      </section>

      <nav className="tabs">
        <button type="button" className={`tab${activeTab === "collection" ? " active" : ""}`} onClick={() => setActiveTab("collection")}>Coleção</button>
        <button type="button" className={`tab${activeTab === "decks" ? " active" : ""}`} onClick={() => setActiveTab("decks")}>Decks</button>
        <button type="button" className={`tab${activeTab === "battles" ? " active" : ""}`} onClick={() => setActiveTab("battles")}>⚔ Batalhas</button>
      </nav>

      {activeTab === "decks" && (
        managingDeck ? (
          <div className="deck-manage">
            <div className="deck-manage-header">
              <button type="button" className="back-btn" onClick={() => setManagingDeck(null)}>← Voltar</button>
              <div className="deck-manage-title">
                {managingDeck.icon_uri
                  ? <img src={managingDeck.icon_uri} className="deck-icon-lg" alt={managingDeck.set_code} />
                  : managingDeck.set_code && <span className="deck-set-code-lg">{managingDeck.set_code.toUpperCase()}</span>
                }
                <div className="deck-manage-title-text">
                  <h2>
                    {managingDeck.name}
                    {managingDeck.commander && <span className="commander-badge">CMD</span>}
                  </h2>
                  <div className="deck-manage-meta">
                    <ColorPips colors={managingDeck.colors} />
                    <span className="total-badge">{deckTotalQuantity} cartas</span>
                    <span className="unique-badge">{deckCards.length} únicas</span>
                    <span className="unique-badge">{unassignedTotal} sem deck</span>
                  </div>
                </div>
              </div>
            </div>

            <div className="deck-inner-tabs">
              <button type="button" className={`deck-inner-tab${deckInnerTab === "cards" ? " active" : ""}`} onClick={() => setDeckInnerTab("cards")}>🃏 Cartas</button>
              <button type="button" className={`deck-inner-tab${deckInnerTab === "eval" ? " active" : ""}`} onClick={() => setDeckInnerTab("eval")}>
                🤖 Avaliação IA
                {managingDeck.evaluated_at && <span className="eval-badge">✓</span>}
              </button>
            </div>

            {deckInnerTab === "cards" && (
              <div className="deck-manage-grid">
                <section className="card list-section">
                  <div className="list-header">
                    <h3 className="section-title">No deck <span className="total-badge">{deckTotalQuantity}</span><span className="unique-badge">{deckCards.length} únicas</span></h3>
                  </div>
                  <div className="list">
                    {deckCards.map((card) => (
                      <div className={`list-item${card.foil ? " is-foil" : ""} item-r-${(card.rarity || "x").toLowerCase()}`} key={card.id}>
                        <div className="list-item-info">
                          <div className="list-item-name">
                            <strong className={card.foil ? "foil-text" : ""}>{card.name}</strong>
                            {card.foil && <span className="foil-text">✦</span>}
                            <CardColorIcons card={card} />
                            {card.rarity && <span className={`rarity r-${card.rarity.toLowerCase()}`}>{card.rarity}</span>}
                          </div>
                          <small>{card.set_code || "—"} · #{card.collection_number || "—"} · {card.language} · ×{card.quantity}</small>
                        </div>
                        <div className="actions">
                          <button type="button" className="danger" onClick={() => handleUnassignCard(card.id)}>Remover</button>
                        </div>
                      </div>
                    ))}
                    {deckCards.length === 0 && <p className="empty">Nenhuma carta no deck.</p>}
                  </div>
                </section>

                <section className="card list-section">
                  <div className="list-header">
                    <div className="list-header-top">
                      <h3 className="section-title">Sem deck <span className="unique-badge">{unassignedTotal}</span></h3>
                    </div>
                    <input className="search-input" type="search" placeholder="Filtrar por nome ou coleção..."
                      value={deckSearch} onChange={(e) => {
                        const q = e.target.value;
                        setDeckSearch(q);
                        clearTimeout(deckSearchTimer.current);
                        deckSearchTimer.current = setTimeout(() => loadUnassigned(1, q), 350);
                      }} />
                  </div>
                  <div className="list">
                    {unassignedCards.map((card) => (
                      <div className={`list-item${card.foil ? " is-foil" : ""} item-r-${(card.rarity || "x").toLowerCase()}`} key={card.id}>
                        <div className="list-item-info">
                          <div className="list-item-name">
                            <strong className={card.foil ? "foil-text" : ""}>{card.name}</strong>
                            {card.foil && <span className="foil-text">✦</span>}
                            <CardColorIcons card={card} />
                            {card.rarity && <span className={`rarity r-${card.rarity.toLowerCase()}`}>{card.rarity}</span>}
                          </div>
                          <small>{card.set_code || "—"} · #{card.collection_number || "—"} · {card.language} · ×{card.quantity}</small>
                        </div>
                        <div className="actions">
                          <button type="button" onClick={() => handleAssignCard(card.id)}>+ Deck</button>
                        </div>
                      </div>
                    ))}
                    {unassignedCards.length === 0 && <p className="empty">Nenhuma carta disponível.</p>}
                  </div>
                  {unassignedTotalPages > 1 && (
                    <div className="pagination">
                      <button type="button" onClick={() => loadUnassigned(unassignedPage - 1, deckSearch)} disabled={unassignedPage <= 1}>‹</button>
                      {Array.from({ length: unassignedTotalPages }, (_, i) => i + 1)
                        .filter((p) => p === 1 || p === unassignedTotalPages || Math.abs(p - unassignedPage) <= 2)
                        .reduce((acc, p, idx, arr) => {
                          if (idx > 0 && p - arr[idx - 1] > 1) acc.push("…");
                          acc.push(p);
                          return acc;
                        }, [])
                        .map((p, i) =>
                          p === "…"
                            ? <span key={`e-${i}`} className="pagination-ellipsis">…</span>
                            : <button key={p} type="button" className={p === unassignedPage ? "active" : ""} onClick={() => loadUnassigned(p, deckSearch)}>{p}</button>
                        )}
                      <button type="button" onClick={() => loadUnassigned(unassignedPage + 1, deckSearch)} disabled={unassignedPage >= unassignedTotalPages}>›</button>
                    </div>
                  )}
                </section>
              </div>
            )}

            {deckInnerTab === "eval" && (
              <div className="eval-panel">
                {evaluating ? (
                  <div className="eval-loading">
                    <div className="eval-spinner">⚙</div>
                    <p className="eval-loading-text">Analisando o deck com IA… pode levar alguns segundos.</p>
                  </div>
                ) : managingDeck.evaluation ? (
                  <>
                    <div className="eval-header">
                      <span className="eval-meta">Gerado em {managingDeck.evaluated_at?.replace("T", " ") ?? ""}</span>
                      <button type="button" className="eval-redo-btn" onClick={handleEvaluate} disabled={evaluating}>♻ Refazer Avaliação</button>
                    </div>
                    <div className="eval-content">{renderEvalMarkdown(managingDeck.evaluation)}</div>
                  </>
                ) : (
                  <div className="eval-empty">
                    <div className="eval-empty-icon">🤖</div>
                    <p>Nenhuma avaliação gerada para este deck ainda.</p>
                    <p className="eval-empty-sub">A IA irá analisar todas as cartas e gerar uma avaliação estratégica completa.</p>
                    <button type="button" className="eval-generate-btn" onClick={handleEvaluate} disabled={evaluating}>
                      ✨ Gerar Avaliação com IA
                    </button>
                  </div>
                )}
              </div>
            )}
          </div>
        ) : (
          <section className="grid">
            <form className="card form" onSubmit={handleDeckCreate}>
              <h2>Novo Deck</h2>
              <label>Nome *<input required value={deckForm.name} onChange={(e) => setDeckForm({ ...deckForm, name: e.target.value })} placeholder="Ex: Eldrazi Unbound" /></label>
              <label>
                Código do set
                <input value={deckForm.set_code} onChange={(e) => setDeckForm({ ...deckForm, set_code: e.target.value })} placeholder="Ex: dmu, bro, mkm" />
                <span className="field-hint">Usado para buscar o ícone do set no Scryfall</span>
              </label>
              <label>Descrição<textarea value={deckForm.description} onChange={(e) => setDeckForm({ ...deckForm, description: e.target.value })} placeholder="Notas sobre o deck..." /></label>
              <label className="checkbox-label">
                <input type="checkbox" checked={deckForm.commander} onChange={(e) => setDeckForm({ ...deckForm, commander: e.target.checked })} />
                Commander
              </label>
              <div className="color-picker">
                <span className="color-picker-label">Cores do deck</span>
                <div className="color-pips-row">
                  {MTG_COLORS.map(({ code, cls }) => (
                    <label key={code} className={`color-pip-check cp-${cls}${(deckForm.colors||"").split(",").filter(Boolean).includes(code) ? " selected" : ""}`}>
                      <input type="checkbox"
                        checked={(deckForm.colors||"").split(",").filter(Boolean).includes(code)}
                        onChange={() => setDeckForm({ ...deckForm, colors: toggleColor(deckForm.colors, code) })} />
                      {code}
                    </label>
                  ))}
                </div>
              </div>
              <DeckColorSelect value={deckForm.theme_color} onChange={(v) => setDeckForm({ ...deckForm, theme_color: v })} />
              <button type="submit">Criar Deck</button>
            </form>

            <section className="card list-section">
              <div className="list-header">
                <div className="list-header-top">
                  <h2>Meus Decks <span className="total-badge">{decks.length}</span></h2>
                  <button type="button" className="import-precon-btn" onClick={() => { setImportModal(true); setImportResult(null); setImportError(""); setImportForm(EMPTY_IMPORT_FORM); }}>Importar Pré-con</button>
                  <button type="button" className="import-precon-btn" onClick={() => { setListModal(true); setListResult(null); setListError(""); setListForm(EMPTY_LIST_FORM); }}>Importar Lista</button>
                </div>
              </div>
              <div className="list deck-list">
                {decks.map((deck) => (
                  <div className="deck-list-item" key={deck.id} style={getDeckItemStyle(deck.theme_color)}>
                    <div className="deck-list-icon">
                      {deck.icon_uri
                        ? <img src={deck.icon_uri} className="set-icon" alt={deck.set_code} />
                        : <span className="set-icon-placeholder">{deck.set_code ? deck.set_code.slice(0,3).toUpperCase() : "—"}</span>
                      }
                    </div>
                    <div className="deck-list-body">
                      <div className="deck-list-name">
                        <strong>{deck.name}</strong>
                        {deck.commander && <span className="commander-badge">CMD</span>}
                        <ColorPips colors={deck.colors} />
                      </div>
                      <div className="deck-list-meta">
                        {deck.set_code && <span className="deck-set-label">{deck.set_code.toUpperCase()}</span>}
                        <span className="total-badge">{deck.card_count} cartas</span>
                        {deck.description && <span className="deck-desc">{deck.description}</span>}
                      </div>
                    </div>
                    <div className="actions">
                      <button type="button" onClick={() => handleManageDeck(deck)}>Cartas</button>
                      <button type="button" onClick={() => setEditDeckModal({ ...deck })}>Editar</button>
                      <button type="button" className="danger" onClick={() => handleDeckDelete(deck.id)}>✕</button>
                    </div>
                  </div>
                ))}
                {decks.length === 0 && <p className="empty">Nenhum deck cadastrado.</p>}
              </div>
            </section>
          </section>
        )
      )}

      {/* ── ABA BATALHAS ── */}
      {activeTab === "battles" && (() => {
        const wins   = battles.filter(b => b.result === "win").length;
        const losses = battles.filter(b => b.result === "loss").length;
        const total  = battles.length;
        const rate   = total > 0 ? Math.round((wins / total) * 100) : 0;
        return (
          <div className="battles-page">

            {/* Stats */}
            <div className="battle-stats">
              <div className="bstat bstat-total">
                <span className="bstat-num">{total}</span>
                <span className="bstat-label">Batalhas</span>
              </div>
              <div className="bstat bstat-win">
                <span className="bstat-num">{wins}</span>
                <span className="bstat-label">⚔ Vitórias</span>
              </div>
              <div className="bstat bstat-loss">
                <span className="bstat-num">{losses}</span>
                <span className="bstat-label">💀 Derrotas</span>
              </div>
              <div className="bstat bstat-rate">
                <span className="bstat-num">{rate}%</span>
                <span className="bstat-label">Taxa de vitória</span>
              </div>
            </div>

            <div className="battles-grid">
              {/* Formulário */}
              <section className="card battle-form-card">
                <h2>Registrar Batalha</h2>
                <form onSubmit={handleBattleSubmit}>

                  {/* Resultado */}
                  <div className="result-toggle">
                    <button type="button"
                      className={`result-btn result-win${battleForm.result === "win" ? " active" : ""}`}
                      onClick={() => setBattleForm({ ...battleForm, result: "win" })}>
                      ⚔ Vitória
                    </button>
                    <button type="button"
                      className={`result-btn result-loss${battleForm.result === "loss" ? " active" : ""}`}
                      onClick={() => setBattleForm({ ...battleForm, result: "loss" })}>
                      💀 Derrota
                    </button>
                  </div>

                  <div className="battle-form-row">
                    <label>Nº de jogadores
                      <input type="number" min="2" max="8" value={battleForm.player_count}
                        onChange={e => {
                          const n = Math.max(2, Math.min(8, +e.target.value));
                          const opp = Array.from({ length: n - 1 }, (_, i) => battleForm.opponents[i] ?? "");
                          setBattleForm({ ...battleForm, player_count: n, opponents: opp });
                        }} />
                    </label>
                    <label>Formato
                      <select value={battleForm.game_style}
                        onChange={e => setBattleForm({ ...battleForm, game_style: e.target.value })}>
                        <option value="Commander">Commander</option>
                        <option value="1v1">1v1</option>
                        <option value="Draft">Draft</option>
                        <option value="Sealed">Sealed</option>
                        <option value="Cube">Cube</option>
                        <option value="Casual">Casual</option>
                      </select>
                    </label>
                  </div>

                  {/* Deck */}
                  <div className="deck-origin-toggle">
                    <button type="button"
                      className={`origin-btn${battleForm.deck_is_mine ? " active" : ""}`}
                      onClick={() => setBattleForm({ ...battleForm, deck_is_mine: true, deck_id: 0, deck_name: "" })}>
                      Meu deck
                    </button>
                    <button type="button"
                      className={`origin-btn${!battleForm.deck_is_mine ? " active" : ""}`}
                      onClick={() => setBattleForm({ ...battleForm, deck_is_mine: false, deck_id: 0, deck_name: "" })}>
                      Deck externo
                    </button>
                  </div>

                  <div className="opponents-list">
                    {battleForm.opponents.map((name, i) => (
                      <label key={i}>
                        Oponente {i + 1}
                        <input placeholder={`Nome do oponente ${i + 1}`} value={name}
                          onChange={e => {
                            const opp = [...battleForm.opponents];
                            opp[i] = e.target.value;
                            setBattleForm({ ...battleForm, opponents: opp });
                          }} />
                      </label>
                    ))}
                  </div>

                  {battleForm.deck_is_mine ? (
                    <label>Deck utilizado
                      <select value={battleForm.deck_id}
                        onChange={e => {
                          const id = +e.target.value;
                          const d = decks.find(d => d.id === id);
                          setBattleForm({ ...battleForm, deck_id: id, deck_name: d ? d.name : "" });
                        }}>
                        <option value={0}>— selecione —</option>
                        {decks.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
                      </select>
                    </label>
                  ) : (
                    <label>Nome do deck externo
                      <input placeholder="Ex: Ur-Dragon precon" value={battleForm.deck_name}
                        onChange={e => setBattleForm({ ...battleForm, deck_name: e.target.value })} />
                    </label>
                  )}

                  <label>Observações
                    <textarea rows={2} placeholder="Opcional…" value={battleForm.notes}
                      onChange={e => setBattleForm({ ...battleForm, notes: e.target.value })} />
                  </label>

                  <button type="submit" className={`submit-battle${battleForm.result === "win" ? " win" : " loss"}`}>
                    {battleForm.result === "win" ? "⚔ Registrar Vitória" : "💀 Registrar Derrota"}
                  </button>
                </form>
              </section>

              {/* Lista */}
              <section className="card list-section">
                <div className="list-header">
                  <div className="list-header-top">
                    <h2>Histórico <span className="total-badge">{total}</span></h2>
                  </div>
                </div>
                <div className="list battles-list">
                  {battles.map(b => {
                    const isWin = b.result === "win";
                    const date = b.played_at ? new Date(b.played_at).toLocaleDateString("pt-BR", { day: "2-digit", month: "short", year: "numeric" }) : "";
                    return (
                      <div key={b.id} className={`battle-item ${isWin ? "battle-win" : "battle-loss"}`}>
                        <div className="battle-result-icon">{isWin ? "⚔" : "💀"}</div>
                        <div className="battle-info">
                          <div className="battle-main">
                            <strong className="battle-deck-name">{b.deck_name || "—"}</strong>
                            {!b.deck_is_mine && <span className="ext-badge">externo</span>}
                            <span className={`battle-result-label ${isWin ? "lbl-win" : "lbl-loss"}`}>
                              {isWin ? "VITÓRIA" : "DERROTA"}
                            </span>
                          </div>
                          <div className="battle-sub">
                            {b.opponents?.filter(Boolean).length > 0 && (
                              <span>vs <em>{b.opponents.filter(Boolean).join(", ")}</em></span>
                            )}
                            {b.game_style && <span className="battle-tag">{b.game_style}</span>}
                            {b.player_count > 0 && <span className="battle-tag">{b.player_count}p</span>}
                            {date && <span className="battle-date">{date}</span>}
                          </div>
                          {b.notes && <div className="battle-notes">{b.notes}</div>}
                        </div>
                        <button type="button" className="danger battle-del" onClick={() => handleBattleDelete(b.id)}>✕</button>
                      </div>
                    );
                  })}
                  {battles.length === 0 && <p className="empty">Nenhuma batalha registrada ainda.</p>}
                </div>
              </section>
            </div>
          </div>
        );
      })()}

      {/* ── MODAL DECK BUILDER IA ── */}
      {deckBuilderModal && (
        <div className="modal-overlay" onClick={() => !deckBuilderLoading && setDeckBuilderModal(false)}>
          <div className="modal deck-builder-modal" onClick={(e) => e.stopPropagation()}>
            <div className="deck-builder-modal-header">
              <h2>✨ Montar Deck com IA</h2>
              {!deckBuilderLoading && (
                <button className="modal-close" onClick={() => setDeckBuilderModal(false)}>✕</button>
              )}
            </div>
            {deckBuilderLoading ? (
              <div className="eval-loading">
                <div className="eval-spinner">⚙</div>
                <p className="eval-loading-text">Analisando suas cartas e montando sugestões de decks…</p>
                <p className="eval-loading-sub">Isso pode levar alguns segundos.</p>
              </div>
            ) : deckBuilderResult?.error ? (
              <div className="eval-empty">
                <div className="eval-empty-icon">⚠️</div>
                <p>{deckBuilderResult.error}</p>
              </div>
            ) : deckBuilderResult ? (
              <>
                <p className="deck-builder-meta">{deckBuilderResult.card_count} cartas únicas analisadas</p>
                <div className="eval-content deck-builder-content">
                  {renderEvalMarkdown(deckBuilderResult.analysis)}
                </div>
              </>
            ) : null}
          </div>
        </div>
      )}

      {/* ── MODAL EDITAR DECK ── */}
      {editDeckModal && (
        <div className="modal-overlay" onClick={() => setEditDeckModal(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => setEditDeckModal(null)}>✕</button>
            <form className="edit-form" onSubmit={handleDeckUpdate}>
              <h2>Editar Deck</h2>
              <div className="edit-grid">
                <label>Nome *<input required value={editDeckModal.name} onChange={(e) => setEditDeckModal({ ...editDeckModal, name: e.target.value })} /></label>
                <label>
                  Código do set
                  <input value={editDeckModal.set_code || ""} onChange={(e) => setEditDeckModal({ ...editDeckModal, set_code: e.target.value })} placeholder="Ex: dmu, bro" />
                </label>
              </div>
              <label>Descrição<textarea value={editDeckModal.description || ""} onChange={(e) => setEditDeckModal({ ...editDeckModal, description: e.target.value })} /></label>
              <label className="checkbox-label">
                <input type="checkbox" checked={editDeckModal.commander || false} onChange={(e) => setEditDeckModal({ ...editDeckModal, commander: e.target.checked })} />
                Commander
              </label>
              <div className="color-picker">
                <span className="color-picker-label">Cores do deck</span>
                <div className="color-pips-row">
                  {MTG_COLORS.map(({ code, cls }) => (
                    <label key={code} className={`color-pip-check cp-${cls}${(editDeckModal.colors||"").split(",").filter(Boolean).includes(code) ? " selected" : ""}`}>
                      <input type="checkbox"
                        checked={(editDeckModal.colors||"").split(",").filter(Boolean).includes(code)}
                        onChange={() => setEditDeckModal({ ...editDeckModal, colors: toggleColor(editDeckModal.colors || "", code) })} />
                      {code}
                    </label>
                  ))}
                </div>
              </div>
              <DeckColorSelect value={editDeckModal.theme_color || ""} onChange={(v) => setEditDeckModal({ ...editDeckModal, theme_color: v })} />
              <div className="modal-actions">
                <button type="button" onClick={() => setEditDeckModal(null)}>Cancelar</button>
                <button type="submit" className="save-btn">Salvar</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── MODAL IMPORTAR LISTA ── */}
      {listModal && (
        <div className="modal-overlay" onClick={() => { if (!listLoading) setListModal(false); }}>
          <div className="modal modal-import" onClick={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => { if (!listLoading) setListModal(false); }}>✕</button>
            <h2>Importar Lista de Cartas</h2>

            {listResult ? (
              <div className="import-result">
                <div className="import-result-icon">✓</div>
                <p><strong>{listResult.imported}</strong> cartas importadas com sucesso!</p>
                <p className="import-result-sub">{listResult.failed > 0 && <span>{listResult.failed} falhas · </span>}{listResult.total_from_api} cartas na lista · Deck: <em>{listResult.deck_name}</em></p>
                <div className="modal-actions">
                  <button type="button" onClick={() => { setListModal(false); setListResult(null); }}>Fechar</button>
                  <button type="button" className="save-btn" onClick={() => { setListResult(null); setListForm(EMPTY_LIST_FORM); }}>Importar outro</button>
                </div>
              </div>
            ) : (
              <form className="edit-form" onSubmit={handleImportList}>
                <div className="edit-grid">
                  <label>
                    Nome do deck *
                    <input required placeholder="Ex: Perrie, the Pulverizer" value={listForm.deck_name}
                      onChange={(e) => setListForm({ ...listForm, deck_name: e.target.value })} />
                  </label>
                  <label>
                    Código do set (opcional)
                    <input placeholder="Ex: ncc, dmu" value={listForm.set_code}
                      onChange={(e) => setListForm({ ...listForm, set_code: e.target.value.toLowerCase().trim() })} />
                  </label>
                </div>
                <label>
                  Idioma
                  <select value={listForm.language} onChange={(e) => setListForm({ ...listForm, language: e.target.value })}>
                    <option value="PT">Português (~15s para 100 cartas)</option>
                    <option value="EN">Inglês (~8s para 100 cartas)</option>
                  </select>
                </label>
                <label className="checkbox-label">
                  <input type="checkbox" checked={listForm.commander}
                    onChange={(e) => setListForm({ ...listForm, commander: e.target.checked })} />
                  Deck Commander
                </label>
                <div className="color-picker">
                  <span className="color-picker-label">Cores do deck</span>
                  <div className="color-pips-row">
                    {MTG_COLORS.map(({ code, cls }) => (
                      <label key={code} className={`color-pip-check cp-${cls}${(listForm.colors||"").split(",").filter(Boolean).includes(code) ? " selected" : ""}`}>
                        <input type="checkbox"
                          checked={(listForm.colors||"").split(",").filter(Boolean).includes(code)}
                          onChange={() => setListForm({ ...listForm, colors: toggleColor(listForm.colors, code) })} />
                        {code}
                      </label>
                    ))}
                  </div>
                </div>
                <DeckColorSelect value={listForm.theme_color} onChange={(v) => setListForm({ ...listForm, theme_color: v })} />
                <label>Descrição<textarea value={listForm.description} onChange={(e) => setListForm({ ...listForm, description: e.target.value })} /></label>
                <label>
                  Lista de cartas *
                  <textarea required rows={12} style={{ fontFamily: "monospace", fontSize: 12 }}
                    placeholder={"Commander\n1 Perrie, the Pulverizer\nCreatures (30)\n1 Aven Courier\n..."}
                    value={listForm.deck_list}
                    onChange={(e) => setListForm({ ...listForm, deck_list: e.target.value })} />
                </label>

                {listLoading && (
                  <div className="import-loading">
                    <span className="import-spinner" />
                    Importando cartas{listForm.language === "PT" ? " em PT…" : "…"} (buscando uma a uma no Scryfall)
                  </div>
                )}
                {listError && <p className="form-error">{listError}</p>}

                <div className="modal-actions">
                  <button type="button" onClick={() => setListModal(false)} disabled={listLoading}>Cancelar</button>
                  <button type="submit" className="save-btn" disabled={listLoading}>
                    {listLoading ? "Importando…" : "Importar"}
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}

      {/* ── MODAL IMPORTAR PRÉ-CON ── */}
      {importModal && (
        <div className="modal-overlay" onClick={() => { if (!importLoading) setImportModal(false); }}>
          <div className="modal modal-import" onClick={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => { if (!importLoading) setImportModal(false); }}>✕</button>
            <h2>Importar Pré-con</h2>

            {importResult ? (
              <div className="import-result">
                <div className="import-result-icon">✓</div>
                <p><strong>{importResult.imported}</strong> cartas importadas com sucesso!</p>
                <p className="import-result-sub">{importResult.failed > 0 && <span>{importResult.failed} falhas · </span>}{importResult.total_from_api} cartas no set · Deck: <em>{importResult.deck_name}</em></p>
                <div className="modal-actions">
                  <button type="button" onClick={() => { setImportModal(false); setImportResult(null); }}>Fechar</button>
                  <button type="button" className="save-btn" onClick={() => { setImportResult(null); setImportForm(EMPTY_IMPORT_FORM); }}>Importar outro</button>
                </div>
              </div>
            ) : (
              <form className="edit-form" onSubmit={handleImportPrecon}>
                <div className="edit-grid">
                  <label>
                    Código do set *
                    <input required placeholder="Ex: ncc, dmu, bro" value={importForm.set_code}
                      onChange={(e) => setImportForm({ ...importForm, set_code: e.target.value.toLowerCase().trim() })} />
                  </label>
                  <label>
                    Nome do deck *
                    <input required placeholder="Ex: New Capenna Commander" value={importForm.deck_name}
                      onChange={(e) => setImportForm({ ...importForm, deck_name: e.target.value })} />
                  </label>
                </div>
                <label>
                  Idioma
                  <select value={importForm.language} onChange={(e) => setImportForm({ ...importForm, language: e.target.value })}>
                    <option value="PT">Português (busca versão PT de cada carta)</option>
                    <option value="EN">Inglês (importação rápida)</option>
                  </select>
                </label>
                <label className="checkbox-label">
                  <input type="checkbox" checked={importForm.commander}
                    onChange={(e) => setImportForm({ ...importForm, commander: e.target.checked })} />
                  Deck Commander
                </label>
                <div className="color-picker">
                  <span className="color-picker-label">Cores do deck</span>
                  <div className="color-pips-row">
                    {MTG_COLORS.map(({ code, cls }) => (
                      <label key={code} className={`color-pip-check cp-${cls}${(importForm.colors||"").split(",").filter(Boolean).includes(code) ? " selected" : ""}`}>
                        <input type="checkbox"
                          checked={(importForm.colors||"").split(",").filter(Boolean).includes(code)}
                          onChange={() => setImportForm({ ...importForm, colors: toggleColor(importForm.colors, code) })} />
                        {code}
                      </label>
                    ))}
                  </div>
                </div>
                <DeckColorSelect value={importForm.theme_color} onChange={(v) => setImportForm({ ...importForm, theme_color: v })} />
                <label>Descrição<textarea value={importForm.description} onChange={(e) => setImportForm({ ...importForm, description: e.target.value })} /></label>

                {importLoading && (
                  <div className="import-loading">
                    <span className="import-spinner" />
                    Importando cartas{importForm.language === "PT" ? " em PT (pode levar ~30s)…" : "…"}
                  </div>
                )}
                {importError && <p className="form-error">{importError}</p>}

                <div className="modal-actions">
                  <button type="button" onClick={() => setImportModal(false)} disabled={importLoading}>Cancelar</button>
                  <button type="submit" className="save-btn" disabled={importLoading}>
                    {importLoading ? "Importando…" : "Importar"}
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}

      {activeTab === "collection" && <section className="grid">
        {/* ── FORMULÁRIO ── */}
        <form className="card form" onSubmit={handleSubmit}>
          <h2>Cadastrar Carta</h2>
          {field("Nome *", "name", { placeholder: "Ex: Lightning Bolt", required: true })}
          <ManaColorPicker value={form.color} onChange={(v) => setForm({ ...form, color: v })} />
          {field("Tipo", "type", { placeholder: "Ex: Criatura, Instant" })}
          {field("Subtítulo", "subtitle", { placeholder: "Ex: Humano Soldado" })}
          {field("Nº na coleção", "collection_number", { placeholder: "Ex: 17" })}
          <label>
            Raridade
            <select value={form.rarity} onChange={(e) => setForm({ ...form, rarity: e.target.value })}>
              <option value="">Selecione</option>
              <option value="L">Land (L)</option>
              <option value="C">Common (C)</option>
              <option value="U">Uncommon (U)</option>
              <option value="R">Rare (R)</option>
              <option value="M">Mythic (M)</option>
              <option value="T">Token (T)</option>
            </select>
          </label>
          {field("Sigla da coleção", "set_code", { placeholder: "Ex: KLD, TMT" })}
          <label>
            Idioma
            <select value={form.language} onChange={(e) => setForm({ ...form, language: e.target.value })}>
              <option value="PT">Português</option>
              <option value="EN">Inglês</option>
              <option value="ES">Espanhol</option>
              <option value="JP">Japonês</option>
              <option value="FR">Francês</option>
              <option value="DE">Alemão</option>
            </select>
          </label>
          {field("Ano", "year", { type: "number", placeholder: "Ex: 2016", min: "1993" })}
          {field("Artista", "artist", { placeholder: "Ex: Ryan Pancoast" })}
          {field("Empresa", "company")}
          <label>
            Condição
            <select value={form.condition} onChange={(e) => setForm({ ...form, condition: e.target.value })}>
              <option value="mint">Mint</option>
              <option value="near_mint">Near Mint</option>
              <option value="played">Played</option>
              <option value="damaged">Damaged</option>
            </select>
          </label>
          <label>
            Quantidade
            <input type="number" value={form.quantity} min="1"
              onChange={(e) => setForm({ ...form, quantity: e.target.value })} />
          </label>
          <label className="checkbox-label">
            <input type="checkbox" checked={form.foil}
              onChange={(e) => setForm({ ...form, foil: e.target.checked })} />
            Foil
          </label>
          <label className="checkbox-label">
            <input type="checkbox" checked={form.prerelease}
              onChange={(e) => setForm({ ...form, prerelease: e.target.checked })} />
            Pré-release
          </label>
          <label className="checkbox-label">
            <input type="checkbox" checked={form.commander}
              onChange={(e) => setForm({ ...form, commander: e.target.checked })} />
            Commander
          </label>
          <label>
            Deck
            <select value={form.deck_id} onChange={(e) => setForm({ ...form, deck_id: Number(e.target.value) })}>
              <option value={0}>— Nenhum —</option>
              {decks.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
            </select>
          </label>
          <label>
            Observações
            <textarea value={form.notes}
              onChange={(e) => setForm({ ...form, notes: e.target.value })}
              placeholder="Ex: carta em bom estado" />
          </label>
          <button type="submit">Cadastrar</button>
        </form>

        {/* ── LISTA ── */}
        <section className="card list-section">
          <div className="list-header">
            <div className="list-header-top">
              <h2>Minha coleção <span className="total-badge">{totalQuantity} cartas</span><span className="unique-badge">{total} únicas</span></h2>
              <div className="export-btns">
                <button type="button" className="export-btn" onClick={handleExportCSV}>↓ CSV</button>
                <button type="button" className="export-btn" onClick={handleExportXLSX}>↓ XLSX</button>
              </div>
            </div>
            <input
              className="search-input"
              type="search"
              placeholder="Buscar por nome, coleção, cor, tipo..."
              value={search}
              onChange={handleSearchChange}
            />
            <div className="filter-bar">
              <select className="filter-select" value={filterDeck} onChange={(e) => { setFilterDeck(e.target.value); setPage(1); }}>
                <option value="">Todos os decks</option>
                <option value="0">Sem deck</option>
              </select>
              <select className="filter-select" value={filterFoil} onChange={(e) => { setFilterFoil(e.target.value); setPage(1); }}>
                <option value="">Todas as cartas</option>
                <option value="1">✦ Somente Foil</option>
              </select>
              <select className="filter-select" value={filterRarity} onChange={(e) => { setFilterRarity(e.target.value); setPage(1); }}>
                <option value="">Todas as raridades</option>
                <option value="L">Land (L)</option>
                <option value="C">Common (C)</option>
                <option value="U">Uncommon (U)</option>
                <option value="R">Rare (R)</option>
                <option value="M">Mythic (M)</option>
                <option value="T">Token (T)</option>
              </select>
            </div>
            {filterDeck === "0" && filterFoil === "" && filterRarity === "" && (
              <button type="button" className="deck-builder-btn" onClick={handleSuggestDecks}>
                <span className="deck-builder-btn-shine" />
                ✨ Montar Deck com IA
                <span className="deck-builder-btn-sub">Analisar {total} carta{total !== 1 ? "s" : ""} sem deck</span>
              </button>
            )}
            <div className="sort-bar">
              {SORT_OPTIONS.map((opt) => (
                <button
                  key={opt.value}
                  type="button"
                  className={`sort-btn${sort === opt.value ? " active" : ""}`}
                  onClick={() => handleSort(opt.value)}
                >
                  {opt.label}{orderIcon(opt.value)}
                </button>
              ))}
            </div>
          </div>

          <div className="list">
            {cards.map((card) => {
              const assignedDeck = card.deck_id > 0 ? decks.find((d) => d.id === card.deck_id) : null;
              return (
                <div
                  className={`list-item${card.foil ? " is-foil" : ""} item-r-${(card.rarity || "x").toLowerCase()}`}
                  key={card.id}
                >
                  <div className="list-item-info">
                    <div className="list-item-name">
                      <strong className={card.foil ? "foil-text" : ""}>
                        {card.name}
                      </strong>
                      {card.foil && <span className="foil-text">✦</span>}
                      <CardColorIcons card={card} />
                      {card.rarity && (
                        <span className={`rarity r-${card.rarity.toLowerCase()}`}>
                          {card.rarity}
                        </span>
                      )}
                      {assignedDeck && (
                        <span className="deck-badge" style={getDeckBadgeStyle(assignedDeck.theme_color)}>
                          {assignedDeck.name}
                        </span>
                      )}
                    </div>
                    <span>{card.set_code || "—"} · #{card.collection_number || "—"} · {card.language || "—"}</span>
                    <small>{card.type || "—"}{card.subtitle ? ` — ${card.subtitle}` : ""} · ×{card.quantity}</small>
                  </div>
                  <div className="actions">
                    <button type="button" onClick={() => handleDetails(card.id)}>Ver</button>
                    <button type="button" className="danger" onClick={() => handleDelete(card.id)}>✕</button>
                  </div>
                </div>
              );
            })}
            {cards.length === 0 && <p className="empty">Nenhuma carta encontrada.</p>}
          </div>

          {totalPages > 1 && (
            <div className="pagination">
              <button type="button" onClick={() => handlePage(page - 1)} disabled={page <= 1}>‹</button>
              {Array.from({ length: totalPages }, (_, i) => i + 1)
                .filter((p) => p === 1 || p === totalPages || Math.abs(p - page) <= 2)
                .reduce((acc, p, idx, arr) => {
                  if (idx > 0 && p - arr[idx - 1] > 1) acc.push("…");
                  acc.push(p);
                  return acc;
                }, [])
                .map((p, i) =>
                  p === "…"
                    ? <span key={`ellipsis-${i}`} className="pagination-ellipsis">…</span>
                    : <button key={p} type="button" className={p === page ? "active" : ""} onClick={() => handlePage(p)}>{p}</button>
                )}
              <button type="button" onClick={() => handlePage(page + 1)} disabled={page >= totalPages}>›</button>
            </div>
          )}
        </section>
      </section>}

      {/* ── MODAL DETALHES ── */}
      {(selectedCard || loadingDetail) && (
        <div className="modal-overlay" onClick={() => { setSelectedCard(null); setEditMode(false); }}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => { setSelectedCard(null); setEditMode(false); }}>✕</button>

            {loadingDetail && <p className="empty">Carregando...</p>}

            {selectedCard && !editMode && (
              <>
                <div className="modal-top">
                  {selectedCard.external?.image_url && (
                    <img src={selectedCard.external.image_url} alt={selectedCard.local.name} />
                  )}
                  <div>
                    <h2>
                      {selectedCard.external?.printed_name || selectedCard.external?.name || selectedCard.local.name}
                      {selectedCard.local.foil ? " ✦" : ""}
                    </h2>
                    {selectedCard.external?.printed_name && (
                      <p className="modal-en-name">{selectedCard.external.name}</p>
                    )}
                    <p className="modal-subtitle">
                      {selectedCard.external?.printed_type || selectedCard.external?.type || selectedCard.local.type}
                      {selectedCard.local.subtitle ? ` — ${selectedCard.local.subtitle}` : ""}
                    </p>
                    {selectedCard.external?.set_name && (
                      <p className="modal-set-name">{selectedCard.external.set_name} ({selectedCard.local.set_code})</p>
                    )}
                  </div>
                </div>

                <div className="modal-divider" />
                <div className="modal-grid">
                  <div className="modal-grid-cell-color">
                    <span>Cor</span>
                    <span className="modal-color-icons">
                      <CardColorIcons card={selectedCard.local} />
                      {!parseCardColorIcons(selectedCard.local).length && (selectedCard.local.color || "—")}
                    </span>
                  </div>
                  <div><span>Raridade</span>{selectedCard.external?.rarity || selectedCard.local.rarity || "—"}</div>
                  <div><span>Coleção</span>{selectedCard.local.set_code || "—"}</div>
                  <div><span>Nº</span>{selectedCard.local.collection_number || "—"}</div>
                  <div><span>Idioma</span>{selectedCard.local.language || "—"}</div>
                  <div><span>Ano</span>{selectedCard.local.year || "—"}</div>
                  <div><span>Artista</span>{selectedCard.external?.artist || selectedCard.local.artist || "—"}</div>
                  <div><span>Empresa</span>{selectedCard.local.company || "—"}</div>
                  <div><span>Custo</span>{selectedCard.external?.mana_cost || selectedCard.local.mana_cost || "—"}</div>
                  <div><span>Quantidade</span>{selectedCard.local.quantity}</div>
                  <div><span>Condição</span>{selectedCard.local.condition || "—"}</div>
                  <div><span>Foil</span>{selectedCard.local.foil ? "Sim" : "Não"}</div>
                  {selectedCard.local.deck_id > 0 && (() => {
                    const d = decks.find(dd => dd.id === selectedCard.local.deck_id);
                    return d ? (
                      <div>
                        <span>Deck</span>
                        <span className="modal-deck-badge" style={getDeckBadgeStyle(d.theme_color)}>{d.name}</span>
                      </div>
                    ) : null;
                  })()}
                  {selectedCard.external?.power && (
                    <div><span>Força/Resistência</span>{selectedCard.external.power}/{selectedCard.external.toughness}</div>
                  )}
                  {selectedCard.external?.prices?.usd && (
                    <div><span>Preço (USD)</span>${selectedCard.external.prices.usd}{selectedCard.external.prices.usd_foil ? ` / foil $${selectedCard.external.prices.usd_foil}` : ""}</div>
                  )}
                  {selectedCard.external?.prices?.eur && (
                    <div><span>Preço (EUR)</span>€{selectedCard.external.prices.eur}{selectedCard.external.prices.eur_foil ? ` / foil €${selectedCard.external.prices.eur_foil}` : ""}</div>
                  )}
                </div>

                {(selectedCard.external?.printed_text || selectedCard.external?.text) && (
                  <p className="modal-notes">
                    <strong>Texto:</strong>{" "}
                    {selectedCard.external.printed_text || selectedCard.external.text}
                  </p>
                )}
                {selectedCard.external?.flavor_text && (
                  <p className="modal-flavor"><em>{selectedCard.external.flavor_text}</em></p>
                )}
                {selectedCard.local.notes && (
                  <p className="modal-notes"><strong>Observações:</strong> {selectedCard.local.notes}</p>
                )}
                {selectedCard.external?.scryfall_uri && (
                  <a className="scryfall-link" href={selectedCard.external.scryfall_uri} target="_blank" rel="noreferrer">
                    Ver no Scryfall ↗
                  </a>
                )}

                <div className="modal-actions">
                  <button type="button" className="edit-btn" onClick={handleEditStart}>Editar</button>
                  <button className="danger" onClick={() => handleDelete(selectedCard.local.id)}>Remover carta</button>
                </div>
              </>
            )}

            {selectedCard && editMode && (
              <form className="edit-form" onSubmit={(e) => { e.preventDefault(); handleEditSave(); }}>
                <h2>Editar carta</h2>
                <div className="edit-grid">
                  <label>Nome *<input required value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })} /></label>
                  <label>Tipo<input value={editForm.type} onChange={(e) => setEditForm({ ...editForm, type: e.target.value })} /></label>
                  <label>Subtítulo<input value={editForm.subtitle} onChange={(e) => setEditForm({ ...editForm, subtitle: e.target.value })} /></label>
                  <label>Nº na coleção<input value={editForm.collection_number} onChange={(e) => setEditForm({ ...editForm, collection_number: e.target.value })} /></label>
                  <label>Raridade
                    <select value={editForm.rarity} onChange={(e) => setEditForm({ ...editForm, rarity: e.target.value })}>
                      <option value="">Selecione</option>
                      <option value="L">Land (L)</option>
                      <option value="C">Common (C)</option>
                      <option value="U">Uncommon (U)</option>
                      <option value="R">Rare (R)</option>
                      <option value="M">Mythic (M)</option>
                      <option value="T">Token (T)</option>
                    </select>
                  </label>
                  <label>Sigla da coleção<input value={editForm.set_code} onChange={(e) => setEditForm({ ...editForm, set_code: e.target.value })} /></label>
                  <label>Idioma
                    <select value={editForm.language} onChange={(e) => setEditForm({ ...editForm, language: e.target.value })}>
                      <option value="PT">Português</option>
                      <option value="EN">Inglês</option>
                      <option value="ES">Espanhol</option>
                      <option value="JP">Japonês</option>
                      <option value="FR">Francês</option>
                      <option value="DE">Alemão</option>
                    </select>
                  </label>
                  <label>Ano<input type="number" min="1993" value={editForm.year} onChange={(e) => setEditForm({ ...editForm, year: e.target.value })} /></label>
                  <label>Artista<input value={editForm.artist} onChange={(e) => setEditForm({ ...editForm, artist: e.target.value })} /></label>
                  <label>Empresa<input value={editForm.company} onChange={(e) => setEditForm({ ...editForm, company: e.target.value })} /></label>
                  <label>Condição
                    <select value={editForm.condition} onChange={(e) => setEditForm({ ...editForm, condition: e.target.value })}>
                      <option value="mint">Mint</option>
                      <option value="near_mint">Near Mint</option>
                      <option value="played">Played</option>
                      <option value="damaged">Damaged</option>
                    </select>
                  </label>
                  <label>Quantidade<input type="number" min="1" value={editForm.quantity} onChange={(e) => setEditForm({ ...editForm, quantity: e.target.value })} /></label>
                  <label className="checkbox-label"><input type="checkbox" checked={editForm.foil} onChange={(e) => setEditForm({ ...editForm, foil: e.target.checked })} />Foil</label>
                  <label className="checkbox-label"><input type="checkbox" checked={editForm.prerelease} onChange={(e) => setEditForm({ ...editForm, prerelease: e.target.checked })} />Pré-release</label>
                  <label className="checkbox-label"><input type="checkbox" checked={editForm.commander} onChange={(e) => setEditForm({ ...editForm, commander: e.target.checked })} />Commander</label>
                  <label>Deck
                    <select value={editForm.deck_id} onChange={(e) => setEditForm({ ...editForm, deck_id: Number(e.target.value) })}>
                      <option value={0}>— Nenhum —</option>
                      {decks.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
                    </select>
                  </label>
                </div>
                <ManaColorPicker value={editForm.color} onChange={(v) => setEditForm({ ...editForm, color: v })} />
                <label>Observações<textarea value={editForm.notes} onChange={(e) => setEditForm({ ...editForm, notes: e.target.value })} /></label>
                <label className="checkbox-label propagate-label">
                  <input type="checkbox" checked={propagate} onChange={(e) => setPropagate(e.target.checked)} />
                  Aplicar campos compartilhados a todas as cópias idênticas
                </label>
                <div className="modal-actions">
                  <button type="button" onClick={() => setEditMode(false)}>Cancelar</button>
                  <button type="submit" className="save-btn">Salvar</button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}
    </main>
  );
}
