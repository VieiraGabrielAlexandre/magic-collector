const BASE_URL = "/api";

export async function listCards({ q = "", page = 1, pageSize = 20, sort = "name", order = "asc" } = {}) {
  const params = new URLSearchParams({ q, page, page_size: pageSize, sort, order });
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

export async function deleteCard(id) {
  const res = await fetch(`${BASE_URL}/cards/${id}`, { method: "DELETE" });
  if (!res.ok) throw new Error("Erro ao remover carta");
  return res.json();
}
