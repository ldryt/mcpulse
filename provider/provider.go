package provider

import (
	"context"
)

type CloudProvider interface {
	CreateServer(ctx context.Context) (cloudID string, ipAddr string, err error)
	DeleteServer(ctx context.Context, cloudID string) error
	IsServerRunning(ctx context.Context, cloudID string) (bool, error)
}
