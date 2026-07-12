variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "app_name" {
  description = "Prefixo usado nos nomes dos recursos AWS"
  type        = string
  default     = "magic-collector"
}

variable "domain_name" {
  description = "Domínio registrado para a aplicação (ex: magic-collector.site)"
  type        = string
  default     = "magic-collector.site"
}

# ── Banco de dados (MySQL externo) ────────────────────────────────────────────
variable "db_host" {
  description = "Host do MySQL externo"
  type        = string
}

variable "db_port" {
  description = "Porta do MySQL"
  type        = string
  default     = "3306"
}

variable "db_user" {
  description = "Usuário do MySQL"
  type        = string
}

variable "db_password" {
  description = "Senha do MySQL"
  type        = string
  sensitive   = true
}

variable "db_name" {
  description = "Nome do banco de dados"
  type        = string
}

# ── Secrets ───────────────────────────────────────────────────────────────────
variable "openai_api_key" {
  description = "Chave da API OpenAI para avaliações de IA"
  type        = string
  sensitive   = true
  default     = ""
}
