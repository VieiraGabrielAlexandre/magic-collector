#!/bin/bash
set -euxo pipefail
exec > /var/log/userdata.log 2>&1

# ── Atualizar sistema ────────────────────────────────────────────────────
export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get upgrade -y

# ── Instalar dependências base ───────────────────────────────────────────
apt-get install -y \
  ca-certificates \
  curl \
  gnupg \
  nginx \
  git \
  unzip \
  awscli

# ── Instalar Docker (repositório oficial) ───────────────────────────────
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
  | tee /etc/apt/sources.list.d/docker.list > /dev/null

apt-get update -y
apt-get install -y \
  docker-ce \
  docker-ce-cli \
  containerd.io \
  docker-buildx-plugin \
  docker-compose-plugin

systemctl enable --now docker
usermod -aG docker ubuntu

# ── Criar swap de 2GB (evita OOM durante builds em t3.micro) ────────────
if [ ! -f /swapfile ]; then
  fallocate -l 2G /swapfile
  chmod 600 /swapfile
  mkswap /swapfile
  swapon /swapfile
  echo '/swapfile none swap sw 0 0' >> /etc/fstab
fi

# ── Criar diretório da aplicação ─────────────────────────────────────────
mkdir -p /opt/magic-collector
chown ubuntu:ubuntu /opt/magic-collector

# ── Configurar Nginx como reverse proxy ─────────────────────────────────
cat > /etc/nginx/sites-available/magic-collector <<'NGINXEOF'
server {
    listen 80 default_server;
    server_name _;

    client_max_body_size 10m;

    # Proxy /api/* → backend (remove prefixo /api)
    location /api/ {
        proxy_pass         http://127.0.0.1:8080/;
        proxy_http_version 1.1;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_connect_timeout 60s;
        proxy_read_timeout    300s;
    }

    # Tudo mais → frontend (React SPA)
    location / {
        proxy_pass         http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }
}
NGINXEOF

ln -sf /etc/nginx/sites-available/magic-collector /etc/nginx/sites-enabled/magic-collector
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl enable --now nginx

echo "=== userdata.sh concluído ==="
