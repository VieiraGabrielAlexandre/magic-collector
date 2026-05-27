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
  value       = "http://${aws_eip.app.public_ip}"
}

output "ssh_command" {
  description = "Comando para conectar via SSH"
  value       = "ssh -i ~/.ssh/id_rsa ubuntu@${aws_eip.app.public_ip}"
}

output "instance_id" {
  description = "ID da instância EC2"
  value       = aws_instance.app.id
}
