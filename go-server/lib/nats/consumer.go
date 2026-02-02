package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/campaign-svc/models"
	"github.com/homingos/campaign-svc/types/consts"
	"github.com/homingos/campaign-svc/utils"
	"github.com/homingos/flam-go-common/authz"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ConsumerName                 = "workflow-complete-consumer"
	Subject                      = consts.WorkflowCompleteSubject
	ProductCatalogueConsumerName = "product-catalogue-consumer"
	ProductCatalogueSubject      = consts.ProductCatalogueSubject
	ProductCatalogueStreamName   = consts.ProductCatalogueStreamName
	WorkflowStore                = "workflow-state"
	BatchSize                    = 10
	FetchWaitTime                = 500 * time.Millisecond
)

var (
	NatsServerURL = os.Getenv("NATS_HOST")
	// NatsServerURL = "nats://localhost:4222"
)

func (cli *Client) Consumer() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer, err := cli.Stream.Consumer(context.Background(), consts.WorkflowStreamName)
	if err != nil {
		log.Println("Consumer not found, creating a new one...")

		consumer, err = cli.Stream.CreateOrUpdateConsumer(context.Background(), jetstream.ConsumerConfig{
			Durable:       ConsumerName,
			AckPolicy:     jetstream.AckExplicitPolicy,
			AckWait:       30 * time.Second,
			MaxAckPending: 100,
			MaxDeliver:    3,
			FilterSubject: Subject,
		})
		if err != nil {
			log.Fatalf("Failed to create consumer: %v", err)
		}
	}

	// Signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Consumer shutting down...")
				return
			default:
				// Fetch messages
				messages, err := consumer.Fetch(BatchSize, jetstream.FetchMaxWait(FetchWaitTime))
				if err != nil {
					log.Printf("Error fetching messages: %v", err)
					time.Sleep(time.Second)
					continue
				}
				for msg := range messages.Messages() {
					// msg.Ack()
					// log.Printf("Received message: %+v\n", msg)
					cli.processMessage(msg)
				}
				// Process messages

			}
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received shutdown signal: %v", sig)
	cancel()
	log.Println("Consumer shutdown complete")
}

// processMessage processes each message.
func (cli *Client) processMessage(msg jetstream.Msg) {
	var wf dtos.WorkflowFinalPubResult
	if err := json.Unmarshal(msg.Data(), &wf); err != nil {
		log.Println("Error unmarshalling message:", err)
		msg.Nak()
		return
	}
	// fmt.Println("WORKFLOW", wf)
	parts := strings.Split(wf.WorkflowId, "_")

	if parts[0] == "remotion" {
		var url, maskedUrl string
		if len(wf.TaskResults) > 0 {
			url = wf.TaskResults[0].Payload.RemotionVideoUrl
			maskedUrl = wf.TaskResults[0].Payload.RemotionMaskedVideoUrl
		} else {
			log.Printf("Error updating remotion video, Task Results is empty: %v", wf.TaskResults)
			msg.Nak()
			return
		}
		err := cli.remotionDao.UpdateRemotionResult(wf.WorkflowId, url, maskedUrl)
		if err != nil {
			log.Printf("Error updating remotion video: %v", err)
			msg.Nak()
			return
		}
	} else if parts[0] == "campaign" {
		shortCode := parts[1]
		updateMap := make(map[string]interface{})
		if len(wf.TaskResults) > 0 {
			url := wf.TaskResults[0].Payload.ScanCompressedImage
			if url == "" {
				log.Printf("nothing to update, scan compressed image url is empty: %v", wf.TaskResults[0].Payload.ScanCompressedImage)
				msg.Ack()
				return
			}
			updateMap["scan_compressed_image_url"] = url
		}
		_, err := cli.campDao.PostbackCampaignDao(shortCode, updateMap)
		if err != nil {
			log.Printf("Error updating campaign assets: %v", err)
			msg.Nak()
			return
		}
		cli.redisClient.ExpireCampaignExperiences(shortCode, false, false)
	} else if parts[0] == "imageqr" {
		err := cli.expDao.UpdateExperienceAssetsWithQrImageDao(wf)
		if err != nil {
			log.Printf("Error updating experience assets with QR: %v", err)
			msg.Nak()
			return
		}
	} else if parts[0] == "regenerate" {
		err := cli.expDao.UpdateRegenerateExperienceAssetsDao(wf)
		if err != nil {
			log.Printf("Error updating regenerate experience assets: %v", err)
			msg.Nak()
			return
		}
	} else {
		err := cli.expDao.UpdateExperienceAssetsDao(wf)
		if err != nil {
			log.Printf("Error updating experience assets: %v", err)
			msg.Nak()
			return
		}
	}
	log.Printf("Message processed successfully: %s at: %v", wf.WorkflowId, time.Now())
	msg.Ack()
}

