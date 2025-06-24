package shreder

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/UnplugCharger/shreder/hash_ring"
)

const (
	replicationHeader = "X-Replication-Request"
)

type CacheServer struct {
	cache    *Cache
	peers    []string
	mu       sync.Mutex
	selfID   string
	hashRing *hash_ring.HashRing
}

func NewCacheServer(peers []string, selfID string) *CacheServer {
	cs := &CacheServer{
		cache:    NewCache(10),
		peers:    peers,
		hashRing: hash_ring.NewHashRing(),
		selfID:   selfID,
	}
	
	// Add self to hash ring first
	cs.hashRing.AddNode(hash_ring.Node{ID: selfID, Address: selfID})
	
	// Add peers to hash ring
	for _, peer := range peers {
		if peer != "" && peer != selfID {
			cs.hashRing.AddNode(hash_ring.Node{ID: peer, Address: peer})
		}
	}
	
	log.Printf("Cache server initialized with selfID: %s, peers: %v", selfID, peers)
	return cs
}

type setRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

func (cs *CacheServer) SetHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received SET request from %s", r.RemoteAddr)
	
	// Read the body first so we can use it for both decoding and forwarding
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	var request setRequest
	if err := json.Unmarshal(bodyBytes, &request); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if this is a replication request to avoid infinite loops
	isReplication := r.Header.Get(replicationHeader) == "true"
	
	targetNode := cs.hashRing.GetNode(request.Key)
	log.Printf("SET: Key=%s, TargetNode=%s, SelfID=%s, IsReplication=%v", 
		request.Key, targetNode.Address, cs.selfID, isReplication)
	
	if targetNode.Address == cs.selfID {
		cs.cache.Set(request.Key, request.Value, 10*time.Minute)
		log.Printf("Stored key %s locally", request.Key)
		// Only replicate if this is not already a replication request
		if !isReplication {
			log.Printf("Starting replication for key %s", request.Key)
			cs.replicaset(request.Key, request.Value)
		}
		w.WriteHeader(http.StatusOK)
	} else {
		log.Printf("Forwarding request for key %s to node %s", request.Key, targetNode.Address)
		// Create a new request with the original body
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		cs.forwardRequest(targetNode, r, w)
	}
}

type getRequest struct {
	Key string `form:"key" binding:"required"`
}

func (cs *CacheServer) GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	targetNode := cs.hashRing.GetNode(key)
	log.Printf("GET: Key=%s, TargetNode=%s, SelfID=%s", key, targetNode.Address, cs.selfID)
	
	if targetNode.Address == cs.selfID {
		value, found := cs.cache.Get(key)
		if !found {
			log.Printf("Key %s not found locally", key)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		log.Printf("Found key %s locally with value %s", key, value)
		w.Write([]byte(value))
	} else {
		log.Printf("Forwarding GET request for key %s to node %s", key, targetNode.Address)
		cs.forwardRequest(targetNode, r, w)
	}
}

func (cs *CacheServer) Start(address string) error {
	http.HandleFunc("/set", cs.SetHandler)
	http.HandleFunc("/get", cs.GetHandler)
	return http.ListenAndServe(address, nil)
}

func (cs *CacheServer) forwardRequest(targetNode hash_ring.Node, r *http.Request, w http.ResponseWriter) {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100, // Adjust based on your load
		},
		Timeout: 5 * time.Second, // Prevent requests from hanging indefinitely
	}
	address := targetNode.Address
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = "http://" + address // Default to HTTP
	}

	// Create a new request based on the method
	var req *http.Request
	var err error

	switch r.Method {
	case http.MethodGet:
		// Forward GET request with query parameters
		getURL := fmt.Sprintf("%s%s?%s", address, r.URL.Path, r.URL.RawQuery)
		req, err = http.NewRequest(r.Method, getURL, nil)
	case http.MethodPost:
		// Forward POST request with body
		postURL := fmt.Sprintf("%s%s", address, r.URL.Path)
		req, err = http.NewRequest(r.Method, postURL, r.Body)
	}

	if err != nil {
		log.Printf("Failed to create forward request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Copy the headers
	req.Header = r.Header

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		// Check for a "connection refused" error
		var urlErr *url.Error
		if errors.As(err, &urlErr) && urlErr.Err != nil {
			var opErr *net.OpError
			if errors.As(urlErr.Err, &opErr) && opErr.Op == "dial" {
				var sysErr *os.SyscallError
				if errors.As(opErr.Err, &sysErr) && sysErr.Syscall == "connect" {
					log.Printf("Connection refused to node %s: %v", targetNode.Address, err)
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
			}
		}
		log.Printf("Failed to forward request to node %s: %v", targetNode.Address, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy response status and headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	io.Copy(w, resp.Body)
}
