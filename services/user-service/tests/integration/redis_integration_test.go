package integration

import (
	"context"
	"testing"

	"bkc_microservice/services/user-service/internal/application/services"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRedisCache_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup Redis container
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "6379")

	rdb := redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
	})
	defer rdb.Close()

	// Test
	service := &services.UserService{RedisClient: rdb}
	userID := "test-user"
	data := map[string]interface{}{"id": userID, "username": "test"}

	// Cache & retrieve
	err = service.CacheUserBundle(ctx, userID, data)
	assert.NoError(t, err)

	cached, _ := service.GetCachedUserBundle(ctx, userID)
	assert.NotNil(t, cached)
	assert.Equal(t, userID, cached["id"])
}
