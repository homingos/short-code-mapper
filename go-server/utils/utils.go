package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	httpClient "github.com/homingos/campaign-svc/lib/http"
	"github.com/homingos/flam-go-common/errors"

	"github.com/golang-jwt/jwt"
	"github.com/homingos/campaign-svc/config"
	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/campaign-svc/models"
	"github.com/homingos/campaign-svc/types/consts"
	"go.uber.org/zap"
)

func IsZeroVal(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func ToPointer[T any](v T) *T {
	return &v
}

type User struct {
	ID       string `json:"user_id"`
	Email    string `json:"email"`
	ClientID string `json:"client_id"`
	Role     string `json:"role"`
}

func StructToMap(item interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	if item == nil {
		return res
	}
	v := reflect.TypeOf(item)
	reflectValue := reflect.ValueOf(item)
	reflectValue = reflect.Indirect(reflectValue)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		tag := strings.Split(v.Field(i).Tag.Get("json"), ",")[0]
		field := reflectValue.Field(i).Interface()
		fieldType := v.Field(i).Type.Kind()
		if fieldType == reflect.Struct {
			nested := StructToMap(field) // <-- use recursive self
			if tag == "" {               // inline struct
				for k, val := range nested {
					res[k] = val
				}
			} else {
				res[tag] = nested
			}
			continue
		}
		if tag == "" || tag == "-" {
			continue
		}
		if !IsZeroVal(field) || fieldType == reflect.Bool {
			res[tag] = field
		}
	}
	return res
}

func DecodeJWT(tokenString string) (*User, error) {
	// Parse the token
	conf := config.GetAppConfig()
	secret := conf.JWT.SecretKey
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user := &User{
			ID:       claims["user_id"].(string),
			Email:    claims["email"].(string),
			ClientID: claims["client_id"].(string),
		}
		if claims["role"] != nil {
			user.Role = claims["role"].(string)
		}
		return user, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func ParseToken(authHeader string) string {
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return authHeader
}

func GetIPAddress(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}
	return ip
}

func GetMACAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, i := range interfaces {
		if i.Flags&net.FlagUp != 0 && !strings.HasPrefix(i.Name, "lo") {
			mac := i.HardwareAddr.String()
			if mac != "" {
				return mac
			}
		}
	}
	return ""
}

func PublishCampaignMail(campaign *models.Campaign, lgr *zap.SugaredLogger, triggerImage, trackType, notificationType string, creditType string, Balance int32, Unlimited bool) error {
	fmt.Println("PublishCampaignMail")
	httpClient := httpClient.NewClient(lgr)
	apiURL := fmt.Sprintf("%s/%s", config.LoadConfig().UserSvcBaseURL, "api/v1/users/service/send/mail")
	headers := map[string]string{
		"Authorization": config.LoadConfig().InterServiceToken,
	}

	fmt.Println("Balance: ", Balance)
	fmt.Println("creditType: ", creditType)

	sendCampaignMailDto := dtos.SendCampaignMailDto{
		CampaignId:       campaign.ShortCode,
		CreatedAt:        int64(campaign.CreatedAt),
		Name:             campaign.Name,
		ClientId:         campaign.ClientId.Hex(),
		Email:            campaign.CreatedBy.Email,
		CreatedBy:        campaign.CreatedBy,
		TriggerImage:     triggerImage,
		NotificationType: notificationType,
		TrackType:        trackType,
		CreditType:       creditType,
		UpdatedAt:        int64(campaign.UpdatedAt),
		Balance:          Balance,
		Unlimited:        Unlimited,
	}
	_, err := httpClient.DoPost(apiURL, sendCampaignMailDto, headers)
	if err != nil {
		lgr.Info("failed to send campaign mail")
	}
	return nil
}