// ProductCatalogueConsumer consumes product catalogue creation messages.
func (cli *Client) ProductCatalogueConsumer() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := CreateStreamIfNotExist(ctx, consts.ProductCatalogueStreamName, ProductCatalogueSubject, jetstream.LimitsPolicy, 0, 0)
	if err != nil {
		log.Fatalf("Failed to create or get stream: %v", err)
	}

	consumer, err := stream.Consumer(context.Background(), ProductCatalogueConsumerName)
	if err != nil {
		cli.lgr.Info("Product catalogue consumer not found, creating a new one...")

		consumer, err = stream.CreateOrUpdateConsumer(context.Background(), jetstream.ConsumerConfig{
			Durable:       ProductCatalogueConsumerName,
			AckPolicy:     jetstream.AckExplicitPolicy,
			AckWait:       10 * time.Minute,
			MaxAckPending: 100,
			MaxDeliver:    3,
			FilterSubject: ProductCatalogueSubject,
		})
		if err != nil {
			cli.lgr.Fatalf("Failed to create product catalogue consumer: %v", err)
		}
	}

	// Signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ctx.Done():
				cli.lgr.Info("Product catalogue consumer shutting down...")
				return
			default:
				// Fetch messages
				messages, err := consumer.Fetch(BatchSize, jetstream.FetchMaxWait(FetchWaitTime))
				if err != nil {
					cli.lgr.Errorf("Error fetching product catalogue messages: %v", err)
					time.Sleep(time.Second)
					continue
				}
				for msg := range messages.Messages() {
					cli.processProductCatalogueMessage(ctx, msg)
				}
			}
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	cli.lgr.Infof("Received shutdown signal: %v", sig)
	cancel()
	cli.lgr.Info("Product catalogue consumer shutdown complete")
}

