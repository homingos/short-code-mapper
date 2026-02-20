package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/flam-go-common/errors"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"go.uber.org/zap"
)

type MilvusDaoImpl struct {
	lgr          *zap.SugaredLogger
	milvusClient client.Client
}

func NewMilvusDao(lgr *zap.SugaredLogger, milvusClient client.Client) *MilvusDaoImpl {
	return &MilvusDaoImpl{lgr: lgr, milvusClient: milvusClient}
}

// L2DistanceToSimilarity converts L2 distance to similarity score
// range: [0, 1], where 1 is most similar

// func l2DistanceToSimilarity(l2Distance float32) float32 {
// 	similarity := 1 / (1 + l2Distance)
// 	return similarity
// }

// NormalizeCOSINESimilarity converts Milvus score to similarity score
// Milvus returns Cosine Similarity [-1, 1] when using entity.COSINE.
func NormalizeCOSINESimilarity(score float32) float32 {
	if score < 0 {
		return 0
	}
	return score
}

func (impl *MilvusDaoImpl) Search(ctx context.Context, embeddings []float32, siteCode string) ([]dtos.SearchResult, error) {
	fmt.Println("searching in milvus")
	searchParams, err := entity.NewIndexFlatSearchParam()
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}
	expr := fmt.Sprintf("catalog_id == '%s'", siteCode)
	Env := os.Getenv("APP_ENV")
	Env = "prod"
	// if Env == "non-prod" {
	// 	Env = "qa"
	// }
	// if Env == "" {
	// 	Env = "dev"
	// }
	milvusColl := fmt.Sprintf("product_vectors_%s", Env)

	// fmt.Println(milvusColl)
	// fmt.Println(embeddings)

	// Ensure collection is loaded (no-op if already loaded)
	loaded, err := impl.milvusClient.GetLoadState(ctx, milvusColl, nil)
	if err != nil || loaded != entity.LoadStateLoaded {
		_ = impl.milvusClient.LoadCollection(ctx, milvusColl, false)
	}

	results, err := impl.milvusClient.Search(ctx,
		milvusColl,
		[]string{},
		expr,
		[]string{"*"},
		[]entity.Vector{entity.FloatVector(embeddings)},
		"vector_information",
		entity.COSINE,
		5,
		searchParams,
	)

	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	var searchResults []dtos.SearchResult

	if len(results) > 0 {
		numHits := results[0].ResultCount

		if numHits == 0 {
			return nil, fmt.Errorf("no hits found")
		}
		for i := 0; i < numHits; i++ {
			id, err := results[0].IDs.GetAsString(i)
			if err != nil {
				continue
			}
			score := results[0].Scores[i]
			catalogID := ""
			clientID := ""
			description := ""
			name := ""

			if len(results[0].Fields) > 0 {
				for _, field := range results[0].Fields {
					switch field.Name() {
					case "catalog_id":
						catalogID, _ = field.GetAsString(i)
					case "client_id":
						clientID, _ = field.GetAsString(i)
					case "description":
						description, _ = field.GetAsString(i)
					case "name":
						name, _ = field.GetAsString(i)
					case "product_name":
						name, _ = field.GetAsString(i)
					case "short_code":
						if name == "" {
							name, _ = field.GetAsString(i)
						}
					default:
						//
					}
				}
			}

			_ = catalogID
			_ = clientID
			_ = description
			candidate := dtos.SearchResult{
				Document: dtos.Document{
					ID:          id,
					Name:        name,
					CatalogID:   catalogID,
					ClientID:    clientID,
					Description: description,
				},
				Score: score,
			}	

			searchResults = append(searchResults, candidate)
		}
	} else {
		return nil, errors.InternalServerError("no results found")
	}

	PrettyPrint(searchResults)

	return searchResults, nil
}

func (impl *MilvusDaoImpl) Delete(ctx context.Context, clientID string, milvusRefID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	Env := os.Getenv("APP_ENV")
	if Env == "non-prod" {
		Env = "qa"
	}
	if Env == "" {
		Env = "prod"
	}
	milvusColl := fmt.Sprintf("product_vectors_%s", Env)
	expr := fmt.Sprintf("id == '%s'", milvusRefID)

	err := impl.milvusClient.Delete(ctx, milvusColl, "", expr)
	if err != nil {
		return errors.InternalServerError(err.Error())
	}
	return nil
}

func PrettyPrint(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling data:", err)
		return
	}
	fmt.Println(string(jsonData))
}
