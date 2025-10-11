package internal

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func RunWorker(cfg *Config) {
	resp, err := http.Get(fmt.Sprintf("%s/ping?token=%s", cfg.Endpoint, cfg.Token))
	if err != nil {
		log.Printf("[ERRO] Comunicação: %v", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	log.Printf("[OK] Resposta: %s", string(body))
}
