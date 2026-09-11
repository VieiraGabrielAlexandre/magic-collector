import { useState, useEffect, useRef, useCallback } from "react";
import { listLoreChapters, getLoreChapter } from "./services/api.js";
import LoreCards from "./LoreCards.jsx";
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

// ── Iconic characters for the gallery ────────────────────────────
const CHARACTERS = [
  { name: "Yawgmoth, Thran Physician", display: "Yawgmoth", era: "Old Phyrexia",       color: "#9966bb",
    desc: "Médico Thran exilado que se tornou o deus-demônio de Phyrexia. Arquiteto da compleação e maior antagonista da era clássica de Dominaria." },
  { name: "Urza, Lord High Artificer",  display: "Urza",     era: "Era de Urza",         color: "#4888cc",
    desc: "O maior artificer de Dominaria. Passou milênios preparando o Multiverso contra Phyrexia após a Brothers' War acender sua Spark." },
  { name: "Karn Liberated",             display: "Karn",     era: "Multiversal",          color: "#c49a4c",
    desc: "Golem de prata criado por Urza com uma Spark transplantada — o único ser incolor com poderes de Planeswalker e criador do plano Mirrodin." },
  { name: "Teferi, Hero of Dominaria",  display: "Teferi",   era: "Dominaria · Tolaria",  color: "#5599dd",
    desc: "Mago temporal incomparável de Tolaria. Sacrificou sua Spark para selar as distorções temporais de Dominaria durante O Mending." },
  { name: "Gerrard, Weatherlight Hero", display: "Gerrard",  era: "Weatherlight Saga",    color: "#d4c070",
    desc: "O Legado vivo de Urza — herdeiro de gerações de manipulação genética. Capitão da Weatherlight que derrotou Yawgmoth sacrificando a própria vida." },
  { name: "Serra the Benevolent",       display: "Serra",    era: "Serra's Realm",        color: "#e8d888",
    desc: "Planeswalker da magia branca que criou seu próprio plano. Sua fé e magia moldaram a filosofia e a história de Dominaria por séculos." },
  { name: "Jace, the Mind Sculptor",    display: "Jace",     era: "Ravnica · Echoverse",  color: "#4499cc",
    desc: "O Mente-Escultor. Cofundador do Gatewatch e o telépata mais poderoso do Multiverso — hoje o controverso arquiteto do Echoverse como The Theorist." },
  { name: "Liliana of the Veil",        display: "Liliana",  era: "Multiversal",          color: "#9966bb",
    desc: "Necromancer que vendeu sua alma a quatro demônios em troca de poder e imortalidade. Sua jornada de redenção custou tudo que ela amou." },
  { name: "Chandra, Torch of Defiance", display: "Chandra",  era: "Kaladesh · Multiversal",color: "#cc4422",
    desc: "Piromante impulsiva e coração de fogo do Gatewatch. Filha de Pia Nalaar — sua Spark acendeu ao escapar de uma execução em Kaladesh." },
  { name: "Nissa, Resurgent Animist",   display: "Nissa",    era: "Zendikar · Multiversal",color: "#338844",
    desc: "Elfa animista de Zendikar com conexão singular às ley lines de cada plano. Cofundadora do Gatewatch ao enfrentar os Titãs Eldrazi." },
];

// ── Encyclopedia entries (planes, places, artifacts, concepts, events) ──
const ENC_CATS = [
  { id: "todos",    label: "Todos" },
  { id: "plano",    label: "Planos" },
  { id: "lugar",    label: "Lugares" },
  { id: "artefato", label: "Artefatos" },
  { id: "conceito", label: "Conceitos" },
  { id: "evento",   label: "Eventos" },
];

