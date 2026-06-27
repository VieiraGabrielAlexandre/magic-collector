import { useEffect, useRef, useState } from "react";
import * as XLSX from "xlsx";
import { acquireWishlistItem, addGameSessionPlayer, assignCardToDeck, createBattle, createCard, createDeck, createGameSession, createToken, createWishlistItem, deleteCard, deleteBattle, deleteDeck, deleteGameSession, deleteGameSessionPlayer, deleteToken, deleteWishlistItem, evaluateDeck, exportCards, fetchDeckIcon, finishGameSession, getCard, getCollectionStats, getGameSession, getMe, importDeckList, importPrecon, listBattles, listCards, listColorCombos, listDecks, listGameSessions, listTokens, listWishlist, logout, previewCard, previewToken, refreshImages, refreshPrices, resetGameSession, restoreGameSession, suggestDecks, updateCard, updateCardQuantity, updateDeck, updateGameSessionPlayer, updateTokenQuantity } from "./services/api";
import LandingPage from "./LandingPage.jsx";
import "./App.css";


const SORT_OPTIONS = [
  { value: "name", label: "Nome" },
  { value: "set_code", label: "Coleção" },
  { value: "color", label: "Cor" },
  { value: "rarity", label: "Raridade" },
  { value: "quantity", label: "Qtd" },
  { value: "price_usd", label: "💰 Preço" },
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

// Converte um card (com colors JSON ou color PT) para string de códigos WUBRG separados por vírgula
function colorToWUBRGCodes(card) {
  if (card.colors && card.colors !== "[]" && card.colors !== "") {
    try {
      const arr = JSON.parse(card.colors);
      if (Array.isArray(arr) && arr.length > 0) return arr.join(",");
    } catch {}
  }
  const PT_TO_CODE = { branco: "W", azul: "U", preto: "B", preta: "B", vermelho: "R", vermelha: "R", verde: "G", incolor: "C" };
  return (card.color || "").split(/[,/]+/).map(s => PT_TO_CODE[s.trim().toLowerCase()]).filter(Boolean).join(",");
}

function ManaColorPicker({ value, onChange }) {
  // value é uma string de códigos WUBRG separados por vírgula: "W,U,B"
  const active = new Set((value || "").split(",").filter(Boolean));
  return (
    <div className="mana-color-picker">
      <span className="mana-color-picker-label">Cor</span>
      <div className="mana-color-picker-row">
        {CARD_COLORS.map(({ icon, pt, code }) => (
          <button key={code} type="button" title={pt} aria-label={pt}
            className={`mana-picker-btn${active.has(code) ? " active" : ""}`}
            onClick={() => {
              const s = new Set(active);
              if (s.has(code)) s.delete(code); else s.add(code);
              onChange([...s].join(","));
            }}
          >
            <img src={`/mana-icons/${icon}.svg`} alt={pt} />
          </button>
        ))}
      </div>
    </div>
  );
}

// ── Rarity display helpers ───────────────────────────────────────────────
const RARITY_LABEL = { M: "Mythic", R: "Rare", U: "Uncommon", C: "Common", L: "Land", T: "Token", "?": "Sem raridade" };
const RARITY_COLOR = { M: "#e07030", R: "#c89b3c", U: "#5ab4cc", C: "#9a9a9a", L: "#78a060", T: "#a878cc", "?": "#5a4830" };
const COLOR_NAME   = { W: "Branco", U: "Azul", B: "Preto", R: "Vermelho", G: "Verde", C: "Incolor" };
const COLOR_HEX    = { W: "#f5e6c8", U: "#5ab4cc", B: "#9080b0", R: "#e05050", G: "#70b860", C: "#a0a0a0" };

function StatsPanel({ stats, loading }) {
  if (loading) return (
    <div className="stats-panel stats-loading">
      <span className="eval-spinner">⚙</span>
      <span>Calculando estatísticas…</span>
    </div>
  );
  if (!stats) return null;

  const winRate = stats.battle_total > 0
    ? Math.round((stats.battle_wins / stats.battle_total) * 100) : null;

  const maxRarityQty = Math.max(...(stats.by_rarity ?? []).map(r => r.quantity), 1);

  return (
    <div className="stats-panel">
      {/* ── Row 1: números grandes ── */}
      <div className="stats-top">
        <div className="stat-box">
          <span className="stat-num">{(stats.total_quantity ?? 0).toLocaleString("pt-BR")}</span>
          <span className="stat-lbl">cartas totais</span>
        </div>
        <div className="stat-box">
          <span className="stat-num">{(stats.unique_cards ?? 0).toLocaleString("pt-BR")}</span>
          <span className="stat-lbl">únicas</span>
        </div>
        <div className="stat-box stat-foil">
          <span className="stat-num">✦ {stats.foil_quantity ?? 0}</span>
          <span className="stat-lbl">foil ({stats.foil_count ?? 0} cartas)</span>
        </div>
        <div className="stat-box stat-value">
          <span className="stat-num">
            {stats.estimated_value_usd > 0
              ? `$${stats.estimated_value_usd.toFixed(2)}`
              : "—"}
          </span>
          <span className="stat-lbl">
            valor est.{stats.priced_cards > 0 && stats.priced_cards < stats.unique_cards
              ? ` (${stats.priced_cards} cartas)` : ""}
          </span>
        </div>
      </div>

      {/* ── Row 2: raridades ── */}
      {stats.by_rarity?.length > 0 && (
        <div className="stats-section">
          <h4 className="stats-section-title">Raridade</h4>
          <div className="stats-rarity-bars">
            {stats.by_rarity.map(r => (
              <div key={r.rarity} className="rarity-row">
                <span className="rarity-row-label" style={{ color: RARITY_COLOR[r.rarity] ?? "#888" }}>
                  {RARITY_LABEL[r.rarity] ?? r.rarity}
                </span>
                <div className="rarity-bar-track">
                  <div className="rarity-bar-fill"
                    style={{ width: `${(r.quantity / maxRarityQty) * 100}%`, background: RARITY_COLOR[r.rarity] ?? "#888" }} />
                </div>
                <span className="rarity-row-count">{r.quantity.toLocaleString("pt-BR")}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ── Row 3: cores ── */}
      {stats.by_color?.length > 0 && (
        <div className="stats-section">
          <h4 className="stats-section-title">Cores</h4>
          <div className="stats-colors">
            {stats.by_color.map(c => (
              <div key={c.color} className="color-chip">
                <img src={`/mana/${c.color.toLowerCase() === "c" ? "incolour" : c.color.toLowerCase()}.svg`}
                  alt={c.color} className="color-chip-icon" onError={e => { e.target.style.display = "none"; }} />
                <span className="color-chip-name" style={{ color: COLOR_HEX[c.color] ?? "#888" }}>
                  {COLOR_NAME[c.color] ?? c.color}
                </span>
                <span className="color-chip-count">{c.count.toLocaleString("pt-BR")}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ── Row 4: top sets ── */}
      {stats.top_sets?.length > 0 && (
        <div className="stats-section">
          <h4 className="stats-section-title">Top Sets</h4>
          <div className="stats-sets">
            {stats.top_sets.map(s => (
              <span key={s.set_code} className="set-chip" title={`${s.count} únicas · ${s.quantity} cópias`}>
                {s.set_code} <em>{s.count}</em>
              </span>
            ))}
          </div>
        </div>
      )}
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
  { id: "",          label: "Nenhuma"                                                                        },
  { id: "pearl",     label: "Branco",    bg: "#181614", border: "#b0a898", text: "#eeeade", swatch: "#ccc4b4" },
  { id: "obsidian",  label: "Preto",     bg: "#07060e", border: "#28203c", text: "#7a6092", swatch: "#0e0b1c" },
  { id: "crimson",   label: "Vermelho",  bg: "#3a0c0c", border: "#7a2020", text: "#ff9090"                   },
  { id: "sapphire",  label: "Azul",      bg: "#081428", border: "#1a4080", text: "#80c0ff"                   },
  { id: "emerald",   label: "Verde",     bg: "#071a0c", border: "#1a5828", text: "#80d890"                   },
  { id: "violet",    label: "Roxo",      bg: "#150a28", border: "#4a2890", text: "#c090f0"                   },
  { id: "gold",      label: "Dourado",   bg: "#221404", border: "#805010", text: "#e8c060"                   },
  { id: "teal",      label: "Turquesa",  bg: "#041820", border: "#0c6878", text: "#60d0c8"                   },
  { id: "ember",     label: "Laranja",   bg: "#280e04", border: "#904010", text: "#f09050"                   },
  { id: "silver",    label: "Prata",     bg: "#0e1218", border: "#384858", text: "#a0b8c8"                   },
  { id: "rose",      label: "Rosa",      bg: "#280a18", border: "#8a1e50", text: "#f080b0"                   },
  { id: "bone",      label: "Marfim",    bg: "#1c1608", border: "#6a5828", text: "#e0d8b0"                   },
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

function DropdownMenu({ label, items, open, onToggle }) {
  const ref = useRef(null);
  useEffect(() => {
    if (!open) return;
    function onOutside(e) { if (ref.current && !ref.current.contains(e.target)) onToggle(); }
    document.addEventListener("mousedown", onOutside);
    return () => document.removeEventListener("mousedown", onOutside);
  }, [open, onToggle]);
  return (
    <div className="toolbar-dropdown" ref={ref}>
      <button type="button" className={`toolbar-menu-btn${open ? " open" : ""}`} onClick={onToggle}>{label}</button>
      {open && (
        <div className="toolbar-menu-items">
          {items.map((item, i) =>
            item.separator
              ? <div key={i} className="toolbar-menu-sep" />
              : <button key={i} type="button"
                  className={`toolbar-menu-item${item.active ? " active" : ""}`}
                  onClick={() => { item.onClick(); onToggle(); }}
                  disabled={item.disabled}>
                  {item.label}
                </button>
          )}
        </div>
      )}
    </div>
  );
}

function DeckColorSelect({ value, onChange }) {
  const theme = getDeckTheme(value);
  return (
    <label>
      Cor do deck
      <div className="deck-color-select-row">
        {theme?.bg && <span className="deck-color-swatch" style={{ background: theme.swatch ?? theme.bg, borderColor: theme.border }} />}
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
  // ── Auth ──────────────────────────────────────────────────────────────
  const [authUser, setAuthUser] = useState(null);
  const [authReady, setAuthReady] = useState(false);
  const [sessionCreatedAt, setSessionCreatedAt] = useState(null);
  const [sessionElapsed, setSessionElapsed] = useState("");
  const [showLanding, setShowLanding] = useState(false);

  useEffect(() => {
    getMe().then((data) => {
      if (data?.user) {
        setAuthUser(data.user);
        setSessionCreatedAt(data.session_created_at);
      }
    }).finally(() => setAuthReady(true));
  }, []);

  useEffect(() => {
    if (!sessionCreatedAt) return;
    function tick() {
      const diff = Math.floor((Date.now() - new Date(sessionCreatedAt).getTime()) / 1000);
      if (diff < 60) setSessionElapsed(`${diff}s`);
      else if (diff < 3600) setSessionElapsed(`${Math.floor(diff / 60)}min`);
      else {
        const h = Math.floor(diff / 3600);
        const m = Math.floor((diff % 3600) / 60);
        setSessionElapsed(m > 0 ? `${h}h ${m}min` : `${h}h`);
      }
    }
    tick();
    const id = setInterval(tick, 30000);
    return () => clearInterval(id);
  }, [sessionCreatedAt]);

  function handleAuthEnter(user, createdAt) {
    setAuthUser(user);
    setSessionCreatedAt(createdAt);
    setAuthReady(true);
  }

  async function handleLogout() {
    await logout();
    setAuthUser(null);
    setSessionCreatedAt(null);
    setSessionElapsed("");
  }

  const [cards, setCards] = useState([]);
  const [total, setTotal] = useState(0);
  const [totalQuantity, setTotalQuantity] = useState(0);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  const [search, setSearch] = useState("");
  const [sort, setSort] = useState("created_at");
  const [order, setOrder] = useState("desc");
  const [filterType, setFilterType] = useState("");
  const [filterFoil, setFilterFoil] = useState("");
  const [filterFullArt, setFilterFullArt] = useState(false);
  const [filterRarity, setFilterRarity] = useState("");
  const [filterDeck, setFilterDeck] = useState("");
  const [filterColors, setFilterColors] = useState("");
  const [availableColors, setAvailableColors] = useState([]);
  const [statsOpen, setStatsOpen] = useState(false);
  const [stats, setStats] = useState(null);
  const [statsLoading, setStatsLoading] = useState(false);
  const [priceRefreshing, setPriceRefreshing] = useState(false);
  const [priceRefreshResult, setPriceRefreshResult] = useState(null);
  const [viewMode, setViewMode] = useState(() => localStorage.getItem("card-view-mode") || "list");
  const [confirmCard, setConfirmCard] = useState(null); // { found, card, formData }
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [deckBuilderModal, setDeckBuilderModal] = useState(false);
  const [deckBuilderLoading, setDeckBuilderLoading] = useState(false);
  const [deckBuilderResult, setDeckBuilderResult] = useState(null);
  const EMPTY_DECK_CONFIG = { format: "auto", goal: "fun", colors: "" };
  const [deckBuilderConfig, setDeckBuilderConfig] = useState(EMPTY_DECK_CONFIG);
  const [deckBuilderStep, setDeckBuilderStep] = useState("config"); // "config" | "loading" | "result"
  const [deckBuilderApproving, setDeckBuilderApproving] = useState(false);

  const [quickAddModal, setQuickAddModal] = useState(false);
  const [quickAddForm, setQuickAddForm] = useState({ set_code: "", collection_number: "", language: "EN", foil: false, quantity: 1 });
  const [openMenu, setOpenMenu] = useState(null);

  const [selectedCard, setSelectedCard] = useState(null);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [detailFromDeck, setDetailFromDeck] = useState(false);
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
  const [deckTypeFilter, setDeckTypeFilter] = useState("");
  const [deckColorFilter, setDeckColorFilter] = useState("");
  const [deckSetFilter, setDeckSetFilter] = useState("");
  const [deckPage, setDeckPage] = useState(1);
  const [deckAnomalyModal, setDeckAnomalyModal] = useState(false);
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

  // ── Wishlist ─────────────────────────────────────────────────────────────
  const EMPTY_WISHLIST_FORM = { set_code: "", collection_number: "", foil: false, reason: "" };
  const EMPTY_ACQUIRE_FORM = { deck_id: 0, condition: "near_mint", commander: false, prerelease: false };
  const [wishlistItems, setWishlistItems] = useState([]);
  const [wishlistForm, setWishlistForm] = useState(EMPTY_WISHLIST_FORM);
  const [wishlistSubmitting, setWishlistSubmitting] = useState(false);
  const [wishlistError, setWishlistError] = useState("");
  const [wishlistDetail, setWishlistDetail] = useState(null);
  const [wishlistAcquireFor, setWishlistAcquireFor] = useState(null);
  const [wishlistAcquireForm, setWishlistAcquireForm] = useState(EMPTY_ACQUIRE_FORM);
  const [wishlistAcquiring, setWishlistAcquiring] = useState(false);

  // ── Pontuação / Life Counter ─────────────────────────────────────────────────
  const EMPTY_SESSION_FORM = { name: "", format: "Commander", starting_life: 40 };
  const EMPTY_PLAYER_ROW = { name: "", short_code: "" };
  const [gameSessions, setGameSessions] = useState([]);
  const [activeSession, setActiveSession] = useState(null);
  const [sessionView, setSessionView] = useState("list");
  const [sessionForm, setSessionForm] = useState(EMPTY_SESSION_FORM);
  const [sessionPlayers, setSessionPlayers] = useState([{ ...EMPTY_PLAYER_ROW }, { ...EMPTY_PLAYER_ROW }]);
  const [sessionLoading, setSessionLoading] = useState(false);
  const [sessionError, setSessionError] = useState("");
  const [addPlayerForm, setAddPlayerForm] = useState(EMPTY_PLAYER_ROW);
  const [showAddPlayer, setShowAddPlayer] = useState(false);
  const playerTimers = useRef({});
  const playerPending = useRef({});

  const EMPTY_LIST_FORM = { deck_name: "", set_code: "", language: "PT", colors: "", commander: false, theme_color: "", description: "", deck_list: "" };
  const [listModal, setListModal] = useState(false);
  const [listForm, setListForm] = useState(EMPTY_LIST_FORM);
  const [listLoading, setListLoading] = useState(false);
  const [listResult, setListResult] = useState(null);
  const [listError, setListError] = useState("");

  // ── Tokens ─────────────────────────────────────────────────────────────────
  const EMPTY_TOKEN_FORM = { set_code: "", collection_number: "", quantity: 1, foil: false, double_faced: false, back_set_code: "", back_collection_number: "" };
  const [tokenList, setTokenList] = useState([]);
  const [tokenAddModal, setTokenAddModal] = useState(false);
  const [tokenForm, setTokenForm] = useState(EMPTY_TOKEN_FORM);
  const [tokenPreviewFront, setTokenPreviewFront] = useState(null); // { found, token }
  const [tokenPreviewBack, setTokenPreviewBack] = useState(null);   // { found, token }
  const [tokenSearchingFront, setTokenSearchingFront] = useState(false);
  const [tokenSearchingBack, setTokenSearchingBack] = useState(false);
  const [tokenSearchError, setTokenSearchError] = useState("");
  const [tokenDetail, setTokenDetail] = useState(null);

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
      full_art: opts.filterFullArt ?? filterFullArt,
      rarity: opts.filterRarity ?? filterRarity,
      colors: opts.filterColors ?? filterColors,
      typeFilter: opts.filterType ?? filterType,
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

  async function loadTokensList() {
    const data = await listTokens();
    setTokenList(data ?? []);
  }

  async function handleTokenSearchFront(e) {
    e.preventDefault();
    setTokenSearchingFront(true);
    setTokenSearchError("");
    setTokenPreviewFront(null);
    try {
      const result = await previewToken({ set_code: tokenForm.set_code, collection_number: tokenForm.collection_number });
      setTokenPreviewFront(result);
      if (!result.found) setTokenSearchError("Frente não encontrada na Scryfall. Verifique o código e número.");
    } catch (err) {
      setTokenSearchError(err.message || "Erro ao buscar frente");
    } finally {
      setTokenSearchingFront(false);
    }
  }

  async function handleTokenSearchBack(e) {
    e.preventDefault();
    setTokenSearchingBack(true);
    setTokenSearchError("");
    setTokenPreviewBack(null);
    try {
      const result = await previewToken({ set_code: tokenForm.back_set_code, collection_number: tokenForm.back_collection_number });
      setTokenPreviewBack(result);
      if (!result.found) setTokenSearchError("Verso não encontrado na Scryfall. Verifique o código e número.");
    } catch (err) {
      setTokenSearchError(err.message || "Erro ao buscar verso");
    } finally {
      setTokenSearchingBack(false);
    }
  }

  async function handleTokenConfirm() {
    const isSaving = tokenSearchingFront || tokenSearchingBack;
    if (isSaving) return;
    setTokenSearchingFront(true);
    try {
      const payload = {
        set_code: tokenForm.set_code,
        collection_number: tokenForm.collection_number,
        quantity: Number(tokenForm.quantity),
        foil: tokenForm.foil,
      };
      if (tokenForm.double_faced && tokenPreviewFront?.found && tokenPreviewBack?.found) {
        payload.back_set_code = tokenForm.back_set_code;
        payload.back_collection_number = tokenForm.back_collection_number;
      }
      await createToken(payload);
      setTokenPreviewFront(null);
      setTokenPreviewBack(null);
      setTokenForm(EMPTY_TOKEN_FORM);
      setTokenAddModal(false);
      await loadTokensList();
    } catch (err) {
      setTokenSearchError(err.message || "Erro ao cadastrar token");
    } finally {
      setTokenSearchingFront(false);
    }
  }

  async function handleTokenDelete(id) {
    await deleteToken(id);
    await loadTokensList();
  }

  async function handleTokenQuantityChange(tok, delta) {
    const newQty = Math.max(1, (tok.quantity || 1) + delta);
    if (newQty === tok.quantity) return;
    setTokenList(prev => prev.map(t => t.id === tok.id ? { ...t, quantity: newQty } : t));
    try {
      await updateTokenQuantity(tok.id, newQty);
    } catch {
      setTokenList(prev => prev.map(t => t.id === tok.id ? { ...t, quantity: tok.quantity } : t));
    }
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

  // ── Wishlist handlers ─────────────────────────────────────────────────────
  async function loadWishlist() {
    const data = await listWishlist();
    setWishlistItems(data ?? []);
  }

  async function handleWishlistCreate(e) {
    e.preventDefault();
    setWishlistSubmitting(true);
    setWishlistError("");
    try {
      await createWishlistItem(wishlistForm);
      setWishlistForm(EMPTY_WISHLIST_FORM);
      await loadWishlist();
    } catch (err) {
      setWishlistError(err.message || "Erro ao adicionar à wishlist");
    } finally {
      setWishlistSubmitting(false);
    }
  }

  async function handleWishlistDelete(id) {
    await deleteWishlistItem(id);
    setWishlistDetail(null);
    await loadWishlist();
  }

  async function handleWishlistAcquire(e) {
    e.preventDefault();
    if (!wishlistAcquireFor) return;
    setWishlistAcquiring(true);
    try {
      await acquireWishlistItem(wishlistAcquireFor.id, wishlistAcquireForm);
      setWishlistAcquireFor(null);
      setWishlistDetail(null);
      await loadWishlist();
      await loadCards();
    } catch (err) {
      setWishlistError(err.message || "Erro ao adquirir carta");
    } finally {
      setWishlistAcquiring(false);
    }
  }

  // ── Game Session handlers ─────────────────────────────────────────────────
  async function loadGameSessions() {
    const data = await listGameSessions();
    setGameSessions(data ?? []);
  }

  async function handleCreateSession(e) {
    e.preventDefault();
    setSessionLoading(true);
    setSessionError("");
    try {
      const validPlayers = sessionPlayers.filter(p => p.name.trim() && p.short_code.trim());
      const session = await createGameSession({
        ...sessionForm,
        starting_life: Number(sessionForm.starting_life),
        players: validPlayers,
      });
      setGameSessions(prev => [session, ...prev]);
      setActiveSession(session);
      setSessionView("play");
      setSessionForm(EMPTY_SESSION_FORM);
      setSessionPlayers([{ ...EMPTY_PLAYER_ROW }, { ...EMPTY_PLAYER_ROW }]);
    } catch (err) {
      setSessionError(err.message);
    } finally {
      setSessionLoading(false);
    }
  }

  async function handleOpenSession(id) {
    const session = await getGameSession(id);
    setActiveSession(session);
    setSessionView("play");
  }

  function handleUpdatePlayer(playerId, field, delta) {
    if (!activeSession || activeSession.status === "finished") return;
    const currentPlayer = activeSession.players.find(p => p.id === playerId);
    if (!currentPlayer) return;

    // Acumula no ref para não depender do closure stale do React state
    const pending = playerPending.current[playerId] ?? {
      life: currentPlayer.life,
      poison: currentPlayer.poison,
      commander_damage_received: currentPlayer.commander_damage_received,
    };

    const oldCmdDmg = pending.commander_damage_received;
    const newCmdDmg = field === "commander_damage_received" ? Math.max(0, oldCmdDmg + delta) : oldCmdDmg;
    const cmdDmgDelta = newCmdDmg - oldCmdDmg;
    const newLife = field === "life" ? pending.life + delta : pending.life - cmdDmgDelta;
    const newPoison = field === "poison" ? Math.max(0, pending.poison + delta) : pending.poison;

    playerPending.current[playerId] = { life: newLife, poison: newPoison, commander_damage_received: newCmdDmg };

    // Atualiza UI imediatamente
    setActiveSession(prev => prev ? ({
      ...prev,
      players: prev.players.map(p => p.id === playerId
        ? { ...p, life: newLife, poison: newPoison, commander_damage_received: newCmdDmg }
        : p),
    }) : prev);

    // Envia para o backend em background após pausa nas alterações
    clearTimeout(playerTimers.current[playerId]);
    const sessionId = activeSession.id;
    playerTimers.current[playerId] = setTimeout(async () => {
      const values = playerPending.current[playerId];
      delete playerPending.current[playerId];
      try {
        const updated = await updateGameSessionPlayer(sessionId, playerId, values);
        // Só sincroniza status de eliminação — nunca reverte os valores exibidos
        setActiveSession(prev => prev ? ({
          ...prev,
          players: prev.players.map(p => p.id === playerId
            ? { ...p, is_eliminated: updated.is_eliminated, eliminated_reason: updated.eliminated_reason }
            : p),
        }) : prev);
      } catch {
        // Falha silenciosa: os valores exibidos estão corretos, sincroniza na próxima ação
      }
    }, 800);
  }

  async function handleAddSessionPlayer(e) {
    e.preventDefault();
    if (!activeSession) return;
    setSessionError("");
    try {
      const player = await addGameSessionPlayer(activeSession.id, addPlayerForm);
      setActiveSession(prev => prev ? ({ ...prev, players: [...prev.players, player] }) : prev);
      setAddPlayerForm({ ...EMPTY_PLAYER_ROW });
      setShowAddPlayer(false);
      await loadGameSessions();
    } catch (err) {
      setSessionError(err.message);
    }
  }

  async function handleRemoveSessionPlayer(playerId) {
    if (!activeSession) return;
    try {
      await deleteGameSessionPlayer(activeSession.id, playerId);
      setActiveSession(prev => prev ? ({
        ...prev,
        players: prev.players.filter(p => p.id !== playerId),
      }) : prev);
      await loadGameSessions();
    } catch (err) {
      setSessionError(err.message);
    }
  }

  async function handleResetSession() {
    if (!activeSession) return;
    const session = await resetGameSession(activeSession.id);
    setActiveSession(session);
    await loadGameSessions();
  }

  async function handleFinishSession() {
    if (!activeSession) return;
    const session = await finishGameSession(activeSession.id);
    setActiveSession(session);
    await loadGameSessions();
  }

  async function handleRestoreSession(sessionId) {
    const session = await restoreGameSession(sessionId);
    setActiveSession(session);
    setSessionView("play");
    await loadGameSessions();
  }

  async function handleDeleteSession(sessionId) {
    if (!confirm("Excluir esta sessão?")) return;
    await deleteGameSession(sessionId);
    if (activeSession?.id === sessionId) {
      setActiveSession(null);
      setSessionView("list");
    }
    await loadGameSessions();
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

  useEffect(() => { loadCards(); }, [sort, order, filterFoil, filterFullArt, filterRarity, filterDeck, filterColors, filterType]);
  useEffect(() => { loadDecks(); }, []);
  useEffect(() => { loadBattles(); }, []);
  useEffect(() => { loadWishlist(); }, []);
  useEffect(() => { loadGameSessions(); }, []);
  useEffect(() => { loadTokensList(); }, []);
  useEffect(() => {
    listColorCombos().then(data => setAvailableColors(data ?? []));
  }, []);

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

  async function handleConfirmCreate() {
    if (!confirmCard) return;
    await createCard(confirmCard.formData);
    setConfirmCard(null);
    setQuickAddForm({ set_code: "", collection_number: "", language: "EN", foil: false, quantity: 1 });
    setPage(1);
    loadCards({ page: 1 });
  }

  async function handleQuickAdd(e) {
    e.preventDefault();
    const payload = { ...quickAddForm, quantity: Number(quickAddForm.quantity) };
    setQuickAddModal(false);
    setConfirmLoading(true);
    try {
      const preview = await previewCard(payload);
      const formData = preview.found && preview.card
        ? { ...payload, name: preview.card.printed_name || preview.card.name || "", type: preview.card.type || "" }
        : payload;
      setConfirmCard({ found: preview.found, card: preview.card ?? null, formData, origin: "quickAdd" });
    } catch {
      setConfirmCard({ found: false, card: null, formData: payload, origin: "quickAdd" });
    } finally {
      setConfirmLoading(false);
    }
  }

  async function handleDetails(id) {
    setDetailFromDeck(false);
    setLoadingDetail(true);
    setSelectedCard(null);
    try {
      const data = await getCard(id);
      setSelectedCard(data);
    } finally {
      setLoadingDetail(false);
    }
  }

  async function handleDeckDetails(id) {
    setDetailFromDeck(true);
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

  async function handleQuantityChange(cardId, delta) {
    const card = cards.find((c) => c.id === cardId);
    if (!card) return;
    const newQty = Math.max(1, (card.quantity || 1) + delta);
    if (newQty === card.quantity) return;
    // Atualização otimista
    setCards((prev) => prev.map((c) => c.id === cardId ? { ...c, quantity: newQty } : c));
    setTotalQuantity((prev) => prev + (newQty - card.quantity));
    try {
      await updateCardQuantity(cardId, newQty);
    } catch {
      // Reverte em caso de erro
      setCards((prev) => prev.map((c) => c.id === cardId ? { ...c, quantity: card.quantity } : c));
      setTotalQuantity((prev) => prev - (newQty - card.quantity));
    }
  }

  function handleEditStart() {
    const c = selectedCard.local;
    setEditForm({
      name: c.name, color: colorToWUBRGCodes(c), type: c.type, subtitle: c.subtitle,
      collection_number: c.collection_number, rarity: c.rarity, set_code: c.set_code,
      language: c.language, year: c.year, artist: c.artist,
      foil: c.foil, full_art: c.full_art || false, prerelease: c.prerelease, commander: c.commander, deck_id: c.deck_id ?? 0, quantity: c.quantity, condition: c.condition, notes: c.notes,
    });
    setEditMode(true);
  }

  async function handleEditSave() {
    const colorArr = editForm.color.split(",").filter(Boolean);
    await updateCard(selectedCard.local.id, { ...editForm, year: Number(editForm.year) || 0, quantity: Number(editForm.quantity) || 1, propagate, colors: JSON.stringify(colorArr) });
    setEditMode(false);
    loadCards();
    handleDetails(selectedCard.local.id);
  }

  const EXPORT_HEADERS = ["name","color","type","subtitle","collection_number","rarity","set_code","language","year","artist","foil","quantity","condition","notes","commander","deck"];

  function cardToRow(c) {
    const deckName = c.deck_id > 0 ? (decks.find(d => d.id === c.deck_id)?.name ?? "") : "";
    return [c.name, c.color, c.type, c.subtitle, c.collection_number, c.rarity,
            c.set_code, c.language, c.year, c.artist,
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
    setDeckTypeFilter("");
    setDeckColorFilter("");
    setDeckSetFilter("");
    setDeckPage(1);
    await Promise.all([loadDeckCards(deck.id), loadUnassigned(1, "")]);
    if (deck.set_code && !deck.icon_uri) {
      const result = await fetchDeckIcon(deck.id);
      if (result?.icon_uri) {
        setManagingDeck((prev) => ({ ...prev, icon_uri: result.icon_uri }));
        loadDecks();
      }
    }
  }

  async function handleRefreshImages() {
    setPriceRefreshing(true);
    setPriceRefreshResult(null);
    try {
      const result = await refreshImages();
      setPriceRefreshResult({ ...result, _type: "images" });
      loadCards({ page: 1 });
    } catch (e) {
      setPriceRefreshResult({ error: e.message });
    } finally {
      setPriceRefreshing(false);
    }
  }

  async function handleRefreshPrices() {
    setPriceRefreshing(true);
    setPriceRefreshResult(null);
    try {
      const result = await refreshPrices();
      setPriceRefreshResult(result);
      setStats(null);
      loadCards({ page: 1 });
    } catch (e) {
      setPriceRefreshResult({ error: e.message });
    } finally {
      setPriceRefreshing(false);
    }
  }

  async function handleRefreshMissingPrices() {
    setPriceRefreshing(true);
    setPriceRefreshResult(null);
    try {
      const result = await refreshPrices({ emptyOnly: true });
      setPriceRefreshResult({ ...result, _type: "missing" });
      setStats(null);
      loadCards({ page: 1 });
    } catch (e) {
      setPriceRefreshResult({ error: e.message });
    } finally {
      setPriceRefreshing(false);
    }
  }

  async function handleOpenStats() {
    setStatsOpen((prev) => {
      if (!prev && !stats) {
        setStatsLoading(true);
        getCollectionStats().then((s) => { setStats(s); setStatsLoading(false); });
      }
      return !prev;
    });
  }

  function handleSuggestDecks() {
    setDeckBuilderModal(true);
    setDeckBuilderStep("config");
    setDeckBuilderResult(null);
    setDeckBuilderConfig(EMPTY_DECK_CONFIG);
  }

  async function handleRunDeckBuilder() {
    setDeckBuilderStep("loading");
    setDeckBuilderLoading(true);
    setDeckBuilderResult(null);
    try {
      const result = await suggestDecks(deckBuilderConfig);
      setDeckBuilderResult(result);
      setDeckBuilderStep("result");
    } catch (e) {
      setDeckBuilderResult({ error: e.message });
      setDeckBuilderStep("result");
    } finally {
      setDeckBuilderLoading(false);
    }
  }

  async function handleRevaluateDeck() {
    setDeckBuilderStep("loading");
    setDeckBuilderLoading(true);
    setDeckBuilderResult(null);
    try {
      const result = await suggestDecks({ ...deckBuilderConfig, revaluate: true });
      setDeckBuilderResult(result);
      setDeckBuilderStep("result");
    } catch (e) {
      setDeckBuilderResult({ error: e.message });
      setDeckBuilderStep("result");
    } finally {
      setDeckBuilderLoading(false);
    }
  }

  async function handleApproveDeck() {
    if (!deckBuilderResult?.deck_list) return;
    setDeckBuilderApproving(true);
    try {
      await importDeckList({
        deck_name: deckBuilderResult.deck_name || "Deck IA",
        deck_list: deckBuilderResult.deck_list,
        colors: deckBuilderResult.deck_colors || "",
        commander: deckBuilderResult.deck_commander || false,
        description: deckBuilderResult.deck_description || "",
        language: "EN",
        set_code: "",
        theme_color: "",
      });
      setDeckBuilderModal(false);
      setActiveTab("decks");
      await loadDecks();
    } catch (e) {
      alert("Erro ao criar deck: " + e.message);
    } finally {
      setDeckBuilderApproving(false);
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

  const orderIcon = (f) => {
    if (sort !== f) return null;
    return order === "asc" ? " ↑" : " ↓";
  };

  const deckTotalQuantity = deckCards.reduce((sum, c) => sum + (c.quantity || 1), 0);

  function deckCardTypeGroup(type) {
    const t = type || "";
    if (t.includes("Land")) return "Terreno";
    if (t.includes("Legendary") && t.includes("Creature")) return "Criatura Lendária";
    if (t.includes("Creature")) return "Criatura";
    if (t.includes("Planeswalker")) return "Planeswalker";
    if (t.includes("Legendary") && t.includes("Enchantment")) return "Encantamento Lendário";
    if (t.includes("Enchantment")) return "Encantamento";
    if (t.includes("Legendary") && t.includes("Artifact")) return "Artefato Lendário";
    if (t.includes("Artifact")) return "Artefato";
    if (t.includes("Instant")) return "Instantâneo";
    if (t.includes("Sorcery")) return "Feitiço";
    if (t.includes("Battle")) return "Batalha";
    return "Outro";
  }

  const _deckTypeCounts = {};
  const _deckColorCounts = {};
  const _deckSetCounts = {};
  for (const c of deckCards) {
    const qty = c.quantity || 1;
    const tg = deckCardTypeGroup(c.type);
    _deckTypeCounts[tg] = (_deckTypeCounts[tg] || 0) + qty;
    try {
      const cols = JSON.parse(c.colors || "[]");
      if (cols.length === 0) { _deckColorCounts["C"] = (_deckColorCounts["C"] || 0) + qty; }
      else { for (const col of cols) _deckColorCounts[col] = (_deckColorCounts[col] || 0) + qty; }
    } catch {}
    const sc = (c.set_code || "").toUpperCase();
    if (sc) _deckSetCounts[sc] = (_deckSetCounts[sc] || 0) + qty;
  }
  const DECK_TYPE_ORDER = ["Criatura Lendária","Criatura","Planeswalker","Encantamento Lendário","Encantamento","Artefato Lendário","Artefato","Instantâneo","Feitiço","Batalha","Terreno","Outro"];
  const deckTypeCounts = DECK_TYPE_ORDER.filter(t => _deckTypeCounts[t]).map(t => ({ label: t, count: _deckTypeCounts[t] }));
  const DECK_COLOR_ORDER = ["W","U","B","R","G","C"];
  const deckColorCounts = DECK_COLOR_ORDER.filter(col => _deckColorCounts[col]).map(col => ({ code: col, count: _deckColorCounts[col] }));
  const deckSetCounts = Object.entries(_deckSetCounts).sort(([,a],[,b]) => b - a).map(([set, count]) => ({ set, count }));

  // Anomalias do deck
  const BASIC_LAND_NAMES = new Set(["Plains","Island","Swamp","Mountain","Forest","Wastes",
    "Snow-Covered Plains","Snow-Covered Island","Snow-Covered Swamp",
    "Snow-Covered Mountain","Snow-Covered Forest"]);
  const deckAnomalies = (() => {
    const anomalies = [];
    const total = deckCards.reduce((s, c) => s + (c.quantity || 1), 0);
    if (total > 100) {
      anomalies.push({ icon: "🃏", label: `${total} cartas no deck`, detail: "Formato Commander permite no máximo 100 cartas." });
    }
    for (const c of deckCards) {
      if (!BASIC_LAND_NAMES.has(c.name) && (c.quantity || 1) > 1) {
        anomalies.push({ icon: "📋", label: `"${c.name}" tem ${c.quantity} cópias`, detail: "Regra singleton: apenas 1 cópia de cada carta não-terreno básico." });
      }
    }
    // Cor fora da identidade — apenas Commander com cores definidas
    if (managingDeck?.commander && managingDeck.colors) {
      const allowed = new Set(managingDeck.colors.split(",").map(x => x.trim()).filter(Boolean));
      for (const c of deckCards) {
        if (c.commander) continue; // o próprio comandante define a identidade
        try {
          const cardColors = JSON.parse(c.colors || "[]");
          const outside = cardColors.filter(col => col !== "C" && !allowed.has(col));
          if (outside.length > 0) {
            anomalies.push({
              icon: "🎨",
              label: `"${c.name}" tem cor fora da identidade`,
              detail: `Cor(es) ${outside.join(", ")} não pertencem à identidade do deck (${managingDeck.colors}).`,
            });
          }
        } catch {}
      }
    }
    return anomalies;
  })();

  // Comandantes sobem para o topo da lista
  const deckCardsSorted = [...deckCards].sort((a, b) => (b.commander ? 1 : 0) - (a.commander ? 1 : 0));

  const filteredDeckCards = deckCardsSorted.filter(c => {
    if (deckTypeFilter && deckCardTypeGroup(c.type) !== deckTypeFilter) return false;
    if (deckColorFilter) {
      try {
        const cols = JSON.parse(c.colors || "[]");
        if (deckColorFilter === "C" ? cols.length !== 0 : !cols.includes(deckColorFilter)) return false;
      } catch { return false; }
    }
    if (deckSetFilter && (c.set_code || "").toUpperCase() !== deckSetFilter) return false;
    return true;
  });
  const DECK_PAGE_SIZE = 20;
  const deckTotalPages = Math.max(1, Math.ceil(filteredDeckCards.length / DECK_PAGE_SIZE));
  const pagedDeckCards = filteredDeckCards.slice((deckPage - 1) * DECK_PAGE_SIZE, deckPage * DECK_PAGE_SIZE);

  // ── Guarda de autenticação (todos os hooks já foram chamados acima) ───
  if (!authReady) return (
    <div className="auth-loading">
      <div className="eval-spinner">⚙</div>
      <p>Carregando…</p>
    </div>
  );

  if (!authUser || showLanding) return (
    <LandingPage
      onEnter={handleAuthEnter}
      onBack={authUser ? () => setShowLanding(false) : null}
    />
  );

  const userInitials = authUser.display_name
    .split(" ").map((w) => w[0]).join("").slice(0, 2).toUpperCase();

  return (
    <>
      {/* ── Session bar ── */}
      <div className="session-bar">
        <div className="session-bar-left">
          <div className="session-avatar">{userInitials}</div>
          <div className="session-info">
            <span className="session-name">Olá, {authUser.display_name}</span>
            {sessionElapsed && (
              <span className="session-time">Logado há {sessionElapsed}</span>
            )}
          </div>
        </div>
        <div className="session-bar-actions">
          <button className="session-home" onClick={() => setShowLanding(true)}>⚔ Início</button>
          <button className="session-logout" onClick={handleLogout}>Sair</button>
        </div>
      </div>

      <main className="app">
      <section className="hero">
        <h1>Magic Collector</h1>
        <p>Cadastre, organize e consulte sua coleção de cartas Magic: The Gathering</p>
      </section>

      <nav className="tabs" role="tablist" aria-label="Navegação principal">
        <button role="tab" type="button" aria-selected={activeTab === "collection"} className={`tab${activeTab === "collection" ? " active" : ""}`} onClick={() => setActiveTab("collection")}>
          <span className="tab-icon" aria-hidden="true">🃏</span><span className="tab-label">Coleção</span>
        </button>
        <button role="tab" type="button" aria-selected={activeTab === "decks"} className={`tab${activeTab === "decks" ? " active" : ""}`} onClick={() => setActiveTab("decks")}>
          <span className="tab-icon" aria-hidden="true">🗂</span><span className="tab-label">Decks</span>
        </button>
        <button role="tab" type="button" aria-selected={activeTab === "battles"} className={`tab${activeTab === "battles" ? " active" : ""}`} onClick={() => setActiveTab("battles")}>
          <span className="tab-icon" aria-hidden="true">⚔</span><span className="tab-label">Batalhas</span>
        </button>
        <button role="tab" type="button" aria-selected={activeTab === "wishlist"} className={`tab tab-wishlist${activeTab === "wishlist" ? " active" : ""}`} onClick={() => setActiveTab("wishlist")}>
          <span className="tab-icon" aria-hidden="true">⭐</span><span className="tab-label">Wishlist</span>
        </button>
        <button role="tab" type="button" aria-selected={activeTab === "tokens"} className={`tab tab-tokens${activeTab === "tokens" ? " active" : ""}`} onClick={() => setActiveTab("tokens")}>
          <span className="tab-icon" aria-hidden="true">🎭</span><span className="tab-label">Tokens</span>
        </button>
        <button role="tab" type="button" aria-selected={activeTab === "score"} className={`tab tab-score${activeTab === "score" ? " active" : ""}`} onClick={() => setActiveTab("score")}>
          <span className="tab-icon" aria-hidden="true">🎮</span><span className="tab-label">Pontuação</span>
        </button>
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
                    {deckAnomalies.length > 0 && (
                      <button type="button" className="deck-anomaly-btn" onClick={() => setDeckAnomalyModal(true)}>
                        ⚠️ {deckAnomalies.length} anomalia{deckAnomalies.length > 1 ? "s" : ""}
                      </button>
                    )}
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
                    <h3 className="section-title">
                      No deck <span className="total-badge">{deckTotalQuantity}</span>
                      <span className="unique-badge">{deckCards.length} únicas</span>
                      {(deckTypeFilter || deckColorFilter || deckSetFilter) && (
                        <span className="unique-badge" style={{color:"var(--gold)"}}>
                          {filteredDeckCards.length} filtradas
                        </span>
                      )}
                    </h3>
                  </div>

                  {deckCards.length > 0 && (
                    <div className="deck-filter-bar">
                      {deckTypeCounts.length > 0 && (
                        <div className="deck-filter-group">
                          <span className="deck-filter-label">Tipo</span>
                          <div className="deck-filter-chips">
                            {deckTypeCounts.map(({label, count}) => (
                              <button key={label} type="button"
                                className={`deck-filter-chip${deckTypeFilter === label ? " active" : ""}`}
                                onClick={() => { setDeckTypeFilter(deckTypeFilter === label ? "" : label); setDeckPage(1); }}>
                                {label} <span className="dfc-count">({count})</span>
                              </button>
                            ))}
                          </div>
                        </div>
                      )}
                      {deckColorCounts.length > 0 && (
                        <div className="deck-filter-group">
                          <span className="deck-filter-label">Cor</span>
                          <div className="deck-filter-chips">
                            {deckColorCounts.map(({code, count}) => (
                              <button key={code} type="button"
                                className={`deck-filter-chip deck-color-chip dc-${code.toLowerCase()}${deckColorFilter === code ? " active" : ""}`}
                                onClick={() => { setDeckColorFilter(deckColorFilter === code ? "" : code); setDeckPage(1); }}>
                                {code} <span className="dfc-count">({count})</span>
                              </button>
                            ))}
                          </div>
                        </div>
                      )}
                      {deckSetCounts.length > 1 && (
                        <div className="deck-filter-group">
                          <span className="deck-filter-label">Coleção</span>
                          <div className="deck-filter-chips">
                            {deckSetCounts.map(({set, count}) => (
                              <button key={set} type="button"
                                className={`deck-filter-chip${deckSetFilter === set ? " active" : ""}`}
                                onClick={() => { setDeckSetFilter(deckSetFilter === set ? "" : set); setDeckPage(1); }}>
                                {set} <span className="dfc-count">({count})</span>
                              </button>
                            ))}
                          </div>
                        </div>
                      )}
                      {(deckTypeFilter || deckColorFilter || deckSetFilter) && (
                        <button type="button" className="deck-filter-clear"
                          onClick={() => { setDeckTypeFilter(""); setDeckColorFilter(""); setDeckSetFilter(""); setDeckPage(1); }}>
                          ✕ Limpar filtros
                        </button>
                      )}
                    </div>
                  )}

                  <div className="list">
                    {pagedDeckCards.map((card) => (
                      <div className={`list-item deck-card-item${card.commander ? " deck-card-commander" : ""}${card.foil ? " is-foil" : ""} item-r-${(card.rarity || "x").toLowerCase()}`} key={card.id}>
                        {card.image_url && (
                          <div className="deck-card-thumb">
                            <img src={card.image_url} alt={card.name} loading="lazy" />
                          </div>
                        )}
                        <div className="list-item-info">
                          <div className="list-item-name">
                            {card.commander && <span className="deck-cmd-crown" title="Comandante">👑</span>}
                            <strong className={card.foil ? "foil-text" : ""}>{card.name}</strong>
                            {card.foil && <span className="foil-text">✦</span>}
                            <CardColorIcons card={card} />
                            {card.rarity && <span className={`rarity r-${card.rarity.toLowerCase()}`}>{card.rarity}</span>}
                            {card.mana_cost && <span className="deck-card-mana">{card.mana_cost}</span>}
                          </div>
                          {card.type && <div className="deck-card-type">{card.type}</div>}
                          <small>
                            {card.set_code || "—"} · #{card.collection_number || "—"} · {card.language} · ×{card.quantity}
                            {card.price_usd > 0 && <span className="deck-card-price"> · ${card.price_usd.toFixed(2)}</span>}
                          </small>
                        </div>
                        <div className="actions">
                          <button type="button" onClick={() => handleDeckDetails(card.id)}>Ver</button>
                          <button type="button" className="danger" onClick={() => handleUnassignCard(card.id)}>Remover</button>
                        </div>
                      </div>
                    ))}
                    {filteredDeckCards.length === 0 && deckCards.length > 0 && (
                      <p className="empty">Nenhuma carta nesse filtro.</p>
                    )}
                    {deckCards.length === 0 && <p className="empty">Nenhuma carta no deck.</p>}
                  </div>

                  {deckTotalPages > 1 && (
                    <div className="pagination" style={{marginTop:"0.6rem"}}>
                      <button disabled={deckPage === 1} onClick={() => setDeckPage(p => p - 1)}>‹</button>
                      {Array.from({length: deckTotalPages}, (_, i) => i + 1).map(p => (
                        <button key={p} className={deckPage === p ? "active" : ""} onClick={() => setDeckPage(p)}>{p}</button>
                      ))}
                      <button disabled={deckPage === deckTotalPages} onClick={() => setDeckPage(p => p + 1)}>›</button>
                    </div>
                  )}
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
                  <div className="deck-import-btns">
                    <button type="button" className="import-precon-btn" onClick={() => { setImportModal(true); setImportResult(null); setImportError(""); setImportForm(EMPTY_IMPORT_FORM); }}>Importar Pré-con</button>
                    <button type="button" className="import-precon-btn" onClick={() => { setListModal(true); setListResult(null); setListError(""); setListForm(EMPTY_LIST_FORM); }}>Importar Lista</button>
                  </div>
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
                      {deck.commanders?.length > 0 && (
                        <div className="deck-commanders-row">
                          {deck.commanders.map(cmd => (
                            <div key={cmd.id} className="deck-cmd-pill">
                              {cmd.image_url && <img src={cmd.image_url} alt={cmd.name} className="deck-cmd-pill-img" />}
                              <span className="deck-cmd-pill-name">👑 {cmd.name}</span>
                            </div>
                          ))}
                        </div>
                      )}
                      <div className="deck-list-meta">
                        {deck.set_code && <span className="deck-set-label">{deck.set_code.toUpperCase()}</span>}
                        <span className="total-badge">{deck.card_count} cartas</span>
                        {deck.battle_total > 0 && (() => {
                          const rate = Math.round((deck.battle_wins / deck.battle_total) * 100);
                          const cls = rate >= 60 ? "wr-high" : rate >= 40 ? "wr-mid" : "wr-low";
                          return (
                            <span className={`win-rate-badge ${cls}`}>
                              {deck.battle_wins}V · {deck.battle_losses}D · {rate}%
                            </span>
                          );
                        })()}
                        {deck.description && <span className="deck-desc">{deck.description}</span>}
                      </div>
                    </div>
                    <div className="actions">
                      <button type="button" onClick={() => handleManageDeck(deck)}>Cartas</button>
                      <button type="button" onClick={() => setEditDeckModal({ ...deck })}>Editar</button>
                      <button type="button" className="danger" aria-label="Remover deck" onClick={() => handleDeckDelete(deck.id)}>✕</button>
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
                        <button type="button" className="danger battle-del" aria-label="Remover batalha" onClick={() => handleBattleDelete(b.id)}>✕</button>
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

      {/* ── ABA WISHLIST ── */}
      {activeTab === "wishlist" && (
        <>
          <section className="grid">

            {/* Formulário */}
            <form className="card form wishlist-form-panel" onSubmit={handleWishlistCreate}>
                <h2>Adicionar à Wishlist</h2>
                <div className="qa-row-2col">
                  <label>
                    Sigla *
                    <input
                      required
                      placeholder="KLD, BRO…"
                      value={wishlistForm.set_code}
                      style={{ textTransform: "uppercase" }}
                      onChange={e => setWishlistForm({ ...wishlistForm, set_code: e.target.value.toUpperCase().trim() })}
                    />
                  </label>
                  <label>
                    Número *
                    <input
                      required
                      placeholder="017, 253a"
                      value={wishlistForm.collection_number}
                      onChange={e => setWishlistForm({ ...wishlistForm, collection_number: e.target.value.trim() })}
                    />
                  </label>
                </div>
                <label className="checkbox-label">
                  <input
                    type="checkbox"
                    checked={wishlistForm.foil}
                    onChange={e => setWishlistForm({ ...wishlistForm, foil: e.target.checked })}
                  />
                  Versão Foil
                </label>
                <label>
                  Motivo
                  <textarea
                    placeholder="Por que você quer essa carta?"
                    value={wishlistForm.reason}
                    onChange={e => setWishlistForm({ ...wishlistForm, reason: e.target.value })}
                  />
                </label>
                {wishlistError && <p className="form-error">{wishlistError}</p>}
                <button type="submit" className="wishlist-submit-btn" disabled={wishlistSubmitting}>
                  {wishlistSubmitting ? "Buscando na Scryfall…" : "＋ Adicionar à Wishlist"}
                </button>
            </form>

            {/* Lista */}
            <section className="card list-section">
              <div className="list-header">
                <h2>
                  Cartas Desejadas
                  <span className="total-badge">{wishlistItems.length}</span>
                  {wishlistItems.filter(i => i.acquired).length > 0 && (
                    <span className="unique-badge">
                      {wishlistItems.filter(i => i.acquired).length} adquiridas
                    </span>
                  )}
                </h2>
              </div>

              {wishlistItems.length === 0 ? (
                <p className="empty">Sua wishlist está vazia. Adicione cartas que deseja adquirir!</p>
              ) : (
                <div className="wishlist-grid">
                  {wishlistItems.map(item => (
                    <div key={item.id} className={`wishlist-card${item.acquired ? " wishlist-acquired" : ""}`}>
                      {item.image_uri ? (
                        <div className="wishlist-card-img-wrap">
                          <img
                            src={item.image_uri}
                            alt={item.name || item.set_code}
                            className="wishlist-card-img"
                            loading="lazy"
                          />
                        </div>
                      ) : (
                        <div className="wishlist-card-img-placeholder">
                          <span>{item.set_code || "?"}</span>
                        </div>
                      )}
                      <div className="wishlist-card-body">
                        <div className="wishlist-card-name">
                          <span className="wishlist-card-name-text">
                            {item.printed_name || item.name || `${item.set_code} #${item.collection_number}`}
                          </span>
                          <div className="wishlist-card-badges">
                            {item.foil && <span className="wishlist-badge-foil">Foil</span>}
                            {item.acquired && <span className="wishlist-badge-acquired">✓</span>}
                          </div>
                        </div>
                        <div className="wishlist-card-meta">
                          <span className="wishlist-meta-set">{item.set_code} #{item.collection_number}</span>
                          {item.rarity && (
                            <span className={`rarity r-${item.rarity.toLowerCase()}`}>{item.rarity}</span>
                          )}
                        </div>
                        {item.artist && <div className="wishlist-card-artist">🎨 {item.artist}</div>}
                        {(item.price_usd > 0 || item.price_usd_foil > 0) && (
                          <div className="wishlist-card-price">
                            {item.price_usd > 0 && <span>${item.price_usd.toFixed(2)}</span>}
                            {item.price_usd_foil > 0 && (
                              <span className="wishlist-price-foil">Foil: ${item.price_usd_foil.toFixed(2)}</span>
                            )}
                          </div>
                        )}
                        {item.reason && (
                          <div className="wishlist-card-reason">"{item.reason}"</div>
                        )}
                        <div className="wishlist-card-actions">
                          <button
                            type="button"
                            className="wishlist-btn-detail"
                            onClick={() => setWishlistDetail(item)}
                          >
                            Detalhes
                          </button>
                          {!item.acquired && (
                            <button
                              type="button"
                              className="wishlist-btn-acquire"
                              onClick={() => {
                                setWishlistAcquireFor(item);
                                setWishlistAcquireForm(EMPTY_ACQUIRE_FORM);
                              }}
                            >
                              Adquirir
                            </button>
                          )}
                          <button
                            type="button"
                            className="danger wishlist-btn-remove"
                            aria-label="Remover da wishlist"
                            onClick={() => handleWishlistDelete(item.id)}
                          >
                            ✕
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </section>
          </section>

          {/* Modal de Detalhes */}
          {wishlistDetail && (
            <div className="modal-overlay" onClick={() => setWishlistDetail(null)}>
              <div className="modal wishlist-modal" onClick={e => e.stopPropagation()}>
                <button className="modal-close" aria-label="Fechar" onClick={() => setWishlistDetail(null)}>✕</button>
                {wishlistDetail.image_uri && (
                  <img
                    src={wishlistDetail.image_uri}
                    alt={wishlistDetail.name}
                    className="wishlist-modal-img"
                  />
                )}
                <div className="wishlist-modal-body">
                  <h2 className="wishlist-modal-name">
                    {wishlistDetail.printed_name || wishlistDetail.name || `${wishlistDetail.set_code} #${wishlistDetail.collection_number}`}
                    {wishlistDetail.foil && <span className="wishlist-badge-foil">Foil</span>}
                  </h2>
                  {wishlistDetail.printed_name && wishlistDetail.name && wishlistDetail.printed_name !== wishlistDetail.name && (
                    <p className="wishlist-modal-en-name">{wishlistDetail.name}</p>
                  )}
                  <div className="wishlist-modal-meta">
                    <span>{wishlistDetail.set_code} #{wishlistDetail.collection_number}</span>
                    {wishlistDetail.rarity && (
                      <span className={`rarity r-${wishlistDetail.rarity.toLowerCase()}`}>{wishlistDetail.rarity}</span>
                    )}
                  </div>
                  {wishlistDetail.artist && (
                    <p className="wishlist-modal-artist">🎨 {wishlistDetail.artist}</p>
                  )}
                  {(wishlistDetail.price_usd > 0 || wishlistDetail.price_usd_foil > 0) && (
                    <div className="wishlist-modal-prices">
                      {wishlistDetail.price_usd > 0 && (
                        <span>Normal: <strong>${wishlistDetail.price_usd.toFixed(2)}</strong></span>
                      )}
                      {wishlistDetail.price_usd_foil > 0 && (
                        <span>Foil: <strong>${wishlistDetail.price_usd_foil.toFixed(2)}</strong></span>
                      )}
                    </div>
                  )}
                  {wishlistDetail.reason && (
                    <blockquote className="wishlist-modal-reason">"{wishlistDetail.reason}"</blockquote>
                  )}
                  {wishlistDetail.acquired && (
                    <p className="wishlist-acquired-label">✓ Carta já adquirida e na coleção</p>
                  )}
                </div>
                <div className="modal-actions">
                  {!wishlistDetail.acquired && (
                    <button
                      type="button"
                      className="wishlist-btn-acquire"
                      onClick={() => {
                        setWishlistAcquireFor(wishlistDetail);
                        setWishlistAcquireForm(EMPTY_ACQUIRE_FORM);
                        setWishlistDetail(null);
                      }}
                    >
                      + Adquirir
                    </button>
                  )}
                  <button
                    type="button"
                    className="danger"
                    onClick={() => handleWishlistDelete(wishlistDetail.id)}
                  >
                    Remover
                  </button>
                  <button type="button" onClick={() => setWishlistDetail(null)}>Fechar</button>
                </div>
              </div>
            </div>
          )}

          {/* Modal de Aquisição */}
          {wishlistAcquireFor && (
            <div className="modal-overlay" onClick={() => setWishlistAcquireFor(null)}>
              <div className="modal wishlist-acquire-modal" onClick={e => e.stopPropagation()}>
                <button className="modal-close" aria-label="Fechar" onClick={() => setWishlistAcquireFor(null)}>✕</button>
                <h2>Adquirir Carta</h2>
                <p className="wishlist-acquire-card-name">
                  {wishlistAcquireFor.printed_name || wishlistAcquireFor.name || `${wishlistAcquireFor.set_code} #${wishlistAcquireFor.collection_number}`}
                  {wishlistAcquireFor.foil && <span className="wishlist-badge-foil">Foil</span>}
                </p>
                <form className="form" onSubmit={handleWishlistAcquire}>
                  <label>
                    Deck
                    <select
                      value={wishlistAcquireForm.deck_id}
                      onChange={e => setWishlistAcquireForm({ ...wishlistAcquireForm, deck_id: +e.target.value })}
                    >
                      <option value={0}>Sem deck</option>
                      {decks.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
                    </select>
                  </label>
                  <label>
                    Condição
                    <select
                      value={wishlistAcquireForm.condition}
                      onChange={e => setWishlistAcquireForm({ ...wishlistAcquireForm, condition: e.target.value })}
                    >
                      <option value="mint">Mint</option>
                      <option value="near_mint">Near Mint</option>
                      <option value="played">Played</option>
                      <option value="damaged">Damaged</option>
                    </select>
                  </label>
                  <label className="checkbox-label">
                    <input
                      type="checkbox"
                      checked={wishlistAcquireForm.commander}
                      onChange={e => setWishlistAcquireForm({ ...wishlistAcquireForm, commander: e.target.checked })}
                    />
                    É comandante do deck
                  </label>
                  {wishlistError && <p className="form-error">{wishlistError}</p>}
                  <div className="modal-actions">
                    <button type="button" onClick={() => setWishlistAcquireFor(null)} disabled={wishlistAcquiring}>
                      Cancelar
                    </button>
                    <button type="submit" className="wishlist-btn-acquire" disabled={wishlistAcquiring}>
                      {wishlistAcquiring ? "Salvando…" : "Confirmar Aquisição"}
                    </button>
                  </div>
                </form>
              </div>
            </div>
          )}
        </>
      )}

      {/* ── ABA TOKENS ── */}
      {activeTab === "tokens" && (
        <div className="tokens-page">
          <div className="tokens-page-header">
            <h2>Tokens <span className="total-badge">{tokenList.length}</span></h2>
            <button type="button" className="token-add-btn" onClick={() => { setTokenAddModal(true); setTokenPreviewFront(null); setTokenPreviewBack(null); setTokenForm(EMPTY_TOKEN_FORM); setTokenSearchError(""); }}>
              ＋ Adicionar Token
            </button>
          </div>

          {tokenList.length === 0 ? (
            <p className="empty">Nenhum token cadastrado. Clique em "Adicionar Token" para buscar na Scryfall.</p>
          ) : (
            <div className="tokens-grid">
              {tokenList.map(tok => (
                <div key={tok.id} className={`token-card${tok.foil ? " token-foil" : ""}${tok.double_faced ? " token-dfc" : ""}`}
                  onClick={() => setTokenDetail(tok)} role="button" tabIndex={0} onKeyDown={e => e.key === "Enter" && setTokenDetail(tok)}>
                  <div className={`token-images${tok.double_faced && tok.back_image_url ? " token-images-dfc" : ""}`}>
                    {tok.image_url
                      ? <img src={tok.image_url} alt={tok.name} className="token-img" loading="lazy" />
                      : <div className="token-img-placeholder">?</div>
                    }
                    {tok.double_faced && tok.back_image_url && (
                      <img src={tok.back_image_url} alt={tok.back_name || tok.name} className="token-img" loading="lazy" />
                    )}
                  </div>
                  <div className="token-body">
                    <div className="token-name">
                      {tok.name}
                      {tok.double_faced && tok.back_name && tok.back_name !== tok.name && (
                        <span className="token-dfc-sep"> ↔ {tok.back_name}</span>
                      )}
                    </div>
                    <div className="token-type">{tok.type_line}</div>
                    {tok.power && <div className="token-pt">{tok.power}/{tok.toughness}</div>}
                    <div className="token-meta">
                      <span className="token-set">{tok.set_code} #{tok.collection_number}</span>
                      {tok.double_faced && <span className="token-dfc-badge">↔</span>}
                      {tok.foil && <span className="token-foil-badge">✦</span>}
                    </div>
                    <div className="token-qty-row" onClick={e => e.stopPropagation()}>
                      <button type="button" className="token-qty-btn" onClick={() => handleTokenQuantityChange(tok, -1)}>−</button>
                      <span className="token-qty">×{tok.quantity}</span>
                      <button type="button" className="token-qty-btn" onClick={() => handleTokenQuantityChange(tok, +1)}>+</button>
                    </div>
                  </div>
                  <button type="button" className="token-delete-btn danger" aria-label="Remover token"
                    onClick={e => { e.stopPropagation(); handleTokenDelete(tok.id); }}>✕</button>
                </div>
              ))}
            </div>
          )}

          {/* Modal detalhe do token */}
          {tokenDetail && (
            <div className="modal-overlay" onClick={() => setTokenDetail(null)}>
              <div className="modal token-detail-modal" onClick={e => e.stopPropagation()}>
                <button className="modal-close" aria-label="Fechar" onClick={() => setTokenDetail(null)}>✕</button>
                <div className="token-detail-images">
                  {tokenDetail.image_url && (
                    <img src={tokenDetail.image_url} alt={tokenDetail.name} className="token-detail-img" />
                  )}
                  {tokenDetail.double_faced && tokenDetail.back_image_url && (
                    <img src={tokenDetail.back_image_url} alt={tokenDetail.back_name || tokenDetail.name} className="token-detail-img" />
                  )}
                </div>
                <div className="token-detail-body">
                  {/* Frente */}
                  <div className="token-detail-face">
                    {tokenDetail.double_faced && <div className="token-face-label">Frente</div>}
                    <h2 className="token-detail-name">{tokenDetail.name}</h2>
                    <div className="token-detail-type">{tokenDetail.type_line}</div>
                    {tokenDetail.power && (
                      <div className="token-detail-pt">{tokenDetail.power} / {tokenDetail.toughness}</div>
                    )}
                    {tokenDetail.oracle_text && (
                      <div className="token-detail-oracle">{tokenDetail.oracle_text}</div>
                    )}
                  </div>

                  {/* Verso (dupla face) */}
                  {tokenDetail.double_faced && tokenDetail.back_name && (
                    <div className="token-detail-face token-detail-back-face">
                      <div className="token-face-label">Verso</div>
                      <h3 className="token-detail-name">{tokenDetail.back_name}</h3>
                      <div className="token-detail-type">{tokenDetail.back_type_line}</div>
                      {tokenDetail.back_power && (
                        <div className="token-detail-pt">{tokenDetail.back_power} / {tokenDetail.back_toughness}</div>
                      )}
                      {tokenDetail.back_oracle_text && (
                        <div className="token-detail-oracle">{tokenDetail.back_oracle_text}</div>
                      )}
                    </div>
                  )}

                  <div className="token-detail-meta-row">
                    <span className="token-set">{tokenDetail.set_code} #{tokenDetail.collection_number}</span>
                    {tokenDetail.artist && <span>🎨 {tokenDetail.artist}</span>}
                    {tokenDetail.foil && <span className="token-foil-badge">✦ Foil</span>}
                    {tokenDetail.double_faced && <span className="token-dfc-badge">↔ Dupla Face</span>}
                    <span>×{tokenDetail.quantity}</span>
                  </div>

                  <div className="token-detail-actions">
                    <div className="token-qty-row">
                      <button type="button" className="token-qty-btn" onClick={() => { handleTokenQuantityChange(tokenDetail, -1); setTokenDetail(prev => prev ? { ...prev, quantity: Math.max(1, prev.quantity - 1) } : prev); }}>−</button>
                      <span className="token-qty">×{tokenDetail.quantity}</span>
                      <button type="button" className="token-qty-btn" onClick={() => { handleTokenQuantityChange(tokenDetail, +1); setTokenDetail(prev => prev ? { ...prev, quantity: prev.quantity + 1 } : prev); }}>+</button>
                    </div>
                    <button type="button" className="danger" onClick={() => { handleTokenDelete(tokenDetail.id); setTokenDetail(null); }}>
                      ✕ Remover token
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Modal adicionar token */}
          {tokenAddModal && (() => {
            const frontDone = tokenPreviewFront?.found;
            const backDone = tokenPreviewBack?.found;
            const canConfirm = tokenForm.double_faced ? (frontDone && backDone) : frontDone;
            const isSaving = tokenSearchingFront || tokenSearchingBack;
            return (
              <div className="modal-overlay" onClick={() => { if (!isSaving) setTokenAddModal(false); }}>
                <div className="modal token-add-modal" onClick={e => e.stopPropagation()}>
                  <button className="modal-close" aria-label="Fechar" onClick={() => setTokenAddModal(false)}>✕</button>
                  <h3>Adicionar Token</h3>
                  <p className="quick-add-hint">
                    Informe a sigla da coleção <strong>sem o "t"</strong> (ex: <code>grn</code>) e o número do token.
                  </p>

                  {/* Opções globais */}
                  <div className="token-options-row">
                    <label>
                      Quantidade
                      <input type="number" min="1" value={tokenForm.quantity}
                        onChange={e => setTokenForm({ ...tokenForm, quantity: Number(e.target.value) })} />
                    </label>
                    <label className="checkbox-label">
                      <input type="checkbox" checked={tokenForm.foil}
                        onChange={e => setTokenForm({ ...tokenForm, foil: e.target.checked })} />
                      ✦ Foil
                    </label>
                    <label className="checkbox-label token-dfc-toggle">
                      <input type="checkbox" checked={tokenForm.double_faced}
                        onChange={e => { setTokenForm({ ...tokenForm, double_faced: e.target.checked }); setTokenPreviewBack(null); }} />
                      ↔ Dupla Face
                    </label>
                  </div>

                  {/* Busca da FRENTE */}
                  <div className="token-face-section">
                    <div className="token-face-section-title">
                      {tokenForm.double_faced ? "Frente" : "Token"}
                      {frontDone && <span className="token-face-ok">✓</span>}
                    </div>
                    {frontDone ? (
                      <div className="token-face-preview">
                        {tokenPreviewFront.token?.image_url && (
                          <img src={tokenPreviewFront.token.image_url} alt={tokenPreviewFront.token.name} className="token-face-preview-img" />
                        )}
                        <div>
                          <strong>{tokenPreviewFront.token?.name}</strong>
                          <div className="token-face-preview-type">{tokenPreviewFront.token?.type_line}</div>
                          {tokenPreviewFront.token?.power && <div>{tokenPreviewFront.token.power}/{tokenPreviewFront.token.toughness}</div>}
                          <div className="token-face-preview-set">{tokenPreviewFront.token?.set_code} #{tokenPreviewFront.token?.collection_number}</div>
                        </div>
                        <button type="button" className="token-face-reset" onClick={() => setTokenPreviewFront(null)}>✎</button>
                      </div>
                    ) : (
                      <form className="token-face-form" onSubmit={handleTokenSearchFront}>
                        <input required placeholder="Sigla (ex: grn)" value={tokenForm.set_code}
                          onChange={e => setTokenForm({ ...tokenForm, set_code: e.target.value.toLowerCase().trim() })} />
                        <input required placeholder="Nº (ex: 1)" value={tokenForm.collection_number}
                          onChange={e => setTokenForm({ ...tokenForm, collection_number: e.target.value.trim() })} />
                        <button type="submit" disabled={tokenSearchingFront}>
                          {tokenSearchingFront ? "…" : "🔍"}
                        </button>
                      </form>
                    )}
                  </div>

                  {/* Busca do VERSO (dupla face) */}
                  {tokenForm.double_faced && (
                    <div className="token-face-section">
                      <div className="token-face-section-title">
                        Verso
                        {backDone && <span className="token-face-ok">✓</span>}
                      </div>
                      {backDone ? (
                        <div className="token-face-preview">
                          {tokenPreviewBack.token?.image_url && (
                            <img src={tokenPreviewBack.token.image_url} alt={tokenPreviewBack.token.name} className="token-face-preview-img" />
                          )}
                          <div>
                            <strong>{tokenPreviewBack.token?.name}</strong>
                            <div className="token-face-preview-type">{tokenPreviewBack.token?.type_line}</div>
                            {tokenPreviewBack.token?.power && <div>{tokenPreviewBack.token.power}/{tokenPreviewBack.token.toughness}</div>}
                            <div className="token-face-preview-set">{tokenPreviewBack.token?.set_code} #{tokenPreviewBack.token?.collection_number}</div>
                          </div>
                          <button type="button" className="token-face-reset" onClick={() => setTokenPreviewBack(null)}>✎</button>
                        </div>
                      ) : (
                        <form className="token-face-form" onSubmit={handleTokenSearchBack}>
                          <input required placeholder="Sigla (ex: grn)" value={tokenForm.back_set_code}
                            onChange={e => setTokenForm({ ...tokenForm, back_set_code: e.target.value.toLowerCase().trim() })} />
                          <input required placeholder="Nº (ex: 2)" value={tokenForm.back_collection_number}
                            onChange={e => setTokenForm({ ...tokenForm, back_collection_number: e.target.value.trim() })} />
                          <button type="submit" disabled={tokenSearchingBack}>
                            {tokenSearchingBack ? "…" : "🔍"}
                          </button>
                        </form>
                      )}
                    </div>
                  )}

                  {/* Preview dupla face */}
                  {tokenForm.double_faced && frontDone && backDone && (
                    <div className="token-dfc-preview-row">
                      <img src={tokenPreviewFront.token?.image_url} alt={tokenPreviewFront.token?.name} className="token-dfc-preview-img" />
                      <span className="token-dfc-arrow">↔</span>
                      <img src={tokenPreviewBack.token?.image_url} alt={tokenPreviewBack.token?.name} className="token-dfc-preview-img" />
                    </div>
                  )}

                  {tokenSearchError && <p className="form-error">{tokenSearchError}</p>}

                  {canConfirm && (
                    <div className="token-preview-actions">
                      <button type="button" className="confirm-yes" disabled={isSaving} onClick={handleTokenConfirm}>
                        {isSaving ? "Salvando…" : "✓ Adicionar à coleção"}
                      </button>
                      <button type="button" className="confirm-no" onClick={() => { setTokenPreviewFront(null); setTokenPreviewBack(null); setTokenSearchError(""); }}>
                        ← Buscar outro
                      </button>
                    </div>
                  )}
                </div>
              </div>
            );
          })()}
        </div>
      )}

      {/* ── ABA PONTUAÇÃO ── */}
      {activeTab === "score" && (
        <>

          {/* Vista: Lista de sessões */}
          {sessionView === "list" && (
            <section className="grid">
              <section className="card form score-form-panel">
                <h2>Pontuação</h2>
                <p className="score-form-intro">Gerencie sessões de jogo e controle de vida.</p>
                <div className="score-form-stats">
                  <div className="score-form-stat">
                    <span className="score-form-stat-num">{gameSessions.length}</span>
                    <span className="score-form-stat-label">Sessões</span>
                  </div>
                  <div className="score-form-stat">
                    <span className="score-form-stat-num">{gameSessions.filter(s => s.status === "active").length}</span>
                    <span className="score-form-stat-label">Ativas</span>
                  </div>
                  <div className="score-form-stat">
                    <span className="score-form-stat-num">{gameSessions.filter(s => s.status === "finished").length}</span>
                    <span className="score-form-stat-label">Encerradas</span>
                  </div>
                </div>
                {sessionError && <p className="form-error">{sessionError}</p>}
                <button type="button" onClick={() => { setSessionError(""); setSessionView("create"); }}>
                  + Nova Sessão
                </button>
              </section>
              <section className="card list-section score-list-panel">
                <div className="list-header">
                  <div className="list-header-top">
                    <h2>Sessões de Jogo <span className="total-badge">{gameSessions.length}</span></h2>
                  </div>
                </div>
                {gameSessions.length === 0 ? (
                  <p className="score-empty-inline">Nenhuma sessão ainda.</p>
                ) : (
                <div className="score-session-cards">
                  {gameSessions.map(s => (
                    <div key={s.id} className={`score-session-card${s.status === "finished" ? " finished" : ""}`}>
                      <div className="score-session-card-header">
                        <span className={`score-badge ${s.status}`}>
                          {s.status === "active" ? "Ativo" : "Encerrado"}
                        </span>
                        <span className="score-format-tag">{s.format}</span>
                        <span className="score-life-tag">♥ {s.starting_life}</span>
                      </div>
                      <h3 className="score-session-name">{s.name}</h3>
                      <div className="score-players-summary">
                        {(s.players || []).map(p => (
                          <span key={p.id} className={`score-player-chip${p.is_eliminated ? " elim" : ""}`}>
                            {p.short_code}{p.is_eliminated ? " ☠" : ""}
                          </span>
                        ))}
                      </div>
                      <p className="score-session-date">
                        {new Date(s.created_at).toLocaleDateString("pt-BR", { day: "2-digit", month: "short", year: "numeric" })}
                        {s.ended_at && ` — Encerrada ${new Date(s.ended_at).toLocaleDateString("pt-BR", { day: "2-digit", month: "short" })}`}
                      </p>
                      <div className="score-session-actions">
                        <button type="button" className="score-btn-primary" onClick={() => handleOpenSession(s.id)}>
                          {s.status === "active" ? "▶ Jogar" : "👁 Ver"}
                        </button>
                        {s.status === "finished" && (
                          <button type="button" className="score-btn-secondary" onClick={() => handleRestoreSession(s.id)}>↩ Restaurar</button>
                        )}
                        <button type="button" className="score-btn-danger" onClick={() => handleDeleteSession(s.id)}>🗑</button>
                      </div>
                    </div>
                  ))}
                </div>
                )}
              </section>
            </section>
          )}

          {/* Vista: Criar sessão */}
          {sessionView === "create" && (
            <div className="score-create-page">
              <button type="button" className="score-back-btn" onClick={() => setSessionView("list")}>← Voltar</button>
              <form className="card form" onSubmit={handleCreateSession}>
                <h2>Nova Sessão de Jogo</h2>
                <label>
                  Nome da sessão *
                  <input
                    required
                    placeholder="Commander sexta-feira"
                    value={sessionForm.name}
                    onChange={e => setSessionForm({ ...sessionForm, name: e.target.value })}
                  />
                </label>
                <div className="score-form-row">
                  <label>
                    Formato
                    <select
                      value={sessionForm.format}
                      onChange={e => {
                        const fmt = e.target.value;
                        const life = fmt === "Casual" ? 20 : 40;
                        setSessionForm({ ...sessionForm, format: fmt, starting_life: life });
                      }}
                    >
                      <option>Commander</option>
                      <option>Casual</option>
                      <option>Standard</option>
                      <option>Modern</option>
                      <option>Pioneer</option>
                      <option>Legacy</option>
                    </select>
                  </label>
                  <label>
                    Vida inicial
                    <input
                      type="number"
                      min={1}
                      value={sessionForm.starting_life}
                      onChange={e => setSessionForm({ ...sessionForm, starting_life: Number(e.target.value) })}
                    />
                  </label>
                </div>

                <div className="score-players-header">
                  <h3>Jogadores <span className="score-player-count">({sessionPlayers.length}/8)</span></h3>
                  {sessionPlayers.length < 8 && (
                    <button
                      type="button"
                      className="score-btn-secondary"
                      onClick={() => setSessionPlayers([...sessionPlayers, { ...EMPTY_PLAYER_ROW }])}
                    >
                      + Jogador
                    </button>
                  )}
                </div>

                {sessionPlayers.map((p, i) => (
                  <div key={i} className="score-player-row">
                    <input
                      placeholder="Nome do jogador"
                      value={p.name}
                      required
                      onChange={e => {
                        const ps = [...sessionPlayers];
                        ps[i] = { ...ps[i], name: e.target.value };
                        setSessionPlayers(ps);
                      }}
                    />
                    <input
                      className="score-short-input"
                      placeholder="Sigla"
                      maxLength={3}
                      value={p.short_code}
                      required
                      onChange={e => {
                        const ps = [...sessionPlayers];
                        ps[i] = { ...ps[i], short_code: e.target.value.toUpperCase() };
                        setSessionPlayers(ps);
                      }}
                    />
                    {sessionPlayers.length > 2 && (
                      <button
                        type="button"
                        className="score-remove-row"
                        onClick={() => setSessionPlayers(sessionPlayers.filter((_, j) => j !== i))}
                      >✕</button>
                    )}
                  </div>
                ))}

                {sessionError && <p className="score-error">{sessionError}</p>}
                <div className="score-form-actions">
                  <button type="button" className="score-btn-secondary" onClick={() => setSessionView("list")}>Cancelar</button>
                  <button type="submit" className="score-btn-primary" disabled={sessionLoading}>
                    {sessionLoading ? "Criando…" : "✓ Criar Sessão"}
                  </button>
                </div>
              </form>
            </div>
          )}

          {/* Vista: Jogar */}
          {sessionView === "play" && activeSession && (
            <div className="score-play-page">
              <div className="score-play-header">
                <button type="button" className="score-back-btn" onClick={() => { setSessionView("list"); setActiveSession(null); }}>
                  ← Sessões
                </button>
                <div className="score-play-meta">
                  <span className="score-play-name">{activeSession.name}</span>
                  <span className="score-format-tag">{activeSession.format}</span>
                  <span className={`score-badge ${activeSession.status}`}>
                    {activeSession.status === "active" ? "Ativo" : "Encerrado"}
                  </span>
                </div>
                <div className="score-play-actions">
                  {activeSession.status === "active" && (
                    <>
                      <button type="button" className="score-btn-secondary" onClick={handleResetSession}>↺ Reset</button>
                      <button type="button" className="score-btn-danger" onClick={handleFinishSession}>■ Encerrar</button>
                    </>
                  )}
                  {activeSession.status === "finished" && (
                    <button type="button" className="score-btn-secondary" onClick={() => handleRestoreSession(activeSession.id)}>↩ Restaurar</button>
                  )}
                </div>
              </div>

              {sessionError && <p className="score-error score-error-inline">{sessionError}</p>}

              {(() => {
                const PLAYER_COLORS = ['#3b82f6','#a855f7','#ec4899','#f59e0b','#10b981','#ef4444','#06b6d4','#84cc16'];
                const alivePlayers = activeSession.players.filter(p => !p.is_eliminated);
                const winnerId = activeSession.status === "active" && activeSession.players.length > 1 && alivePlayers.length === 1
                  ? alivePlayers[0].id : null;
                const isFinished = activeSession.status === "finished";
                return (
                  <div className="score-players-list">
                    {activeSession.players.map((player, idx) => {
                      const color = PLAYER_COLORS[idx % PLAYER_COLORS.length];
                      const isWinner = player.id === winnerId;
                      const isElim = player.is_eliminated;
                      return (
                        <div
                          key={player.id}
                          className={`score-prow${isElim ? " prow-elim" : ""}${isWinner ? " prow-winner" : ""}`}
                          style={{ '--pcolor': color }}
                        >
                          <div className="score-prow-head">
                            <div className="score-prow-badge">
                              {isElim ? '💀' : isWinner ? '👑' : player.short_code}
                            </div>
                            <div className="score-prow-nameblock">
                              <span className="score-prow-name">{player.name}</span>
                              {isElim && (
                                <span className="score-prow-elim-tag">
                                  {player.eliminated_reason === "life" && "☠ Sem vida"}
                                  {player.eliminated_reason === "commander_damage" && "⚔ Cmd damage"}
                                </span>
                              )}
                              {isWinner && <span className="score-prow-winner-tag">🏆 Vencedor!</span>}
                            </div>
                            {!isElim && (
                              <div className="score-life-header">
                                <span className="score-life-icon">❤️</span>
                                <span className={`score-life-big${player.life <= 0 ? " dead" : player.life <= 5 ? " low" : ""}`}>
                                  {player.life}
                                </span>
                              </div>
                            )}
                            {!isFinished && activeSession.players.length > 2 && (
                              <button type="button" className="score-remove-player" title="Remover" onClick={() => handleRemoveSessionPlayer(player.id)}>✕</button>
                            )}
                          </div>

                          {!isElim && (
                            <div className="score-controls">
                              <div className="score-life-row">
                                <button type="button" disabled={isFinished} onClick={() => handleUpdatePlayer(player.id, "life", -5)}>−5</button>
                                <button type="button" disabled={isFinished} onClick={() => handleUpdatePlayer(player.id, "life", -1)}>−1</button>
                                <button type="button" disabled={isFinished} onClick={() => handleUpdatePlayer(player.id, "life", +1)}>+1</button>
                                <button type="button" disabled={isFinished} onClick={() => handleUpdatePlayer(player.id, "life", +5)}>+5</button>
                              </div>
                              <div className="score-cmd-row">
                                <span className="score-cmd-label">⚔️ Cmd</span>
                                <button type="button" disabled={isFinished} onClick={() => handleUpdatePlayer(player.id, "commander_damage_received", -5)}>−5</button>
                                <button type="button" disabled={isFinished} onClick={() => handleUpdatePlayer(player.id, "commander_damage_received", -1)}>−1</button>
                                <span className={`score-cmd-val${player.commander_damage_received >= 21 ? " dead" : player.commander_damage_received >= 15 ? " low" : ""}`}>
                                  {player.commander_damage_received}<small>/21</small>
                                </span>
                                <button type="button" disabled={isFinished} onClick={() => handleUpdatePlayer(player.id, "commander_damage_received", +1)}>+1</button>
                                <button type="button" disabled={isFinished} onClick={() => handleUpdatePlayer(player.id, "commander_damage_received", +5)}>+5</button>
                              </div>
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </div>
                );
              })()}

              {/* Adicionar jogador */}
              {activeSession.status === "active" && (
                <div className="score-add-player-area">
                  {showAddPlayer ? (
                    <form className="score-add-player-form" onSubmit={handleAddSessionPlayer}>
                      <input
                        placeholder="Nome"
                        value={addPlayerForm.name}
                        required
                        onChange={e => setAddPlayerForm({ ...addPlayerForm, name: e.target.value })}
                      />
                      <input
                        className="score-short-input"
                        placeholder="Sigla"
                        maxLength={3}
                        value={addPlayerForm.short_code}
                        required
                        onChange={e => setAddPlayerForm({ ...addPlayerForm, short_code: e.target.value.toUpperCase() })}
                      />
                      <button type="submit" className="score-btn-primary">Adicionar</button>
                      <button type="button" className="score-btn-secondary" onClick={() => setShowAddPlayer(false)}>Cancelar</button>
                    </form>
                  ) : (
                    activeSession.players.length < 8 && (
                      <button type="button" className="score-btn-secondary" onClick={() => setShowAddPlayer(true)}>
                        + Adicionar Jogador
                      </button>
                    )
                  )}
                </div>
              )}
            </div>
          )}
        </>
      )}

      {/* ── MODAL BUSCA RÁPIDA ── */}
      {quickAddModal && (
        <div className="modal-overlay" onClick={() => setQuickAddModal(false)}>
          <div className="modal quick-add-modal" onClick={(e) => e.stopPropagation()}>
            <button className="modal-close" aria-label="Fechar" onClick={() => setQuickAddModal(false)}>✕</button>
            <h2>⚡ Busca Rápida</h2>
            <p className="quick-add-hint">Informe a sigla e o número da coleção para buscar a carta no Scryfall automaticamente.</p>
            <form className="quick-add-form" onSubmit={handleQuickAdd}>
              <div className="qa-row-2col">
                <label>
                  Sigla *
                  <input required autoFocus placeholder="KLD, BRO…"
                    value={quickAddForm.set_code}
                    style={{ textTransform: "uppercase" }}
                    onChange={(e) => setQuickAddForm({ ...quickAddForm, set_code: e.target.value.toUpperCase().trim() })} />
                </label>
                <label>
                  Número *
                  <input required placeholder="017, 253a"
                    value={quickAddForm.collection_number}
                    onChange={(e) => setQuickAddForm({ ...quickAddForm, collection_number: e.target.value.trim() })} />
                </label>
              </div>
              <label>
                Idioma
                <select value={quickAddForm.language} onChange={(e) => setQuickAddForm({ ...quickAddForm, language: e.target.value })}>
                  <option value="EN">Inglês</option>
                  <option value="PT">Português</option>
                  <option value="ES">Espanhol</option>
                  <option value="JP">Japonês</option>
                  <option value="FR">Francês</option>
                  <option value="DE">Alemão</option>
                </select>
              </label>
              <label>
                Quantidade
                <input type="number" min="1" value={quickAddForm.quantity}
                  onChange={(e) => setQuickAddForm({ ...quickAddForm, quantity: e.target.value })} />
              </label>
              <label>
                Deck
                <select value={quickAddForm.deck_id || 0} onChange={(e) => setQuickAddForm({ ...quickAddForm, deck_id: Number(e.target.value) })}>
                  <option value={0}>— Nenhum —</option>
                  {decks.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
                </select>
              </label>
              <label className="checkbox-label">
                <input type="checkbox" checked={quickAddForm.foil}
                  onChange={(e) => setQuickAddForm({ ...quickAddForm, foil: e.target.checked })} />
                ✦ Foil
              </label>
              <button type="submit" className="quick-add-submit">Buscar e Confirmar →</button>
            </form>
          </div>
        </div>
      )}

      {/* ── MODAL CONFIRMAR CADASTRO ── */}
      {(confirmCard || confirmLoading) && (
        <div className="modal-overlay" onClick={() => !confirmLoading && setConfirmCard(null)}>
          <div className="modal confirm-card-modal" onClick={(e) => e.stopPropagation()}>
            {confirmLoading ? (
              <div className="eval-loading">
                <div className="eval-spinner">⚙</div>
                <p className="eval-loading-text">Buscando na Scryfall…</p>
              </div>
            ) : (
              <>
                <button className="modal-close" aria-label="Fechar" onClick={() => setConfirmCard(null)}>✕</button>
                {confirmCard?.found && confirmCard.card ? (
                  <>
                    <div className="confirm-card-top">
                      <img
                        src={confirmCard.card.image_url}
                        alt={confirmCard.card.name}
                        className="confirm-card-img"
                      />
                      <div className="confirm-card-info">
                        <h2 className="confirm-card-title">{confirmCard.card.printed_name || confirmCard.card.name}</h2>
                        {confirmCard.card.printed_name && <p className="confirm-card-en">{confirmCard.card.name}</p>}
                        <div className="confirm-card-grid">
                          <div><span>Tipo</span><strong>{confirmCard.card.type || "—"}</strong></div>
                          <div><span>Custo</span><strong>{confirmCard.card.mana_cost || "—"}</strong></div>
                          <div><span>Set</span><strong>{confirmCard.card.set || "—"} #{confirmCard.card.number || "—"}</strong></div>
                          <div><span>Raridade</span><strong className={`rarity r-${(confirmCard.card.rarity || "x")[0].toLowerCase()}`}>{confirmCard.card.rarity || "—"}</strong></div>
                          <div><span>Artista</span><strong>{confirmCard.card.artist || "—"}</strong></div>
                          {confirmCard.card.prices?.usd && <div><span>Preço USD</span><strong className="price-tag">${confirmCard.card.prices.usd}</strong></div>}
                        </div>
                        {(confirmCard.card.printed_text || confirmCard.card.text) && (
                          <p className="confirm-card-text">{confirmCard.card.printed_text || confirmCard.card.text}</p>
                        )}
                      </div>
                    </div>
                    <p className="confirm-card-question">Esta é a carta que você quer cadastrar?</p>
                    <div className="confirm-card-actions">
                      <button type="button" className="confirm-yes" onClick={handleConfirmCreate}>✓ Sim, cadastrar esta carta</button>
                      <button type="button" className="confirm-no" onClick={() => { if (confirmCard?.origin === "quickAdd") setQuickAddModal(true); setConfirmCard(null); }}>← Voltar e editar</button>
                    </div>
                  </>
                ) : (
                  <>
                    <div className="confirm-not-found">
                      <span className="confirm-not-found-icon">🔍</span>
                      <h3>Carta não encontrada na Scryfall</h3>
                      <p>Não foi possível encontrar esta carta automaticamente. Você pode cadastrá-la assim mesmo ou voltar e verificar os dados.</p>
                    </div>
                    <div className="confirm-card-actions">
                      <button type="button" className="confirm-yes" onClick={handleConfirmCreate}>✓ Cadastrar assim mesmo</button>
                      <button type="button" className="confirm-no" onClick={() => { if (confirmCard?.origin === "quickAdd") setQuickAddModal(true); setConfirmCard(null); }}>← Voltar e editar</button>
                    </div>
                  </>
                )}
              </>
            )}
          </div>
        </div>
      )}

      {/* ── MODAL ANOMALIAS DO DECK ── */}
      {deckAnomalyModal && (
        <div className="modal-overlay" onClick={() => setDeckAnomalyModal(false)}>
          <div className="modal anomaly-modal" onClick={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => setDeckAnomalyModal(false)}>✕</button>
            <h2 className="anomaly-modal-title">⚠️ Anomalias do deck</h2>
            <p className="anomaly-modal-sub">{managingDeck?.name}</p>
            <div className="anomaly-list">
              {deckAnomalies.map((a, i) => (
                <div key={i} className="anomaly-item">
                  <span className="anomaly-icon">{a.icon}</span>
                  <div className="anomaly-text">
                    <strong>{a.label}</strong>
                    <span>{a.detail}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* ── MODAL DECK BUILDER IA ── */}
      {deckBuilderModal && (
        <div className="modal-overlay" onClick={() => !deckBuilderLoading && !deckBuilderApproving && setDeckBuilderModal(false)}>
          <div className="modal deck-builder-modal" onClick={(e) => e.stopPropagation()}>
            <div className="deck-builder-modal-header">
              <h2>✨ Montar Deck com IA</h2>
              {!deckBuilderLoading && !deckBuilderApproving && (
                <button className="modal-close" aria-label="Fechar" onClick={() => setDeckBuilderModal(false)}>✕</button>
              )}
            </div>

            {/* ── STEP: config ── */}
            {deckBuilderStep === "config" && (
              <div className="db-config">
                <p className="deck-builder-meta">{total} carta{total !== 1 ? "s" : ""} sem deck serão analisadas</p>

                <div className="db-config-section">
                  <span className="db-config-label">Formato</span>
                  <div className="db-option-row">
                    {[
                      { id: "auto",       icon: "🤖", label: "Auto",       desc: "IA decide o melhor" },
                      { id: "casual60",   icon: "📄", label: "60 cartas",  desc: "Casual / Standard / Modern" },
                      { id: "commander",  icon: "👑", label: "Commander",  desc: "100 cartas singleton" },
                    ].map(f => (
                      <button key={f.id} type="button"
                        className={`db-option-btn${deckBuilderConfig.format === f.id ? " active" : ""}`}
                        onClick={() => setDeckBuilderConfig({ ...deckBuilderConfig, format: f.id })}>
                        <span className="db-option-icon">{f.icon}</span>
                        <strong>{f.label}</strong>
                        <small>{f.desc}</small>
                      </button>
                    ))}
                  </div>
                </div>

                <div className="db-config-section">
                  <span className="db-config-label">Objetivo</span>
                  <div className="db-option-row">
                    {[
                      { id: "fun",          icon: "🎉", label: "Diversão",    desc: "Sinergias e combos divertidos" },
                      { id: "competitive",  icon: "⚔",  label: "Competitivo", desc: "Máxima eficiência" },
                    ].map(g => (
                      <button key={g.id} type="button"
                        className={`db-option-btn${deckBuilderConfig.goal === g.id ? " active" : ""}`}
                        onClick={() => setDeckBuilderConfig({ ...deckBuilderConfig, goal: g.id })}>
                        <span className="db-option-icon">{g.icon}</span>
                        <strong>{g.label}</strong>
                        <small>{g.desc}</small>
                      </button>
                    ))}
                  </div>
                </div>

                <div className="db-config-section">
                  <span className="db-config-label">Cores preferidas <em>(opcional — deixe vazio para a IA escolher)</em></span>
                  <ManaColorPicker
                    value={deckBuilderConfig.colors}
                    onChange={(v) => setDeckBuilderConfig({ ...deckBuilderConfig, colors: v })}
                  />
                </div>

                <button type="button" className="db-run-btn" onClick={handleRunDeckBuilder}>
                  ✨ Analisar e Montar Deck →
                </button>
              </div>
            )}

            {/* ── STEP: loading ── */}
            {deckBuilderStep === "loading" && (
              <div className="eval-loading">
                <div className="eval-spinner">⚙</div>
                <p className="eval-loading-text">Montando o deck com IA…</p>
                <p className="eval-loading-sub">Analisando sinergias, curva de mana e estratégia. Pode levar alguns segundos.</p>
              </div>
            )}

            {/* ── STEP: result ── */}
            {deckBuilderStep === "result" && deckBuilderResult && (
              (deckBuilderResult.error || deckBuilderResult.error_ia) ? (
                <>
                  <div className="eval-empty">
                    <div className="eval-empty-icon">{deckBuilderResult.error_ia ? "🚫" : "⚠️"}</div>
                    <p>{deckBuilderResult.error_ia
                      ? `Não foi possível montar o deck: ${deckBuilderResult.error_ia}`
                      : deckBuilderResult.error}</p>
                  </div>
                  <div className="db-actions">
                    <button type="button" className="db-revaluate-btn" onClick={handleRevaluateDeck}>♻ Tentar com outros parâmetros</button>
                    <button type="button" className="db-reject-btn" onClick={() => setDeckBuilderModal(false)}>✕ Fechar</button>
                  </div>
                </>
              ) : (
                <>
                  {deckBuilderResult.deck_name && (
                    <div className="db-result-header">
                      <div className="db-deck-name-block">
                        <span className="db-deck-name-label">Deck sugerido</span>
                        <strong className="db-deck-name-value">{deckBuilderResult.deck_name}</strong>
                        {deckBuilderResult.deck_description && (
                          <p className="db-deck-desc">{deckBuilderResult.deck_description}</p>
                        )}
                      </div>
                      <div className="db-deck-meta">
                        {deckBuilderResult.deck_colors && (
                          <span className="db-deck-colors">
                            {deckBuilderResult.deck_colors.split(",").filter(Boolean).map(c => (
                              <img key={c} src={`/mana-icons/${({W:"white",U:"blue",B:"black",R:"red",G:"green",C:"incolour"}[c]||"incolour")}.svg`}
                                className="mana-icon" alt={c} style={{width:18,height:18}} />
                            ))}
                          </span>
                        )}
                        {deckBuilderResult.deck_commander && <span className="commander-badge">CMD</span>}
                        <span className="db-card-count">{deckBuilderResult.card_count} cartas analisadas</span>
                      </div>
                    </div>
                  )}

                  <div className="modal-divider" />

                  <div className="eval-content deck-builder-content">
                    {renderEvalMarkdown(deckBuilderResult.analysis)}
                  </div>

                  {deckBuilderResult.card_roles && (
                    <>
                      <div className="modal-divider" />
                      <div className="db-card-roles">
                        <h3 className="db-card-roles-title">Cartas no Deck</h3>
                        <div className="db-card-roles-list">
                          {(deckBuilderResult.card_roles.nao_terrenos || []).map((card) => (
                            <div key={card.nome} className="db-card-role-item">
                              <span className="db-card-role-name">{card.nome}</span>
                              <span className="db-card-role-papel">{card.papel}</span>
                            </div>
                          ))}
                        </div>
                        {deckBuilderResult.card_roles.terrenos?.total > 0 && (
                          <div className="db-terrenos-info">
                            <span className="db-terrenos-label">
                              Terrenos ({deckBuilderResult.card_roles.terrenos.total})
                            </span>
                            <span className="db-card-role-papel">{deckBuilderResult.card_roles.terrenos.motivo}</span>
                          </div>
                        )}
                      </div>
                    </>
                  )}

                  <div className="db-actions">
                    {deckBuilderResult.deck_list && (
                      <button type="button" className="db-approve-btn"
                        onClick={handleApproveDeck} disabled={deckBuilderApproving}>
                        {deckBuilderApproving ? "⏳ Criando deck…" : "✓ Aprovar e criar deck"}
                      </button>
                    )}
                    <button type="button" className="db-revaluate-btn" onClick={handleRevaluateDeck} disabled={deckBuilderApproving}>
                      ♻ Re-avaliar
                    </button>
                    <button type="button" className="db-reject-btn" onClick={() => setDeckBuilderModal(false)} disabled={deckBuilderApproving}>
                      ✕ Recusar
                    </button>
                  </div>
                </>
              )
            )}
          </div>
        </div>
      )}

      {/* ── MODAL EDITAR DECK ── */}
      {editDeckModal && (
        <div className="modal-overlay" onClick={() => setEditDeckModal(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <button className="modal-close" aria-label="Fechar" onClick={() => setEditDeckModal(null)}>✕</button>
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
            <button className="modal-close" aria-label="Fechar" onClick={() => { if (!listLoading) setListModal(false); }}>✕</button>
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
            <button className="modal-close" aria-label="Fechar" onClick={() => { if (!importLoading) setImportModal(false); }}>✕</button>
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

      {activeTab === "collection" && <section className="card list-section collection-full">
        {/* ── LISTA ── */}
          <div className="list-header">
            <div className="list-header-top">
              <h2>Minha coleção <span className="total-badge">{totalQuantity} cartas</span><span className="unique-badge">{total} únicas</span></h2>
              <div className="toolbar-group">
                <button type="button" className="quick-add-btn-toolbar" onClick={() => setQuickAddModal(true)}>
                  ⚡ Busca Rápida
                </button>
                <button type="button" className="tokens-tab-btn" onClick={() => setActiveTab("tokens")}>
                  🎭 Tokens
                </button>
                <button type="button" className={`stats-toggle-btn${statsOpen ? " active" : ""}`} onClick={handleOpenStats}>
                  📊 Stats
                </button>
                <DropdownMenu
                  label={priceRefreshing ? "⏳…" : "🔄 Atualizar"}
                  open={openMenu === "update"}
                  onToggle={() => setOpenMenu(openMenu === "update" ? null : "update")}
                  items={[
                    { label: "💰 Todos os preços", onClick: handleRefreshPrices, disabled: priceRefreshing },
                    { label: "💰 Só faltantes", onClick: handleRefreshMissingPrices, disabled: priceRefreshing },
                    { separator: true },
                    { label: "🖼 Imagens", onClick: handleRefreshImages, disabled: priceRefreshing },
                  ]}
                />
                <DropdownMenu
                  label={viewMode === "grid" ? "⊞ Exibição" : "☰ Exibição"}
                  open={openMenu === "view"}
                  onToggle={() => setOpenMenu(openMenu === "view" ? null : "view")}
                  items={[
                    { label: "☰ Lista", active: viewMode === "list", onClick: () => { setViewMode("list"); localStorage.setItem("card-view-mode", "list"); } },
                    { label: "⊞ Grid", active: viewMode === "grid", onClick: () => { setViewMode("grid"); localStorage.setItem("card-view-mode", "grid"); } },
                  ]}
                />
                <DropdownMenu
                  label="↓ Exportar"
                  open={openMenu === "export"}
                  onToggle={() => setOpenMenu(openMenu === "export" ? null : "export")}
                  items={[
                    { label: "↓ CSV", onClick: handleExportCSV },
                    { label: "↓ XLSX", onClick: handleExportXLSX },
                  ]}
                />
              </div>
            </div>
            {priceRefreshResult && (
              <div className={`price-refresh-banner${priceRefreshResult.error ? " error" : ""}`}>
                {priceRefreshResult.error
                  ? `⚠ Erro: ${priceRefreshResult.error}`
                  : priceRefreshResult._type === "images"
                    ? `✓ ${priceRefreshResult.updated} imagens atualizadas · ${priceRefreshResult.skipped} sem dados · total: ${priceRefreshResult.total}`
                    : priceRefreshResult._type === "missing"
                    ? `✓ ${priceRefreshResult.updated} preços faltantes preenchidos · ${priceRefreshResult.skipped} sem dados na Scryfall · total: ${priceRefreshResult.total}`
                    : `✓ ${priceRefreshResult.updated} preços atualizados · ${priceRefreshResult.skipped} sem dados · total: ${priceRefreshResult.total}`
                }
                <button type="button" className="banner-close" aria-label="Fechar aviso" onClick={() => setPriceRefreshResult(null)}>✕</button>
              </div>
            )}
            {statsOpen && <StatsPanel stats={stats} loading={statsLoading} />}
            <input
              className="search-input"
              type="search"
              placeholder="Buscar por nome, coleção, nº, cor, tipo..."
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
              <button
                type="button"
                className={`filter-select filter-fullart-btn${filterFullArt ? " active" : ""}`}
                onClick={() => { setFilterFullArt(v => !v); setPage(1); }}
              >
                ◈ Full Art
              </button>
              <select className="filter-select" value={filterRarity} onChange={(e) => { setFilterRarity(e.target.value); setPage(1); }}>
                <option value="">Todas as raridades</option>
                <option value="L">Land (L)</option>
                <option value="C">Common (C)</option>
                <option value="U">Uncommon (U)</option>
                <option value="R">Rare (R)</option>
                <option value="M">Mythic (M)</option>
                <option value="T">Token (T)</option>
              </select>
              <select className="filter-select" value={filterType} onChange={(e) => { setFilterType(e.target.value); setPage(1); }}>
                <option value="">Todos os tipos</option>
                <option value="Creature">Criatura</option>
                <option value="Instant">Instantâneo</option>
                <option value="Sorcery">Feitiço</option>
                <option value="Enchantment">Encantamento</option>
                <option value="Artifact">Artefato</option>
                <option value="Land">Terreno</option>
                <option value="Planeswalker">Planeswalker</option>
                <option value="Battle">Batalha</option>
                <option value="Legendary">Lendário</option>
              </select>
            </div>
            {availableColors.length > 0 && (
              <div className="color-filter-bar">
                <button
                  type="button"
                  className={`color-filter-chip${filterColors === "" ? " active" : ""}`}
                  onClick={() => { setFilterColors(""); setPage(1); }}
                >
                  Todas
                </button>
                {availableColors.map((combo) => {
                  const isActive = filterColors === combo.codes;
                  const isNone = combo.codes === "none";
                  const codes = isNone ? [] : combo.codes.split(",").filter(Boolean);
                  return (
                    <button
                      key={combo.codes}
                      type="button"
                      className={`color-filter-chip${isActive ? " active" : ""}${isNone ? " no-color" : ""}`}
                      title={`${isNone ? "Sem cor" : combo.codes} — ${combo.count} carta${combo.count !== 1 ? "s" : ""}`}
                      aria-label={`${isNone ? "Sem cor" : `Cor ${combo.codes}`} — ${combo.count} carta${combo.count !== 1 ? "s" : ""}`}
                      aria-pressed={isActive}
                      onClick={() => { setFilterColors(isActive ? "" : combo.codes); setPage(1); }}
                    >
                      {isNone
                        ? <span className="color-filter-none-icon">?</span>
                        : codes.map((c) => (
                            <img key={c} src={`/mana-icons/${MTG_CODE_TO_ICON[c] || "incolour"}.svg`}
                              className="mana-icon" alt={c} />
                          ))
                      }
                      <span className="color-filter-count">{combo.count}</span>
                    </button>
                  );
                })}
              </div>
            )}
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

          {viewMode === "grid" ? (
            <div className="cards-grid">
              {cards.map((card) => (
                <div
                  key={card.id}
                  className={`card-grid-item item-r-${(card.rarity || "x").toLowerCase()}${card.foil ? " is-foil" : ""}${card.full_art ? " is-full-art" : ""}`}
                  onClick={() => handleDetails(card.id)}
                  title={card.name}
                >
                  {card.image_url
                    ? <img src={card.image_url} alt={card.name} loading="lazy" className="card-grid-img" />
                    : <div className="card-grid-placeholder">
                        <CardColorIcons card={card} />
                        <span className="card-grid-placeholder-name">{card.name}</span>
                      </div>
                  }
                  <div className="card-grid-overlay">
                    <div className="card-grid-name">
                      {card.foil && <span className="foil-text">✦ </span>}
                      {card.name}
                    </div>
                    <div className="card-grid-meta">
                      {card.rarity && <span className={`rarity r-${card.rarity.toLowerCase()}`}>{card.rarity}</span>}
                      {card.price_usd > 0 && <span className="price-tag">${card.price_usd.toFixed(2)}</span>}
                    </div>
                  </div>
                  <div className="card-grid-qty" onClick={(e) => e.stopPropagation()}>
                    <button type="button" className="qty-btn" aria-label="Reduzir quantidade" onClick={() => handleQuantityChange(card.id, -1)} disabled={card.quantity <= 1}>−</button>
                    <span className="qty-display">{card.quantity}</span>
                    <button type="button" className="qty-btn" aria-label="Aumentar quantidade" onClick={() => handleQuantityChange(card.id, +1)}>+</button>
                  </div>
                </div>
              ))}
              {cards.length === 0 && <p className="empty">Nenhuma carta encontrada.</p>}
            </div>
          ) : (
            <div className="list">
              {cards.map((card) => {
                const assignedDeck = card.deck_id > 0 ? decks.find((d) => d.id === card.deck_id) : null;
                return (
                  <div
                    className={`list-item${card.foil ? " is-foil" : ""}${card.full_art ? " is-full-art" : ""} item-r-${(card.rarity || "x").toLowerCase()}`}
                    key={card.id}
                  >
                    <div className="list-item-info">
                      <div className="list-item-name">
                        <strong className={card.foil ? "foil-text" : ""}>
                          {card.name}
                        </strong>
                        {card.foil && <span className="foil-text">✦</span>}
                        {card.full_art && <span className="full-art-badge">◈ Full Art</span>}
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
                      <div className="list-item-meta">
                        <span>{card.set_code || "—"} · #{card.collection_number || "—"} · {card.language || "—"} · ×{card.quantity}</span>
                        {card.price_usd > 0 && (
                          <span className="price-tag">${card.price_usd.toFixed(2)}</span>
                        )}
                      </div>
                      <small>{card.type || "—"}{card.subtitle ? ` — ${card.subtitle}` : ""}</small>
                    </div>
                    <div className="actions">
                      <div className="qty-ctrl">
                        <button type="button" className="qty-btn" aria-label="Reduzir quantidade" onClick={() => handleQuantityChange(card.id, -1)} disabled={card.quantity <= 1}>−</button>
                        <span className="qty-display">{card.quantity}</span>
                        <button type="button" className="qty-btn" aria-label="Aumentar quantidade" onClick={() => handleQuantityChange(card.id, +1)}>+</button>
                      </div>
                      <button type="button" onClick={() => handleDetails(card.id)}>Ver</button>
                      <button type="button" className="danger" aria-label="Remover carta" onClick={() => handleDelete(card.id)}>✕</button>
                    </div>
                  </div>
                );
              })}
              {cards.length === 0 && <p className="empty">Nenhuma carta encontrada.</p>}
            </div>
          )}

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
      }

      {/* ── MODAL DETALHES ── */}
      {(selectedCard || loadingDetail) && (
        <div className="modal-overlay" onClick={() => { setSelectedCard(null); setEditMode(false); setDetailFromDeck(false); }}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <button className="modal-close" aria-label="Fechar" onClick={() => { setSelectedCard(null); setEditMode(false); setDetailFromDeck(false); }}>✕</button>

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
                  <div><span>Custo</span>{selectedCard.external?.mana_cost || selectedCard.local.mana_cost || "—"}</div>
                  <div><span>Quantidade</span>{selectedCard.local.quantity}</div>
                  <div><span>Condição</span>{selectedCard.local.condition || "—"}</div>
                  <div><span>Foil</span>{selectedCard.local.foil ? "Sim" : "Não"}</div>
                  <div><span>Full Art</span>{selectedCard.local.full_art ? <span className="full-art-badge">◈ Sim</span> : "Não"}</div>
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
                  {detailFromDeck ? (
                    <button className="danger" onClick={async () => {
                      await handleUnassignCard(selectedCard.local.id);
                      setSelectedCard(null);
                      setDetailFromDeck(false);
                    }}>Remover do deck</button>
                  ) : (
                    <>
                      <button type="button" className="edit-btn" onClick={handleEditStart}>Editar</button>
                      <button className="danger" onClick={() => handleDelete(selectedCard.local.id)}>Remover carta</button>
                    </>
                  )}
                </div>
              </>
            )}

            {selectedCard && editMode && (
              <form className="edit-form" onSubmit={(e) => { e.preventDefault(); handleEditSave(); }}>
                <h2>Editar carta</h2>
                <div className="edit-grid">
                  <label className="eg-full">Nome *<input required value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })} /></label>
                  <label>Tipo<input value={editForm.type} onChange={(e) => setEditForm({ ...editForm, type: e.target.value })} /></label>
                  <label>Subtítulo<input value={editForm.subtitle} onChange={(e) => setEditForm({ ...editForm, subtitle: e.target.value })} /></label>
                  <label>Nº<input value={editForm.collection_number} onChange={(e) => setEditForm({ ...editForm, collection_number: e.target.value })} /></label>
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
                  <label>Sigla<input value={editForm.set_code} onChange={(e) => setEditForm({ ...editForm, set_code: e.target.value })} /></label>
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
                  <label>Ano<input type="number" placeholder="Ex: 2016" value={editForm.year || ""} onChange={(e) => setEditForm({ ...editForm, year: e.target.value })} /></label>
                  <label>Artista<input value={editForm.artist} onChange={(e) => setEditForm({ ...editForm, artist: e.target.value })} /></label>
                  <label>Condição
                    <select value={editForm.condition} onChange={(e) => setEditForm({ ...editForm, condition: e.target.value })}>
                      <option value="mint">Mint</option>
                      <option value="near_mint">Near Mint</option>
                      <option value="played">Played</option>
                      <option value="damaged">Damaged</option>
                    </select>
                  </label>
                  <label>Qtd<input type="number" min="1" value={editForm.quantity} onChange={(e) => setEditForm({ ...editForm, quantity: e.target.value })} /></label>
                  <label>Deck
                    <select value={editForm.deck_id} onChange={(e) => setEditForm({ ...editForm, deck_id: Number(e.target.value) })}>
                      <option value={0}>— Nenhum —</option>
                      {decks.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
                    </select>
                  </label>
                  <label className="checkbox-label"><input type="checkbox" checked={editForm.foil} onChange={(e) => setEditForm({ ...editForm, foil: e.target.checked })} />Foil</label>
                  <label className="checkbox-label"><input type="checkbox" checked={editForm.full_art || false} onChange={(e) => setEditForm({ ...editForm, full_art: e.target.checked })} />Full Art</label>
                  <label className="checkbox-label"><input type="checkbox" checked={editForm.prerelease} onChange={(e) => setEditForm({ ...editForm, prerelease: e.target.checked })} />Pré-release</label>
                  <label className="checkbox-label"><input type="checkbox" checked={editForm.commander} onChange={(e) => setEditForm({ ...editForm, commander: e.target.checked })} />Commander</label>
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
      {/* ── FAB MOBILE ── */}
      {activeTab === "collection" && (
        <button type="button" className="mobile-fab" onClick={() => setQuickAddModal(true)} aria-label="Adicionar carta" title="Busca Rápida">+</button>
      )}
      {activeTab === "tokens" && (
        <button type="button" className="mobile-fab mobile-fab-token" onClick={() => { setTokenAddModal(true); setTokenPreviewFront(null); setTokenPreviewBack(null); setTokenForm(EMPTY_TOKEN_FORM); setTokenSearchError(""); }} aria-label="Adicionar token" title="Adicionar Token">+</button>
      )}
      </main>
    </>
  );
}
