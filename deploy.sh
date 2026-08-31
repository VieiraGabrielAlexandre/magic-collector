#!/bin/bash
# deploy.sh — faz deploy do Magic Collector (Lambda + S3 + CloudFront)
# Uso: ./deploy.sh [--backend-only | --frontend-only | --infra-only]
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TERRAFORM_DIR="$ROOT_DIR/terraform"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"
LAMBDA_ZIP="$TERRAFORM_DIR/lambda_package.zip"

# ── Cores ─────────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
step()  { echo -e "\n${GREEN}==>${NC} $*"; }
warn()  { echo -e "${YELLOW}AVISO:${NC} $*"; }
die()   { echo -e "${RED}ERRO:${NC} $*" >&2; exit 1; }

# ── Modo de deploy ────────────────────────────────────────────────────────────
MODE="all"
case "${1:-}" in
  --backend-only)  MODE="backend"  ;;
  --frontend-only) MODE="frontend" ;;
  --infra-only)    MODE="infra"    ;;
  "")              MODE="all"      ;;
  *) die "Uso: ./deploy.sh [--backend-only | --frontend-only | --infra-only]" ;;
esac

# ── Pré-requisitos ────────────────────────────────────────────────────────────
step "Verificando pré-requisitos..."
command -v go        >/dev/null || die "go não encontrado"
command -v aws       >/dev/null || die "aws CLI não encontrado (brew install awscli)"
command -v terraform >/dev/null || die "terraform não encontrado (brew install terraform)"

if [[ "$MODE" == "all" || "$MODE" == "frontend" ]]; then
  command -v npm >/dev/null || die "npm não encontrado"
fi

# ── Build do backend Lambda ───────────────────────────────────────────────────
build_backend() {
  step "Compilando backend Go para arm64 (Graviton)..."
  cd "$BACKEND_DIR"
  GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o bootstrap \
    ./cmd/lambda
  zip -j "$LAMBDA_ZIP" bootstrap
  rm -f bootstrap
  echo "   lambda_package.zip: $(du -sh "$LAMBDA_ZIP" | cut -f1)"
}

# ── Build do frontend React ───────────────────────────────────────────────────
build_frontend() {
  step "Build do frontend React..."
  cd "$FRONTEND_DIR"
  npm ci --silent
  npm run build
  echo "   dist/: $(du -sh dist | cut -f1)"
}

# ── Terraform apply ───────────────────────────────────────────────────────────
infra_apply() {
  step "Aplicando infraestrutura com Terraform..."
  cd "$TERRAFORM_DIR"
  terraform init -upgrade -input=false -no-color 2>&1 | grep -E "provider|initialized|error" || true
  terraform apply -auto-approve -input=false
}

# ── Deploy rápido: só atualiza o código Lambda (sem terraform) ────────────────
deploy_backend_fast() {
  local FUNCTION_NAME REGION
  FUNCTION_NAME=$(cd "$TERRAFORM_DIR" && terraform output -raw cloudfront_distribution_id 2>/dev/null && echo "magic-collector" || echo "magic-collector")
  REGION="us-east-1"

  step "Atualizando código do Lambda diretamente (fast deploy)..."
  aws lambda update-function-code \
    --function-name magic-collector \
    --zip-file "fileb://$LAMBDA_ZIP" \
    --region "$REGION" \
    --output text --query 'FunctionName' \
    | xargs -I{} echo "   Função atualizada: {}"

  # Aguarda a atualização ficar ativa
  aws lambda wait function-updated \
    --function-name magic-collector \
    --region "$REGION"
}

# ── Deploy rápido: só atualiza o frontend no S3 ───────────────────────────────
deploy_frontend_fast() {
  local BUCKET DIST_ID
  BUCKET=$(cd "$TERRAFORM_DIR" && terraform output -raw frontend_bucket 2>/dev/null) \
    || die "Bucket não encontrado. Rode ./deploy.sh primeiro para criar a infra."
  DIST_ID=$(cd "$TERRAFORM_DIR" && terraform output -raw cloudfront_distribution_id 2>/dev/null)

  step "Sincronizando frontend para S3 (s3://$BUCKET)..."
  aws s3 sync "$FRONTEND_DIR/dist/" "s3://$BUCKET/" \
    --delete \
    --region us-east-1

  step "Invalidando cache do CloudFront ($DIST_ID)..."
  aws cloudfront create-invalidation \
    --distribution-id "$DIST_ID" \
    --paths "/*" \
    --region us-east-1 \
    --output text --query 'Invalidation.Id' \
    | xargs -I{} echo "   Invalidação: {}"
}

# ── Resumo final ──────────────────────────────────────────────────────────────
print_summary() {
  local APP_URL
  APP_URL=$(cd "$TERRAFORM_DIR" && terraform output -raw app_url 2>/dev/null || echo "https://magic-collector.site")

  echo ""
  echo -e "${GREEN}✓ Deploy concluído!${NC}"
  echo "  App:    $APP_URL"
  echo "  Health: $APP_URL/api/health"
  echo ""
  echo "  Logs do Lambda:"
  echo "  aws logs tail /aws/lambda/magic-collector --since 5m --follow --region us-east-1"
  echo ""
}

# ── Execução ──────────────────────────────────────────────────────────────────
case "$MODE" in
  all)
    build_backend
    build_frontend
    infra_apply
    print_summary
    ;;
  backend)
    build_backend
    # Se a infra já existe, fast deploy; senão, terraform apply
    if [[ -f "$TERRAFORM_DIR/terraform.tfstate" ]] && \
       cd "$TERRAFORM_DIR" && terraform output frontend_bucket &>/dev/null; then
      deploy_backend_fast
    else
      warn "Infra não encontrada, rodando terraform apply..."
      infra_apply
    fi
    print_summary
    ;;
  frontend)
    build_frontend
    if [[ -f "$TERRAFORM_DIR/terraform.tfstate" ]] && \
       cd "$TERRAFORM_DIR" && terraform output frontend_bucket &>/dev/null; then
      deploy_frontend_fast
    else
      warn "Infra não encontrada, rodando terraform apply..."
      infra_apply
    fi
    print_summary
    ;;
  infra)
    infra_apply
    print_summary
    ;;
esac
