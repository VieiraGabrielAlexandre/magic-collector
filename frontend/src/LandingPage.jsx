import { useState } from "react";
import { login } from "./services/api.js";

const API_ROUTES = [
  { method: "GET",    path: "/cards",                    desc: "Listar cartas (filtros, paginação, ordenação)" },
  { method: "POST",   path: "/cards",                    desc: "Cadastrar carta (busca automática na Scryfall)" },
  { method: "GET",    path: "/cards/:id",                desc: "Detalhes da carta + dados externos Scryfall" },
  { method: "PUT",    path: "/cards/:id",                desc: "Editar carta" },
  { method: "DELETE", path: "/cards/:id",                desc: "Remover carta" },
  { method: "PATCH",  path: "/cards/:id/quantity",       desc: "Atualizar quantidade" },
  { method: "PATCH",  path: "/cards/:id/deck",           desc: "Atribuir carta a um deck" },
  { method: "GET",    path: "/cards/export",             desc: "Exportar toda a coleção (sem paginação)" },
  { method: "GET",    path: "/cards/colors",             desc: "Combinações de cores disponíveis na coleção" },
  { method: "GET",    path: "/cards/stats",              desc: "Estatísticas da coleção" },
  { method: "POST",   path: "/cards/preview",            desc: "Preview de carta na Scryfall sem salvar" },
  { method: "POST",   path: "/cards/suggest-decks",      desc: "Sugestão de deck com IA (GPT-4o)" },
  { method: "POST",   path: "/cards/refresh-prices",     desc: "Atualizar preços em USD via Scryfall" },
  { method: "POST",   path: "/cards/refresh-images",     desc: "Atualizar imagens via Scryfall" },
  { method: "GET",    path: "/decks",                    desc: "Listar decks" },
  { method: "POST",   path: "/decks",                    desc: "Criar deck" },
  { method: "PUT",    path: "/decks/:id",                desc: "Editar deck" },
  { method: "DELETE", path: "/decks/:id",                desc: "Remover deck" },
  { method: "POST",   path: "/decks/:id/evaluate",       desc: "Avaliar deck com IA" },
  { method: "POST",   path: "/decks/import-precon",      desc: "Importar deck pré-construído via Scryfall" },
  { method: "POST",   path: "/decks/import-list",        desc: "Importar lista de deck (formato texto)" },
  { method: "GET",    path: "/battles",                  desc: "Histórico de partidas" },
  { method: "POST",   path: "/battles",                  desc: "Registrar partida" },
  { method: "DELETE", path: "/battles/:id",              desc: "Remover registro de partida" },
  { method: "POST",   path: "/auth/login",               desc: "Autenticar usuário" },
  { method: "POST",   path: "/auth/logout",              desc: "Encerrar sessão" },
  { method: "GET",    path: "/auth/me",                  desc: "Dados do usuário autenticado" },
];

const FEATURES = [
  { icon: "🃏", title: "Coleção completa", desc: "Cadastre suas cartas com busca automática na Scryfall. Nome, arte, tipo, custo de mana, preço e imagem preenchidos automaticamente." },
  { icon: "🗂️", title: "Organização em decks", desc: "Agrupe suas cartas em decks, importe pré-cons completos ou listas em texto, e visualize a identidade de cor de cada deck." },
  { icon: "⚔️", title: "Histórico de batalhas", desc: "Registre suas partidas, acompanhe vitórias e derrotas por deck e identifique qual estilo de jogo você domina." },
  { icon: "🤖", title: "IA para deck building", desc: "Com um clique, o GPT-4o analisa sua coleção e sugere o melhor deck possível considerando sinergias, curva de mana e estratégia." },
  { icon: "🌐", title: "Scryfall integrado", desc: "Acesso direto à maior base de dados de MTG. Preços atualizados, imagens em alta resolução e suporte a cartas em PT, JP, ES e mais." },
  { icon: "📊", title: "Estatísticas", desc: "Veja sua coleção por raridade, cor, set e valor total em USD. Exporte tudo para Excel com um clique." },
];

