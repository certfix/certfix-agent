package main

import (
    "log"
    "time"
)

func main() {
    log.Println("[certfix-agent] Starting agent...")

    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            log.Println("[certfix-agent] Heartbeat: agent is alive")
        }
    }
}
