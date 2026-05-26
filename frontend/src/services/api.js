const BASE_URL = "/api";

export async function listCards({ q = "", page = 1, pageSize = 20, sort = "name", order = "asc", deckId } = {}) {
  const params = new URLSearchParams({ q, page, page_size: pageSize, sort, order });
  if (deckId !== undefined) params.set("deck_id", deckId);
  const res = await fetch(`${BASE_URL}/cards?${params}`);
  if (!res.ok) return { data: [], total: 0, page: 1, page_size: pageSize, total_pages: 1 };
  return res.json();
}

export async function createCard(data) {
  const res = await fetch(`${BASE_URL}/cards`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Erro ao cadastrar carta");
  return res.json();
}

export async function getCard(id) {
  const res = await fetch(`${BASE_URL}/cards/${id}`);
  if (!res.ok) throw new Error("Carta não encontrada");
  return res.json();
}

export async function updateCard(id, data) {
  const res = await fetch(`${BASE_URL}/cards/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Erro ao atualizar carta");
  return res.json();
}

export async function assignCardToDeck(cardId, deckId) {
  const res = await fetch(`${BASE_URL}/cards/${cardId}/deck`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ deck_id: deckId }),
  });
  if (!res.ok) throw new Error("Erro ao atribuir deck");
  return res.json();
}

export async function deleteCard(id) {
  const res = await fetch(`${BASE_URL}/cards/${id}`, { method: "DELETE" });
  if (!res.ok) throw new Error("Erro ao remover carta");
  return res.json();
}

export async function exportCards() {
  const res = await fetch(`${BASE_URL}/cards/export`);
  if (!res.ok) throw new Error("Erro ao exportar coleção");
  return res.json();
}

export async function listDecks() {
  const res = await fetch(`${BASE_URL}/decks`);
  if (!res.ok) return [];
  return res.json();
}

export async function createDeck(data) {
  const res = await fetch(`${BASE_URL}/decks`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Erro ao criar deck");
  return res.json();
}

export async function updateDeck(id, data) {
  const res = await fetch(`${BASE_URL}/decks/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Erro ao atualizar deck");
  return res.json();
}

export async function deleteDeck(id) {
  const res = await fetch(`${BASE_URL}/decks/${id}`, { method: "DELETE" });
  if (!res.ok) throw new Error("Erro ao remover deck");
  return res.json();
}