const METHOD_COLORS = { GET: "#4ade80", POST: "#60a5fa", PUT: "#fbbf24", PATCH: "#a78bfa", DELETE: "#f87171" };

export default function LandingPage({ onEnter, onBack = null }) {
  const [showLogin, setShowLogin] = useState(false);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [loginError, setLoginError] = useState("");
  const [loginLoading, setLoginLoading] = useState(false);
  const [showRoutes, setShowRoutes] = useState(false);

  async function handleLogin(e) {
    e.preventDefault();
    setLoginError("");
    setLoginLoading(true);
    try {
      const data = await login(username, password);
      localStorage.setItem("auth_token", data.token);
      localStorage.setItem("auth_session_created_at", data.session_created_at);
      onEnter(data.user, data.session_created_at);
    } catch (err) {
      setLoginError(err.message);
    } finally {
      setLoginLoading(false);
    }
  }

  return (
    <div className="landing">
      {/* ── NAV ── */}
      <nav className="landing-nav">
        <span className="landing-logo">⚔ Magic Collector</span>
        {onBack
          ? <button className="landing-enter-btn" onClick={onBack}>← Voltar à coleção</button>
          : <button className="landing-enter-btn" onClick={() => setShowLogin(true)}>Acessar coleção →</button>
        }
      </nav>

      {/* ── HERO ── */}
      <section className="landing-hero">
        <div className="landing-hero-content">
          <div className="landing-hero-badge">Open source · Feito com ❤️ para colecionadores</div>
          <h1 className="landing-title">
            Sua coleção de<br />
            <span className="landing-title-gold">Magic: The Gathering</span><br />
            organizada e inteligente
          </h1>
          <p className="landing-subtitle">
            Gerencie cartas, monte decks, registre batalhas e deixe a IA sugerir
            as melhores estratégias — tudo integrado com a Scryfall API.
          </p>
          <div className="landing-hero-actions">
            <button className="landing-cta-btn" onClick={() => setShowLogin(true)}>
              Acessar coleção
            </button>
            <a className="landing-gh-btn" href="https://github.com" target="_blank" rel="noreferrer">
              Ver no GitHub ↗
            </a>
          </div>
        </div>
        <div className="landing-hero-visual">
          <div className="landing-card-stack">
            {["#c0a060", "#8060c0", "#4080c0"].map((c, i) => (
              <div key={i} className="landing-card-ghost" style={{ "--card-color": c, "--card-i": i }} />
            ))}
            <div className="landing-card-main">
              <div className="landing-card-inner">
                <span className="landing-card-icon">⚔</span>
                <strong>Magic Collector</strong>
                <small>v2.0 · magic-collector.site</small>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── FEATURES ── */}
      <section className="landing-section">
        <h2 className="landing-section-title">Tudo que você precisa</h2>
        <div className="landing-features">
          {FEATURES.map((f) => (
            <div key={f.title} className="landing-feature-card">
              <span className="landing-feature-icon">{f.icon}</span>
              <h3>{f.title}</h3>
              <p>{f.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* ── COMO USAR ── */}
      <section className="landing-section landing-section-alt">
        <h2 className="landing-section-title">Como usar</h2>
        <div className="landing-steps">
          {[
            { n: "1", title: "Cadastre suas cartas", desc: "Digite o set e número da carta. O sistema busca automaticamente nome, imagem, custo de mana, raridade e preço na Scryfall." },
            { n: "2", title: "Organize em decks", desc: "Crie decks, atribua cartas e importe listas completas no formato texto. Suporte a Commander e formatos de 60 cartas." },
            { n: "3", title: "Use a IA", desc: "Clique em 'Sugerir deck com IA' e o GPT-4o analisa toda sua coleção e propõe o deck mais sinérgico e divertido para você." },
            { n: "4", title: "Registre suas partidas", desc: "Após cada jogo, registre o resultado, oponentes e estilo. Acompanhe sua evolução e descubra qual deck você domina melhor." },
          ].map((s) => (
            <div key={s.n} className="landing-step">
              <div className="landing-step-num">{s.n}</div>
              <div>
                <h3>{s.title}</h3>
                <p>{s.desc}</p>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* ── API ROUTES ── */}
      <section className="landing-section">
        <h2 className="landing-section-title">API REST disponível</h2>
        <p className="landing-section-sub">
          Todas as funcionalidades expostas como API JSON. Autenticação via Bearer token.
          Base URL: <code>https://magic-collector.site/api</code>
        </p>
        <button className="landing-toggle-routes" onClick={() => setShowRoutes((v) => !v)}>
          {showRoutes ? "▲ Ocultar rotas" : "▼ Ver todas as rotas"}
        </button>
        {showRoutes && (
          <div className="landing-routes">
            {API_ROUTES.map((r) => (
              <div key={r.path + r.method} className="landing-route">
                <span className="landing-route-method" style={{ color: METHOD_COLORS[r.method] || "#fff" }}>
                  {r.method}
                </span>
                <code className="landing-route-path">{r.path}</code>
                <span className="landing-route-desc">{r.desc}</span>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* ── CONTRIBUIR ── */}
      <section className="landing-section landing-section-alt">
        <h2 className="landing-section-title">Contribua com o projeto</h2>
        <div className="landing-contribute">
          <div className="landing-contribute-text">
            <p>
              O Magic Collector é um projeto pessoal open source criado por{" "}
              <strong>Gabriel</strong> e <strong>Juliana</strong> para resolver um
              problema real: gerenciar uma coleção crescente de cartas MTG sem depender
              de planilhas ou apps engessados.
            </p>
            <p>
              O projeto usa <strong>Go + React</strong>, integra com a{" "}
              <strong>Scryfall API</strong> e roda na AWS com infraestrutura
              gerenciada por Terraform. Sugestões, issues e PRs são bem-vindos!
            </p>
            <div className="landing-contribute-actions">
              <a className="landing-cta-btn" href="https://github.com" target="_blank" rel="noreferrer">
                ★ Star no GitHub
              </a>
              <a className="landing-gh-btn" href="https://github.com" target="_blank" rel="noreferrer">
                Abrir issue ↗
              </a>
            </div>
          </div>
          <div className="landing-stack">
            {[
              { label: "Backend", val: "Go 1.25 · Gin · MySQL" },
              { label: "Frontend", val: "React 18 · Vite · JSX" },
              { label: "Infra", val: "AWS EC2 · Terraform" },
              { label: "API MTG", val: "Scryfall (sem chave)" },
              { label: "IA", val: "OpenAI GPT-4o" },
              { label: "SSL", val: "Let's Encrypt · Certbot" },
            ].map((s) => (
              <div key={s.label} className="landing-stack-item">
                <span className="landing-stack-label">{s.label}</span>
                <span className="landing-stack-val">{s.val}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── FOOTER ── */}
      <footer className="landing-footer">
        <span>⚔ Magic Collector · magic-collector.site</span>
        <span>Feito com ❤️ · Não afiliado à Wizards of the Coast</span>
      </footer>

      {/* ── MODAL LOGIN ── */}
      {showLogin && (
        <div className="modal-overlay" onClick={() => setShowLogin(false)}>
          <div className="modal landing-login-modal" onClick={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => setShowLogin(false)}>✕</button>
            <div className="landing-login-header">
              <span className="landing-login-icon">⚔</span>
              <h2>Entrar no Magic Collector</h2>
              <p>Use suas credenciais para acessar a coleção</p>
            </div>
            <form className="landing-login-form" onSubmit={handleLogin}>
              <label>
                Usuário
                <input
                  autoFocus
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="gabriel ou juliana"
                  required
                />
              </label>
              <label>
                Senha
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  required
                />
              </label>
              {loginError && <p className="landing-login-error">{loginError}</p>}
              <button type="submit" className="landing-cta-btn" disabled={loginLoading}>
                {loginLoading ? "Entrando…" : "Entrar →"}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
