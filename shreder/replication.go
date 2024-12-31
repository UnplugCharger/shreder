package shreder

import (
	"bytes"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (cs *CacheServer) replicaset(key string, value string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	request := struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}{
		Key:   key,
		Value: value,
	}

	data, jsonerror := json.Marshal(request)
	if jsonerror != nil {
		log.Error().Msg("Error marshalling json")
		return
	}

	for _, peer := range cs.peers {
		go func(peer string) {
			client := &http.Client{}
			req, err := http.NewRequest("POST", peer+"/set", bytes.NewBuffer(data))
			if err != nil {
				log.Error().Msg("Error creating request")
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(replicationHeader, "true")

			_, err = client.Do(req)
			if err != nil {
				log.Error().Msg("Error sending request")
				return
			}

			log.Info().Msg("Replication successful")
		}(peer)
	}
}
