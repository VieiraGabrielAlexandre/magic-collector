# ── Build do binário Go para arm64 (Graviton — 20% mais barato que x86) ───────
# O null_resource é disparado sempre que qualquer arquivo .go muda.
resource "null_resource" "build_lambda" {
  triggers = {
    src_hash = sha256(join("", [
      for f in sort(fileset("${path.module}/../backend", "**/*.go")) :
      filesha256("${path.module}/../backend/${f}")
    ]))
  }

  provisioner "local-exec" {
    command = <<-EOT
      set -e
      echo "==> Compilando backend para arm64..."
      cd "${path.module}/../backend"
      GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build \
        -trimpath \
        -ldflags="-s -w" \
        -o bootstrap \
        ./cmd/lambda
      zip -j "${path.module}/lambda_package.zip" bootstrap
      rm -f bootstrap
      echo "==> lambda_package.zip gerado."
    EOT
    interpreter = ["/bin/bash", "-c"]
  }
}

# ── IAM: role de execução do Lambda ───────────────────────────────────────────
data "aws_iam_policy_document" "lambda_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda" {
  name               = "${var.app_name}-lambda-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json

  tags = { Project = var.app_name }
}

resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# ── CloudWatch Log Group (retenção 30 dias) ────────────────────────────────────
resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${var.app_name}"
  retention_in_days = 30

  tags = { Project = var.app_name }
}

# ── Função Lambda ──────────────────────────────────────────────────────────────
resource "aws_lambda_function" "app" {
  function_name = var.app_name
  description   = "Magic Collector — backend Go/Gin"

  filename         = "${path.module}/lambda_package.zip"
  source_code_hash = fileexists("${path.module}/lambda_package.zip") ? filebase64sha256("${path.module}/lambda_package.zip") : ""

  # provided.al2023 + arm64 = Graviton2 (melhor custo/perf para Go)
  runtime       = "provided.al2023"
  architectures = ["arm64"]
  handler       = "bootstrap"

  role = aws_iam_role.lambda.arn

  # 256 MB é suficiente para o Go + Gin; timeout longo para chamadas de IA (~90s)
  memory_size = 256
  timeout     = 300

  environment {
    variables = {
      DB_HOST        = var.db_host
      DB_PORT        = var.db_port
      DB_USER        = var.db_user
      DB_PASSWORD    = var.db_password
      DB_NAME        = var.db_name
      OPENAI_API_KEY = var.openai_api_key
      GIN_MODE       = "release"
    }
  }

  depends_on = [
    null_resource.build_lambda,
    aws_cloudwatch_log_group.lambda,
  ]

  tags = { Project = var.app_name }
}

# Lambda Function URL removida — a conta AWS bloqueia acesso público a Function URLs.
# O backend é exposto via API Gateway HTTP API (ver apigateway.tf).
