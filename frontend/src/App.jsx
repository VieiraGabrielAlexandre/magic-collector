import { useEffect, useRef, useState } from "react";
import * as XLSX from "xlsx";
import { assignCardToDeck, createCard, createDeck, deleteCard, deleteDeck, exportCards, fetchDeckIcon, getCard, listCards, listDecks, updateCard, updateDeck } from "./services/api";
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

  const [selectedCard, setSelectedCard] = useState(null);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [editForm, setEditForm] = useState({});
  const [propagate, setPropagate] = useState(true);

  const [activeTab, setActiveTab] = useState("collection");
  const [decks, setDecks] = useState([]);
  const [deckForm, setDeckForm] = useState({ name: "", description: "", commander: false, colors: "", set_code: "" });
  const [editDeckModal, setEditDeckModal] = useState(null);

  const [managingDeck, setManagingDeck] = useState(null);
  const [deckCards, setDeckCards] = useState([]);
  const [unassignedCards, setUnassignedCards] = useState([]);
  const [deckSearch, setDeckSearch] = useState("");

  const searchTimer = useRef(null);

  async function loadCards(opts = {}) {
    const result = await listCards({
      q: opts.q ?? search,
      page: opts.page ?? page,
      sort: opts.sort ?? sort,
      order: opts.order ?? order,
      pageSize: 20,
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

  async function refreshDeckManagement(deckId) {
    const [dc, uc] = await Promise.all([
      listCards({ deckId, pageSize: 500 }),
      listCards({ deckId: 0, pageSize: 500 }),
    ]);
    setDeckCards(dc.data ?? []);
    setUnassignedCards(uc.data ?? []);
  }

  useEffect(() => { loadCards(); }, [sort, order]);
  useEffect(() => { loadDecks(); }, []);

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
    setDeckForm({ name: "", description: "", commander: false, colors: "", set_code: "" });
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
    });
    setEditDeckModal(null);
    loadDecks();
  }

  async function handleManageDeck(deck) {
    setManagingDeck(deck);
    setDeckSearch("");
    await refreshDeckManagement(deck.id);
    if (deck.set_code && !deck.icon_uri) {
      const result = await fetchDeckIcon(deck.id);
      if (result?.icon_uri) {
        setManagingDeck((prev) => ({ ...prev, icon_uri: result.icon_uri }));
        loadDecks();
      }
    }
  }

  async function handleAssignCard(cardId) {
    await assignCardToDeck(cardId, managingDeck.id);
    await refreshDeckManagement(managingDeck.id);
    loadDecks();
  }

  async function handleUnassignCard(cardId) {
    await assignCardToDeck(cardId, 0);
    await refreshDeckManagement(managingDeck.id);
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

  const filteredUnassigned = deckSearch
    ? unassignedCards.filter((c) =>
        c.name.toLowerCase().includes(deckSearch.toLowerCase()) ||
        (c.set_code || "").toLowerCase().includes(deckSearch.toLowerCase())
      )
    : unassignedCards;

  return (
    <main className="app">
      <section className="hero">
        <h1>Magic Collector</h1>
        <p>Cadastre, organize e consulte sua coleção de cartas Magic: The Gathering</p>
      </section>

      <nav className="tabs">
        <button type="button" className={`tab${activeTab === "collection" ? " active" : ""}`} onClick={() => setActiveTab("collection")}>Coleção</button>
        <button type="button" className={`tab${activeTab === "decks" ? " active" : ""}`} onClick={() => setActiveTab("decks")}>Decks</button>
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
                    <span className="total-badge">{deckCards.length} cartas no deck</span>
                    <span className="unique-badge">{unassignedCards.length} sem deck</span>
                  </div>
                </div>
              </div>
            </div>

            <div className="deck-manage-grid">
              <section className="card list-section">
                <div className="list-header">
                  <h3 className="section-title">No deck <span className="total-badge">{deckCards.length}</span></h3>
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
                        <small>{card.set_code || "—"} · {card.language} · ×{card.quantity}</small>
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
                    <h3 className="section-title">Sem deck <span className="unique-badge">{unassignedCards.length}</span></h3>
                  </div>
                  <input className="search-input" type="search" placeholder="Filtrar por nome ou coleção..."
                    value={deckSearch} onChange={(e) => setDeckSearch(e.target.value)} />
                </div>
                <div className="list">
                  {filteredUnassigned.map((card) => (
                    <div className={`list-item${card.foil ? " is-foil" : ""} item-r-${(card.rarity || "x").toLowerCase()}`} key={card.id}>
                      <div className="list-item-info">
                        <div className="list-item-name">
                          <strong className={card.foil ? "foil-text" : ""}>{card.name}</strong>
                          {card.foil && <span className="foil-text">✦</span>}
                          <CardColorIcons card={card} />
                          {card.rarity && <span className={`rarity r-${card.rarity.toLowerCase()}`}>{card.rarity}</span>}
                        </div>
                        <small>{card.set_code || "—"} · {card.language} · ×{card.quantity}</small>
                      </div>
                      <div className="actions">
                        <button type="button" onClick={() => handleAssignCard(card.id)}>+ Deck</button>
                      </div>
                    </div>
                  ))}
                  {filteredUnassigned.length === 0 && <p className="empty">Nenhuma carta disponível.</p>}
                </div>
              </section>
            </div>
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
              <button type="submit">Criar Deck</button>
            </form>

            <section className="card list-section">
              <div className="list-header">
                <div className="list-header-top">
                  <h2>Meus Decks <span className="total-badge">{decks.length}</span></h2>
                </div>
              </div>
              <div className="list deck-list">
                {decks.map((deck) => (
                  <div className="deck-list-item" key={deck.id}>
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
              <div className="modal-actions">
                <button type="button" onClick={() => setEditDeckModal(null)}>Cancelar</button>
                <button type="submit" className="save-btn">Salvar</button>
              </div>
            </form>
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
            {cards.map((card) => (
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
                  </div>
                  <span>{card.set_code || "—"} · #{card.collection_number || "—"} · {card.language || "—"}</span>
                  <small>{card.type || "—"}{card.subtitle ? ` — ${card.subtitle}` : ""} · ×{card.quantity}</small>
                </div>
                <div className="actions">
                  <button type="button" onClick={() => handleDetails(card.id)}>Ver</button>
                  <button type="button" className="danger" onClick={() => handleDelete(card.id)}>✕</button>
                </div>
              </div>
            ))}
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
                  <div><span>Cor</span>
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
                  {selectedCard.local.deck_id > 0 && (
                    <div><span>Deck</span>{decks.find(d => d.id === selectedCard.local.deck_id)?.name ?? "—"}</div>
                  )}
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
