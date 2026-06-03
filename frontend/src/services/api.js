const BASE_URL = "/api";

// ── Auth helpers ─────────────────────────────────────────────────────────────

function authHeaders(extra = {}) {
  const token = localStorage.getItem("auth_token");
  const headers = { "Content-Type": "application/json", ...extra };
  if (token) headers["Authorization"] = `Bearer ${token}`;
  return headers;
}

function authFetch(url, options = {}) {
  const token = localStorage.getItem("auth_token");
  const headers = { ...(options.headers || {}) };
  if (token) headers["Authorization"] = `Bearer ${token}`;
  return fetch(url, { ...options, headers });
}

// ── Auth API ─────────────────────────────────────────────────────────────────

export async function login(username, password) {
  const res = await fetch(`${BASE_URL}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  const json = await res.json();
  if (!res.ok) throw new Error(json.error || "Credenciais inválidas");
  return json; // { token, user, session_created_at }
}

export async function logout() {
  await authFetch(`${BASE_URL}/auth/logout`, { method: "POST" });
  localStorage.removeItem("auth_token");
  localStorage.removeItem("auth_session_created_at");
}

export async function getMe() {
  const res = await authFetch(`${BASE_URL}/auth/me`);
  if (!res.ok) return null;
  return res.json(); // { user, session_created_at }
}

// ── Cards ────────────────────────────────────────────────────────────────────

export async function listCards({ q = "", page = 1, pageSize = 20, sort = "name", order = "asc", deckId, foil, rarity, colors } = {}) {
  const params = new URLSearchParams({ q, page, page_size: pageSize, sort, order });
  if (deckId !== undefined) params.set("deck_id", deckId);
  if (foil) params.set("foil", "1");
  if (rarity) params.set("rarity", rarity);
  if (colors) params.set("colors", colors);
  const res = await authFetch(`${BASE_URL}/cards?${params}`);
  if (!res.ok) return { data: [], total: 0, page: 1, page_size: pageSize, total_pages: 1 };
  return res.json();
}

export async function createCard(data) {
  const res = await authFetch(`${BASE_URL}/cards`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Erro ao cadastrar carta");
  return res.json();
}

export async function getCard(id) {
  const res = await authFetch(`${BASE_URL}/cards/${id}`);
  if (!res.ok) throw new Error("Carta não encontrada");
  return res.json();
}

export async function updateCard(id, data) {
  const res = await authFetch(`${BASE_URL}/cards/${id}`, {
    method: "PUT",
    headers: authHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Erro ao atualizar carta");
  return res.json();
}

export async function assignCardToDeck(cardId, deckId) {
  const res = await authFetch(`${BASE_URL}/cards/${cardId}/deck`, {
    method: "PATCH",
    headers: authHeaders(),
    body: JSON.stringify({ deck_id: deckId }),
  });
  if (!res.ok) throw new Error("Erro ao atribuir deck");
  return res.json();
}

export async function deleteCard(id) {
  const res = await authFetch(`${BASE_URL}/cards/${id}`, { method: "DELETE" });
  if (!res.ok) throw new Error("Erro ao remover carta");
  return res.json();
}

export async function exportCards() {
  const res = await authFetch(`${BASE_URL}/cards/export`);
  if (!res.ok) throw new Error("Erro ao exportar coleção");
  return res.json();
}

export async function listColorCombos() {
  const res = await authFetch(`${BASE_URL}/cards/colors`);
  if (!res.ok) return [];
  return res.json();
}

export async function updateCardQuantity(id, quantity) {
  await authFetch(`${BASE_URL}/cards/${id}/quantity`, {
    method: "PATCH",
    headers: authHeaders(),
    body: JSON.stringify({ quantity }),
  });
}

export async function previewCard(data) {
  const res = await authFetch(`${BASE_URL}/cards/preview`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify(data),
  });
  const json = await res.json();
  if (!res.ok) throw new Error(json.error || "Erro ao buscar carta");
  return json;
}

export async function refreshImages() {
  const res = await authFetch(`${BASE_URL}/cards/refresh-images`, { method: "POST" });
  const json = await res.json();
  if (!res.ok) throw new Error(json.error || "Erro ao atualizar imagens");
  return json;
}

export async function refreshPrices({ emptyOnly = false } = {}) {
  const url = `${BASE_URL}/cards/refresh-prices${emptyOnly ? "?empty_only=1" : ""}`;
  const res = await authFetch(url, { method: "POST" });
  const json = await res.json();
  if (!res.ok) throw new Error(json.error || "Erro ao atualizar preços");
  return json;
}

export async function getCollectionStats() {
  const res = await authFetch(`${BASE_URL}/cards/stats`);
  if (!res.ok) return null;
  return res.json();
}

export async function suggestDecks(params = {}) {
  const res = await authFetch(`${BASE_URL}/cards/suggest-decks`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify(params),
  });
  const json = await res.json();
  if (!res.ok) throw new Error(json.error || "Erro ao gerar sugestão");
  return json;
}

// ── Decks ─────────────────────────────────────────────────────────────────────

export async function listDecks() {
  const res = await authFetch(`${BASE_URL}/decks`);
  if (!res.ok) return [];
  return res.json();
}

export async function createDeck(data) {
  const res = await authFetch(`${BASE_URL}/decks`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Erro ao criar deck");
  return res.json();
}

export async function updateDeck(id, data) {
  const res = await authFetch(`${BASE_URL}/decks/${id}`, {
    method: "PUT",
    headers: authHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Erro ao atualizar deck");
  return res.json();
}

export async function deleteDeck(id) {
  const res = await authFetch(`${BASE_URL}/decks/${id}`, { method: "DELETE" });
  if (!res.ok) throw new Error("Erro ao remover deck");
  return res.json();
}

export async function fetchDeckIcon(id) {
  const res = await authFetch(`${BASE_URL}/decks/${id}/icon`, { method: "PATCH" });
  if (!res.ok) return null;
  return res.json();
}

export async function evaluateDeck(id) {
  const res = await authFetch(`${BASE_URL}/decks/${id}/evaluate`, { method: "POST" });
  const json = await res.json();
  if (!res.ok) throw new Error(json.error || "Erro ao gerar avaliação");
  return json;
}

export async function importDeckList(data) {
  const res = await authFetch(`${BASE_URL}/decks/import-list`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify(data),
  });
  const json = await res.json();
  if (!res.ok) throw new Error(json.error || "Erro ao importar lista");
  return json;
}

export async function importPrecon(data) {
  const res = await authFetch(`${BASE_URL}/decks/import-precon`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify(data),
  });
  const json = await res.json();
  if (!res.ok) throw new Error(json.error || "Erro ao importar pré-con");
  return json;
}

// ── Battles ───────────────────────────────────────────────────────────────────

export async function listBattles() {
  const res = await authFetch(`${BASE_URL}/battles`);
  if (!res.ok) return [];
  return res.json();
}

export async function createBattle(data) {
  const res = await authFetch(`${BASE_URL}/battles`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Erro ao registrar batalha");
  return res.json();
}

export async function deleteBattle(id) {
  const res = await authFetch(`${BASE_URL}/battles/${id}`, { method: "DELETE" });
  if (!res.ok) throw new Error("Erro ao remover batalha");
  return res.json();
}
