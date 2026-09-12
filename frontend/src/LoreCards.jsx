import { useState, useEffect, useCallback } from "react";
import { listCardShowcases, getCardShowcase } from "./services/api.js";

// ── Parse the card showcase .md format ────────────────────────────
function parseCardFile(content) {
  const rawTitle = content.match(/^#\s+(.+)/m);
  const title = rawTitle ? rawTitle[1].trim() : "";

  const blocks = content.split(/\n?<-\.\.\.->\n?/);
  const cards = [];

  for (const block of blocks) {
    const lines = block.trim().split("\n");
    const fields = {};
    let lastKey = null;

    for (const line of lines) {
      if (line.startsWith("#")) continue;
      const colon = line.indexOf(":");
      if (colon > 0 && colon <= 25 && !line.startsWith(" ")) {
        const rawKey = line.slice(0, colon).trim().toLowerCase()
          .normalize("NFD").replace(/[̀-ͯ]/g, "")
          .replace(/[^a-z0-9\s_]/g, "").trim()
          .replace(/\s+/g, "_");
        const val = line.slice(colon + 1).trim();
        fields[rawKey] = val;
        lastKey = rawKey;
      } else if (lastKey && line.trim()) {
        fields[lastKey] += " " + line.trim();
      }
    }

    const name = fields["card_name"];
    if (name) cards.push({ ...fields, _name: name });
  }

  return { title, cards };
}

// ── Mana symbols ──────────────────────────────────────────────────
const MANA_COLORS = {
  W: { bg: "#f0e8c8", fg: "#4a3820" },
  U: { bg: "#3a6ea8", fg: "#ffffff" },
  B: { bg: "#1a1020", fg: "#c0a0d0" },
  R: { bg: "#c83820", fg: "#ffffff" },
  G: { bg: "#2a7038", fg: "#ffffff" },
  C: { bg: "#b0a898", fg: "#2a2018" },
  X: { bg: "#686868", fg: "#ffffff" },
};

function ManaSymbol({ sym }) {
  const upper = sym.toUpperCase();
  const style = MANA_COLORS[upper] || { bg: "#686868", fg: "#fff" };
  return (
    <span className="lore-card-mana-sym" style={{ background: style.bg, color: style.fg }}>
      {upper}
    </span>
  );
}

function ManaCost({ cost }) {
  if (!cost) return null;
  const syms = [...cost.matchAll(/\{([^}]+)\}/g)].map(m => m[1]);
  if (!syms.length) return <span className="lore-card-mana-text">{cost}</span>;
  return (
    <span className="lore-card-mana-cost">
      {syms.map((s, i) => <ManaSymbol key={i} sym={s} />)}
    </span>
  );
}

// ── Card thumbnail ────────────────────────────────────────────────
function CardItem({ card, image, onClick }) {
  return (
    <button className="lore-card-item" onClick={onClick} title={card._name}>
      {image
        ? <img src={image} alt={card._name} className="lore-card-img" loading="lazy" />
        : (
          <div className="lore-card-img-placeholder">
            <span>✦</span>
            <span className="lore-card-ph-name">{card._name}</span>
          </div>
        )}
      <div className="lore-card-hover-hint">Detalhes</div>
    </button>
  );
}

