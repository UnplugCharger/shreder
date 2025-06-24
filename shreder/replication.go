package shreder

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

func (cs *CacheServer) replicaset(key string, value string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	req := struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		Key:   key,
		Value: value,
	}

	data, _ := json.Marshal(req)

	for _, peer := range cs.peers {
		if peer != "" && peer != cs.selfID {
			go func(peer string) {
				// Ensure peer URL has proper format
				peerURL := peer
				if !strings.HasPrefix(peerURL, "http://") && !strings.HasPrefix(peerURL, "https://") {
					peerURL = "http://" + peerURL
				}
				
				client := &http.Client{}
				req, err := http.NewRequest("POST", peerURL+"/set", bytes.NewReader(data))
				if err != nil {
					log.Printf("Failed to create replication request: %v", err)
					return
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set(replicationHeader, "true")
				
				resp, err := client.Do(req)
				if err != nil {
					log.Printf("Failed to replicate to peer %s: %v", peer, err)
					return
				}
				defer resp.Body.Close()
				
				if resp.StatusCode != http.StatusOK {
					log.Printf("Replication to peer %s failed with status: %d", peer, resp.StatusCode)
				} else {
					log.Printf("Successfully replicated key %s to peer %s", key, peer)
				}
			}(peer)
		}
	}
}
