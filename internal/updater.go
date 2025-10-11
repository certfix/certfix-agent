package internal

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "os/exec"

    "github.com/blang/semver/v4"
)


const (
	RepoAPI = "https://api.github.com/repos/certfix/certfix-agent/releases/latest"
	BinaryPath = "/usr/local/bin/certfix-agent"
)

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		BrowserDownloadURL string `json:"browser_download_url"`
		Name               string `json:"name"`
	} `json:"assets"`
}

func CheckForUpdates(cfg *Config) {
	if !cfg.AutoUpdate {
		return
	}

	resp, err := http.Get(RepoAPI)
	if err != nil {
		log.Printf("Falha ao verificar atualização: %v", err)
		return
	}
	defer resp.Body.Close()

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		log.Printf("Erro ao parsear release: %v", err)
		return
	}

	current, _ := semver.Make(cfg.CurrentVer)
	latest, _ := semver.Make(rel.TagName)

	if latest.GT(current) {
		log.Printf("Nova versão disponível: %s -> %s", current, latest)
		url := rel.Assets[0].BrowserDownloadURL
		updateBinary(url)
	}
}

func updateBinary(url string) {
	tmp := "/tmp/certfix-agent"
	log.Printf("Baixando nova versão de %s...", url)
	cmd := exec.Command("curl", "-fsSL", "-o", tmp, url)
	if err := cmd.Run(); err != nil {
		log.Printf("Erro ao baixar update: %v", err)
		return
	}

	if err := os.Rename(tmp, BinaryPath); err != nil {
		log.Printf("Erro ao substituir binário: %v", err)
		return
	}

	if err := os.Chmod(BinaryPath, 0755); err != nil {
		log.Printf("Erro ao aplicar permissão: %v", err)
	}

	log.Println("Atualização concluída. Reiniciando serviço...")
	exec.Command("systemctl", "restart", "certfix-agent").Run()
}
