# ── ACM: certificado TLS (obrigatório em us-east-1 para CloudFront) ────────────
resource "aws_acm_certificate" "app" {
  domain_name               = var.domain_name
  subject_alternative_names = ["www.${var.domain_name}"]
  validation_method         = "DNS"

  lifecycle {
    create_before_destroy = true
  }

  tags = { Project = var.app_name }
}

# Registros DNS de validação do certificado
resource "aws_route53_record" "cert_validation" {
  for_each = {
    for dvo in aws_acm_certificate.app.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  zone_id = aws_route53_zone.main.zone_id
  name    = each.value.name
  type    = each.value.type
  ttl     = 60
  records = [each.value.record]
}

resource "aws_acm_certificate_validation" "app" {
  certificate_arn         = aws_acm_certificate.app.arn
  validation_record_fqdns = [for r in aws_route53_record.cert_validation : r.fqdn]
}

# ── CloudFront Function: remove prefixo /api antes de encaminhar ao Lambda ─────
resource "aws_cloudfront_function" "api_rewrite" {
  name    = "${var.app_name}-api-rewrite"
  runtime = "cloudfront-js-2.0"
  comment = "Strip /api prefix para a origin Lambda"
  publish = true

  code = <<-JS
    function handler(event) {
      var request = event.request;
      var uri = request.uri;
      // Remove o prefixo /api antes de encaminhar ao Lambda
      if (uri === '/api' || uri === '/api/') {
        request.uri = '/';
      } else if (uri.slice(0, 4) === '/api') {
        request.uri = uri.slice(4);
      }
      return request;
    }
  JS
}

# ── CloudFront Distribution ────────────────────────────────────────────────────
locals {
  # API Gateway invoke_url: "https://XXXX.execute-api.us-east-1.amazonaws.com"
  # CloudFront exige só o hostname, sem protocolo nem barra final.
  api_origin_domain = trimsuffix(trimprefix(aws_apigatewayv2_stage.default.invoke_url, "https://"), "/")
}

resource "aws_cloudfront_distribution" "app" {
  enabled             = true
  is_ipv6_enabled     = true
  comment             = "Magic Collector — ${var.domain_name}"
  default_root_object = "index.html"
  aliases             = [var.domain_name, "www.${var.domain_name}"]
  price_class         = "PriceClass_100" # EUA + Europa — mais barato

  # ── Origin 1: S3 (frontend estático) ────────────────────────────────────────
  origin {
    origin_id                = "s3-frontend"
    domain_name              = aws_s3_bucket.frontend.bucket_regional_domain_name
    origin_access_control_id = aws_cloudfront_origin_access_control.frontend.id
  }

  # ── Origin 2: API Gateway (backend API) ───────────────────────────────────
  origin {
    origin_id   = "lambda-api"
    domain_name = local.api_origin_domain

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
      origin_read_timeout      = 60
      origin_keepalive_timeout = 60
    }
  }

  # ── Behavior padrão: frontend S3 ─────────────────────────────────────────
  default_cache_behavior {
    target_origin_id       = "s3-frontend"
    viewer_protocol_policy = "redirect-to-https"
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    compress               = true

    # Cache longo para assets estáticos
    cache_policy_id = "658327ea-f89d-4fab-a63d-7e88639e58f6" # CachingOptimized (AWS managed)
  }

  # ── Behavior /api/*: Lambda (sem cache, forwarda Authorization) ───────────
  ordered_cache_behavior {
    path_pattern           = "/api/*"
    target_origin_id       = "lambda-api"
    viewer_protocol_policy = "redirect-to-https"
    allowed_methods        = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods         = ["GET", "HEAD"]
    compress               = true

    # Sem cache para chamadas de API dinâmicas
    cache_policy_id          = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # CachingDisabled (AWS managed)
    origin_request_policy_id = "b689b0a8-53d0-40ab-baf2-68738e2966ac" # AllViewerExceptHostHeader

    function_association {
      event_type   = "viewer-request"
      function_arn = aws_cloudfront_function.api_rewrite.arn
    }
  }

  # Sem custom_error_response: esses blocos se aplicam a TODOS os origins,
  # então 403/404 da Lambda seriam servidos como index.html — quebrando a API.
  # O app usa roteamento por estado (não por URL), então o default_root_object
  # acima é suficiente para o SPA.

  viewer_certificate {
    acm_certificate_arn      = aws_acm_certificate_validation.app.certificate_arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  tags = { Project = var.app_name }
}