const ENCYCLOPEDIA = [
  // ── Planos ──
  { cat:"plano",    icon:"🌍", color:"#c87840",
    name:"Dominaria",           sub:"O Plano Central",
    def:"O plano mais importante da cosmologia MTG. Cenário de milênios de conflitos — dos Thran ao Mending. Lar de Urza, Gerrard e Teferi." },
  { cat:"plano",    icon:"⚙️", color:"#9966bb",
    name:"Old Phyrexia",        sub:"Nove Esferas · Criação de Yawgmoth",
    def:"Plano artificial de nove esferas remodelado por Yawgmoth como paraíso da perfeição mecânica. Destruído pelos Nine Titans durante o Apocalypse." },
  { cat:"plano",    icon:"🦠", color:"#88aa44",
    name:"New Phyrexia",        sub:"Mirrodin Corrompido",
    def:"O plano metálico Mirrodin criado por Karn, corrompido pelo óleo Phyrexiano. Governado por cinco Praetores — destruído na invasão multiversal." },
  { cat:"plano",    icon:"🏛️", color:"#c04488",
    name:"Ravnica",             sub:"Cidade-Mundo das Guildas",
    def:"Um único plano urbano governado por dez guildas. Cenário do despertar de Jace e do Guildpact que moldou o equilíbrio do Multiverso." },
  { cat:"plano",    icon:"⚡", color:"#338844",
    name:"Zendikar",            sub:"Plano Vivo",
    def:"Plano onde a terra é uma arma viva — prisão dos Eldrazi Titãs por milênios. Lar de Nissa e palco da primeira batalha do Gatewatch." },
  { cat:"plano",    icon:"🌸", color:"#e8c860",
    name:"Kamigawa",            sub:"Plano dos Espíritos",
    def:"Plano de inspiração japonesa dividido entre o mundo material e o espiritual (kami). A guerra entre mortais e kami devastou o plano por décadas." },
  { cat:"plano",    icon:"☀️", color:"#f0e8b0",
    name:"Serra's Realm",       sub:"Plano Artificial de Serra",
    def:"Plano criado pela Planeswalker Serra como paraíso de magia branca. Consumido para alimentar o Legacy de Urza após a morte de Serra." },
  { cat:"plano",    icon:"🔩", color:"#446688",
    name:"Rath",                sub:"Plano Phyrexiano de Invasão",
    def:"Plano artificial criado por Yawgmoth para estagiar tropas. Sobreposto a Dominaria durante a Invasão via Rathi Overlay, inserindo legiões Phyrexianas diretamente." },
  { cat:"plano",    icon:"🏔️", color:"#aa8844",
    name:"Otaria",              sub:"Continente de Dominaria",
    def:"Continente isolado de Dominaria que surgiu após o Apocalypse. Palco dos arcos de Odyssey, Onslaught e das histórias de Kamahl, Phage e Karona." },
  // ── Lugares ──
  { cat:"lugar",    icon:"📚", color:"#4888cc",
    name:"Tolaria",             sub:"Ilha Acadêmica de Dominaria",
    def:"Academia de magia temporal fundada por Urza. A explosão da Mana Bomb criou bolsões de tempo acelerado e lento — formando Teferi, Jhoira e toda uma geração de magos." },
  { cat:"lugar",    icon:"🌲", color:"#338844",
    name:"Argoth",              sub:"Floresta Sagrada de Dominaria",
    def:"Floresta sagrada destruída durante a Brothers' War. O Golgothian Sylex foi detonado aqui, encerrando a guerra e desencadeando a Ice Age que durou milênios." },
  { cat:"lugar",    icon:"🌑", color:"#5a3a6a",
    name:"Urborg",              sub:"Pântano de Dominaria",
    def:"Pântano sombrio de Dominaria, sede do Cabal Patriarch e da organização Cabal. Lar de Phage e centro do controle da Otaria no período pós-Apocalypse." },
  { cat:"lugar",    icon:"⚓", color:"#4466aa",
    name:"Mercadia",            sub:"Cidade Mercantil Invertida",
    def:"Cidade no topo de uma montanha invertida — o centro comercial de um plano peculiar. A Weatherlight passou por aqui durante sua fuga de Rath." },
  { cat:"lugar",    icon:"🏛️", color:"#8866aa",
    name:"Koilos",              sub:"Cavernas dos Thran",
    def:"Cavernas de Dominaria onde os Thran esconderam seu arsenal de artefatos. O local onde Urza e Mishra encontraram o Mightstone e o Weakstone — o gatilho da Brothers' War." },
  // ── Artefatos ──
  { cat:"artefato", icon:"🚢", color:"#d4c070",
    name:"Weatherlight",        sub:"Nave do Legacy de Urza",
    def:"Nave voadora e viajante de planos, parte do Legacy de Urza. Propulsionada por um cristal de mana e ligada ao Spark de Gerrard — central na luta contra Yawgmoth." },
  { cat:"artefato", icon:"💥", color:"#cc4422",
    name:"Golgothian Sylex",    sub:"Artefato Thran Detonado em Argoth",
    def:"Detonado por Urza em Argoth, encerrou a Brothers' War, acendeu a Spark de Urza e causou a Ice Age que dominou Dominaria por milênios." },
  { cat:"artefato", icon:"💎", color:"#c49a4c",
    name:"Mightstone & Weakstone", sub:"Powerstones de Glacian",
    def:"Duas metades da Powerstone de Glacian. Urza obteve a Mightstone, Mishra a Weakstone — o conflito por elas iniciou a Brothers' War. Tornaram-se os olhos de Urza." },
  { cat:"artefato", icon:"🌿", color:"#88aa44",
    name:"Realmbreaker",        sub:"Árvore da Invasão de Elesh Norn",
    def:"Estrutura biomecânica criada por Elesh Norn para perfurar simultaneamente dezenas de planos — o vetor da invasão multiversal de New Phyrexia." },
  { cat:"artefato", icon:"⚗️", color:"#7c3aed",
    name:"Legacy de Urza",      sub:"Arsenal do Grande Artificer",
    def:"Coleção de artefatos poderosos que Urza passou milênios reunindo como arma contra Phyrexia. Incluía a Weatherlight, o Thran Tome e diversas relíquias mágicas." },
  // ── Conceitos ──
  { cat:"conceito", icon:"✨", color:"#c49a4c",
    name:"Planeswalker Spark",  sub:"O Dom dos Viajantes",
    def:"Potencial raro que, aceso por trauma extremo, transforma seu portador em Planeswalker — capaz de viajar entre planos através das Blind Eternities." },
  { cat:"conceito", icon:"🌌", color:"#7c3aed",
    name:"Blind Eternities",    sub:"O Espaço Entre Planos",
    def:"O espaço caótico e hostil entre os planos do Multiverso. Fatal para a maioria dos seres — somente Planeswalkers e Eldrazi o atravessam." },
  { cat:"conceito", icon:"⚙️", color:"#9966bb",
    name:"Compleation",         sub:"Perfeição Phyrexiana",
    def:"Processo de transformação de seres orgânicos em criaturas mecânico-orgânicas. New Phyrexia aperfeiçoou a técnica para compleat Planeswalkers preservando suas Sparks." },
  { cat:"conceito", icon:"🫧", color:"#88aa44",
    name:"Glistening Oil",      sub:"O Óleo Phyrexiano",
    def:"Substância de Old Phyrexia que corrompe qualquer ser ao contato prolongado, iniciando a Phyresis. O principal vetor da expansão Phyrexiana por todo o Multiverso." },
  { cat:"conceito", icon:"🚪", color:"#44aa88",
    name:"Omenpaths",           sub:"Caminhos Interplanares Pós-March",
    def:"Passagens entre planos abertas após o colapso do Realmbreaker em March of the Machine. Permitem viagem interplanar sem necessitar de Spark." },
  { cat:"conceito", icon:"🔮", color:"#5599dd",
    name:"Echoverse",           sub:"A Realidade Alternativa de Jace",
    def:"Realidade construída por Jace como The Theorist — sem Phyrexia, sem Eldrazi, sem Nicol Bolas. Ameaça substituir o Multiverso existente." },
  { cat:"conceito", icon:"⚔️", color:"#c49a4c",
    name:"Gatewatch",           sub:"Os Guardiões do Multiverso",
    def:"Gideon, Jace, Chandra, Nissa e Liliana — cinco Planeswalkers unidos por juramento de proteger o Multiverso de ameaças existenciais." },
  // ── Eventos ──
  { cat:"evento",   icon:"⚔️", color:"#cc4422",
    name:"Brothers' War",       sub:"Dominaria · ~5000 A.R.",
    def:"A guerra devastadora entre Urza e Mishra que destruiu continentes inteiros de Dominaria, culminando no Sylex Blast e na Ice Age que se seguiu por milênios." },
  { cat:"evento",   icon:"💀", color:"#9966bb",
    name:"Phyrexian Invasion",  sub:"Dominaria · ~4205 A.R.",
    def:"A invasão de Dominaria por Yawgmoth — o clímax da Weatherlight Saga. Terminou com a morte de Yawgmoth, de Gerrard e a destruição final de Old Phyrexia." },
  { cat:"evento",   icon:"❄️", color:"#aaccee",
    name:"Ice Age",             sub:"Dominaria · ~450 A.R.",
    def:"A era glacial que cobriu Dominaria por milênios após o Sylex Blast. Período de isolamento e declínio que moldou civilizações inteiras antes de seu fim gradual." },
  { cat:"evento",   icon:"🔥", color:"#e05828",
    name:"March of the Machine", sub:"Multiversal · Era Atual",
    def:"A invasão simultânea de todos os planos por New Phyrexia usando o Realmbreaker. Encerrada com a destruição de New Phyrexia — e a abertura dos Omenpaths." },
  { cat:"evento",   icon:"🕰️", color:"#5599dd",
    name:"O Mending",           sub:"Dominaria · Era Recente",
    def:"Evento desencadeado por Teferi e Jeska para reparar as distorções temporais de Dominaria. Reduziu drasticamente o poder de todos os Planeswalkers do Multiverso." },
];

