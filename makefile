# Run via docker
run:
	@docker compose up --remove-orphans --build -d

# Generates http server from openapi spec.
oapi.gen.servers:
	@oapi-codegen -config configs/urlshortener.oapi.server.yaml api/openapi/v1/urlshortener.openapi.yaml

test:
	@go test ./...