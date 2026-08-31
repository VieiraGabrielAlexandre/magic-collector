.PHONY: build-lambda build-frontend deploy plan destroy logs

# ── Backend Lambda ────────────────────────────────────────────────────────────
build-lambda:
	@echo "==> Compilando Go para arm64 (Graviton)..."
	cd backend && \
	  GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build \
	    -trimpath -ldflags="-s -w" \
	    -o bootstrap ./cmd/lambda
	zip -j terraform/lambda_package.zip backend/bootstrap
	rm -f backend/bootstrap
	@echo "==> lambda_package.zip criado."

# ── Frontend React ────────────────────────────────────────────────────────────
build-frontend:
	@echo "==> Build do React..."
	cd frontend && npm ci --silent && npm run build
	@echo "==> dist/ gerado."

# ── Deploy completo ───────────────────────────────────────────────────────────
deploy: build-lambda build-frontend
	@echo "==> Aplicando Terraform..."
	cd terraform && terraform apply

# ── Somente plan (sem aplicar) ────────────────────────────────────────────────
plan: build-lambda
	cd terraform && terraform plan

# ── Forçar redeploy do backend (sem rebuild de infra) ─────────────────────────
deploy-backend: build-lambda
	aws lambda update-function-code \
	  --function-name magic-collector \
	  --zip-file fileb://terraform/lambda_package.zip \
	  --region us-east-1

# ── Forçar redeploy do frontend ───────────────────────────────────────────────
deploy-frontend: build-frontend
	@BUCKET=$$(cd terraform && terraform output -raw frontend_bucket); \
	DIST_ID=$$(cd terraform && terraform output -raw cloudfront_distribution_id); \
	aws s3 sync frontend/dist/ s3://$$BUCKET/ --delete --region us-east-1; \
	aws cloudfront create-invalidation --distribution-id $$DIST_ID --paths "/*" --region us-east-1

# ── Logs do Lambda (últimas 5 min) ────────────────────────────────────────────
logs:
	aws logs tail /aws/lambda/magic-collector --since 5m --follow --region us-east-1

# ── Destroy (CUIDADO!) ────────────────────────────────────────────────────────
destroy:
	cd terraform && terraform destroy
