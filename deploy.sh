#!/bin/bash
# deploy.sh — sobe o Magic Collector no servidor AWS
# Uso: ./deploy.sh <ip-do-servidor>
# Exemplo: ./deploy.sh 54.123.45.67

set -euo pipefail

SERVER_IP="${1:-}"
SSH_KEY="${SSH_KEY:-$HOME/.ssh/id_rsa}"
REMOTE_USER="ubuntu"
REMOTE_DIR="/opt/magic-collector"

if [[ -z "$SERVER_IP" ]]; then
  echo "Uso: $0 <ip-do-servidor>"
  echo "Dica: obtenha o IP com: cd terraform && terraform output public_ip"
  exit 1
fi

if [[ ! -f .env ]]; then
  echo "ERRO: arquivo .env não encontrado."
  echo "Crie o .env com as credenciais do banco de dados Locaweb."
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
  --exclude 'data/' \
  --exclude '.env' \
  -e "ssh $SSH_OPTS" \
  . "$REMOTE:$REMOTE_DIR/"

echo "==> Enviando .env (credenciais do banco)..."
scp $SSH_OPTS .env "$REMOTE:$REMOTE_DIR/.env"

echo "==> Buildando containers (output completo)..."
ssh $SSH_OPTS "$REMOTE" bash <<'SSHEOF'
set -euo pipefail
cd /opt/magic-collector

docker compose -f docker-compose.prod.yml build --progress=plain 2>&1
SSHEOF

echo "==> Subindo containers..."
ssh $SSH_OPTS "$REMOTE" bash <<'SSHEOF'
set -euo pipefail
cd /opt/magic-collector

docker compose -f docker-compose.prod.yml up -d --remove-orphans

echo "--- Status dos containers: ---"
docker compose -f docker-compose.prod.yml ps
SSHEOF

AUTH_USER=$(grep -E '^AUTH_USER=' .env | cut -d'=' -f2- | tr -d '"'"'" 2>/dev/null || true)
AUTH_PASS=$(grep -E '^AUTH_PASS=' .env | cut -d'=' -f2- | tr -d '"'"'" 2>/dev/null || true)

if [[ -n "$AUTH_USER" && -n "$AUTH_PASS" ]]; then
  echo "==> Configurando autenticação (HTTP Basic Auth)..."
  # Escapa aspas simples no password para uso no heredoc
  ESCAPED_PASS="${AUTH_PASS//\'/\'\\\'\'}"
  ssh $SSH_OPTS "$REMOTE" bash << SSHEOF
set -e
echo '${AUTH_USER}':"\$(openssl passwd -apr1 '${ESCAPED_PASS}')" | sudo tee /etc/nginx/.htpasswd > /dev/null
sudo chmod 640 /etc/nginx/.htpasswd
sudo chown root:www-data /etc/nginx/.htpasswd

sudo tee /etc/nginx/sites-available/magic-collector > /dev/null << 'NGINXEOF'
server {
    listen 80 default_server;
    server_name _;
    server_tokens off;

    client_max_body_size 10m;

    auth_basic " ";
    auth_basic_user_file /etc/nginx/.htpasswd;

    location /api/ {
        proxy_pass         http://127.0.0.1:8080/;
        proxy_http_version 1.1;
        proxy_set_header   Host              \$host;
        proxy_set_header   X-Real-IP         \$remote_addr;
        proxy_set_header   X-Forwarded-For   \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;
        proxy_connect_timeout 60s;
        proxy_read_timeout    300s;
    }

    location / {
        proxy_pass         http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header   Host              \$host;
        proxy_set_header   X-Real-IP         \$remote_addr;
        proxy_set_header   X-Forwarded-For   \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;
    }
}
NGINXEOF

sudo nginx -t && sudo nginx -s reload
echo "  Auth configurada para usuário: ${AUTH_USER}"
SSHEOF
else
  echo "  (AUTH_USER/AUTH_PASS não definidos no .env — sem autenticação)"
fi

echo ""
echo "✓ Deploy concluído!"
echo "  App: http://$SERVER_IP"
echo "  Backend health: http://$SERVER_IP/api/health"
echo ""
echo "  Para ver logs:"
echo "  ssh ubuntu@$SERVER_IP 'docker compose -f /opt/magic-collector/docker-compose.prod.yml logs -f'"
