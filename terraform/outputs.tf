output "public_ip" {
  description = "IP público do servidor (Elastic IP)"
  value       = aws_eip.app.public_ip
}

output "public_dns" {
  description = "DNS público do servidor"
  value       = aws_eip.app.public_dns
}

output "app_url" {
  description = "URL da aplicação"
  value       = "https://${var.domain_name}"
}

output "ssh_command" {
  description = "Comando para conectar via SSH"
  value       = "ssh -i ~/.ssh/id_rsa ubuntu@${aws_eip.app.public_ip}"
}

output "instance_id" {
  description = "ID da instância EC2"
  value       = aws_instance.app.id
}

output "route53_nameservers" {
  description = "Nameservers do Route 53 — configure esses NS no seu registrador de domínio"
  value       = aws_route53_zone.main.name_servers
}

output "next_steps" {
  description = "Próximos passos após o terraform apply"
  value       = <<-EOT
    1. Copie os nameservers acima e configure-os no seu registrador (magic-collector.site).
    2. Aguarde a propagação DNS (5–30 min). Teste com: dig magic-collector.site NS
    3. Execute o script de SSL na instância:
         scp scripts/setup-ssl.sh ubuntu@${aws_eip.app.public_ip}:~
         ssh ubuntu@${aws_eip.app.public_ip} "bash ~/setup-ssl.sh"
  EOT
}