func CampaignPushNotif(campaign *models.Campaign, lgr *zap.SugaredLogger, Type string) error {
	httpClient := httpClient.NewClient(lgr)
	apiURL := fmt.Sprintf("%s/%s", config.LoadConfig().UserSvcBaseURL, "api/v1/users/service/notification")
	headers := map[string]string{
		"Authorization": config.LoadConfig().InterServiceToken,
	}

	var sendPush dtos.SendCampaignPushNotifDto
	if Type == "new_template_added" {
		sendPush = dtos.SendCampaignPushNotifDto{
			RecipientType: "all_users",
			SendPush:      false,
			NotifType:     Type,
		}
	} else {
		sendPush = dtos.SendCampaignPushNotifDto{
			CampaignId:    campaign.ShortCode, // not in handler, of old user-svc
			ShortCode:     campaign.ShortCode, // not in handler, of old user-svc
			UserID:        campaign.CreatedBy.ID,
			RecipientType: "user",
			CreatedAt:     time.Now().UnixMilli(),
			Name:          campaign.Name, // not in handler, of old user-svc
			ClientId:      campaign.ClientId.Hex(),
			SendPush:      true,
			Variables: map[string]interface{}{
				"campaign_name": campaign.Name,
				"campaign_id":   campaign.ID.Hex(),
				"short_code":    campaign.ShortCode,
			},
			NotifType: Type,
		}
	}

	_, err := httpClient.DoPost(apiURL, sendPush, headers)
	if err != nil {
		lgr.Info("failed to send campaign push notification")
	}
	return nil
}

func GetClientCampaignExpiry(RegUserID string) (*dtos.ExpiryDurationWithUserName, error) {
	httpClient := http.Client{
		Timeout: time.Second * 30,
	}
	apiURL := fmt.Sprintf("%s/%s", config.LoadConfig().UserSvcBaseURL, "api/v1/users/service/user_details?register_user_id="+RegUserID)

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}
	req.Header.Set(consts.Authorization, config.LoadConfig().InterServiceToken)
	req.Header.Set(consts.ContentType, consts.ApplicationJson)

	res, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("API request failure: %v", err)
		return nil, errors.InternalServerError(err.Error())
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Failed to read response body: ", err)
		return nil, errors.InternalServerError(err.Error())
	}

	var response dtos.GatewayResponse[dtos.ClientExpiryRespDto]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Status != 200 {
		return nil, fmt.Errorf("unexpected status: %d, message: %s", response.Status, response.Message)
	}

	duration, err := AddDurationToCurrentTime(response.Data.Validity)
	if err != nil {
		return nil, fmt.Errorf("failed to add duration to current time: %w", err)
	}

	return &dtos.ExpiryDurationWithUserName{
		Duration: duration,
		UserName: response.Data.Name,
	}, nil
}

