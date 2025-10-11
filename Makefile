APP_NAME=certfix-agent
CONTAINER=certfix-agent-dev

# Build local (usando Go direto)
build:
	@echo "Compilando para Linux..."
	GOOS=linux GOARCH=amd64 go build -o build/$(APP_NAME) ./cmd

# Build dentro do container
docker-build:
	docker exec $(CONTAINER) go build -o build/$(APP_NAME) ./cmd
	@echo "Build dentro do container concluído!"

# Executa o agent dentro do container
docker-run:
	docker exec -it $(CONTAINER) ./build/$(APP_NAME)

# Sobe o ambiente de desenvolvimento
docker-up:
	cd docker && docker-compose up -d --build

# Entra no shell do container
docker-shell:
	docker exec -it $(CONTAINER) /bin/bash

docker-down:
	cd docker && docker-compose down
