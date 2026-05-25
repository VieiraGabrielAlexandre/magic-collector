import { useEffect, useRef, useState } from "react";
import { createCard, deleteCard, getCard, listCards } from "./services/api";
import "./App.css";

const EMPTY_FORM = {
  name: "", color: "", type: "", subtitle: "", collection_number: "",
  rarity: "", set_code: "", language: "PT", year: "", artist: "",
  company: "Wizards of the Coast", foil: false, quantity: 1,
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

export default function App() {
  const [form, setForm] = useState(EMPTY_FORM);

  const [cards, setCards] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  const [search, setSearch] = useState("");
  const [sort, setSort] = useState("name");
  const [order, setOrder] = useState("asc");

  const [selectedCard, setSelectedCard] = useState(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

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
    setPage(result.page ?? 1);
    setTotalPages(result.total_pages ?? 1);
  }

  useEffect(() => { loadCards(); }, [sort, order]);

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

  const field = (label, key, extra = {}) => (
    <label>
      {label}
      <input value={form[key]} onChange={(e) => setForm({ ...form, [key]: e.target.value })} {...extra} />
    </label>
  );

  const orderIcon = (field) => {
    if (sort !== field) return null;
    return order === "asc" ? " ↑" : " ↓";
  };

  return (
    <main className="app">
      <section className="hero">
        <h1>Magic Collection</h1>
        <p>Cadastre, organize e consulte sua coleção de cartas Magic.</p>
      </section>

      <section className="grid">
        {/* ── FORMULÁRIO ── */}
        <form className="card form" onSubmit={handleSubmit}>
          <h2>Cadastrar carta</h2>
          {field("Nome *", "name", { placeholder: "Ex: Lightning Bolt", required: true })}
          {field("Cor", "color", { placeholder: "Ex: Vermelha, Azul, Incolor" })}
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
            <h2>Minha coleção <span className="total-badge">{total}</span></h2>
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
              <div className="list-item" key={card.id}>
                <div className="list-item-info">
                  <strong>{card.name}{card.foil ? " ✦" : ""}</strong>
                  <span>{card.set_code || "—"} #{card.collection_number || "—"} · {card.rarity || "—"} · {card.language || "—"}</span>
                  <small>{card.color || "—"} · {card.type || "—"}{card.subtitle ? ` — ${card.subtitle}` : ""} · Qtd: {card.quantity}</small>
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
      </section>

      {/* ── MODAL DETALHES ── */}
      {(selectedCard || loadingDetail) && (
        <div className="modal-overlay" onClick={() => setSelectedCard(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => setSelectedCard(null)}>✕</button>

            {loadingDetail && <p className="empty">Carregando...</p>}

            {selectedCard && (
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

                <div className="modal-grid">
                  <div><span>Cor</span>{selectedCard.local.color || "—"}</div>
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

                <button className="danger full-width" onClick={() => handleDelete(selectedCard.local.id)}>
                  Remover carta
                </button>
              </>
            )}
          </div>
        </div>
      )}
    </main>
  );
}