func CheckEscrowBalance(ClientID string) (*[]dtos.CheckEscrowDto, error) {
	httpClient := httpClient.NewClient(nil)
	apiURL := fmt.Sprintf("%s/%s", config.LoadConfig().PaymentSvcBaseURL, "api/v1/credit/allowance")
	fmt.Println(apiURL)
	headers := map[string]string{
		"Authorization": config.LoadConfig().InterServiceToken,
	}
	fmt.Println(headers)
	checkBalanceRequest := map[string]interface{}{
		"client_id": ClientID,
	}
	res, err := httpClient.DoPost(apiURL, checkBalanceRequest, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to check escrow balance1: %w", err)
	}

	var response dtos.CheckEscrowResDto
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error {
		return nil, fmt.Errorf("failed to check escrow balance: %s", response.Message)
	}

	return &response.Data, nil
}
func AdjustEscrowCredits(EscrowCreditRequest dtos.AdjustEscrowCreditsRequest, lgr *zap.SugaredLogger) (string, error) {
	httpClient := httpClient.NewClient(lgr)
	apiURL := fmt.Sprintf("%s/%s", config.LoadConfig().PaymentSvcBaseURL, "api/v1/credit/escrow")
	headers := map[string]string{
		"Authorization": config.LoadConfig().InterServiceToken,
	}
	res, err := httpClient.DoPost(apiURL, EscrowCreditRequest, headers)
	if err != nil {
		if err.Error() == "failed to consume credit" {
			return "", err
		}
		fmt.Println("error consume credit request: ", err)
		return "", fmt.Errorf("%s", err.Error())
	}

	var response dtos.AdjustEscrowCreditsResponse
	if err := json.Unmarshal(res, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error {
		if response.Message == "no credits available" {
			return "", fmt.Errorf("no credits available")
		}
		return "", fmt.Errorf("failed to adjust escrow credits: %s", response.Message)
	}

	return response.Data.CreditAllowanceID, nil
}

func ConsumeCredit(data *dtos.CampaignExperienceDto, CreditAllowanceID string, lgr *zap.SugaredLogger, UserID string) (*dtos.ConsumeCreditResDto, error) {
	httpClient := httpClient.NewClient(lgr)
	apiURL := fmt.Sprintf("%s/%s", config.LoadConfig().PaymentSvcBaseURL, "api/v1/credit/consume")
	headers := map[string]string{
		"Authorization": config.LoadConfig().InterServiceToken,
	}
	consumeCreditRequest := &dtos.ConsumeCreditRequest{
		RefID:             data.Campaign.ShortCode,
		RefType:           "CAMPAIGN",
		RefName:           data.Campaign.Name,
		CreditAllowanceID: CreditAllowanceID,
		UserID:            UserID,
	}
	res, err := httpClient.DoPost(apiURL, consumeCreditRequest, headers)
	if err != nil {
		lgr.Infof("Error consuming credit: %v", err)
		return nil, err
	}
	var response dtos.ConsumeCreditsResponse
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error {
		if response.Message == "no credits available" {
			return nil, fmt.Errorf("no credits available")
		}
		return nil, fmt.Errorf("failed to conume credits: %s", response.Message)
	}

	return &response.Data, nil
}

func ValidateTempalateFitTypes(typ string) error {
	validTypes := []string{"INSIDE", "OUTSIDE"}

	if !slices.Contains(validTypes, typ) {
		return fmt.Errorf("invalid fit type: %s. Valid types are: %s", typ, strings.Join(validTypes, ", "))
	}

	return nil
}

func AddDurationToCurrentTime(validity dtos.Validity) (int64, error) {
	currentTime := time.Now().UTC()
	var futureTime time.Time

	switch validity.Unit {
	case "YEAR":
		futureTime = currentTime.AddDate(int(validity.Value), 0, 0) // Add years
	case "MONTH":
		futureTime = currentTime.AddDate(0, int(validity.Value), 0) // Add months
	case "DAY":
		futureTime = currentTime.AddDate(0, 0, int(validity.Value)) // Add days
	default:
		return 0, fmt.Errorf("invalid unit: %s", validity.Unit)
	}

	Duration := futureTime.UnixMilli()

	return Duration, nil
}

func PrepareAndCaptureAssets(db *mongo.Database, updateMap map[string]interface{}, updatedExp, oldExp *models.Experience) {
	images := []models.Image{}
	videos := []models.Video{}
	glbs := []models.GLB{}

	if url, ok := updateMap["image_url"].(string); ok && url != "" {
		images = append(images, models.Image{K: "original", V: url})
	}

	if url, ok := updateMap["video_url"].(string); ok && url != "" {
		videos = append(videos, models.Video{K: "original", V: url})
	}

	if maskUrl, ok := updateMap["mask_url"].(string); ok && maskUrl != "" {
		// If no new video URL is provided, use existing experience video URL if available
		for _, video := range updatedExp.Videos {
			if video.K == "original" {
				videos = append(videos, models.Video{K: "original", V: video.V})
				break
			}
		}
		videos = append(videos, models.Video{K: "mask", V: maskUrl})
	}

	// add only new glbs
	for _, oldGlb := range oldExp.GLBs {
		switch oldGlb.K {
		case "original_glb":
			if url, ok := updateMap["glb"].(string); ok && url != "" {
				if oldGlb.V == url {
					delete(updateMap, "glb")
				}
			}
		case "original_usdz":
			if url, ok := updateMap["usdz"].(string); ok && url != "" {
				if oldGlb.V == url {
					delete(updateMap, "usdz")
				}
			}
		case "original_obj":
			if url, ok := updateMap["obj"].(string); ok && url != "" {
				if oldGlb.V == url {
					delete(updateMap, "obj")
				}
			}
		}
	}

	if url, ok := updateMap["obj"].(string); ok && url != "" {
		glbs = append(glbs, models.GLB{K: "original_obj", V: url})
	}

	if url, ok := updateMap["usdz"].(string); ok && url != "" {
		glbs = append(glbs, models.GLB{K: "original_usdz", V: url})
	}

	if url, ok := updateMap["glb"].(string); ok && url != "" {
		glbs = append(glbs, models.GLB{K: "original_glb", V: url})
	}

	if len(images) > 0 || len(videos) > 0 || len(glbs) > 0 {
		CaptureAssets(db, updatedExp.CampaignID, videos, images, glbs)
	}
}

func CaptureAssets(db *mongo.Database, campaignID primitive.ObjectID, videos []models.Video, images []models.Image, glbs []models.GLB) {
	// fetch shortcode from db
	campaignCollection := db.Collection("campaigns")
	campaign := models.Campaign{}
	err := campaignCollection.FindOne(context.Background(), bson.M{"_id": campaignID}).Decode(&campaign)
	if err != nil {
		fmt.Println("Error fetching campaign:", err)
		return
	}

	objectsToUpdate := make([]interface{}, 0)

	// Capture videos
	var videoDoc, imageDoc models.Asset

	for _, video := range videos {
		if video.K == "original" {
			videoDoc = createAssetDoc(video.V, "VIDEO", campaign.ShortCode, campaignID, campaign.ClientId, campaign.CreatedBy)
			continue
		}

		if video.K == "mask" {
			videoDoc.MaskedUrl = &video.V
			continue
		}
	}

	if videoDoc.ID != primitive.NilObjectID {
		objectsToUpdate = append(objectsToUpdate, videoDoc)
	}

	// Capture images
	for _, image := range images {
		if image.K == "original" {
			imageDoc = createAssetDoc(image.V, "IMAGE", campaign.ShortCode, campaignID, campaign.ClientId, campaign.CreatedBy)
			objectsToUpdate = append(objectsToUpdate, imageDoc)
			break
		}

	}

	var glbDoc, usdzDoc, objDoc, blendDoc, textureDoc models.Asset

	// Capture GLBs
	for _, glb := range glbs {
		assetDoc := createAssetDoc(
			glb.V,
			"3D",
			campaign.ShortCode,
			campaignID,
			campaign.ClientId,
			campaign.CreatedBy,
		)

		// only capture original 3D assets, and in future we may insert identifier for all these types
		switch glb.K {
		case "original_glb":
			glbDoc = assetDoc
			objectsToUpdate = append(objectsToUpdate, glbDoc)
		case "original_usdz":
			usdzDoc = assetDoc
			objectsToUpdate = append(objectsToUpdate, usdzDoc)
		case "original_obj":
			objDoc = assetDoc
			objectsToUpdate = append(objectsToUpdate, objDoc)
		case "blend_file":
			blendDoc = assetDoc
			objectsToUpdate = append(objectsToUpdate, blendDoc)
		case "texture_file":
			textureDoc = assetDoc
			objectsToUpdate = append(objectsToUpdate, textureDoc)
		}
	}

	// Insert all the assets into asset DB
	assetCollection := db.Collection("assets")
	_, err = assetCollection.InsertMany(context.Background(), objectsToUpdate)
	if err != nil {
		fmt.Println("Error inserting assets:", err)
		return
	}

}

func createAssetDoc(url, assetType, shortCode string, campaignID, clientID primitive.ObjectID, createdBy *models.User) models.Asset {
	return models.Asset{
		ID:         primitive.NewObjectID(),
		Url:        url,
		Type:       assetType,
		ShortCode:  shortCode,
		CampaignID: campaignID,
		ClientID:   clientID,
		CreatedAt:  time.Duration(time.Now().UnixMilli()),
		CreatedBy: &models.User{
			ID:    createdBy.ID,
			Email: createdBy.Email,
		},
	}
}

func GetClientCreditsVisibilityByID(ClientId string) (*dtos.CreditVisibilityDto, error) {
	httpClient := http.Client{
		Timeout: time.Second * 30,
	}
	apiURL := fmt.Sprintf("%s/%s", config.LoadConfig().UserSvcBaseURL, "api/v1/client/service?client_id="+ClientId)

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}
	req.Header.Set(consts.Authorization, config.LoadConfig().InterServiceToken)
	req.Header.Set(consts.ContentType, consts.ApplicationJson)

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	var response dtos.CreditVisibilityDto
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Status != 200 {
		return nil, fmt.Errorf("unexpected status: %d, message: %s", response.Status, response.Message)
	}

	return &response, nil
}