function eraKey(slug) { return slug.slice(0, 2); }

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
  let listItems = [], bqLines = [], key = 0;

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
    nodes.push(<blockquote key={key++} className="md-blockquote">{bqLines.join(" ")}</blockquote>);
    bqLines = [];
  }

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i].trimEnd();
    if (line.startsWith("# ")) {
      flushList(); flushBq();
      const text = line.slice(2);
      const parts = text.split(/\s+---\s+|\s+—\s+/);
      const title = parts.length > 1 ? parts.slice(1).join(" — ") : text;
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
    if (!line.trim()) { flushList(); flushBq(); continue; }
    flushList(); flushBq();
    let para = line;
    while (
      i + 1 < lines.length &&
      lines[i + 1].trim() !== "" &&
      !lines[i + 1].startsWith("#") &&
      !lines[i + 1].startsWith(">") &&
      !/^-\s+/.test(lines[i + 1]) &&
      !/^---+$/.test(lines[i + 1].trim())
    ) { i++; para += " " + lines[i].trimEnd(); }
    nodes.push(<p key={key++} className="md-p" dangerouslySetInnerHTML={{ __html: parseLine(para) }} />);
  }
  flushList(); flushBq();
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
      const count = Math.floor((canvas.width * canvas.height) / 9000);
      particles = Array.from({ length: count }, () => ({
        x: rand(0, canvas.width), y: rand(0, canvas.height),
        r: rand(0.4, 1.6), speed: rand(0.04, 0.18),
        opacity: rand(0.1, 0.5), drift: rand(-0.04, 0.04),
      }));
      sparks = Array.from({ length: 18 }, () => ({
        x: rand(0, canvas.width), y: rand(0, canvas.height),
        r: rand(1.2, 3), color: MANA_COLORS[Math.floor(rand(0, MANA_COLORS.length))],
        speed: rand(0.06, 0.22), opacity: rand(0.05, 0.22),
        drift: rand(-0.06, 0.06), pulse: rand(0, Math.PI * 2),
      }));
    }

    function draw() {
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      particles.forEach(p => {
        p.y -= p.speed; p.x += p.drift;
        if (p.y < -2) { p.y = canvas.height + 2; p.x = rand(0, canvas.width); }
        if (p.x < -2) p.x = canvas.width + 2;
        if (p.x > canvas.width + 2) p.x = -2;
        ctx.beginPath();
        ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(220,200,160,${p.opacity})`;
        ctx.fill();
      });
      sparks.forEach(s => {
        s.y -= s.speed; s.x += s.drift; s.pulse += 0.04;
        if (s.y < -4) { s.y = canvas.height + 4; s.x = rand(0, canvas.width); }
        if (s.x < -4) s.x = canvas.width + 4;
        if (s.x > canvas.width + 4) s.x = -4;
        const glow = ctx.createRadialGradient(s.x, s.y, 0, s.x, s.y, s.r * (2 + Math.sin(s.pulse) * 0.5));
        glow.addColorStop(0, s.color + "55");
        glow.addColorStop(1, "transparent");
        ctx.beginPath();
        ctx.arc(s.x, s.y, s.r * 3, 0, Math.PI * 2);
        ctx.fillStyle = glow;
        ctx.fill();
        ctx.beginPath();
        ctx.arc(s.x, s.y, s.r, 0, Math.PI * 2);
        ctx.fillStyle = s.color + Math.floor(s.opacity * 255).toString(16).padStart(2, "0");
        ctx.fill();
      });
      rafId = requestAnimationFrame(draw);
    }

    resize(); initParticles(); draw();
    const observer = new ResizeObserver(() => { resize(); initParticles(); });
    observer.observe(document.documentElement);
    return () => { cancelAnimationFrame(rafId); observer.disconnect(); };
  }, [canvasRef, active]);
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

// ── Main component ─────────────────────────────────────────────────
export default function LorePage({ onClose }) {
  const canvasRef      = useRef(null);
  const chapRefs       = useRef({});
  const timelineRef    = useRef(null);
  const [chapters, setChapters]       = useState([]);
  const [openSlugs, setOpenSlugs]     = useState({});
  const [chContent, setChContent]     = useState({});
  const [loadingChs, setLoadingChs]   = useState({});
  const [loading, setLoading]         = useState(true);
  const [activeNavSlug, setActiveNavSlug] = useState(null);
  const [charImages, setCharImages]   = useState({});
  const [encFilter, setEncFilter]     = useState("todos");
  const [activeView, setActiveView]   = useState("lore"); // "lore" | "cards"

  useCosmosCanvas(canvasRef, true);

  // Load chapter list
  useEffect(() => {
    listLoreChapters()
      .then(setChapters)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  // Fetch character card art from Scryfall (batch)
  useEffect(() => {
    const identifiers = CHARACTERS.map(c => ({ name: c.name }));
    fetch("https://api.scryfall.com/cards/collection", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ identifiers }),
    })
      .then(r => r.json())
      .then(result => {
        const imgs = {};
        (result.data || []).forEach(card => {
          const match = CHARACTERS.find(
            c => card.name.toLowerCase() === c.name.toLowerCase()
          );
          if (match) {
            const url =
              card.image_uris?.art_crop ||
              card.card_faces?.[0]?.image_uris?.art_crop;
            if (url) imgs[match.display] = url;
          }
        });
        setCharImages(imgs);
      })
      .catch(console.error);
  }, []);

  // IntersectionObserver for chapter grid highlight
  useEffect(() => {
    if (!chapters.length) return;
    const obs = new IntersectionObserver(
      entries => entries.forEach(e => { if (e.isIntersecting) setActiveNavSlug(e.target.dataset.slug); }),
      { threshold: 0.1, rootMargin: "-10% 0px -65% 0px" }
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
    // Open chapter first, then scroll after a tick so DOM is ready
    const willOpen = !openSlugs[slug];
    setOpenSlugs(prev => ({ ...prev, [slug]: true }));
    if (willOpen && !chContent[slug] && !loadingChs[slug]) {
      setLoadingChs(prev => ({ ...prev, [slug]: true }));
      getLoreChapter(slug)
        .then(ch => setChContent(prev => ({ ...prev, [slug]: ch.content })))
        .catch(console.error)
        .finally(() => setLoadingChs(prev => ({ ...prev, [slug]: false })));
    }
    setTimeout(() => {
      const el = chapRefs.current[slug];
      if (el) el.scrollIntoView({ behavior: "smooth", block: "start" });
    }, 60);
  }, [openSlugs, chContent, loadingChs]);

  if (loading) {
    return (
      <div className="lore-root" style={{ display:"flex", alignItems:"center", justifyContent:"center" }}>
        <span style={{ fontFamily:"'Cinzel',serif", fontSize:10, letterSpacing:"0.3em", textTransform:"uppercase", color:"#4a4038" }}>
          Carregando Multiverso…
        </span>
      </div>
    );
  }

  // ── Cards view ─────────────────────────────────────────────────
  if (activeView === "cards") {
    return (
      <div className="lore-root">
        <canvas ref={canvasRef} className="lore-canvas" />
        <LoreCards onBack={() => setActiveView("lore")} onClose={onClose} />
      </div>
    );
  }

  return (
    <div className="lore-root">
      <canvas ref={canvasRef} className="lore-canvas" />

      {onClose && (
        <button className="lore-close" onClick={onClose}>
          ← Voltar
        </button>
      )}

      <div className="lore-wrap">
        {/* ── Hero ──────────────────────────────────────── */}
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

        <main className="lore-main">

          {/* ── Characters gallery ────────────────────── */}
          <section className="lore-chars-section">
            <p className="lore-chars-eyebrow">Personagens Canônicos</p>
            <h2 className="lore-section-title">Figuras do Multiverso</h2>
            <p className="lore-section-intro">
              Os planeswalkers, artificers e heróis que moldaram a história de Magic: The Gathering.
            </p>
            <div className="lore-chars-grid">
              {CHARACTERS.map(ch => (
                <div key={ch.display} className="lore-char-card" style={{ "--era-c": ch.color }}>
                  <div className="lore-char-art">
                    {charImages[ch.display]
                      ? <img src={charImages[ch.display]} alt={ch.display} loading="lazy" />
                      : <div className="lore-char-art-placeholder">✦</div>}
                    <div className="lore-char-art-fade" />
                  </div>
                  <div className="lore-char-info">
                    <span className="lore-char-era-label">{ch.era}</span>
                    <div className="lore-char-name">{ch.display}</div>
                    <div className="lore-char-desc">{ch.desc}</div>
                  </div>
                </div>
              ))}
            </div>
          </section>

          {/* ── Encyclopedia ─────────────────────────── */}
          <section className="lore-enc-section">
            <p className="lore-chars-eyebrow">Enciclopédia do Multiverso</p>
            <h2 className="lore-section-title">Planos, Lugares & Conceitos</h2>
            <p className="lore-section-intro">
              Os planos, lugares, artefatos, conceitos e eventos que definem a história de Magic: The Gathering.
            </p>

            {/* Filter tabs */}
            <div className="lore-enc-filters">
              {ENC_CATS.map(c => (
                <button
                  key={c.id}
                  className={`lore-enc-filter-btn${encFilter === c.id ? " active" : ""}`}
                  onClick={() => setEncFilter(c.id)}
                >
                  {c.label}
                </button>
              ))}
            </div>

            <div className="lore-enc-grid">
              {ENCYCLOPEDIA.filter(e => encFilter === "todos" || e.cat === encFilter).map(e => (
                <div key={e.name} className="lore-enc-card" style={{ "--enc-c": e.color }}>
                  <div className="lore-enc-visual">
                    <div className="lore-enc-icon">{e.icon}</div>
                    <div className="lore-enc-glow" />
                  </div>
                  <div className="lore-enc-body">
                    <div className="lore-enc-cat">{ENC_CATS.find(c => c.id === e.cat)?.label}</div>
                    <div className="lore-enc-name">{e.name}</div>
                    <div className="lore-enc-sub">{e.sub}</div>
                    <div className="lore-enc-def">{e.def}</div>
                  </div>
                </div>
              ))}
            </div>
          </section>

          {/* ── Cards showcase entry ──────────────────── */}
          <section className="lore-cards-entry">
            <div className="lore-cards-entry-inner">
              <div className="lore-cards-entry-text">
                <p className="lore-chars-eyebrow">Vitrines de Cartas</p>
                <h2 className="lore-selector-title">Cartas em Destaque</h2>
                <p className="lore-selector-sub">
                  Arte oficial do Scryfall, mana costs, curiosidades e lore — em vitrines curáveis por você via <code>.md</code>.
                </p>
              </div>
              <button className="lore-cards-entry-btn" onClick={() => setActiveView("cards")}>
                <span>✦</span> Explorar Cartas
              </button>
            </div>
          </section>

          {/* ── Chapter selector grid ─────────────────── */}
          <section className="lore-ch-selector">
            <h2 className="lore-selector-title">Capítulos</h2>
            <p className="lore-selector-sub">
              Selecione um capítulo para abrir a narrativa completa na cronologia abaixo.
            </p>
            <div className="lore-ch-grid">
              {chapters.map((ch, i) => {
                const k   = eraKey(ch.slug);
                const era = ERA_DATA[k] || {};
                return (
                  <button
                    key={ch.slug}
                    className={`lore-ch-card${activeNavSlug === ch.slug ? " active-ch" : ""}`}
                    style={{ "--era-c": era.color || "#c49a4c" }}
                    onClick={() => scrollToChapter(ch.slug)}
                  >
                    <div className="lore-ch-card-top">
                      <span className="lore-ch-card-num">{toRoman(i + 1)}</span>
                      {era.label && <span className="lore-ch-card-era">{era.label}</span>}
                    </div>
                    <div className="lore-ch-card-title">{ch.title}</div>
                    {era.chars && (
                      <div className="lore-ch-card-chars">
                        {era.chars.slice(0, 3).map(c => <span key={c}>{c}</span>)}
                      </div>
                    )}
                  </button>
                );
              })}
            </div>
          </section>

          {/* ── Timeline header ───────────────────────── */}
          <div className="lore-timeline-hd">
            <h2 className="lore-selector-title">Cronologia</h2>
            <p className="lore-selector-sub">Clique em um capítulo para expandir a narrativa.</p>
          </div>

          {/* ── Timeline accordion ────────────────────── */}
          <div className="lore-timeline" ref={timelineRef}>
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
                    {loadingChs[ch.slug] && <div className="lore-ch-loading">Carregando…</div>}
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

          <footer className="lore-footer">
            Magic: The Gathering · Lore Canônica Oficial · Compilação até setembro de 2026
          </footer>
        </main>
      </div>
    </div>
  );
}
