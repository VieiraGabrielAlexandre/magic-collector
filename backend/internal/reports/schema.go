package reports

// JSONSchema is the JSON Schema (draft-07) that describes CollectionReport.
// Served at GET /reports/collection/schema.
const JSONSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://magic-collector/schemas/collection-report/1.0",
  "title": "CollectionReport",
  "description": "Relatório completo da coleção Magic: The Gathering",
  "type": "object",
  "required": ["generated_at", "schema_version", "summary", "cards_without_deck_by_color",
               "top_expensive_cards", "recent_decks", "recent_cards", "message"],
  "properties": {
    "message": {
      "type": "string",
      "minLength": 1,
      "maxLength": 4096,
      "description": "Texto pronto para WhatsApp (emojis e marcação *negrito*/_itálico_), enviado sem alteração pelo ArcanePulse"
    },
    "generated_at": {
      "type": "string",
      "format": "date-time",
      "description": "Momento de geração do relatório (RFC 3339 / ISO 8601)"
    },
    "schema_version": {
      "type": "string",
      "description": "Versão do schema deste relatório",
      "example": "1.0"
    },
    "summary": {
      "type": "object",
      "description": "Números agregados de toda a coleção",
      "required": ["total_cards", "total_unique_cards", "total_decks",
                   "total_value_usd", "cards_without_deck", "cards_in_deck",
                   "foil_count", "rarity_breakdown"],
      "properties": {
        "total_cards": {
          "type": "integer",
          "description": "Soma de quantity de todas as cartas (cópias totais)"
        },
        "total_unique_cards": {
          "type": "integer",
          "description": "Número de registros únicos na coleção"
        },
        "total_decks": {
          "type": "integer",
          "description": "Número de decks cadastrados"
        },
        "total_value_usd": {
          "type": "number",
          "description": "Valor total estimado da coleção em USD (price_usd × quantity)"
        },
        "cards_without_deck": {
          "type": "integer",
          "description": "Cartas únicas sem deck associado (deck_id = 0)"
        },
        "cards_in_deck": {
          "type": "integer",
          "description": "Cartas únicas com deck associado"
        },
        "foil_count": {
          "type": "integer",
          "description": "Número de cartas foil na coleção"
        },
        "rarity_breakdown": {
          "type": "object",
          "description": "Contagem de cartas por raridade (M/R/U/C/L/T)",
          "additionalProperties": { "type": "integer" },
          "example": { "M": 12, "R": 48, "U": 130, "C": 220, "L": 25, "T": 8 }
        }
      }
    },
    "cards_without_deck_by_color": {
      "type": "array",
      "description": "Cartas sem deck agrupadas por cor principal",
      "items": {
        "type": "object",
        "required": ["color", "color_pt", "count", "quantity"],
        "properties": {
          "color":    { "type": "string", "description": "Código da cor (W/U/B/R/G/Multi/C)" },
          "color_pt": { "type": "string", "description": "Nome da cor em português" },
          "count":    { "type": "integer", "description": "Cartas únicas nessa cor sem deck" },
          "quantity": { "type": "integer", "description": "Total de cópias nessa cor sem deck" }
        }
      }
    },
    "top_expensive_cards": {
      "type": "array",
      "description": "As cartas mais caras da coleção (máximo 10, ordenadas por price_usd desc)",
      "items": {
        "type": "object",
        "required": ["id", "name", "set_code", "rarity", "price_usd", "foil", "image_url", "deck_id"],
        "properties": {
          "id":        { "type": "integer" },
          "name":      { "type": "string" },
          "set_code":  { "type": "string" },
          "rarity":    { "type": "string", "enum": ["M","R","U","C","L","T",""] },
          "price_usd": { "type": "number", "minimum": 0 },
          "foil":      { "type": "boolean" },
          "image_url": { "type": "string", "format": "uri" },
          "deck_id":   { "type": "integer", "description": "0 = sem deck" },
          "deck_name": { "type": "string",  "description": "Nome do deck, se associado" }
        }
      }
    },
    "recent_decks": {
      "type": "array",
      "description": "Os 5 decks criados mais recentemente",
      "items": {
        "type": "object",
        "required": ["id", "name", "colors", "card_count", "estimated_value_usd", "created_at"],
        "properties": {
          "id":                   { "type": "integer" },
          "name":                 { "type": "string" },
          "colors":               { "type": "string", "description": "Códigos separados por vírgula, ex: W,U,B" },
          "card_count":           { "type": "integer" },
          "estimated_value_usd":  { "type": "number" },
          "created_at":           { "type": "string", "format": "date-time" }
        }
      }
    },
    "recent_cards": {
      "type": "array",
      "description": "As 10 cartas adicionadas mais recentemente",
      "items": {
        "type": "object",
        "required": ["id", "name", "set_code", "rarity", "foil", "image_url", "added_at"],
        "properties": {
          "id":        { "type": "integer" },
          "name":      { "type": "string" },
          "set_code":  { "type": "string" },
          "rarity":    { "type": "string", "enum": ["M","R","U","C","L","T",""] },
          "foil":      { "type": "boolean" },
          "image_url": { "type": "string", "format": "uri" },
          "added_at":  { "type": "string", "format": "date-time" }
        }
      }
    }
  }
}`
