import { useState, useEffect, useRef, useCallback } from "react";
import { listLoreChapters, getLoreChapter } from "./services/api.js";
import "./lore.css";

// ── Era metadata ──────────────────────────────────────────────────
const ERA_DATA = {
  "01": { color: "#9966bb", label: "Antiguidade · Antes de Urza",
          chars: ["Yawgmoth","Rebbec","Glacian","Dyfed"] },
  "02": { color: "#cc4422", label: "Era Antiga · Dominaria",
          chars: ["Urza","Mishra","Tocasia","Tawnos","Ashnod","Gix"] },
  "03": { color: "#4888cc", label: "Era de Urza · Dominaria",
          chars: ["Urza","Xantcha","Serra","Barrin","Karn","Jhoira","Teferi"] },
  "04": { color: "#d4c070", label: "Weatherlight Saga · Dominaria",
          chars: ["Gerrard","Sisay","Karn","Crovax","Mirri","Ertai"] },
  "05": { color: "#58b840", label: "Invasão Phyrexiana · Dominaria",
          chars: ["Urza","Gerrard","Yawgmoth","Karn","Eladamri","Lin Sivvi"] },
  "06": { color: "#3a9a60", label: "Old Phyrexia · Nine Titans",
          chars: ["Urza","Tevesh Szat","Commodore Guff","Teferi","Kristina"] },
  "07": { color: "#7888cc", label: "Tolaria · Sacrifício",
          chars: ["Hanna","Barrin","Rayne","Multani","Jhoira"] },
  "08": { color: "#e05828", label: "Apocalypse · Fim de Old Phyrexia",
          chars: ["Gerrard","Yawgmoth","Karn","Urza","Crovax","Teferi"] },
  "09": { color: "#a07840", label: "Odyssey · Otaria",
          chars: ["Kamahl","Chainer","Laquatus","Mirari","Cabal Patriarch"] },
  "10": { color: "#c04838", label: "Onslaught · Otaria",
          chars: ["Ixidor","Akroma","Phage","Kamahl","Braids"] },
  "11": { color: "#c49a4c", label: "Legions & Scourge · Karona",
          chars: ["Karona","Numena","Ixidor","Phage","Kamahl"] },
  "12": { color: "#e8c878", label: "Kamigawa · Kami War",
          chars: ["Konda","Toshiro","Michiko","O-Kagachi","Mochi"] },
  "13": { color: "#c04488", label: "Ravnica · Guildpact",
          chars: ["Jace","Niv-Mizzet","Szadek","Agrus Kos","Jarad"] },
  "14": { color: "#4888cc", label: "Time Spiral · O Mending",
          chars: ["Teferi","Jhoira","Urza","Karn","Venser","Radha"] },
  "99": { color: "#c49a4c", label: "Glossário · Referência",
          chars: ["Todos os personagens e planos"] },
};

function eraKey(slug) {
  return slug.slice(0, 2);
}

// ── Markdown parser ───────────────────────────────────────────────
function parseLine(line) {
  return line
    .replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>")
    .replace(/\*(.+?)\*/g, "<em>$1</em>")
    .replace(/`(.+?)`/g, "<code>$1</code>");
}