// processProductCatalogueMessage processes product catalogue creation messages.
func (cli *Client) processProductCatalogueMessage(ctx context.Context, msg jetstream.Msg) {
	var catalogueDto dtos.ProductCatalogueDto
	if err := json.Unmarshal(msg.Data(), &catalogueDto); err != nil {
		cli.lgr.Errorw("Error unmarshalling product catalogue message", "error", err)
		msg.Nak()
		return
	}
	fmt.Println("catalogueDto", catalogueDto.CategoryID)

	sessionCtx, err := cli.txManager.BeginTx(ctx)
	if err != nil {
		cli.lgr.Errorw("Error starting transaction", "error", err)
		msg.Nak()
		return
	}

	var txError error
	defer func() {
		if txError != nil {
			if abortErr := sessionCtx.AbortTransaction(ctx); abortErr != nil {
				cli.lgr.Errorw("Error aborting transaction", "error", abortErr)
			}
		}
		sessionCtx.EndSession(ctx)
	}()

	template, err := cli.templateDao.GetAlphaTemplateDao()
	if err != nil {
		cli.lgr.Errorw("Error getting alpha template", "error", err)
		txError = err
		msg.Nak()
		return
	}

	clientObjID, err := primitive.ObjectIDFromHex(catalogueDto.ClientID)
	if err != nil {
		cli.lgr.Errorw("Error converting client id to object id", "error", err)
		msg.Nak()
		return
	}

	categoryCampaigns := map[string][]string{}

	campaignGroupReq := &dtos.CampaignGroupCreateDto{
		ClientID:  catalogueDto.ClientID,
		GroupName: "Product Catalogue - " + catalogueDto.SiteCode,
		CreatedBy: catalogueDto.CreatedBy,
	}

	expShortCodeMap := make(map[string]string)

	campaignGroup, daoErr := cli.campDao.CreateCampaignGroupDao(nil, &sessionCtx, campaignGroupReq)
	if daoErr != nil {
		txError = daoErr
		cli.lgr.Errorw("Failed to create campaign group", "error", daoErr)
		msg.Nak()
		return
	}

	var campaigns []*models.Campaign
	var experiences []*models.Experience
	var workflows []dtos.Workflow

	for _, product := range catalogueDto.Products {
		imageURL, err := utils.GetImageFromURL(product.ImageUrl)
		if err != nil {
			txError = err
			cli.lgr.Errorw("Error getting image from URL", "error", err)
			msg.Nak()
			return
		}
		product.ImageUrl = imageURL

		code, err := utils.GenerateShortCode(product.Name)
		if err != nil {
			txError = err
			cli.lgr.Errorw("Error generating short code", "error", err)
			msg.Nak()
			return
		}
		categoryCampaigns[product.Category] = append(categoryCampaigns[product.Category], code)

		experienceID := primitive.NewObjectID()
		workflowID := fmt.Sprintf("%s_%s", experienceID.Hex(), primitive.NewObjectID().Hex())

		newCampaign := &models.Campaign{
			ID:        primitive.NewObjectID(),
			Name:      product.Name + " - " + product.ID,
			ClientId:  clientObjID,
			ShortCode: code,
			GroupID:   campaignGroup.ID,
			GroupName: campaignGroup.Name,
			Status:    consts.Created,
			TrackType: "CARD",
			Scan: models.Scan{
				ScanText:           consts.DefaultCardScanText,
				ImageUrl:           "",
				CompressedImageUrl: "",
			},
			IsActive: true,
			AirTracking: utils.ToPointer(true),
			Publish:  false,
			CopyRight: models.CopyRight{
				Show:    false,
				Content: consts.CopyRightContent,
			},
			FeatureFlags: models.FeatureFlags{
				EnableNextar:            true,
				EnableGeoVideos:         false,
				EnableVideoFullscreen:   true,
				EnableQRButton:          false,
				EnableRecording:         false,
				EnableIosStreaming:      true,
				EnableAndroidStreaming:  true,
				EnableAdaptiveStreaming: true,
				EnableScreenCapture:     true,
				EnableAirBoard:          true,
				EnableAutoPlay:          true,
			},
			CreatedAt: time.Duration(time.Now().UnixMilli()),
			UpdatedAt: time.Duration(time.Now().UnixMilli()),
			CreatedBy: catalogueDto.CreatedBy,
		}

		campaigns = append(campaigns, newCampaign)

		newExperience := &models.Experience{
			ID:         experienceID,
			CampaignID: newCampaign.ID,
			Canvas: models.Canvas{
				IOS:     2100,
				Android: 0,
			},
			IsActive: true,
			Variant: models.Variant{
				Class:     1,
				TrackType: "CARD",
				IsAlpha:   utils.ToPointer(true),
				Offset: models.ThreeDCoordinates{
					TwoDCoordinates: models.TwoDCoordinates{
						X: 0,
						Y: 0,
					},
					Z: 0,
				},
				ScaleAxis: models.ThreeDCoordinates{
					TwoDCoordinates: models.TwoDCoordinates{
						X: 1,
						Y: 1,
					},
					Z: 0,
				},
			},
			PlaybackScale: 1,
			Status:        consts.Processing,
			Images: []models.Image{
				{
					K: "original",
					V: product.ImageUrl,
				},
				{
					K: "original_input",
					V: product.ImageUrl,
				},
			},
			Videos:          []models.Video{},
			UIElements:      &models.UIElements{},
			CreatedAt:       time.Duration(time.Now().UnixMilli()),
			UpdatedAt:       time.Duration(time.Now().UnixMilli()),
			CreatedBy:       catalogueDto.CreatedBy,
			TemplateDetails: template,
			QrPanel:         &models.QrPanel{},
			CatalogueDetails: &models.CatalogueDetails{
				ProductID:   product.ID,
				Category:    product.Category,
				Name:        product.Name,
				Description: product.Description,
				Currency:    product.Currency,
				ProductUrl:  product.ProductUrl,
				Price:       product.Price,
				ImageURL:    product.ImageUrl,
			},
			WorkflowID: workflowID,
		}

		if product.Currency == "USD" {
			newExperience.CatalogueDetails.Currency = "$"
		} else if product.Currency == "INR" {
			newExperience.CatalogueDetails.Currency = "₹"
		} else if product.Currency == "EUR" {
			newExperience.CatalogueDetails.Currency = "€"
		} else if product.Currency == "GBP" {
			newExperience.CatalogueDetails.Currency = "£"
		} else if product.Currency == "JPY" {
			newExperience.CatalogueDetails.Currency = "¥"
		}

		expShortCodeMap[newExperience.ID.Hex()] = newCampaign.ShortCode
		experiences = append(experiences, newExperience)

		tasks := []dtos.Task{}
		mediaProcess := models.MediaProcess{
			Experience:          *newExperience,
			ShortCode:           newCampaign.ShortCode,
			ClientId:            clientObjID,
			GenerateGreenScreen: true,
			ImageVectorLLMProductJob: &models.ImageVectorLLMProductJob{
				ID:          product.ID,
				SiteCode:    catalogueDto.SiteCode,
				ClientID:    catalogueDto.ClientID,
				Name:        product.Name,
				Category:    product.Category,
				Price:       product.Price,
				Currency:    product.Currency,
				ImageURL:    product.ImageUrl,
				ProductURL:  product.ProductUrl,
				Description: product.Description,
			},
			GenStudiJob: &models.FalVideoJob{
				Type:                  "video",
				Prompt:                "Create a green screen video with the product showcase",
				VideoGenerationSource: "alpha",
				MediaReferences:       []models.MediaReference{},
				UserID:                catalogueDto.CreatedBy.ID,
				EnableAudio:           true,
				VideoCategory:         product.Category,
			},
		}
		masterDependencies := map[string][]string{}

		imageVectorLLMTask := utils.CreateImageRagAITask(&mediaProcess, workflowID)
		falVideoTask := utils.CreateFalVideoTask(&mediaProcess, workflowID)
		alphaVideoTask := utils.CreateAlphaVideoTask(&mediaProcess, workflowID)
		mainImageTask := utils.CreateImageTask(&mediaProcess, workflowID)
		mainVideoTask := utils.CreateVideoTaskWithDependencies(&mediaProcess, workflowID)

		masterDependencies[falVideoTask.Id] = append(masterDependencies[falVideoTask.Id], mainImageTask.Id)
		masterDependencies[alphaVideoTask.Id] = append(masterDependencies[alphaVideoTask.Id], falVideoTask.Id)
		masterDependencies[mainVideoTask.Id] = append(masterDependencies[mainVideoTask.Id], alphaVideoTask.Id)

		tasks = append(tasks, *mainImageTask, *mainVideoTask, *imageVectorLLMTask, *falVideoTask, *alphaVideoTask)
		workflows = append(workflows, utils.CreateExperienceWorkflow(tasks, false, workflowID, masterDependencies))
	}

	// Array of category and code for map
	categoryCampaignsArray := []models.Categories{}
	for category, campaignCodes := range categoryCampaigns {
		categoryCampaignsArray = append(categoryCampaignsArray, models.Categories{
			Name:      category,
			Campaigns: campaignCodes,
		})
	}

	_, err = cli.categoryDao.UpdateCategoryDao(nil, &sessionCtx, catalogueDto.CategoryID, &dtos.UpdateCategoryDto{
		Categories: categoryCampaignsArray,
		ClientID:   catalogueDto.ClientID,
	})
	if err != nil {
		txError = err
		cli.lgr.Errorw("Error updating category", "error", err)
		msg.Nak()
		return
	}

	_, err = cli.campDao.CreateBulkCampaignDao(&sessionCtx, campaigns)
	if err != nil {
		txError = err
		cli.lgr.Errorw("Error creating bulk campaigns", "error", err)
		msg.Nak()
		return
	}

	_, err = cli.expDao.CreateBulkExperienceDao(&sessionCtx, experiences)
	if err != nil {
		txError = err
		cli.lgr.Errorw("Error creating bulk experiences", "error", err)
		msg.Nak()
		return
	}

	// Create tuples for Campaigns
	var campaignTuples []authz.Tuple
	for _, camp := range campaigns {
		campaignTuples = append(campaignTuples, authz.Tuple{
			UserType:   authz.ObjectTypeBrand,
			User:       camp.ClientId.Hex(),
			Relation:   authz.RelationParent,
			ObjectType: authz.ObjectTypeCampaign,
			Object:     camp.ID.Hex(),
		})
	}
	err = cli.fgaClient.AssignAndRevokePermissions(ctx, campaignTuples, []authz.Tuple{})
	if err != nil {
		txError = err
		cli.lgr.Errorw("Failed to store permission in FGA", "error", err)
		msg.Nak()
		return
	}

	for _, workflow := range workflows {
		barr, err := json.Marshal(workflow)
		if err != nil {
			txError = err
			cli.lgr.Errorw("Error marshaling workflow", "error", err)
			msg.Nak()
			return
		}
		err = cli.PublishMsg("workflow.submit", barr)
		if err != nil {
			cli.lgr.Errorw("Error publishing workflow", "error", err)
			// Don't fail the whole operation if workflow publish fails
		}
	}

	if err = sessionCtx.CommitTransaction(ctx); err != nil {
		txError = err
		cli.lgr.Errorw("Error committing transaction", "error", err)
		msg.Nak()
		return
	}

	// Publish edit logs asynchronously
	go func() {
		for _, camp := range campaigns {
			createLogs := dtos.EditLogDto{
				ClientId:       camp.ClientId,
				CampaignId:     camp.ID,
				BeforeCampaign: camp,
				ShortCode:      camp.ShortCode,
			}

			barr, err := json.Marshal(createLogs)
			if err != nil {
				cli.lgr.Errorw("Error marshaling campaign edit log", "error", err)
				continue
			}
			err = cli.PublishMsg(consts.EditLogsSubject, barr)
			if err != nil {
				cli.lgr.Errorw("Error publishing campaign edit log", "error", err)
			}
		}
		for _, exp := range experiences {
			createLogs := dtos.EditLogDto{
				ClientId:     clientObjID,
				ExperienceId: exp.ID,
				Before:       exp,
				ShortCode:    expShortCodeMap[exp.ID.Hex()],
			}

			barr, err := json.Marshal(createLogs)
			if err != nil {
				cli.lgr.Errorw("Error marshaling experience edit log", "error", err)
				continue
			}
			err = cli.PublishMsg(consts.EditLogsSubject, barr)
			if err != nil {
				cli.lgr.Errorw("Error publishing experience edit log", "error", err)
			}
		}
	}()

	cli.lgr.Infow("Product catalogue processed successfully", "siteCode", catalogueDto.SiteCode, "clientID", catalogueDto.ClientID)
	msg.Ack()
}

// Stop gracefully shuts down the NATS client.
func (cli *Client) Stop() {
	log.Println("Closing NATS connection...")
	cli.nc.Close()
	log.Println("NATS connection closed.")
}

func CreateStreamIfNotExist(ctx context.Context, name, subject string, retention jetstream.RetentionPolicy, maxMessages, maxBytes int64) (jetstream.Stream, error) {
	natsURL := os.Getenv("NATS_HOST")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		fmt.Printf("connecting to %s:", natsURL)
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("Failed to get JetStream context: %v", err)
	}

	stream, err := js.Stream(ctx, name)
	if err == jetstream.ErrStreamNotFound {
		stream, err = js.CreateStream(ctx, jetstream.StreamConfig{
			Name:       name,
			Subjects:   []string{subject},
			Retention:  retention,
			MaxMsgs:    maxMessages,
			MaxBytes:   maxBytes,
			Duplicates: 20 * time.Minute,
		})
		if err != nil {
			return nil, err
		}
		return stream, nil
	}
	return stream, err
}
