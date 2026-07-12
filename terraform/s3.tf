# ── S3: bucket privado para o frontend estático ───────────────────────────────
resource "aws_s3_bucket" "frontend" {
  bucket = "${replace(var.domain_name, ".", "-")}-frontend"

  tags = { Project = var.app_name }
}

resource "aws_s3_bucket_public_access_block" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Habilita versionamento para facilitar rollback de deploys
resource "aws_s3_bucket_versioning" "frontend" {
  bucket = aws_s3_bucket.frontend.id
  versioning_configuration {
    status = "Enabled"
  }
}

# ── OAC: somente o CloudFront pode ler o bucket ────────────────────────────────
resource "aws_cloudfront_origin_access_control" "frontend" {
  name                              = "${var.app_name}-s3-oac"
  description                       = "OAC para bucket frontend do Magic Collector"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

# Política do bucket: acesso apenas via CloudFront OAC
data "aws_iam_policy_document" "s3_frontend_policy" {
  statement {
    sid    = "AllowCloudFrontServicePrincipal"
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["cloudfront.amazonaws.com"]
    }
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.frontend.arn}/*"]
    condition {
      test     = "StringEquals"
      variable = "AWS:SourceArn"
      values   = [aws_cloudfront_distribution.app.arn]
    }
  }
}

resource "aws_s3_bucket_policy" "frontend" {
  bucket = aws_s3_bucket.frontend.id
  policy = data.aws_iam_policy_document.s3_frontend_policy.json

  depends_on = [aws_s3_bucket_public_access_block.frontend]
}

# ── Build + deploy do frontend para o S3 ──────────────────────────────────────
resource "null_resource" "deploy_frontend" {
  triggers = {
    # Redeploy quando o bucket ou a distribuição mudar, ou ao forçar com -replace
    bucket          = aws_s3_bucket.frontend.bucket
    distribution_id = aws_cloudfront_distribution.app.id
    src_hash = sha256(join("", [
      for f in sort(fileset("${path.module}/../frontend/src", "**")) :
      filesha256("${path.module}/../frontend/src/${f}")
    ]))
  }

  provisioner "local-exec" {
    command = <<-EOT
      set -e
      echo "==> Build do frontend React..."
      cd "${path.module}/../frontend"
      npm ci --silent
      npm run build

      echo "==> Sync para S3..."
      aws s3 sync dist/ s3://${aws_s3_bucket.frontend.bucket}/ \
        --delete \
        --region ${var.aws_region}

      echo "==> Invalidando cache do CloudFront..."
      aws cloudfront create-invalidation \
        --distribution-id ${aws_cloudfront_distribution.app.id} \
        --paths "/*" \
        --region us-east-1 \
        --output text --query 'Invalidation.Id'

      echo "==> Deploy concluído."
    EOT
    interpreter = ["/bin/bash", "-c"]
  }

  depends_on = [
    aws_s3_bucket_policy.frontend,
    aws_cloudfront_distribution.app,
  ]
}
