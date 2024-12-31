package shreder

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

const (
	replicationHeader = "X-Replication-Request"
)

type CacheServer struct {
	cache  *Cache
	router *gin.Engine
	peers  []string
	mu     sync.Mutex
}

func NewCacheServer(peers []string) *CacheServer {
	return &CacheServer{
		cache: NewCache(10),
		peers: peers,
	}
}

type setRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

func (cs *CacheServer) SetHandler(ctx *gin.Context) {
	var request setRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cs.cache.Set(request.Key, request.Value, 30*time.Second)

	// check if the request is a replication request
	if ctx.GetHeader(replicationHeader) == "" {
		go cs.replicaset(request.Key, request.Value)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})

}

type getRequest struct {
	Key string `form:"key" binding:"required"`
}

func (cs *CacheServer) GetHandler(ctx *gin.Context) {
	var request getRequest
	if err := ctx.ShouldBindQuery(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	value, found := cs.cache.Get(request.Key)
	if !found {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"value": value})
}

func (cs *CacheServer) Start(address string) error {
	cs.router = gin.Default()
	cs.router.POST("/set", cs.SetHandler)
	cs.router.GET("/get", cs.GetHandler)
	return http.ListenAndServe(address, cs.router)
}
