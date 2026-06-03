#!/bin/bash
# deploy.sh — faz deploy do Magic Collector no servidor AWS
# Uso: ./deploy.sh <ip-do-servidor>
# Exemplo: ./deploy.sh 34.196.63.122

set -euo pipefail

SERVER_IP="${1:-34.196.63.122}"
SSH_KEY="${SSH_KEY:-$HOME/.ssh/id_rsa}"
REMOTE_USER="ubuntu"
REMOTE_DIR="/opt/magic-collector"

if [[ ! -f .env ]]; then
  echo "ERRO: arquivo .env não encontrado."
  exit 1
fi

SSH_OPTS="-i $SSH_KEY -o StrictHostKeyChecking=accept-new -o ConnectTimeout=15"
REMOTE="$REMOTE_USER@$SERVER_IP"

echo "==> Aguardando servidor estar pronto..."
for i in $(seq 1 20); do
  if ssh $SSH_OPTS "$REMOTE" "echo ok" &>/dev/null; then break; fi
  echo "    tentativa $i/20, aguardando 15s..."
  sleep 15
done

echo "==> Enviando arquivos para $SERVER_IP..."
rsync -az --progress \
  --exclude '.git' \
  --exclude 'node_modules' \
  --exclude 'terraform/.terraform' \
  --exclude 'terraform/terraform.tfvars' \
  --exclude 'terraform/terraform.tfstate*' \
  --exclude 'data/' \
  --exclude '.env' \
  -e "ssh $SSH_OPTS" \
  . "$REMOTE:$REMOTE_DIR/"

echo "==> Enviando .env..."
scp $SSH_OPTS .env "$REMOTE:$REMOTE_DIR/.env"

echo "==> Buildando e subindo containers..."
ssh $SSH_OPTS "$REMOTE" bash <<'SSHEOF'
set -euo pipefail
cd /opt/magic-collector
docker compose -f docker-compose.prod.yml up --build -d --remove-orphans
echo "--- Containers: ---"
docker compose -f docker-compose.prod.yml ps
SSHEOF

echo ""
echo "✓ Deploy concluído!"
echo "  https://magic-collector.site"
echo "  Health: https://magic-collector.site/api/health"
echo ""
echo "  Logs: ssh ubuntu@$SERVER_IP 'docker compose -f /opt/magic-collector/docker-compose.prod.yml logs -f'"
