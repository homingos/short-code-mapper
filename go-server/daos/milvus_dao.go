package dao

import (
	"context"

	"github.com/homingos/campaign-svc/dtos"
)

type MilvusDao interface {
	Search(ctx context.Context, embeddings []float32, clientID string) ([]dtos.SearchResult, error)
}