function renderMarkdown(md) {
  if (!md) return [];
  const lines = md.split("\n");
  const nodes = [];
  let listItems = [];
  let bqLines = [];
  let key = 0;

  function flushList() {
    if (!listItems.length) return;
    nodes.push(
      <ul key={key++} className="md-ul">
        {listItems.map((it, i) => (
          <li key={i} className="md-li" dangerouslySetInnerHTML={{ __html: it }} />
        ))}
      </ul>
    );
    listItems = [];
  }
  function flushBq() {
    if (!bqLines.length) return;
    nodes.push(
      <blockquote key={key++} className="md-blockquote">
        {bqLines.join(" ")}
      </blockquote>
    );
    bqLines = [];
  }

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i].trimEnd();

    if (line.startsWith("# ")) {
      flushList(); flushBq();
      const text = line.slice(2);
      const parts = text.split(" --- ");
      const title = parts.length > 1 ? parts.slice(1).join(" --- ") : text;
      nodes.push(<h1 key={key++} className="md-h1" dangerouslySetInnerHTML={{ __html: parseLine(title) }} />);
      continue;
    }
    if (line.startsWith("## ")) {
      flushList(); flushBq();
      nodes.push(<h2 key={key++} className="md-h2" dangerouslySetInnerHTML={{ __html: parseLine(line.slice(3)) }} />);
      continue;
    }
    if (line.startsWith("### ")) {
      flushList(); flushBq();
      nodes.push(<h3 key={key++} className="md-h3" dangerouslySetInnerHTML={{ __html: parseLine(line.slice(4)) }} />);
      continue;
    }
    if (line.startsWith(">")) {
      flushList();
      bqLines.push(parseLine(line.replace(/^>\s?/, "")));
      continue;
    }
    if (/^-\s+/.test(line)) {
      flushBq();
      listItems.push(parseLine(line.replace(/^-\s+/, "")));
      continue;
    }
    if (/^---+$/.test(line.trim())) {
      flushList(); flushBq();
      nodes.push(<hr key={key++} className="md-hr" />);
      continue;
    }
    if (!line.trim()) {
      flushList(); flushBq();
      continue;
    }

    flushList(); flushBq();
    let para = line;
    while (
      i + 1 < lines.length &&
      lines[i + 1].trim() !== "" &&
      !lines[i + 1].startsWith("#") &&
      !lines[i + 1].startsWith(">") &&
      !/^-\s+/.test(lines[i + 1]) &&
      !/^---+$/.test(lines[i + 1].trim())
    ) {
      i++;
      para += " " + lines[i].trimEnd();
    }
    nodes.push(<p key={key++} className="md-p" dangerouslySetInnerHTML={{ __html: parseLine(para) }} />);
  }

  flushList();
  flushBq();
  return nodes;
}

// ── Canvas Blind Eternities animation ─────────────────────────────
const MANA_COLORS = ["#f0e8c0","#4888cc","#9966bb","#cc4422","#338844","#c49a4c","#7c3aed"];

