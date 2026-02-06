# Run via docker
run:
	@docker compose up --remove-orphans --build -d

# Generates http server from openapi spec.
gen.oapi.servers:
	@oapi-codegen -config configs/urlshortener.oapi.server.yaml api/openapi/v1/urlshortener.openapi.yaml

# Generate mocks:
gen.mockery:
	@mockery

test:
	@go test ./...