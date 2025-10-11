# certfix-agent

```
make docker-up
```

```
make docker-build
```

```
make docker-run
```

```
make docker-down
```

## Dentro do Container

```
make docker-shell
```

git tag v0.1.0
git push origin v0.1.0

curl -fsSL https://raw.githubusercontent.com/certfix/certfix-agent/main/scripts/install.sh -o install.sh
chmod +x install.sh
sudo ./install.sh

sudo systemctl status certfix-agent
sudo journalctl -u certfix-agent -f
