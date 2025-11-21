# 🛡️ certfix-agent

**certfix-agent** é um agente leve e multiplataforma para gerenciamento e automação de certificados digitais.  
O projeto é **open source** e pode ser facilmente compilado, testado e implantado em produção.

---

## 🚀 Início Rápido

### Instalação e Configuração

```bash
# 1. Baixar e instalar
curl -fsSL https://raw.githubusercontent.com/certfix/certfix-agent/main/scripts/install.sh | sudo bash

# 2. Configurar com seu token de API
sudo certfix-agent configure --token "seu-token-api" --endpoint "https://api.certfix.com/api"

# 3. Verificar configuração (não requer sudo)
certfix-agent config

# 4. Iniciar o serviço
sudo systemctl start certfix-agent

# 5. Verificar status
sudo systemctl status certfix-agent
```

---

## Ambiente de Desenvolvimento

### Usando Docker

O ambiente Docker é recomendado para desenvolvimento isolado e reproduzível.

```
# Subir o ambiente Docker
make docker-up

# Build e execução do agente no container
make docker-run

# Entrar no container para depuração
make docker-shell

# Finalizar o ambiente
make docker-down
```

Builds no Docker

```
# Build para arquitetura atual
make docker-build

# Build para todas as arquiteturas suportadas
make docker-build-all
```

### Desenvolvimento Local

Para compilar e testar o agente diretamente na sua máquina:

```
# Build para a plataforma atual
make build-dev

# Executar localmente
make run

# Executar testes automatizados
make test

# Limpar diretórios de build
make clean
```

## Builds e Releases

### Build para produção

```
# Compilar binários para todas as arquiteturas suportadas
make build-all

# Preparar release (empacotamento e verificação)
make prepare-release
```

### Criar um novo release

Ao criar e enviar uma nova tag Git, o pipeline gera automaticamente os binários de release.

```
git tag v0.1.0
git push origin v0.1.0
```

Binários gerados para:

- Linux x86_64 (amd64)
- Linux ARM64 (aarch64)
- Linux ARMv7 (32 bits)

### Instalação

```
# Baixar e executar o instalador
curl -fsSL https://raw.githubusercontent.com/certfix/certfix-agent/main/scripts/install.sh -o install.sh
chmod +x install.sh
sudo ./install.sh
```

### Configuração

Após a instalação, configure o agente com seu token de API e endpoint:

```bash
# Configurar o agente
sudo certfix-agent configure --token "seu-token-api" --endpoint "https://api.example.com/api"
```

Isso criará o arquivo de configuração em `/etc/certfix-agent/config.json`:

```json
{
  "token": "seu-token-api",
  "endpoint": "https://api.example.com/api",
  "current_version": "0.1.0",
  "architecture": "amd64"
}
```

**Nota sobre Versão:** A versão do agente é armazenada no arquivo de configuração (`current_version`). Quando você reconfigura o agente, a versão existente é preservada. Para atualizar a versão, edite o arquivo de configuração ou use o script de atualização.

### Comandos Disponíveis

```bash
# Configurar o agente (requer sudo)
sudo certfix-agent configure --token <api-key> --endpoint <url>

# Mostrar configuração atual (não requer sudo)
certfix-agent config

# Iniciar o agente (requer sudo quando executado manualmente)
certfix-agent start

# Ver versão (não requer sudo)
certfix-agent version

# Ver ajuda (não requer sudo)
certfix-agent help
```

**Nota:** Apenas os comandos `configure` e `start` requerem sudo. Os comandos de consulta (`config`, `version`, `help`) podem ser executados sem privilégios elevados.

### Verificar Instalação

```
# Ver status do serviço
sudo systemctl status certfix-agent

# Visualizar logs em tempo real
sudo journalctl -u certfix-agent -f
```

### Desinstalar

```
curl -fsSL https://raw.githubusercontent.com/certfix/certfix-agent/main/scripts/uninstall.sh -o uninstall.sh
chmod +x uninstall.sh
sudo ./uninstall.sh
```

### Arquiteturas Suportadas

- Linux x86_64 (Intel/AMD 64-bit)
- Linux ARM64 (aarch64)
- Linux ARMv7 (32-bit ARM)

### Gerenciamento dos Serviços

```
# Iniciar serviço
sudo systemctl start certfix-agent

# Parar serviço
sudo systemctl stop certfix-agent

# Reiniciar serviço
sudo systemctl restart certfix-agent

# Habilitar inicialização automática
sudo systemctl enable certfix-agent

# Desabilitar inicialização automática
sudo systemctl disable certfix-agent

# Ver status do serviço
sudo systemctl status certfix-agent

# Visualizar logs em tempo real
sudo journalctl -u certfix-agent -f
```

### Remoção Manual

Se desejar remover manualmente o agente e seus arquivos:

```
sudo systemctl stop certfix-agent
sudo systemctl disable certfix-agent

sudo rm -f /etc/systemd/system/certfix-agent.service
sudo systemctl daemon-reload

sudo rm -f /usr/local/bin/certfix-agent
sudo rm -rf /etc/certfix-agent

sudo systemctl reset-failed certfix-agent
```

### Atualizações

#### Atualização Automática (sem confirmação)

```
curl -fsSL https://raw.githubusercontent.com/certfix/certfix-agent/main/scripts/update.sh | sudo bash -s -- --yes
```

#### Atualização Manual

```
curl -fsSL https://raw.githubusercontent.com/certfix/certfix-agent/main/scripts/update.sh -o update.sh
chmod +x update.sh
sudo ./update.sh
```

O script de atualização realiza automaticamente:

- Verificação de novas versões
- Download do binário correto para a arquitetura
- Backup da versão atual
- Atualização segura do serviço
- Rollback automático em caso de falha

### Ajuda

Para visualizar todos os comandos disponíveis:

```
make help
```