function useCosmosCanvas(canvasRef, active) {
  useEffect(() => {
    if (!active) return;
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    let particles = [], sparks = [], rafId;

    function rand(a, b) { return a + Math.random() * (b - a); }

    function resize() {
      canvas.width  = canvas.offsetWidth;
      canvas.height = canvas.offsetHeight;
    }

    function initParticles() {
      const { width: W, height: H } = canvas;
      const count = Math.min(160, Math.floor(W * H / 9000));
      particles = Array.from({ length: count }, () => ({
        x: rand(0, W), y: rand(0, H),
        r: rand(0.3, 1.3),
        speed: rand(0.05, 0.18),
        opacity: rand(0.08, 0.35),
        tw: rand(0, Math.PI * 2),
        twS: rand(0.008, 0.024),
      }));
    }

    function spawnSpark() {
      if (Math.random() < 0.012) {
        sparks.push({
          x: rand(0, canvas.width), y: rand(0, canvas.height),
          color: MANA_COLORS[Math.floor(Math.random() * MANA_COLORS.length)],
          r: rand(1.5, 3),
          life: 1,
          decay: rand(0.012, 0.025),
          vx: rand(-0.4, 0.4),
          vy: rand(-1, -0.3),
        });
      }
    }

    function draw() {
      const { width: W, height: H } = canvas;
      ctx.clearRect(0, 0, W, H);

      const g = ctx.createRadialGradient(W * 0.5, H * 0.4, 0, W * 0.5, H * 0.4, Math.max(W, H) * 0.7);
      g.addColorStop(0, "rgba(50,16,90,0.06)");
      g.addColorStop(1, "rgba(0,0,0,0)");
      ctx.fillStyle = g;
      ctx.fillRect(0, 0, W, H);

      for (const p of particles) {
        p.tw += p.twS;
        p.y -= p.speed;
        if (p.y < -2) p.y = H + 2;
        const alpha = p.opacity * (0.65 + 0.35 * Math.sin(p.tw));
        ctx.beginPath();
        ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(220,208,188,${alpha})`;
        ctx.fill();
      }

      spawnSpark();
      sparks = sparks.filter(s => s.life > 0);
      for (const s of sparks) {
        s.x += s.vx; s.y += s.vy; s.life -= s.decay;
        ctx.beginPath();
        ctx.arc(s.x, s.y, s.r * s.life, 0, Math.PI * 2);
        ctx.fillStyle = s.color + Math.floor(s.life * 180).toString(16).padStart(2, "0");
        ctx.shadowColor = s.color;
        ctx.shadowBlur = 8 * s.life;
        ctx.fill();
        ctx.shadowBlur = 0;
      }

      rafId = requestAnimationFrame(draw);
    }

    resize();
    initParticles();
    draw();

    const observer = new ResizeObserver(() => { resize(); initParticles(); });
    observer.observe(document.documentElement);

    return () => {
      cancelAnimationFrame(rafId);
      observer.disconnect();
    };
  }, [canvasRef, active]);
}

// ── Glossary data ──────────────────────────────────────────────────
const GLOSSARY = [
  { term: "Planeswalker",   def: "Ser cuja Spark se acendeu, capaz de atravessar as Blind Eternities entre planos." },
  { term: "Spark",          def: "Potencial raro que, aceso por trauma extremo, transforma alguém em Planeswalker." },
  { term: "Blind Eternities",def: "Espaço caótico entre planos. Viagem normal é fatal para a maioria dos seres." },
  { term: "Compleation",    def: "Transformação Phyrexiana de seres. New Phyrexia aprendeu a compleat Planeswalkers preservando Sparks." },
  { term: "Powerstone",     def: "Cristal capaz de armazenar grandes quantidades de energia. Tecnologia central dos Thran." },
  { term: "Old Phyrexia",   def: "Plano artificial de nove esferas remodelado por Yawgmoth como centro de sua civilização." },
  { term: "New Phyrexia",   def: "Antigo Mirrodin/Argentum transformado pelo óleo Phyrexiano. Destruída após a invasão multiversal." },
  { term: "Mightstone & Weakstone", def: "Duas metades da powerstone de Glacian. Gatilho da Brothers' War e, mais tarde, olhos de Urza." },
  { term: "O Mending",      def: "Reparação das distorções temporais de Dominaria — reduziu drasticamente o poder dos Planeswalkers." },
  { term: "Weatherlight",   def: "Nave voadora e planar ligada ao Legacy de Urza. Centro da saga que terminou com a morte de Yawgmoth." },
  { term: "Eldrazi",        def: "Entidades cósmicas sem cor. Os Titãs Ulamog, Kozilek e Emrakul distorcem a realidade ao redor." },
  { term: "Omenpaths",      def: "Caminhos interplanares abertos após March of the Machine, permitindo viagem sem Spark." },
  { term: "Realmbreaker",   def: "The Invasion Tree — estrutura de Elesh Norn que perfurou simultaneamente inúmeros planos." },
  { term: "Echoverse",      def: "Realidade construída por Jace/The Theorist sem Phyrexia, Eldrazi e Bolas — ameaça substituir o Multiverso." },
  { term: "Rathi Overlay",  def: "Operação que sobrepôs o plano artificial Rath a Dominaria, inserindo forças Phyrexianas diretamente." },
  { term: "Golgothian Sylex",def: "Artefato detonado por Urza que encerrou a Brothers' War, acendeu sua Spark e desencadeou a Ice Age." },
  { term: "Gatewatch",      def: "Grupo de Planeswalkers — Gideon, Jace, Chandra, Nissa, Liliana — unidos por juramento de proteger o Multiverso." },
  { term: "The Theorist",   def: "Identidade adotada por Jace como arquiteto do Echoverse — calculado, deliberado, traumatizado." },
];

// ── Main component ─────────────────────────────────────────────────
export default function LorePage({ onClose }) {
  const canvasRef  = useRef(null);
  const chapRefs   = useRef({});
  const [chapters, setChapters] = useState([]);
  const [openSlugs, setOpenSlugs] = useState({});
  const [chContent, setChContent] = useState({});
  const [loadingChs, setLoadingChs] = useState({});
  const [loading, setLoading] = useState(true);
  const [activeNavSlug, setActiveNavSlug] = useState(null);

  useCosmosCanvas(canvasRef, true);

  // Load chapter list
  useEffect(() => {
    listLoreChapters()
      .then(setChapters)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  // IntersectionObserver for nav highlight
  useEffect(() => {
    if (!chapters.length) return;
    const obs = new IntersectionObserver(
      (entries) => {
        entries.forEach(e => {
          if (e.isIntersecting) setActiveNavSlug(e.target.dataset.slug);
        });
      },
      { threshold: 0.15, rootMargin: "-5% 0px -70% 0px" }
    );
    Object.values(chapRefs.current).forEach(el => { if (el) obs.observe(el); });
    return () => obs.disconnect();
  }, [chapters]);

  const toggleChapter = useCallback((slug) => {
    const willOpen = !openSlugs[slug];
    setOpenSlugs(prev => ({ ...prev, [slug]: willOpen }));

    if (willOpen && !chContent[slug] && !loadingChs[slug]) {
      setLoadingChs(prev => ({ ...prev, [slug]: true }));
      getLoreChapter(slug)
        .then(ch => setChContent(prev => ({ ...prev, [slug]: ch.content })))
        .catch(console.error)
        .finally(() => setLoadingChs(prev => ({ ...prev, [slug]: false })));
    }
  }, [openSlugs, chContent, loadingChs]);

  const scrollToChapter = useCallback((slug) => {
    const el = chapRefs.current[slug];
    if (el) el.scrollIntoView({ behavior: "smooth", block: "start" });
    if (!openSlugs[slug]) toggleChapter(slug);
  }, [chapRefs, openSlugs, toggleChapter]);

  if (loading) {
    return (
      <div className="lore-root" style={{ display:"flex", alignItems:"center", justifyContent:"center" }}>
        <span style={{ fontFamily:"'Cinzel',serif", fontSize:10, letterSpacing:"0.3em", textTransform:"uppercase", color:"#4a4038" }}>
          Carregando Multiverso…
        </span>
      </div>
    );
  }

  return (
    <div className="lore-root">
      <canvas ref={canvasRef} className="lore-canvas" />

      {onClose && (
        <button className="lore-close" onClick={onClose}>
          ← Voltar ao app
        </button>
      )}

      <div className="lore-wrap">
        {/* ── Hero ─────────────────────────────────────── */}
        <section className="lore-hero">
          <p className="lore-eyebrow">Lore Oficial — Magic: The Gathering</p>
          <h1 className="lore-hero-title">Crônicas do<br />Multiverso</h1>
          <p className="lore-hero-sub">
            Da ascensão do Império Thran ao colapso das realidades em Reality Fracture — a história canônica completa.
          </p>
          <div className="lore-stats">
            <div className="lore-stat"><strong>{chapters.length}</strong>Eras Cronológicas</div>
            <div className="lore-stat-div" />
            <div className="lore-stat"><strong>5</strong>Cores de Mana</div>
            <div className="lore-stat-div" />
            <div className="lore-stat"><strong>∞</strong>Planos do Multiverso</div>
          </div>
          <div className="lore-mana-bar">
            {["w","u","b","r","g"].map(c => <div key={c} className={`lore-mp ${c}`} />)}
          </div>
        </section>

        {/* ── Dedicated chapter nav (menu exclusivo) ──── */}
        <nav className="lore-chapter-nav" aria-label="Capítulos da lore">
          <div className="lore-chapter-nav-inner">
            {chapters.map((ch, i) => {
              const k = eraKey(ch.slug);
              const era = ERA_DATA[k];
              return (
                <button
                  key={ch.slug}
                  className={`lore-nav-chapter-btn${activeNavSlug === ch.slug ? " active" : ""}`}
                  style={activeNavSlug === ch.slug && era ? { color: era.color, borderBottomColor: era.color } : {}}
                  onClick={() => scrollToChapter(ch.slug)}
                >
                  {toRoman(i + 1)} · {ch.title.split(",")[0]}
                </button>
              );
            })}
            <button
              className={`lore-nav-chapter-btn${activeNavSlug === "gloss" ? " active" : ""}`}
              onClick={() => {
                const el = chapRefs.current["gloss"];
                if (el) el.scrollIntoView({ behavior: "smooth", block: "start" });
              }}
            >
              Glossário
            </button>
          </div>
        </nav>

        {/* ── Timeline ─────────────────────────────────── */}
        <main className="lore-main">
          <div className="lore-timeline">
            {chapters.map((ch, i) => {
              const k     = eraKey(ch.slug);
              const era   = ERA_DATA[k] || {};
              const open  = !!openSlugs[ch.slug];
              const style = era.color ? { "--era-c": era.color } : {};

              return (
                <article
                  key={ch.slug}
                  id={`ch-${ch.slug}`}
                  data-slug={ch.slug}
                  ref={el => { chapRefs.current[ch.slug] = el; }}
                  className={`lore-ch${open ? " open" : ""}`}
                  style={style}
                >
                  <div className="lore-ch-dot" />

                  <div
                    className="lore-ch-head"
                    role="button"
                    tabIndex={0}
                    aria-expanded={open}
                    onClick={() => toggleChapter(ch.slug)}
                    onKeyDown={e => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); toggleChapter(ch.slug); }}}
                  >
                    <span className="lore-ch-num">{toRoman(i + 1)}</span>
                    <div className="lore-ch-meta">
                      {era.label && <span className="lore-era-tag">{era.label}</span>}
                      <div className="lore-ch-title">{ch.title}</div>
                      {era.chars && (
                        <div className="lore-ch-chars">
                          {era.chars.map(c => <span key={c} className="lore-char">{c}</span>)}
                        </div>
                      )}
                    </div>
                    <span className="lore-ch-arrow">▾</span>
                  </div>

                  <div className="lore-ch-body">
                    {loadingChs[ch.slug] && (
                      <div className="lore-ch-loading">Carregando…</div>
                    )}
                    {chContent[ch.slug] && (
                      <div className="lore-md" style={style}>
                        {renderMarkdown(chContent[ch.slug])}
                      </div>
                    )}
                  </div>
                </article>
              );
            })}
          </div>

          {/* ── Glossary ─────────────────────────────────── */}
          <section
            className="lore-gloss"
            id="lore-gloss"
            data-slug="gloss"
            ref={el => { chapRefs.current["gloss"] = el; }}
          >
            <h2 className="lore-section-title">Glossário</h2>
            <p className="lore-section-intro">Personagens, planos e conceitos fundamentais do Multiverso.</p>
            <div className="lore-gloss-grid">
              {GLOSSARY.map(g => (
                <div key={g.term} className="lore-gloss-card">
                  <div className="lore-gloss-term">{g.term}</div>
                  <div className="lore-gloss-def">{g.def}</div>
                </div>
              ))}
            </div>
          </section>

          <footer className="lore-footer">
            Magic: The Gathering · Lore Canônica Oficial · Compilação até setembro de 2026
          </footer>
        </main>
      </div>
    </div>
  );
}

// ── Roman numerals helper ──────────────────────────────────────────
function toRoman(n) {
  const vals = [10,9,5,4,1];
  const syms = ["X","IX","V","IV","I"];
  let result = "";
  for (let i = 0; i < vals.length; i++) {
    while (n >= vals[i]) { result += syms[i]; n -= vals[i]; }
  }
  return result;
}