func PopulateButtonSegments(inp dtos.SegmentData) (models.Segments, []models.InteractiveButton, dtos.SegmentInfo, error) {
	segments := models.Segments{}

	SegmentInfo := dtos.SegmentInfo{}
	SegmentImageInfo := []dtos.UpdatedVariantInfo{}
	SegmentVideoInfo := []dtos.UpdatedVariantInfo{}
	AllVideoInfo := []dtos.SegmentVideo{}
	buttons := []models.InteractiveButton{}
	segmentMarkers := []models.SegmentMarker{}
	var DefaultMarkerID string
	for i, segment := range inp.ButtonSegments {
		var Next string
		mkrID := primitive.NewObjectID().Hex()
		if segment.MarkerID != "" {
			mkrID = segment.MarkerID
		}
		inp.ButtonSegments[i].MarkerID = mkrID
		if DefaultMarkerID == "" {
			DefaultMarkerID = mkrID
		}
		if i < len(inp.ButtonSegments)-1 {
			if inp.ButtonSegments[i+1].MarkerID != "" {
				Next = inp.ButtonSegments[i+1].MarkerID
			} else {
				Next = primitive.NewObjectID().Hex()
				inp.ButtonSegments[i+1].MarkerID = Next
			}
		} else {
			Next = inp.ButtonSegments[0].MarkerID
		}
		btn := models.InteractiveButton{
			ID:            primitive.NewObjectID().Hex(),
			Type:          segment.ButtonType,
			AssetFileName: segment.AssetFileName,
			AssetUrl:      segment.AssetURL,
			Color:         segment.Color,
			MarkerId:      mkrID,
			Position:      segment.Position,
			MaskId:        segment.MaskID,
		}
		SegmentImageInfo = append(SegmentImageInfo, dtos.UpdatedVariantInfo{
			MarkerId: mkrID,
			AssetURL: segment.AssetURL,
			Type:     "IMAGE",
		})
		AllVideoInfo = append(AllVideoInfo, dtos.SegmentVideo{
			MarkedID:    mkrID,
			OriginalURL: segment.OriginalVideoURL,
			MaskURL:     segment.MaskURL,
		})
		buttons = append(buttons, btn)
		segmentMarker := models.SegmentMarker{
			Id:           mkrID,
			Next:         Next,
			ShowElements: segment.ShowElements,
			Videos: models.VideoObject{
				Original:   segment.OriginalVideoURL,
				Mask:       segment.MaskURL,
				MergeVideo: segment.MergeVideo,
			},
		}

		if segment.Orientation != nil {
			segmentMarker.Videos.Orientation = *segment.Orientation
		}

		if len(segment.ShowElements) == 0 {
			segmentMarker.ShowElements = []int{}
		}
		SegmentVideoInfo = append(SegmentVideoInfo, dtos.UpdatedVariantInfo{
			MarkerId: mkrID,
			VideoURL: segment.OriginalVideoURL,
			MaskURL:  segment.MaskURL,
			Type:     "VIDEO",
		})
		segmentMarkers = append(segmentMarkers, segmentMarker)
	}
	segments.UseSegmentedElements = inp.UseSegmentedElement
	segments.UseMarkerVideo = true
	segments.Markers = segmentMarkers
	segments.BackColor = "#FFFFFF"
	segments.FlushColor = "#000000"
	segments.Default = DefaultMarkerID
	SegmentInfo.ImageInfo = SegmentImageInfo
	SegmentInfo.VideoInfo = SegmentVideoInfo
	SegmentInfo.VideoUrls = AllVideoInfo
	SegmentInfo.ProcessStitchVideo = true

	return segments, buttons, SegmentInfo, nil
}

