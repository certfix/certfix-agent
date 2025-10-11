package main

import (
    "log"
    "time"
)

func main() {
    log.Println("[linux-agent] Starting agent...")

    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            log.Println("[linux-agent] Heartbeat: agent is alive")
        }
    }
}
