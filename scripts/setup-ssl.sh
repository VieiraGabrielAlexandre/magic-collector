#!/bin/bash
# Execute este script no servidor após o DNS propagar para o IP do servidor.
# Pré-requisito: certbot e python3-certbot-nginx instalados (já no userdata.sh).
#
# Uso:
#   scp scripts/setup-ssl.sh ubuntu@<IP>:~
#   ssh ubuntu@<IP> "bash ~/setup-ssl.sh"

set -euo pipefail

DOMAIN="magic-collector.site"
EMAIL="gabrielvieira840@gmail.com"

echo "=== Verificando se o DNS já aponta para este servidor ==="
SERVER_IP=$(curl -s ifconfig.me)
DOMAIN_IP=$(dig +short "$DOMAIN" | tail -1)

if [ "$SERVER_IP" != "$DOMAIN_IP" ]; then
  echo "AVISO: $DOMAIN ainda aponta para $DOMAIN_IP, mas este servidor é $SERVER_IP"
  echo "Aguarde a propagação DNS e execute novamente."
  exit 1
fi

echo "=== DNS ok — $DOMAIN → $SERVER_IP ==="

# Instalar certbot caso não esteja presente (instâncias antigas)
if ! command -v certbot &>/dev/null; then
  apt-get update -y
  apt-get install -y certbot python3-certbot-nginx
fi

# Atualizar nginx para incluir server_name correto (caso seja instância antiga)
if ! grep -q "$DOMAIN" /etc/nginx/sites-available/magic-collector 2>/dev/null; then
  echo "=== Atualizando nginx para usar server_name $DOMAIN ==="
  sed -i "s/server_name _;/server_name $DOMAIN www.$DOMAIN;/" \
    /etc/nginx/sites-available/magic-collector
  nginx -t && systemctl reload nginx
fi

echo "=== Obtendo certificado Let's Encrypt ==="
certbot --nginx \
  --non-interactive \
  --agree-tos \
  --email "$EMAIL" \
  -d "$DOMAIN" \
  -d "www.$DOMAIN"

echo "=== Configurando renovação automática ==="
systemctl enable --now certbot.timer 2>/dev/null || true
# Fallback: cron de renovação
(crontab -l 2>/dev/null; echo "0 3 * * * certbot renew --quiet") | crontab -

echo "=== SSL configurado com sucesso! ==="
echo "Acesse: https://$DOMAIN"