func UpdateExperienceModelV2(existingExp *models.Experience, inp dtos.SegmentData) ([]models.InteractiveButton, []models.SegmentMarker, dtos.SegmentInfo) {
	SegmentInfo := dtos.SegmentInfo{}
	AllVideoInfo := []dtos.SegmentVideo{}
	updatedImageVariant := []dtos.UpdatedVariantInfo{}
	updatedVideoVariant := []dtos.UpdatedVariantInfo{}

	existingBtns := make(map[string]models.InteractiveButton)
	for _, b := range existingExp.Variant.Buttons {
		existingBtns[b.MarkerId] = b
	}

	var existMarkerLength int
	existingMarkers := make(map[string]models.SegmentMarker)
	if existingExp.Variant.Segments != nil {
		for _, m := range existingExp.Variant.Segments.Markers {
			existingMarkers[m.Id] = m
		}
		existMarkerLength = len(existingExp.Variant.Segments.Markers)
	}

	var updatedBtns []models.InteractiveButton
	var updatedMarkers []models.SegmentMarker

	for i, newInp := range inp.ButtonSegments {
		mkrID := newInp.MarkerID
		if mkrID == "" {
			mkrID = primitive.NewObjectID().Hex()
		}
		inp.ButtonSegments[i].MarkerID = mkrID
		var Next string
		if i < len(inp.ButtonSegments)-1 {
			if inp.ButtonSegments[i+1].MarkerID != "" {
				Next = inp.ButtonSegments[i+1].MarkerID
			} else {
				Next = primitive.NewObjectID().Hex()
				inp.ButtonSegments[i+1].MarkerID = Next
			}
		} else {
			Next = inp.ButtonSegments[0].MarkerID
		}
		var btn models.InteractiveButton
		if oldBtn, ok := existingBtns[mkrID]; ok {
			btn = oldBtn

			if oldBtn.AssetUrl != newInp.AssetURL {
				updatedImageVariant = append(updatedImageVariant, dtos.UpdatedVariantInfo{
					MarkerId: mkrID,
					AssetURL: newInp.AssetURL,
					Type:     "IMAGE",
				})
				btn.AssetUrl = newInp.AssetURL
				btn.CompressedAssetUrl = ""
			}
			btn.Color = newInp.Color
			btn.AssetFileName = newInp.AssetFileName
			btn.Type = newInp.ButtonType
			btn.Position = newInp.Position
			btn.Scale = newInp.Scale
			btn.MaskId = newInp.MaskID
		} else {
			updatedImageVariant = append(updatedImageVariant, dtos.UpdatedVariantInfo{
				MarkerId: mkrID,
				AssetURL: newInp.AssetURL,
				Type:     "IMAGE",
			})
			btn = models.InteractiveButton{
				ID:            primitive.NewObjectID().Hex(),
				Type:          newInp.ButtonType,
				AssetFileName: newInp.AssetFileName,
				AssetUrl:      newInp.AssetURL,
				Color:         newInp.Color,
				MarkerId:      mkrID,
				Position:      newInp.Position,
				Scale:         newInp.Scale,
				MaskId:        newInp.MaskID,
			}
		}
		updatedBtns = append(updatedBtns, btn)

		var marker models.SegmentMarker
		if oldMkr, ok := existingMarkers[mkrID]; ok {
			marker = oldMkr
			marker.ShowElements = newInp.ShowElements
			if len(newInp.ShowElements) == 0 {
				marker.ShowElements = []int{}
			}
			if oldMkr.Videos.Original != newInp.OriginalVideoURL || oldMkr.Videos.Mask != newInp.MaskURL {
				updatedVideoVariant = append(updatedVideoVariant, dtos.UpdatedVariantInfo{
					MarkerId: mkrID,
					VideoURL: newInp.OriginalVideoURL,
					MaskURL:  newInp.MaskURL,
					Type:     "VIDEO",
				})
				marker.Videos = models.VideoObject{
					Original:   newInp.OriginalVideoURL,
					Mask:       newInp.MaskURL,
					MergeVideo: newInp.MergeVideo,
				}
				marker.Stime = 0
				marker.Etime = 0
				SegmentInfo.ProcessStitchVideo = true
			}
			// if only merge video is changed and not original or mask video
			if newInp.MergeVideo != "" && oldMkr.Videos.MergeVideo != newInp.MergeVideo {
				marker.Videos.MergeVideo = newInp.MergeVideo
			}
			if newInp.Orientation != nil {
				marker.Videos.Orientation = *newInp.Orientation
			}

		} else {
			updatedVideoVariant = append(updatedVideoVariant, dtos.UpdatedVariantInfo{
				MarkerId: mkrID,
				VideoURL: newInp.OriginalVideoURL,
				MaskURL:  newInp.MaskURL,
				Type:     "VIDEO",
			})
			marker = models.SegmentMarker{
				Id:           mkrID,
				Next:         mkrID,
				ShowElements: newInp.ShowElements,
				Videos: models.VideoObject{
					Original:   newInp.OriginalVideoURL,
					Mask:       newInp.MaskURL,
					MergeVideo: newInp.MergeVideo,
				},
			}
			SegmentInfo.ProcessStitchVideo = true
			if len(newInp.ShowElements) == 0 {
				marker.ShowElements = []int{}
			}
			if newInp.Orientation != nil {
				marker.Videos.Orientation = *newInp.Orientation
			}
		}
		AllVideoInfo = append(AllVideoInfo, dtos.SegmentVideo{
			MarkedID:    mkrID,
			OriginalURL: marker.Videos.Original,
			MaskURL:     marker.Videos.Mask,
		})
		marker.Next = Next
		updatedMarkers = append(updatedMarkers, marker)
	}

	if existMarkerLength != len(inp.ButtonSegments) {
		SegmentInfo.ProcessStitchVideo = true
	}

	SegmentInfo.VideoUrls = AllVideoInfo
	SegmentInfo.ImageInfo = updatedImageVariant
	SegmentInfo.VideoInfo = updatedVideoVariant
	return updatedBtns, updatedMarkers, SegmentInfo
}
