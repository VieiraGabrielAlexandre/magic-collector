output "app_url" {
  description = "URL pública da aplicação (via CloudFront)"
  value       = "https://${var.domain_name}"
}

output "cloudfront_domain" {
  description = "Domínio do CloudFront (use para debug se o domínio próprio não propagar)"
  value       = aws_cloudfront_distribution.app.domain_name
}

output "cloudfront_distribution_id" {
  description = "ID da distribuição CloudFront (útil para invalidações manuais)"
  value       = aws_cloudfront_distribution.app.id
}

output "api_gateway_url" {
  description = "URL direta do API Gateway (para debug — use /api/* em produção via CloudFront)"
  value       = aws_apigatewayv2_stage.default.invoke_url
}

output "frontend_bucket" {
  description = "Nome do bucket S3 com o frontend"
  value       = aws_s3_bucket.frontend.bucket
}

output "route53_nameservers" {
  description = "Nameservers do Route 53 — configure esses NS no seu registrador de domínio"
  value       = aws_route53_zone.main.name_servers
}

output "estimated_monthly_cost" {
  description = "Custo mensal estimado (uso baixo, dentro do free tier Lambda/CloudFront)"
  value       = "~$0.51/mês (Route53 $0.50 + S3 <$0.01 + Lambda/CloudFront: free tier)"
}

output "next_steps" {
  description = "Próximos passos após o terraform apply"
  value       = <<-EOT
    1. Se os nameservers mudaram, atualize no registrador do domínio.
       Verifique: dig ${var.domain_name} NS
    2. O certificado TLS é emitido automaticamente via ACM + Route53.
    3. Acesse: https://${var.domain_name}
    4. Para forçar redeploy do backend: terraform apply -replace=null_resource.build_lambda
    5. Para forçar redeploy do frontend: terraform apply -replace=null_resource.deploy_frontend
  EOT
}
