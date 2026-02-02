package main

import (
	"context"
	"encoding/json"
	"encoding/csv"
	"strings"
	"fmt"
	"log"
	"os"
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/homingos/campaign-svc/config"
	daos "github.com/homingos/campaign-svc/daos"
	"github.com/homingos/campaign-svc/handlers"
	"github.com/homingos/campaign-svc/lib/nats"
	redisStorage "github.com/homingos/campaign-svc/lib/redis"
	"github.com/homingos/campaign-svc/lib/transaction"
	"github.com/homingos/flam-go-common/authz"
	"go.uber.org/zap"
)

type ResultItem struct {
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Score float32 `json:"score"`
}

type ShortCodeMapping struct {
	ShortCode  string `bson:"short_code" json:"short_code"`
	Name       string `bson:"name" json:"name"`
	CampaignID string `bson:"_id" json:"campaign_id,omitempty"`
}

type MappingData struct {
	SiteCode    string             `json:"site_code"`
	Mappings    []ShortCodeMapping `json:"mappings"`
	LastUpdated string             `json:"last_updated"`
	Count       int                `json:"count"`
}

func appendToJSON(fileName string, text string, newData []byte) {
	var existingData map[string]interface{}
	fileContent, err := os.ReadFile(fileName)

	if err == nil && len(fileContent) > 0 {
		_ = json.Unmarshal(fileContent, &existingData)
	}
    if existingData == nil {
        existingData = make(map[string]interface{})
    }
	existingData[text] = json.RawMessage(newData)

	updatedData, err := json.MarshalIndent(existingData, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling updated data:", err)
		return
	}

	err = os.WriteFile(fileName, updatedData, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func appendToCSV(fileName string, text string, results []ResultItem) {
    file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Println("Error opening CSV file:", err)
        return
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Add header if file is new
    info, err := file.Stat()
    if err == nil && info.Size() == 0 {
        writer.Write([]string{"Topic", "Codes", "Names", "Scores"})
    }

    var codes, names, scores []string
    for _, r := range results {
        codes = append(codes, r.Code)
        names = append(names, r.Name)
        scores = append(scores, fmt.Sprintf("%.4f", r.Score))
    }

    writer.Write([]string{
        text,
        strings.Join(codes, ","),
        strings.Join(names, ","),
        strings.Join(scores, ","),
    })
}

func main() {
	const mappingFile = "mapping.json"

	logger, _ := zap.NewProduction()
	defer logger.Sync()
	lgr := logger.Sugar()

	appConfig := config.GetAppConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// init db conn
	db := config.GetDBConn(ctx, lgr, appConfig.DB)

	redisClient := redisStorage.NewRedisClient(lgr)

	// init mlivus client
	milvusClient := config.ConfigureMilvusDatabase(appConfig.Milvus.Host, appConfig.Milvus.Key)

	// init daos
	campaignDao := daos.NewCampaignDao(lgr, db)
	categoryDao := daos.NewCategoryDao(lgr, db)
	experienceDao := daos.NewExperienceDao(lgr, db, redisClient)
	templateDao := daos.NewTemplateDao(lgr, db)
	milvusDao := daos.NewMilvusDao(lgr, milvusClient.Client)

	// init transaction manager
	mongoClient := db.Client() // Get the underlying mongo client from database
	txManager := transaction.NewTransactionManager(mongoClient)

	// init NATS client
	var natsClient *nats.Client
	natsClient, err := nats.NewClient(experienceDao, campaignDao, nil, redisClient, categoryDao, templateDao, txManager, nil, lgr)
	if err != nil {
		lgr.Warnf("Failed to initialize NATS client: %v", err)
	} else {
		lgr.Info("NATS client initialized successfully")
	}

	fgaClient := &authz.OpenFGAClient{} // Configure based on your OpenFGA setup

	categorySvc := handlers.NewCategorySvc(
		lgr,
		redisClient,
		milvusClient.Client,
		categoryDao,
		campaignDao,
		experienceDao,
		templateDao,
		natsClient,
		txManager,
		fgaClient,
		milvusDao,
	)

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Health check!")
	})

	app.Get("/generate-mappings/:siteCode", func(c *fiber.Ctx) error {
		siteCode := c.Params("siteCode")
		if siteCode == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "siteCode parameter is required",
			})
		}

		categoryData, appErr := categorySvc.GetCategoriesBySiteCodeSvc(c.Context(), siteCode, "")
		if appErr != nil {
			return c.Status(appErr.StatusCode).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Error(),
			})
		}

		var mappings []ShortCodeMapping

		if categoryDataMap, ok := categoryData.(map[string]interface{}); ok {
			if categories, ok := categoryDataMap["categories"].([]interface{}); ok {
				for _, cat := range categories {
					if catMap, ok := cat.(map[string]interface{}); ok {
						if campaigns, ok := catMap["campaigns"].([]interface{}); ok {
							for _, camp := range campaigns {
								if campMap, ok := camp.(map[string]interface{}); ok {
									mapping := ShortCodeMapping{
										ShortCode:  fmt.Sprintf("%v", campMap["short_code"]),
										Name:       fmt.Sprintf("%v", campMap["name"]),
										CampaignID: fmt.Sprintf("%v", campMap["_id"]),
									}
									mappings = append(mappings, mapping)
								}
							}
						}
					}
				}
			}
		}

		if len(mappings) == 0 {
			return c.Status(404).JSON(fiber.Map{
				"error": "No campaigns found for site code: " + siteCode,
			})
		}

		mappingData := MappingData{
			SiteCode:    siteCode,
			Mappings:    mappings,
			LastUpdated: time.Now().Format(time.RFC3339),
			Count:       len(mappings),
		}

		mappingFile := "mapping_" + siteCode + ".json"

		jsonData, err := json.MarshalIndent(mappingData, "", "  ")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Failed to marshal mappings",
				"details": err.Error(),
			})
		}

		err = os.WriteFile(mappingFile, jsonData, 0644)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Failed to write mapping file",
				"details": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message":      "Mappings generated successfully",
			"file":         mappingFile,
			"site_code":    siteCode,
			"count":        len(mappings),
			"last_updated": mappingData.LastUpdated,
		})
	})

	app.Get("/campaigns/:sitecode", func(c *fiber.Ctx) error {
		siteCode := c.Params("sitecode")
		text := c.Query("text", "")

		mappingFile := "mapping_" + siteCode + ".json"
		mappingData, err := os.ReadFile(mappingFile)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Mapping file not found. Please generate mappings first.",
			})
		}

		var mappingInfo MappingData
		if err := json.Unmarshal(mappingData, &mappingInfo); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to parse mapping file",
			})
		}

		shortCodeToName := make(map[string]string)
		for _, mapping := range mappingInfo.Mappings {
			shortCodeToName[mapping.ShortCode] = mapping.Name
		}

		var results []ResultItem
		queryKey := text
		if queryKey == "" {
			queryKey = "all_campaigns"
		}

		if text != "" {
			embeddings, err := handlers.GetEmbeddings(text)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to get embeddings"})
			}

			milvusDocs, err := milvusDao.Search(c.Context(), embeddings, siteCode)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Milvus search failed"})
			}

			// Map Milvus results back to short codes and names
			for _, doc := range milvusDocs {
				// Find short code for this milvus reference
				scList, err := campaignDao.GetShortcodesByMilvusRefID([]string{doc.ID})
				if err == nil && len(scList) > 0 {
					shortCode := scList[0]
					results = append(results, ResultItem{
						Code:  shortCode,
						Name:  shortCodeToName[shortCode],
						Score: doc.Score,
					})
				}
			}
		} else {
			for _, m := range mappingInfo.Mappings {
				results = append(results, ResultItem{
					Code:  m.ShortCode,
					Name:  m.Name,
					Score: 0,
				})
			}
		}

		resultsJSON, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			lgr.Errorf("Failed to marshal results: %v", err)
		} else {
            appendToJSON("short_code_output.json",text,resultsJSON)
			appendToCSV("short_code_output.csv", text, results)
        }

		return c.JSON(map[string][]ResultItem{
			queryKey: results,
		})
	})
	log.Fatal(app.Listen(":3000"))
}