// ── Card detail modal ─────────────────────────────────────────────
function CardModal({ card, image, onClose }) {
  // Close on Escape
  useEffect(() => {
    function onKey(e) { if (e.key === "Escape") onClose(); }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  return (
    <div className="lore-modal-overlay" onClick={onClose}>
      <div className="lore-modal-box" onClick={e => e.stopPropagation()}>
        <button className="lore-modal-close" onClick={onClose}>✕</button>

        <div className="lore-modal-inner">
          {/* Card image */}
          <div className="lore-modal-art">
            {image
              ? <img src={image} alt={card._name} className="lore-modal-img" />
              : (
                <div className="lore-card-img-placeholder lore-modal-placeholder">
                  <span>✦</span>
                  <span className="lore-card-ph-name">{card._name}</span>
                </div>
              )}
          </div>

          {/* Details */}
          <div className="lore-modal-info">
            <div className="lore-modal-name-row">
              <div className="lore-modal-name">{card._name}</div>
            </div>

            {card.tipo && <div className="lore-modal-type">{card.tipo}</div>}
            {card.set  && <div className="lore-modal-set">— {card.set} —</div>}

            {card.descricao && (
              <p className="lore-modal-desc">{card.descricao}</p>
            )}

            {card.curiosidade && (
              <div className="lore-modal-fact">
                <span className="lore-modal-fact-label">✦ Curiosidade</span>
                <p>{card.curiosidade}</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

// ── Main card showcase view ────────────────────────────────────────
export default function LoreCards({ onBack, onClose }) {
  const [showcases, setShowcases]       = useState([]);
  const [loading, setLoading]           = useState(true);
  const [selected, setSelected]         = useState(null);
  const [cardImages, setCardImages]     = useState({});
  const [loadingShowcase, setLoadingShowcase] = useState(false);
  const [activeCard, setActiveCard]     = useState(null);

  useEffect(() => {
    listCardShowcases()
      .then(setShowcases)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (!selected?.cards?.length) return;
    setCardImages({});
    const names = selected.cards.map(c => c._name).filter(Boolean);
    if (!names.length) return;

    fetch("https://api.scryfall.com/cards/collection", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ identifiers: names.map(n => ({ name: n })) }),
    })
      .then(r => r.json())
      .then(result => {
        const imgs = {};
        (result.data || []).forEach(card => {
          const url = card.image_uris?.normal || card.card_faces?.[0]?.image_uris?.normal;
          if (url) imgs[card.name.toLowerCase()] = url;
        });
        setCardImages(imgs);
      })
      .catch(console.error);
  }, [selected]);

  async function openShowcase(slug) {
    setLoadingShowcase(true);
    try {
      const raw = await getCardShowcase(slug);
      const { title, cards } = parseCardFile(raw.content || "");
      setSelected({ title: raw.title || title, cards });
    } catch (e) {
      console.error(e);
    } finally {
      setLoadingShowcase(false);
    }
  }

  const getImage = useCallback((card) => (
    cardImages[card._name.toLowerCase()] || null
  ), [cardImages]);

  const closeModal = useCallback(() => setActiveCard(null), []);

  return (
    <div className="lore-cards-view">
      <div className="lore-cards-topbar">
        <button className="lore-cards-back-btn" onClick={selected ? () => setSelected(null) : onBack}>
          ← {selected ? "Vitrine" : "Lore"}
        </button>
        {onClose && (
          <button className="lore-close" style={{ position: "relative", top: "auto", right: "auto" }}
            onClick={onClose}>
            ← App
          </button>
        )}
      </div>

      {loading ? (
        <div className="lore-cards-loading">Carregando vitrines…</div>
      ) : !selected ? (
        <div className="lore-cards-main">
          <p className="lore-chars-eyebrow">Vitrines de Cartas</p>
          <h2 className="lore-section-title">Cartas em Destaque</h2>
          <p className="lore-section-intro">
            Selecione uma vitrine para explorar cartas icônicas, curiosidades e a arte oficial do Scryfall.
          </p>

          {loadingShowcase ? (
            <div className="lore-cards-loading">Carregando vitrine…</div>
          ) : (
            <div className="lore-showcase-list">
              {showcases.map(sc => (
                <button key={sc.slug} className="lore-showcase-entry"
                  onClick={() => openShowcase(sc.slug)}>
                  <span className="lore-showcase-icon">✦</span>
                  <span className="lore-showcase-title">{sc.title}</span>
                  <span className="lore-showcase-arrow">→</span>
                </button>
              ))}
              {!showcases.length && (
                <p style={{ color: "#4a4038", fontFamily: "'Cinzel', serif", fontSize: 11, letterSpacing: "0.15em" }}>
                  Nenhuma vitrine encontrada.
                </p>
              )}
            </div>
          )}

          <details className="lore-cards-format-hint">
            <summary>Como criar uma vitrine</summary>
            <pre>{`# Título da Vitrine\n\ncard name: Black Lotus\nmanas: {0}\ntipo: Artifact\nset: Alpha (1993)\ndescricao: Descrição aqui.\ncuriosidade: Fato curioso aqui.\n\n<-...->  ← separa os cards\n\ncard name: Próxima Carta\n...`}</pre>
            <p>Salve em <code>backend/internal/lore/data/cards/</code> — sem rebuild necessário.</p>
          </details>
        </div>
      ) : (
        <div className="lore-cards-main">
          <h2 className="lore-showcase-active-title">{selected.title}</h2>
          <p className="lore-section-intro">
            {selected.cards.length} {selected.cards.length === 1 ? "carta" : "cartas"} · Clique em uma carta para ver detalhes
          </p>
          <div className="lore-card-grid">
            {selected.cards.map((card, i) => (
              <CardItem
                key={i}
                card={card}
                image={getImage(card)}
                onClick={() => setActiveCard(card)}
              />
            ))}
          </div>
        </div>
      )}

      {/* Modal */}
      {activeCard && (
        <CardModal
          card={activeCard}
          image={getImage(activeCard)}
          onClose={closeModal}
        />
      )}
    </div>
  );
}
