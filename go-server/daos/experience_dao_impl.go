package dao

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/homingos/campaign-svc/dtos"
	redisStorage "github.com/homingos/campaign-svc/lib/redis"
	"github.com/homingos/campaign-svc/models"
	"github.com/homingos/campaign-svc/types/consts"
	"github.com/homingos/campaign-svc/utils"
	"github.com/homingos/flam-go-common/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type ExperienceDaoImpl struct {
	lgr         *zap.SugaredLogger
	db          *mongo.Database
	redisClient *redisStorage.RedisClient
}

func createExperienceIndexes(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := db.Collection(consts.ExperienceCollection)
	indexes := []mongo.IndexModel{{
		Keys:    bson.D{{Key: "campaign_id", Value: 1}, {Key: "image_hash", Value: 1}},
		Options: options.Index().SetUnique(true),
	}}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	coll.Indexes().CreateMany(ctx, indexes, opts)
}

func NewExperienceDao(lgr *zap.SugaredLogger, db *mongo.Database, redisClient *redisStorage.RedisClient) *ExperienceDaoImpl {
	createExperienceIndexes(db)
	return &ExperienceDaoImpl{lgr, db, redisClient}
}

func (impl *ExperienceDaoImpl) CreateExperienceDao(expCreationDto *dtos.ExperienceCreationDto) (*models.MediaProcess, error) {
	campaignID, err := primitive.ObjectIDFromHex(expCreationDto.CampaignID)
	if err != nil {
		return nil, err
	}
	if expCreationDto.UIElements != nil {
		for i, element := range expCreationDto.UIElements.Elements {
			if element.Type == "BANNER" {
				BannerDto := &dtos.GetBannerDto{
					Element:  element,
					ImageUrl: expCreationDto.ImageURL,
					QrCode:   expCreationDto.QrCode,
				}
				variant := impl.GetElementBannerVariant(BannerDto)
				BannerDto.Element.Variant = variant
				expCreationDto.UIElements.Elements[i] = BannerDto.Element
			}
		}
	}

	if expCreationDto.UIElements == nil {
		expCreationDto.UIElements = &models.UIElements{}
	}

	images := []models.Image{}
	videos := []models.Video{}
	glbs := []models.GLB{}
	var exp models.Experience
	exp.ID = primitive.NewObjectID()
	exp.Name = expCreationDto.Name
	exp.Canvas.IOS = 2100
	exp.CampaignID = campaignID
	exp.Status = expCreationDto.Status
	exp.Variant = *expCreationDto.Variant
	exp.IsActive = true
	exp.UIElements = expCreationDto.UIElements
	exp.QrCode = expCreationDto.QrCode
	exp.ImageHash = expCreationDto.ImageHash
	exp.CreatedBy = expCreationDto.CreatedBy
	exp.TemplateDetails = expCreationDto.TemplateDetails
	exp.TemplateCategory = expCreationDto.TemplateCategory
	exp.PlaybackScale = expCreationDto.PlaybackScale
	if exp.PlaybackScale == 0 {
		exp.PlaybackScale = 1
	}
	if expCreationDto.Rewards == nil {
		exp.Rewards = models.Rewards{
			Enabled: false,
		}
	} else {
		exp.Rewards = *expCreationDto.Rewards
	}
	if expCreationDto.ImageURL != "" {
		images = append(images, models.Image{K: "original", V: expCreationDto.ImageURL})
	}
	if expCreationDto.OriginalInputUrl != "" {
		images = append(images, models.Image{K: "original_input", V: expCreationDto.OriginalInputUrl})
	}
	if expCreationDto.SpawnImage != "" {
		images = append(images, models.Image{K: "spawn", V: expCreationDto.SpawnImage})
	}
	if expCreationDto.MaskedPhotoURL != "" {
		url := expCreationDto.MaskedPhotoURL
		if !strings.Contains(url, "zingcam-dev") {
			url = strings.Replace(url, "storage.googleapis.com/zingcam", "zingcam.cdn.flamapp.com", 1)
		}
		images = append(images, models.Image{K: "masked_photo", V: url})
	}
	if expCreationDto.AudioURL != "" {
		exp.Audios = []models.Audio{{K: "original", V: expCreationDto.AudioURL}}
	}
	if expCreationDto.GLB != "" {
		glbs = append(glbs, models.GLB{K: "original_glb", V: expCreationDto.GLB})
	}
	if expCreationDto.USDZ != "" {
		glbs = append(glbs, models.GLB{K: "original_usdz", V: expCreationDto.USDZ})
	}
	if expCreationDto.OBJ != "" {
		glbs = append(glbs, models.GLB{K: "original_obj", V: expCreationDto.OBJ})
	}
	if expCreationDto.TextureFile != "" {
		glbs = append(glbs, models.GLB{K: "texture_file", V: expCreationDto.TextureFile})
	}
	if expCreationDto.BlendFile != "" {
		glbs = append(glbs, models.GLB{K: "blend_file", V: expCreationDto.BlendFile})
	}
	if expCreationDto.VideoURL != "" {
		videos = append(videos, models.Video{K: "original", V: expCreationDto.VideoURL})
		// videos = append(videos, models.Video{K: "playback", V: expCreationDto.VideoURL})
	}
	if expCreationDto.WebmUrl != "" {
		videos = append(videos, models.Video{K: "webm", V: expCreationDto.WebmUrl})
	}
	if expCreationDto.MaskUrl != "" {
		videos = append(videos, models.Video{K: "mask", V: expCreationDto.MaskUrl})
	}
	exp.Videos = videos
	exp.Images = images
	exp.GLBs = glbs
	go utils.CaptureAssets(impl.db, exp.CampaignID, exp.Videos, exp.Images, exp.GLBs)
	if exp.Variant.ScaleAxis.X == 0 {
		exp.Variant.ScaleAxis.X = 1
	}
	if exp.Variant.ScaleAxis.Y == 0 {
		exp.Variant.ScaleAxis.Y = 1
	}
	if expCreationDto.EngagmentOptions != nil {
		exp.EngagmentOptions = expCreationDto.EngagmentOptions
	}
	exp.Scene = expCreationDto.Scene
	exp.Overlay = expCreationDto.Overlay
	exp.QrPanel = expCreationDto.QrPanel
	exp.CreatedAt = time.Duration(time.Now().UnixMilli())
	exp.UpdatedAt = time.Duration(time.Now().UnixMilli())
	if exp.Status == consts.Processing && expCreationDto.Variant.Class == 2 && exp.Variant.TrackType == "GROUND" {
		exp.Status = consts.Processed
	}
	coll := impl.db.Collection(consts.ExperienceCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = coll.InsertOne(ctx, exp)
	if err != nil {
		return nil, err
	}

	var campaign models.Campaign
	query := bson.M{"_id": campaignID}
	coll = impl.db.Collection(consts.CampaignCollection)
	Update := bson.M{
		"$set": bson.M{
			"updated_at": time.Duration(time.Now().UnixMilli()),
		},
	}
	result := coll.FindOneAndUpdate(ctx, query, Update)
	if err := result.Err(); err != nil {
		return nil, err
	}
	err = result.Decode(&campaign)
	if err != nil {
		return nil, err
	}
	mediaProcess := models.MediaProcess{
		Experience: exp,
		ShortCode:  campaign.ShortCode,
		Name:       campaign.Name,
		CreatedBy:  *campaign.CreatedBy,
		ClientId:   campaign.ClientId,
	}
	return &mediaProcess, nil
}

func (impl *ExperienceDaoImpl) CreateBulkExperienceDao(sessionCtx *mongo.SessionContext, Experiences []*models.Experience) (string, error) {
	coll := impl.db.Collection(consts.ExperienceCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	document := make([]interface{}, len(Experiences))
	for i, Exp := range Experiences {
		document[i] = Exp
	}
	if sessionCtx != nil {
		ctx = *sessionCtx
	}
	_, err := coll.InsertMany(ctx, document)
	if err != nil {
		return "", err
	}
	return "Campaign Created Successfully", nil
}

func (impl *ExperienceDaoImpl) DuplicateExperienceWithNewCampaignIdDao(ctx context.Context, Exp *models.Experience) (*models.Experience, error) {
	newExp := *Exp
	newExp.ID = primitive.NewObjectID()
	newExp.CreditDeduct = true
	newExp.CreatedAt = time.Duration(time.Now().UnixMilli())
	newExp.UpdatedAt = time.Duration(time.Now().UnixMilli())

	coll := impl.db.Collection(consts.ExperienceCollection)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := coll.InsertOne(ctx, newExp)
	if err != nil {
		return nil, err
	}

	return &newExp, nil
}

func (impl *ExperienceDaoImpl) GetExperienceDao(objID primitive.ObjectID) (interface{}, error) {
	filter := bson.M{
		"_id":       objID,
		"is_active": true,
	}
	match := bson.M{
		"$match": filter,
	}
	project := bson.M{
		"$project": bson.M{
			"_id":               0,
			"id":                "$_id",
			"name":              1,
			"canvas":            1,
			"is_active":         1,
			"created_at":        1,
			"scan_text":         1,
			"status":            1,
			"ui_elements":       1,
			"share_meta":        1,
			"variant":           1,
			"campaign_id":       1,
			"reward_enabled":    1,
			"rewards":           1,
			"playback_scale":    1,
			"overlay":           1,
			"template_details":  1,
			"mask":              1,
			"scene":             1,
			"workflow_error":    1,
			"template_category": 1,
			"total_task":        1,
			"engagment_options": 1,
			"workflow_id":       1,
			"images":            bson.M{"$arrayToObject": "$images"},
			"videos":            bson.M{"$arrayToObject": "$videos"},
			"audios":            bson.M{"$arrayToObject": "$audios"},
			"3d_assets":         bson.M{"$arrayToObject": "$3d_assets"},
		},
	}
	pipelines := []bson.M{match, project}
	var data []map[string]interface{}
	coll := impl.db.Collection(consts.ExperienceCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := coll.Aggregate(ctx, pipelines)
	defer cursor.Close(ctx)

	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &data); err != nil {
		return nil, err
	}
	if len(data) > 0 {
		return data[0], nil
	}
	return nil, nil
}

func (impl *ExperienceDaoImpl) UpdateExperienceDao(objID primitive.ObjectID, updateMap map[string]interface{}, OptionalVariant ...dtos.ExperienceUpdateDto) (*dtos.UpdateResponseDto, error) {

	filter := bson.M{"_id": objID}
	var existExperience models.Experience
	updateMap["updated_at"] = time.Now().UnixMilli()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	arrayFilters := []bson.M{}
	setOperations := bson.M{}
	pullOperations := bson.M{}
	coll := impl.db.Collection(consts.ExperienceCollection)
	result := coll.FindOne(ctx, filter)
	err := result.Decode(&existExperience)
	if err != nil {
		return nil, err
	}

	Status := consts.Processing
	if status, ok := updateMap["status"]; ok {
		Status = status.(string)
	} else if existExperience.Status == consts.Failed || existExperience.Status == consts.TimedOut {
		updateMap["status"] = Status
	}

	// var updateVariantInfo []dtos.UpdatedVariantInfo
	var SegmentInfo *dtos.SegmentInfo
	if len(OptionalVariant) > 0 {
		expUpdateDto := OptionalVariant[0]
		if expUpdateDto.Variant != nil && len(expUpdateDto.SegmentData.ButtonSegments) > 0 {
			Buttons, Markers, Segmentinfo := utils.UpdateExperienceModelV2(&existExperience, OptionalVariant[0].SegmentData)
			if expUpdateDto.Variant.Segments == nil {
				expUpdateDto.Variant.Segments = &models.Segments{
					Markers:              Markers,
					BackColor:            "#FFFFFF",
					FlushColor:           "#000000",
					UseSegmentedElements: expUpdateDto.SegmentData.UseSegmentedElement,
					Default:              Markers[0].Id,
				}
			} else {
				expUpdateDto.Variant.Segments.Markers = Markers
				if len(Markers) > 0 {
					expUpdateDto.Variant.Segments.Default = Markers[0].Id
				}
				expUpdateDto.Variant.Segments.UseSegmentedElements = expUpdateDto.SegmentData.UseSegmentedElement
			}
			SegmentInfo = &Segmentinfo
			if SegmentInfo.ProcessStitchVideo {
				expUpdateDto.Variant.Segments.UseMarkerVideo = true
				pullOperations["videos"] = bson.M{}
			}
			expUpdateDto.Variant.Buttons = Buttons
			updateMap["variant"] = utils.StructToMap(expUpdateDto.Variant)
			if len(Segmentinfo.ImageInfo) > 0 || len(Segmentinfo.VideoInfo) > 0 {
				updateMap["status"] = Status
			}
		}
	}
	delete(updateMap, "segment_data")

	//Edit logs
	var EditLog dtos.EditLogDto
	EditLog.ExperienceId = existExperience.ID
	EditLog.Before = &existExperience

	if deleteImage, ok := updateMap["delete_image"]; ok && deleteImage.(bool) {
		if existExperience.Status != consts.Draft {
			return nil, fmt.Errorf("assets can only be deleted if status is draft")
		}
		pullOperations["images"] = bson.M{}
	}
	pullOperationsKeys := []string{}
	if deleteSpawnImage, ok := updateMap["delete_spawn"]; ok && deleteSpawnImage.(bool) {
		if _, oksI := updateMap["spawn_image"]; oksI {
			return nil, fmt.Errorf("conflict while updating")
		}
		pullOperationsKeys = append(pullOperationsKeys, "spawn")
		pullOperationsKeys = append(pullOperationsKeys, "compressed_spawn")
		pullOperations["images"] = bson.M{
			"k": bson.M{
				"$in": pullOperationsKeys,
			},
		}
	}

	if deleteMaskPtr, ok := updateMap["delete_mask"].(*bool); ok && deleteMaskPtr != nil && *deleteMaskPtr {
		if _, okm := updateMap["mask_url"]; okm {
			return nil, fmt.Errorf("conflict while updating")
		}
		pullOperations["videos"] = bson.M{
			"k": bson.M{
				"$in": []string{"mask", "compressed", "hls", "dash", "compressed_playback", "webm"},
			},
		}
		updateMap["status"] = Status
	}

	if deleteVideo, ok := updateMap["delete_video"]; ok && deleteVideo.(bool) {
		if existExperience.Status != consts.Draft {
			return nil, fmt.Errorf("assets can only be deleted if status is draft")
		}
		if _, okm := updateMap["video_url"]; okm {
			return nil, fmt.Errorf("conflict while updating")
		}
		pullOperations["videos"] = bson.M{}
	}

	delete(updateMap, "delete_video")
	delete(updateMap, "delete_image")

	if imageURL, ok := updateMap["image_url"]; ok {
		for _, image := range existExperience.Images {
			if image.K == "original" && image.V == imageURL {
				delete(updateMap, "image_url")
			}
		}
	}

	if originalInputURL, ok := updateMap["orignal_input_url"]; ok {
		for _, image := range existExperience.Images {
			if image.K == "original_input" && image.V == originalInputURL {
				delete(updateMap, "orignal_input_url")
			}
		}
	}
	if spawnImageUrl, ok := updateMap["spawn_image"]; ok {
		for _, image := range existExperience.Images {
			if image.K == "spawn" && image.V == spawnImageUrl {
				delete(updateMap, "spawn_image")
			}
		}
	}

	if videoURL, ok := updateMap["video_url"]; ok {
		for _, video := range existExperience.Videos {
			if video.K == "original" && video.V == videoURL {
				delete(updateMap, "video_url")
			}
		}
	}

	if webmURL, ok := updateMap["webm_url"]; ok {
		for _, video := range existExperience.Videos {
			if video.K == "webm" && video.V == webmURL {
				delete(updateMap, "webm_url")
			}
		}
	}

	if variant, ok := updateMap["variant"].(*models.Variant); ok {
		updateMap["variant"] = variant
	}

	if engagmentOptions, ok := updateMap["engagement_options"].(*models.EngagmentOptions); ok {
		if engagmentOptions.Sharable != nil {
			updateMap["engagement_options.sharable"] = engagmentOptions.Sharable
		}
		if engagmentOptions.Name != "" {
			updateMap["engagement_options.name"] = engagmentOptions.Name
		}
		if engagmentOptions.ShadowGeneration != nil && engagmentOptions.ShadowGeneration.Enabled != nil {
			updateMap["engagement_options.shadow_generation.enabled"] = engagmentOptions.ShadowGeneration.Enabled
		}
		if engagmentOptions.ImageHarmonization != nil {
			if engagmentOptions.ImageHarmonization.Enabled != nil {
				updateMap["engagement_options.image_harmonization.enabled"] = engagmentOptions.ImageHarmonization.Enabled
			}
			if len(engagmentOptions.ImageHarmonization.ModelList) > 0 {
				updateMap["engagement_options.image_harmonization.model_list"] = engagmentOptions.ImageHarmonization.ModelList
			}
		}
		if engagmentOptions.BorderSmoothing != nil && engagmentOptions.BorderSmoothing.Enabled != nil {
			updateMap["engagement_options.border_smoothing.enabled"] = engagmentOptions.BorderSmoothing.Enabled
		}
		if engagmentOptions.PostProcessFilter != nil {
			if engagmentOptions.PostProcessFilter.Enabled != nil {
				updateMap["engagement_options.post_process_filter.enabled"] = engagmentOptions.PostProcessFilter.Enabled
			}
			if len(engagmentOptions.PostProcessFilter.FilterList) > 0 {
				updateMap["engagement_options.post_process_filter.filter_list"] = engagmentOptions.PostProcessFilter.FilterList
			}
		}
	}

	if imageURL, ok := updateMap["image_url"]; ok {
		arrayFilters = append(arrayFilters, bson.M{"ielem.k": "original"})
		setOperations["images.$[ielem].v"] = imageURL
		updateMap["status"] = Status
	}

	if OriginalInputUrl, ok := updateMap["original_input_url"]; ok {
		arrayFilters = append(arrayFilters, bson.M{"oielem.k": "original_input"})
		setOperations["images.$[oielem].v"] = OriginalInputUrl
	}

	if spawnImageUrl, ok := updateMap["spawn_image"]; ok {
		arrayFilters = append(arrayFilters, bson.M{"selem.k": "spawn"})
		setOperations["images.$[selem].v"] = spawnImageUrl
		pullOperationsKeys = append(pullOperationsKeys, "compressed", "compressed_spawn", "color_compressed", "std_compressed", "feature_image")
		pullOperations["images"] = bson.M{
			"k": bson.M{
				"$in": pullOperationsKeys,
			},
		}
		updateMap["status"] = Status
	}

	if featureImageURL, ok := updateMap["feature_image_url"]; ok {
		arrayFilters = append(arrayFilters, bson.M{"felem.k": "feature_image"})
		setOperations["images.$[felem].v"] = featureImageURL
	}

	if maskedPhotoURL, ok := updateMap["masked_photo_url"]; ok {
		arrayFilters = append(arrayFilters, bson.M{"melem.k": "masked_photo"})
		url := maskedPhotoURL.(string)
		if !strings.Contains(url, "zingcam-dev") {
			url = strings.Replace(url, "storage.googleapis.com/zingcam", "zingcam.cdn.flamapp.com", 1)
		}
		setOperations["images.$[melem].v"] = url
	}

	if videoURL, ok := updateMap["video_url"]; ok {
		updateMap["status"] = Status
		arrayFilters = append(arrayFilters, bson.M{"velem.k": "original"})
		setOperations["videos.$[velem].v"] = videoURL
	}

	if overlay, ok := updateMap["overlay"].(models.Overlay); ok {
		if overlay.Type != consts.OverlayTrasparent {
			updateMap["status"] = Status
		}
	}

	if maskURL, ok := updateMap["mask_url"]; ok {
		for _, video := range existExperience.Videos {
			if video.K == "mask" && video.V == maskURL {
				delete(updateMap, "mask_url")
				break
			} else if video.K == "mask" {
				updateMap["status"] = Status
				arrayFilters = append(arrayFilters, bson.M{"vmelem.k": "mask"})
				setOperations["videos.$[vmelem].v"] = maskURL
				break
			}
		}
	}

	if webmURL, ok := updateMap["webm_url"]; ok {
		arrayFilters = append(arrayFilters, bson.M{"welem.k": "webm"})
		setOperations["videos.$[welem].v"] = webmURL
	}

	if playbackURL, ok := updateMap["playback_url"]; ok {
		updateMap["status"] = Status
		arrayFilters = append(arrayFilters, bson.M{"pelem.k": "playback"})
		setOperations["videos.$[pelem].v"] = playbackURL
	}

	if audioURL, ok := updateMap["audio_url"]; ok {
		arrayFilters = append(arrayFilters, bson.M{"aelem.k": "original"})
		setOperations["audios.$[aelem].v"] = audioURL
	}

	if glbURL, ok := updateMap["glb"]; ok {
		arrayFilters = append(arrayFilters, bson.M{"gelem.k": "original_glb"})
		setOperations["3d_assets.$[gelem].v"] = glbURL
	}

	if usdzURL, ok := updateMap["usdz"]; ok {
		arrayFilters = append(arrayFilters, bson.M{"uelem.k": "original_usdz"})
		setOperations["3d_assets.$[uelem].v"] = usdzURL
	}

	if objURL, ok := updateMap["obj"]; ok {
		arrayFilters = append(arrayFilters, bson.M{"oelem.k": "original_obj"})
		setOperations["3d_assets.$[oelem].v"] = objURL
	}

	if blendFile, ok := updateMap["blend_file"]; ok {
		arrayFilters = append(arrayFilters, bson.M{"belem.k": "blend_file"})
		setOperations["3d_assets.$[belem].v"] = blendFile
	}

	if textureFile, ok := updateMap["texture_file"]; ok {
		arrayFilters = append(arrayFilters, bson.M{"telem.k": "texture_file"})
		setOperations["3d_assets.$[telem].v"] = textureFile
	}

	for key, value := range updateMap {
		if key != "original_input_url" && key != "delete_spawn" && key != "publish" && key != "feature_image_url" && key != "image_url" && key != "masked_photo_url" && key != "video_url" && key != "audio_url" && key != "3d_asset" && key != "playback_url" && key != "mask_url" && key != "delete_mask" && key != "spawn_image" && key != "delete_share_meta" && key != "glb" && key != "usdz" && key != "obj" && key != "blend_file" && key != "texture_file" && key != "delete_texture" && key != "delete_obj" && key != "delete_glb" && key != "delete_usdz" && key != "delete_blend" {
			setOperations[key] = value
		}
	}

	after := options.After
	if len(setOperations) > 0 {
		updateExisting := bson.M{}
		if len(setOperations) > 0 {
			updateExisting["$set"] = setOperations
		}
		opt := options.FindOneAndUpdateOptions{
			ReturnDocument: &after,
		}

		if len(arrayFilters) > 0 {
			arrayFiltersInterface := make([]interface{}, len(arrayFilters))
			for i, filter := range arrayFilters {
				arrayFiltersInterface[i] = filter
			}
			opt.ArrayFilters = &options.ArrayFilters{Filters: arrayFiltersInterface}
		}

		var experience models.Experience
		coll := impl.db.Collection(consts.ExperienceCollection)
		result := coll.FindOneAndUpdate(ctx, filter, updateExisting, &opt)
		if err := result.Err(); err != nil && err != mongo.ErrNoDocuments {
			return nil, err
		}

		if err := result.Decode(&experience); err != nil && err != mongo.ErrNoDocuments {
			return nil, err
		}
	}

	var experience models.Experience
	coll = impl.db.Collection(consts.ExperienceCollection)
	if err := coll.FindOne(ctx, filter).Decode(&experience); err != nil {
		return nil, err
	}
	setOperations = bson.M{}
	addToSetOperations := bson.M{}
	addToSetImages := []models.Image{}
	addToSetVideos := []models.Video{}
	addToset3dAssets := []models.GLB{}
	if ImageUrl, ok := updateMap["image_url"]; ok {
		alreadyExists := false
		for _, image := range experience.Images {
			if image.K == "original" {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			addToSetImages = append(addToSetImages, models.Image{K: "original", V: ImageUrl.(string)})
		}
	}
	if OriginalInputUrl, ok := updateMap["original_input_url"]; ok {
		alreadyExists := false
		for _, image := range experience.Images {
			if image.K == "original_input" {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			addToSetImages = append(addToSetImages, models.Image{K: "original_input", V: OriginalInputUrl.(string)})
		}
	}
	if SpawnImageUrl, ok := updateMap["spawn_image"]; ok {
		alreadyExists := false
		for _, image := range experience.Images {
			if image.K == "spawn" {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			addToSetImages = append(addToSetImages, models.Image{K: "spawn", V: SpawnImageUrl.(string)})
		}
	}

	if MaskedPhotoURL, ok := updateMap["masked_photo_url"]; ok {
		alreadyExists := false
		for _, image := range experience.Images {
			if image.K == "masked_photo" {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			url := MaskedPhotoURL.(string)
			if !strings.Contains(url, "zingcam-dev") {
				url = strings.Replace(url, "storage.googleapis.com/zingcam", "zingcam.cdn.flamapp.com", 1)
			}
			addToSetImages = append(addToSetImages, models.Image{K: "masked_photo", V: url})
		}
	}

	if videoURL, ok := updateMap["video_url"]; ok {
		alreadyExists := false
		for _, video := range experience.Videos {
			if video.K == "original" {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			addToSetVideos = append(addToSetVideos, models.Video{K: "original", V: videoURL.(string)})
		}
	}

	if maskURL, ok := updateMap["mask_url"]; ok {
		alreadyExists := false
		for _, video := range experience.Videos {
			if video.K == "mask" {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			setOperations["status"] = Status
			addToSetVideos = append(addToSetVideos, models.Video{K: "mask", V: maskURL.(string)})
		}
	}

	if webmURL, ok := updateMap["webm_url"]; ok {
		alreadyExists := false
		for _, video := range experience.Videos {
			if video.K == "webm" {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			addToSetVideos = append(addToSetVideos, models.Video{K: "webm", V: webmURL.(string)})
		}
	}

	if PlaybackURL, ok := updateMap["playback_url"]; ok {
		alreadyExists := false
		for _, video := range experience.Videos {
			if video.K == "playback" {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			addToSetVideos = append(addToSetVideos, models.Video{K: "playback", V: PlaybackURL.(string)})
		}
	}

	if audioURL, ok := updateMap["audio_url"]; ok {
		alreadyExists := false
		for _, audio := range experience.Audios {
			if audio.K == "original" {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			addToSetOperations["audios"] = models.Audio{K: "original", V: audioURL.(string)}
		}
	}

	if glbURL, ok := updateMap["blend_file"]; ok {
		alreadyExists := false
		for _, glb := range experience.GLBs {
			if glb.K == "blend_file" {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			addToset3dAssets = append(addToset3dAssets, models.GLB{K: "blend_file", V: glbURL.(string)})
		}
	}

	if textureFile, ok := updateMap["texture_file"]; ok {
		alreadyExists := false
		for _, glb := range experience.GLBs {
			if glb.K == "texture_file" {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			addToset3dAssets = append(addToset3dAssets, models.GLB{K: "texture_file", V: textureFile.(string)})
		}
	}

	if glbURL, ok := updateMap["glb"]; ok {
		alreadyExists := false
		for _, glb := range experience.GLBs {
			if glb.K == "original_glb" {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			addToset3dAssets = append(addToset3dAssets, models.GLB{K: "original_glb", V: glbURL.(string)})
		}
	}

	if usdzURL, ok := updateMap["usdz"]; ok {
		alreadyExists := false
		for _, glb := range experience.GLBs {
			if glb.K == "original_usdz" {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			addToset3dAssets = append(addToset3dAssets, models.GLB{K: "original_usdz", V: usdzURL.(string)})
		}
	}

	if objURL, ok := updateMap["obj"]; ok {
		alreadyExists := false
		for _, glb := range experience.GLBs {
			if glb.K == "original_obj" {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			addToset3dAssets = append(addToset3dAssets, models.GLB{K: "original_obj", V: objURL.(string)})
		}
	}

	if len(addToSetImages) > 0 {
		addToSetOperations["images"] = bson.M{"$each": addToSetImages}
	}
	if len(addToSetVideos) > 0 {
		addToSetOperations["videos"] = bson.M{"$each": addToSetVideos}
	}
	if len(addToset3dAssets) > 0 {
		addToSetOperations["3d_assets"] = bson.M{"$each": addToset3dAssets}
	}
	if variant, ok := updateMap["variant"].(*models.Variant); ok && variant.Class != 2 {
		pullOperationsKeys = append(pullOperationsKeys, "masked_photo")
		pullOperations["images"] = bson.M{
			"k": bson.M{
				"$in": pullOperationsKeys,
			},
		}
	}

	if _, ok := updateMap["video_url"]; ok {
		if len(existExperience.Videos) > 0 {
			pullOperations["videos"] = bson.M{
				"k": bson.M{
					"$in": []string{"mask", "compressed", "hls", "dash", "compressed_playback", "webm"},
				},
			}
		}

	}

	if _, ok := updateMap["mask_url"]; ok {
		if len(existExperience.Videos) > 0 {
			pullOperations["videos"] = bson.M{
				"k": bson.M{
					"$in": []string{"compressed", "hls", "dash", "compressed_playback", "webm"},
				},
			}
		}

	}
	unsetOperations := bson.M{}
	if updateMap["workflow_error"] == nil {
		unsetOperations["workflow_error"] = 1
	}
	if DeleteShareMeta, ok := updateMap["delete_share_meta"].(bool); ok {
		if DeleteShareMeta {
			unsetOperations["share_meta"] = 1
		}
	}

	if DeleteTexture, ok := updateMap["delete_texture"].(bool); ok {
		if DeleteTexture {
			pullOperationsKeys = append(pullOperationsKeys, "texture_file")
			pullOperations["3d_assets"] = bson.M{
				"k": bson.M{
					"$in": pullOperationsKeys,
				},
			}
		}
	}

	if DeleteObj, ok := updateMap["delete_obj"].(bool); ok {
		if DeleteObj {
			pullOperationsKeys = append(pullOperationsKeys, "original_obj")
			pullOperations["3d_assets"] = bson.M{
				"k": bson.M{
					"$in": pullOperationsKeys,
				},
			}
		}
	}

	if DeleteGlb, ok := updateMap["delete_glb"].(bool); ok {
		if DeleteGlb {
			pullOperationsKeys = append(pullOperationsKeys, "original_glb")
			pullOperations["3d_assets"] = bson.M{
				"k": bson.M{
					"$in": pullOperationsKeys,
				},
			}
		}
	}

	if DeleteUsdz, ok := updateMap["delete_usdz"].(bool); ok {
		if DeleteUsdz {
			pullOperationsKeys = append(pullOperationsKeys, "original_usdz")
			pullOperations["3d_assets"] = bson.M{
				"k": bson.M{
					"$in": pullOperationsKeys,
				},
			}
		}
	}

	if DeleteBlend, ok := updateMap["delete_blend"].(bool); ok {
		if DeleteBlend {
			pullOperationsKeys = append(pullOperationsKeys, "blend_file")
			pullOperations["3d_assets"] = bson.M{
				"k": bson.M{
					"$in": pullOperationsKeys,
				},
			}
		}
	}

	if _, ok := updateMap["image_url"]; ok {
		if len(existExperience.Images) > 0 {
			pullOperationsKeys = append(pullOperationsKeys, "compressed", "compressed_spawn", "color_compressed", "std_compressed", "feature_image", "fdb")
			pullOperations["images"] = bson.M{
				"k": bson.M{
					"$in": pullOperationsKeys,
				},
			}
		}
		if existExperience.Overlay != nil {
			setOperations["overlay.compressed_image"] = ""
		}
		if existExperience.Mask != nil {
			unsetOperations["mask"] = 1
		}
	}

	if len(addToSetOperations) > 0 || len(pullOperations) > 0 || len(setOperations) > 0 || len(unsetOperations) > 0 {
		if len(pullOperations) > 0 {
			updateAdd := bson.M{}
			updateAdd["$pull"] = pullOperations
			_, err := coll.UpdateOne(ctx, filter, updateAdd)
			if err != nil {
				return nil, err
			}
		}
		if len(addToSetOperations) > 0 {
			updateAdd := bson.M{}
			updateAdd["$addToSet"] = addToSetOperations
			_, err := coll.UpdateOne(ctx, filter, updateAdd)
			if err != nil {
				return nil, err
			}
		}
		if len(setOperations) > 0 {
			updateAdd := bson.M{}
			updateAdd["$set"] = setOperations
			_, err := coll.UpdateOne(ctx, filter, updateAdd)
			if err != nil {
				return nil, err
			}
		}
		if len(unsetOperations) > 0 {
			updateAdd := bson.M{}
			updateAdd["$unset"] = unsetOperations
			_, err := coll.UpdateOne(ctx, filter, updateAdd)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := coll.FindOne(ctx, filter).Decode(&experience); err != nil {
		return nil, err
	}

	var campaign models.Campaign
	query := bson.M{"_id": experience.CampaignID}
	coll = impl.db.Collection(consts.CampaignCollection)
	Update := bson.M{
		"$set": bson.M{
			"updated_at": time.Duration(time.Now().UnixMilli()),
		},
	}
	if _, ok := updateMap["status"]; ok {
		Update["$set"].(bson.M)["publish"] = false
	}
	result = coll.FindOneAndUpdate(ctx, query, Update)
	decodeErr := result.Decode(&campaign)
	if decodeErr != nil {
		return nil, decodeErr
	}

	//Edit logs
	EditLog.After = &experience
	EditLog.ClientId = campaign.ClientId
	EditLog.ShortCode = campaign.ShortCode
	mediaProcess := models.MediaProcess{
		Experience: experience,
		ShortCode:  campaign.ShortCode,
		IsEdited:   true,
		Name:       campaign.Name,
		CreatedBy:  *campaign.CreatedBy,
		ClientId:   campaign.ClientId,
	}

	Response := dtos.UpdateResponseDto{
		MediaProcess: &mediaProcess,
		EditLog:      EditLog,
		SegmentInfo:  SegmentInfo,
	}

	go utils.PrepareAndCaptureAssets(impl.db, updateMap, &experience, &existExperience)

	return &Response, nil
}

func (impl *ExperienceDaoImpl) PostbackExperienceDao(objID primitive.ObjectID, updateMap map[string]interface{}) (*dtos.PostbackResponseDto, error) {
	var exp models.Experience
	filter := bson.M{"_id": objID}
	coll := impl.db.Collection(consts.ExperienceCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result := coll.FindOne(ctx, filter)
	err := result.Decode(&exp)
	if err != nil {
		return nil, err
	}

	EditLog := dtos.EditLogDto{
		ExperienceId: exp.ID,
		Before:       &exp,
	}

	updateOperation := map[string]interface{}{}
	updates := bson.M{"$set": updateOperation}
	updateMap["updated_at"] = time.Now().UnixMilli()

	if updateMap["edited_by"] != nil {
		updateOperation["edited_by"] = updateMap["edited_by"]
	}

	//todo: template
	vids := []map[string]interface{}{}
	var hls, dash, comVideo, playback bool
	for _, video := range exp.Videos {
		switch video.K {
		case "compressed":
			comVideo = true
			if updateMap["compressed_video"] != nil {
				vids = append(vids, map[string]interface{}{"k": "compressed", "v": updateMap["compressed_video"]})
				continue
			}
		case "hls":
			hls = true
			if updateMap["hls_url"] != nil {
				vids = append(vids, map[string]interface{}{"k": "hls", "v": updateMap["hls_url"]})
				continue
			}
		case "dash":
			dash = true
			if updateMap["dash_url"] != nil {
				vids = append(vids, map[string]interface{}{"k": "dash", "v": updateMap["dash_url"]})
				continue
			}
		case "compressed_playback":
			playback = true
			if updateMap["compressed_video"] != nil {
				vids = append(vids, map[string]interface{}{"k": "compressed_playback", "v": updateMap["compressed_video"]})
				continue
			}
		}
		vids = append(vids, map[string]interface{}{"k": video.K, "v": video.V})
	}
	if !comVideo {
		if updateMap["compressed_video"] != nil {
			comVideo = true
			vids = append(vids, map[string]interface{}{"k": "compressed", "v": updateMap["compressed_video"]})
		}
	}
	if !hls {
		if updateMap["hls_url"] != nil {
			hls = true
			vids = append(vids, map[string]interface{}{"k": "hls", "v": updateMap["hls_url"]})
		}
	}
	if !dash {
		if updateMap["dash_url"] != nil {
			dash = true
			vids = append(vids, map[string]interface{}{"k": "dash", "v": updateMap["dash_url"]})
		}
	}
	if !playback {
		if updateMap["compressed_video"] != nil {
			vids = append(vids, map[string]interface{}{"k": "compressed_playback", "v": updateMap["compressed_video"]})
		}
	}
	updateOperation["videos"] = vids

	updateOperation["updated_at"] = time.Now().UnixMilli()
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	var NewExp models.Experience
	result = coll.FindOneAndUpdate(ctx, filter, updates, &opt)
	if err := result.Err(); err != nil {
		return nil, err
	}
	if err := result.Decode(&NewExp); err != nil {
		return nil, err
	}

	var campaign models.Campaign
	query := bson.M{"_id": NewExp.CampaignID}
	coll = impl.db.Collection(consts.CampaignCollection)
	result = coll.FindOne(ctx, query)
	err = result.Decode(&campaign)
	if err != nil {
		return nil, err
	}

	EditLog.After = &NewExp
	EditLog.ClientId = campaign.ClientId
	EditLog.ShortCode = campaign.ShortCode

	PostbackResponse := &dtos.PostbackResponseDto{
		EditLog:    EditLog,
		Campaign:   &campaign,
		Experience: &NewExp,
	}

	return PostbackResponse, nil
}

func (impl *ExperienceDaoImpl) GetElementBannerVariant(bannerDto *dtos.GetBannerDto) int {
	if bannerDto.Element.PrimaryColor == "" {
		bannerDto.Element.PrimaryColor = "#ffffff"
	}

	if bannerDto.Element.SecondaryColor == "" {
		bannerDto.Element.SecondaryColor = "#007AFF"
	}

	if bannerDto.Element.ShareUrl == "" && bannerDto.QrCode {
		bannerDto.Element.ShareUrl = bannerDto.ImageUrl
	}
	var banners = bannerDto.Element
	conditions := map[int]map[string]bool{
		1: {"Title": true, "SubTitle": true, "ShareText": true, "RedirectionUrl": false},
		2: {"Title": true, "SubTitle": true, "ShareText": false, "RedirectionUrl": true},
		3: {"Title": true, "SubTitle": false, "ShareText": false, "RedirectionUrl": true},
		4: {"Title": true, "SubTitle": true, "ShareText": true, "RedirectionUrl": true},
		5: {"Title": true, "SubTitle": false, "ShareText": true, "RedirectionUrl": false},
		6: {"Title": true, "SubTitle": false, "ShareText": true, "RedirectionUrl": true},
	}

	fieldPresence := map[string]bool{
		"Title":          banners.Title != "",
		"SubTitle":       banners.SubTitle != "",
		"ShareText":      banners.ShareText != "",
		"RedirectionUrl": banners.RedirectionUrl != "",
	}

	for version, conds := range conditions {
		matched := true
		for field, required := range conds {
			if fieldPresence[field] != required {
				matched = false
				break
			}
		}
		if matched {
			return version
		}
	}
	return 0
}

func (impl *ExperienceDaoImpl) GetImageUrls(uniqImage *dtos.UniqueImageDto) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	campaignObjID, err := primitive.ObjectIDFromHex(uniqImage.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("invalid campaign ID: %v", err)
	}

	coll := impl.db.Collection(consts.ExperienceCollection)

	match := bson.M{
		"$match": bson.M{
			"campaign_id": campaignObjID,
			"is_active":   true,
		},
	}
	if uniqImage.ExperienceID != "" {
		expID, err := primitive.ObjectIDFromHex(uniqImage.ExperienceID)
		if err != nil {
			return nil, fmt.Errorf("invalid experience ID: %v", err)
		}
		match = bson.M{
			"$match": bson.M{
				"campaign_id":   campaignObjID,
				"experience_id": bson.M{"$ne": expID, "$exists": true},
				"is_active":     true,
			},
		}
	}
	pipeline := []bson.M{
		match,
		{"$unwind": "$images"},
		{"$match": bson.M{"images.k": "original"}},
		{
			"$group": bson.M{
				"_id":             "$campaign_id",
				"original_images": bson.M{"$push": "$images.v"},
			},
		},
	}

	// Execute the aggregation
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Iterate through the results
	var results []struct {
		ID             primitive.ObjectID `bson:"_id"`
		OriginalImages []string           `bson:"original_images"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		log.Fatal(err)
	}
	var imageURLs []string
	if len(results) > 0 {
		imageURLs = results[0].OriginalImages
	}
	return imageURLs, nil
}

func (impl *ExperienceDaoImpl) GetExperienceByID(id primitive.ObjectID) (*models.Experience, error) {
	var exp models.Experience
	query := bson.M{"_id": id}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := impl.db.Collection(consts.ExperienceCollection)
	result := coll.FindOne(ctx, query)
	decodeErr := result.Decode(&exp)
	if decodeErr != nil {
		return nil, decodeErr
	}
	return &exp, nil
}

func (impl *ExperienceDaoImpl) GetExperienceByCampaignID(campaignID primitive.ObjectID) (*models.Experience, error) {
	var exp models.Experience
	query := bson.M{"campaign_id": campaignID, "is_active": true}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := impl.db.Collection(consts.ExperienceCollection)
	result := coll.FindOne(ctx, query)
	decodeErr := result.Decode(&exp)
	if decodeErr != nil {
		return nil, decodeErr
	}
	return &exp, nil
}

func (impl *ExperienceDaoImpl) GetCampaignExperiencesDao(campaignID string) (*dtos.CampaignExperienceDto, error) {
	campaignObjId, err := primitive.ObjectIDFromHex(campaignID)
	if err != nil {
		return nil, fmt.Errorf("invalid campaign ID: %v", err)
	}
	query := bson.M{"campaign_id": campaignObjId, "credit_deduct": true, "is_active": true}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var exps []models.Experience
	coll := impl.db.Collection(consts.ExperienceCollection)
	cursor, err := coll.Find(ctx, query)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &exps); err != nil {
		return nil, err
	}

	var campaign models.Campaign
	CampaignQuery := bson.M{"_id": campaignObjId}
	coll = impl.db.Collection(consts.CampaignCollection)
	result := coll.FindOne(ctx, CampaignQuery)
	err = result.Decode(&campaign)
	if err != nil {
		return nil, err
	}

	query = bson.M{"campaign_id": campaignObjId, "is_active": true}
	var Experience models.Experience
	coll = impl.db.Collection(consts.ExperienceCollection)
	result = coll.FindOne(ctx, query)
	err = result.Decode(&Experience)
	if err != nil {
		return nil, err
	}

	response := dtos.CampaignExperienceDto{
		Campaign:    campaign,
		Experiences: exps,
		Experience:  &Experience,
	}

	return &response, nil
}
func (impl *ExperienceDaoImpl) GetCampaignExperiencesStatus(campaignID string) (*bool, error) {
	campaignObjId, err := primitive.ObjectIDFromHex(campaignID)
	if err != nil {
		return nil, fmt.Errorf("invalid campaign ID: %v", err)
	}

	query := bson.M{
		"campaign_id": campaignObjId,
		"is_active":   true,
		"status":      bson.M{"$ne": consts.Processed},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := impl.db.Collection(consts.ExperienceCollection)
	count, err := coll.CountDocuments(ctx, query)
	if err != nil {
		return nil, err
	}

	// If count > 0, return false; otherwise, return true
	result := count == 0
	return &result, nil
}

func (impl *ExperienceDaoImpl) GetLogsDao(filter *dtos.EditLogsFilterDto, optionalParams ...string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := impl.db.Collection(consts.EditLogsCollection)

	query := bson.M{}

	if filter.ClientId != primitive.NilObjectID {
		query["client_id"] = filter.ClientId
	}

	if filter.ExperienceId != primitive.NilObjectID {
		query["experience_id"] = filter.ExperienceId
	}

	if !filter.StartDate.IsZero() || !filter.EndDate.IsZero() {
		dateFilter := bson.M{}

		if !filter.StartDate.IsZero() {
			dateFilter["$gte"] = filter.StartDate.UTC()
		}

		if !filter.EndDate.IsZero() {
			dateFilter["$lte"] = filter.EndDate.UTC()
		}

		if len(dateFilter) > 0 {
			query["created_at"] = dateFilter
		}
	}

	if filter.ShortCode != "" {
		query["short_code"] = primitive.Regex{Pattern: filter.ShortCode}
	}

	match := bson.M{"$match": query}
	sort := bson.M{"$sort": bson.M{"created_at": -1}}

	add := bson.M{
		"$set": bson.M{
			"after.images": bson.M{"$arrayToObject": bson.M{"$map": bson.M{
				"input": "$after.images",
				"as":    "image",
				"in": bson.M{
					"k": "$$image.k",
					"v": "$$image.v",
				},
			}}},
			"after.videos": bson.M{"$arrayToObject": bson.M{"$map": bson.M{
				"input": "$after.videos",
				"as":    "video",
				"in": bson.M{
					"k": "$$video.k",
					"v": "$$video.v",
				},
			}}},
			"before.images": bson.M{"$arrayToObject": bson.M{"$map": bson.M{
				"input": "$before.images",
				"as":    "image",
				"in": bson.M{
					"k": "$$image.k",
					"v": "$$image.v",
				},
			}}},
			"before.videos": bson.M{"$arrayToObject": bson.M{"$map": bson.M{
				"input": "$before.videos",
				"as":    "video",
				"in": bson.M{
					"k": "$$video.k",
					"v": "$$video.v",
				},
			}}},
		},
	}

	project := bson.M{
		"$project": bson.M{
			"_id":             0,
			"id":              "$_id",
			"created_at":      1,
			"experience_id":   1,
			"client_id":       1,
			"after":           1,
			"before":          1,
			"before_campaign": 1,
			"after_campaign":  1,
			"campaign_id":     1,
			"short_code":      1,
			"network_info":    1,
		},
	}

	// Pagination stages
	skip := bson.M{"$skip": filter.Page * filter.PageSize}
	limit := bson.M{"$limit": filter.PageSize}

	facet := bson.M{
		"$facet": bson.M{
			"count": []bson.M{
				{"$count": "totalDocs"},
			},
			"data": []bson.M{
				sort,
				skip,
				limit,
				add,
				project,
			},
		},
	}

	projectFinal := bson.M{
		"$project": bson.M{
			"logs": "$data",
			"total_document": bson.M{"$ifNull": []interface{}{
				bson.M{"$arrayElemAt": bson.A{"$count.totalDocs", 0}},
				0,
			}},
		},
	}

	pipeline := []bson.M{match, facet, projectFinal}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregation failed: %w", err)
	}
	defer cursor.Close(ctx)

	var result []map[string]interface{}
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("failed to decode cursor: %w", err)
	}
	return result[0], nil
}

func (impl *ExperienceDaoImpl) UpdateExperienceAssetsDao(WfResult dtos.WorkflowFinalPubResult) error {
	parts := strings.Split(WfResult.WorkflowId, "_")
	var exp models.Experience
	fmt.Println(parts)
	ExpObjId, err := primitive.ObjectIDFromHex(parts[0])
	if err != nil {
		log.Printf("Error converting workflow ID to ObjectID: %v", err)
		return err
	}
	filter := bson.M{"_id": ExpObjId}
	ExpColl := impl.db.Collection(consts.ExperienceCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result := ExpColl.FindOne(ctx, filter)
	err = result.Decode(&exp)
	if err != nil {
		log.Printf("Error decoding experience: %v", err)
		return err
	}
	var campaign models.Campaign
	query := bson.M{"_id": exp.CampaignID}
	CampaignColl := impl.db.Collection(consts.CampaignCollection)
	result = CampaignColl.FindOne(ctx, query)
	err = result.Decode(&campaign)
	if err != nil {
		log.Printf("Error decoding campaign: %v", err)
		return err
	}

	updateOperation := map[string]interface{}{}
	unsetOperation := map[string]interface{}{}
	updates := bson.M{
		"$set":   updateOperation,
		"$unset": unsetOperation,
	}

	var CreditConsume bool
	var ConsumeCreditRes *dtos.ConsumeCreditResDto
	if WfResult.Status == dtos.Completed && WfResult.Publish {
		data, err := impl.GetCampaignExperiencesDao(campaign.ID.Hex())
		if err != nil {
			return err
		}
		if len(data.Experiences) == 0 {
			var UserID string
			if data.Experience.EditedBy != nil {
				UserID = data.Experience.EditedBy.ID
			} else if data.Experience.CreatedBy != nil {
				UserID = data.Experience.CreatedBy.ID
			}
			ConsumeCreditRes, err = utils.ConsumeCredit(data, data.Experience.CreditAllowanceID.Hex(), impl.lgr, UserID)
			if err != nil {
				if err.Error() == "failed to consume credit" {
					WfResult.Status = dtos.NoCredit
					WfResult.WorkflowError = models.WorkflowError{
						ConsumerType: "credit",
						Msg:          "Server Error: failed to consume credit",
					}
				} else {
					WfResult.Status = dtos.NoCredit
					WfResult.WorkflowError = models.WorkflowError{
						Msg:          err.Error(),
						ConsumerType: "credit",
					}
				}
				AdjustEscrowCreditRequest := dtos.AdjustEscrowCreditsRequest{
					Reverse:           true,
					CreditAllowanceID: data.Experience.CreditAllowanceID.Hex(),
				}
				_, err = utils.AdjustEscrowCredits(AdjustEscrowCreditRequest, impl.lgr)
				if err != nil {
					impl.lgr.Infof("Error adjusting escrow credit: %v", err)
				}
				unsetOperation["credit_allowance_id"] = 1
			} else {
				CreditConsume = true
				impl.lgr.Infof("Credit consume successfully: %s", data.Campaign.ID)
			}
		}
	} else if WfResult.Publish {
		// adjust +1  escrow credit but before this
		// check if any experience has deducted credit. if none of them have deducted then only adjust.
		data, err := impl.GetCampaignExperiencesDao(campaign.ID.Hex())
		if err != nil {
			return err
		}
		if len(data.Experiences) == 0 {
			AdjustEscrowCreditRequest := dtos.AdjustEscrowCreditsRequest{
				Reverse:           true,
				CreditAllowanceID: data.Experience.CreditAllowanceID.Hex(),
			}
			_, err = utils.AdjustEscrowCredits(AdjustEscrowCreditRequest, impl.lgr)
			if err != nil {
				impl.lgr.Errorf("Error adjusting escrow credit: %v", err)
			}
		}
	}
	Images := []models.Image{}
	Videos := []models.Video{}
	ogImage := false
	var ogImageUrl string
	for _, image := range exp.Images {
		if image.K == "original" {
			ogImageUrl = image.V
		}
		Images = append(Images, models.Image{K: image.K, V: image.V})
	}

	for _, video := range exp.Videos {
		Videos = append(Videos, models.Video{K: video.K, V: video.V})
	}
	updateOperation["status"] = consts.Processed
	updateOperation["updated_at"] = time.Now().UnixMilli()
	if CreditConsume {
		updateOperation["credit_deduct"] = true
	}
	updateCampaign := map[string]interface{}{}
	var WindowRatio float64
	segmentVideo := make(map[string]interface{})
	segmentsStartEndTime := make(map[string]interface{})
	button := make(map[string]interface{})
	IsStitchWf := false
	if WfResult.Status == dtos.Completed || WfResult.Status == dtos.NoCredit {
		for _, task := range WfResult.TaskResults {
			taskID := task.TaskId
			data := task.Payload
			if data.ImageAspectRatio != 0 {
				WindowRatio = data.ImageAspectRatio
			} else if data.VideoAspectRatio != 0 && WindowRatio == 0 {
				ratio := 1 / data.VideoAspectRatio
				formattedRatio := math.Round(ratio*1000) / 1000
				WindowRatio = formattedRatio
			}
			if data.VideoAspectRatio != 0 {
				updateOperation["aspect_ratio"] = data.VideoAspectRatio
			}
			if data.IsHorizontal != nil {
				updateOperation["variant.is_horizontal"] = data.IsHorizontal
			}
			if data.TemplateMaskUrl != "" {
				updateOperation["mask"] = models.Mask{
					URL:           data.TemplateMaskUrl,
					Offset:        models.ThreeDCoordinates{},
					Scale:         1,
					CompressedUrl: data.TemplateMaskUrl,
				}
			}
			switch {
			case taskID == "main_overlay":
				if data.OverlayCompressed != "" {
					updateOperation["overlay.compressed_image"] = data.OverlayCompressed
				}
			case taskID == "main_fal":
				if data.GenStudioOutput != nil && data.GenStudioOutput.Value != "" {
					Videos = append(Videos, models.Video{K: "green_screen", V: data.GenStudioOutput.Value})
				}
			case taskID == "main_image":
				ogImage = true
				if data.CompressedImage != "" {
					Images = append(Images, models.Image{K: "compressed", V: data.CompressedImage})
				}
				if data.ColorCompressedImage != "" {
					updateCampaign["scan.compressed_image_url"] = data.ColorCompressedImage
					updateCampaign["icon_url"] = data.ColorCompressedImage
					Images = append(Images, models.Image{K: "color_compressed", V: data.ColorCompressedImage})
				}
				if data.StdCompressedImage != "" {
					Images = append(Images, models.Image{K: "std_compressed", V: data.StdCompressedImage})
				}
				if data.FeatureImage != "" {
					Images = append(Images, models.Image{K: "feature_image", V: data.FeatureImage})
				}
				if data.SpawnCompressedImage != "" {
					Images = append(Images, models.Image{K: "compressed_spawn", V: data.SpawnCompressedImage})
				}
				if data.OriginalGreenScreenIMGURL != "" {
					Images = append(Images, models.Image{K: "original_green_screen", V: data.OriginalGreenScreenIMGURL})
				}
			case taskID == "main_video":
				if data.CompressedVideo != "" {
					Videos = append(Videos, models.Video{K: "compressed", V: data.CompressedVideo})
				}
				if data.CompressedVideo != "" {
					Videos = append(Videos, models.Video{K: "compressed_playback", V: data.CompressedVideo})
				}
				if data.HlsUrl != "" {
					Videos = append(Videos, models.Video{K: "hls", V: data.HlsUrl})
				}
				if data.DashUrl != "" {
					Videos = append(Videos, models.Video{K: "dash", V: data.DashUrl})
				}
				if data.WebMUrl != "" {
					Videos = append(Videos, models.Video{K: "webm", V: data.WebMUrl})
				}
				if data.RGBVideoUrl != "" {
					Videos = append(Videos, models.Video{K: "original", V: data.RGBVideoUrl})
				}
				if data.MaskVideoUrl != "" {
					Videos = append(Videos, models.Video{K: "mask", V: data.MaskVideoUrl})
				}
			case taskID == "main_image_vector_llm":
				if data.MilvusRefID != "" {
					updateCampaign["milvus_ref_id"] = data.MilvusRefID
				}
				if data.ProductDescription != "" {
					updateOperation["catalogue_details.description"] = data.ProductDescription
				}
			default:
				if strings.Contains(taskID, "parallaxId_") && strings.Contains(taskID, "_planeId_") {
					parts := strings.Split(taskID, "_")
					if len(parts) >= 5 {
						parallaxID := parts[1]
						planeID := parts[3]
						mediaType := parts[4]
						for i, parallax := range exp.Scene.Parallax {
							if parallax.ID.Hex() == parallaxID {
								for j, plane := range parallax.Planes {
									if plane.ID.Hex() == planeID {
										switch mediaType {
										case "image":
											if data.StdCompressedImage != "" {
												exp.Scene.Parallax[i].Planes[j].Compressed = data.StdCompressedImage
											} else {
												exp.Scene.Parallax[i].Planes[j].Compressed = data.ColorCompressedImage
											}
										case "video":
											exp.Scene.Parallax[i].Planes[j].Compressed = data.CompressedVideo
											exp.Scene.Parallax[i].Planes[j].Hls = data.HlsUrl
											exp.Scene.Parallax[i].Planes[j].Dash = data.DashUrl
											exp.Scene.Parallax[i].Planes[j].IsHorizontal = data.IsHorizontal
										}
									}
								}
							}
						}
					}
				}
				if strings.Contains(taskID, "parallaxId_") && strings.Contains(taskID, "_mask_") {
					parts := strings.Split(taskID, "_")
					if len(parts) >= 1 {
						parallaxID := parts[1]
						for i, parallax := range exp.Scene.Parallax {
							if parallax.ID.Hex() == parallaxID {
								exp.Scene.Parallax[i].Mask.CompressedUrl = data.ColorCompressedImage
							}
						}
					}
				}
				if strings.Contains(taskID, "markerId_") {
					markerAndType := strings.Split(taskID, "_")
					var markerID, markerType string
					if len(markerAndType) >= 3 {
						markerID = markerAndType[1]
						markerType = markerAndType[2]
					}
					if markerType == "image" {
						if markerID != "" {
							button[markerID] = map[string]string{
								"marker_id":        task.Payload.MarkerID,
								"compressed_image": task.Payload.ColorCompressedImage,
							}
						}
					}
					if markerType == "video" {
						if markerID != "" {
							segmentVideo[markerID] = map[string]string{
								"marker_id":        task.Payload.MarkerID,
								"compressed_video": task.Payload.CompressedVideo,
								"hls":              task.Payload.HlsUrl,
								"dash":             task.Payload.DashUrl,
								"webm":             task.Payload.WebMUrl,
							}

							if task.Payload.IsHorizontal != nil {
								segmentVideo[markerID].(map[string]string)["is_horizontal"] = fmt.Sprintf("%v", *task.Payload.IsHorizontal)
							}
						}
						if task.Payload.IsHorizontal != nil {
							updateOperation["variant.is_horizontal"] = *task.Payload.IsHorizontal
						}

					}
				}
				if strings.Contains(taskID, "stitchsegment_") {
					IsStitchWf = true
					Videos = []models.Video{}
					updateOperation["status"] = exp.Status
					if data.OriginalVideo != "" {
						Videos = append(Videos, models.Video{K: "original", V: data.OriginalVideo})
					}
					if data.CompressedVideo != "" {
						Videos = append(Videos, models.Video{K: "compressed", V: data.CompressedVideo})
					}
					if data.HlsUrl != "" {
						Videos = append(Videos, models.Video{K: "hls", V: data.HlsUrl})
					}
					if data.DashUrl != "" {
						Videos = append(Videos, models.Video{K: "dash", V: data.DashUrl})
					}
					if data.MaskVideo != "" {
						Videos = append(Videos, models.Video{K: "mask", V: data.MaskVideo})
					}
					if data.WebMUrl != "" {
						Videos = append(Videos, models.Video{K: "webm", V: data.WebMUrl})
					}

					for _, marker := range task.Payload.SegmentInfo {
						if marker.MarkerID != "" {
							segmentsStartEndTime[marker.MarkerID] = map[string]int64{
								"start_time": marker.StartTime,
								"end_time":   marker.EndTime,
							}
						}
					}
				}
			}
		}
		if len(button) > 0 {
			for i, btn := range exp.Variant.Buttons {
				_, ok := button[btn.MarkerId]
				if ok {
					exp.Variant.Buttons[i].CompressedAssetUrl = button[btn.MarkerId].(map[string]string)["compressed_image"]
				}
			}
			updateOperation["variant.buttons"] = exp.Variant.Buttons
		}

		if len(segmentVideo) > 0 {
			for i, marker := range exp.Variant.Segments.Markers {
				_, ok := segmentVideo[marker.Id]
				if ok {
					newProcessedUrls := models.VideoObject{
						Compressed:  segmentVideo[marker.Id].(map[string]string)["compressed_video"],
						Hls:         segmentVideo[marker.Id].(map[string]string)["hls"],
						Dash:        segmentVideo[marker.Id].(map[string]string)["dash"],
						WebM:        segmentVideo[marker.Id].(map[string]string)["webm"],
						Original:    marker.Videos.Original,
						Mask:        marker.Videos.Mask,
						MergeVideo:  marker.Videos.MergeVideo,
						Orientation: marker.Videos.Orientation,
					}

					exp.Variant.Segments.Markers[i].IsHorizontal = utils.ToPointer(segmentVideo[marker.Id].(map[string]string)["is_horizontal"] == "true")
					exp.Variant.Segments.Markers[i].Videos = newProcessedUrls
				}
			}
			updateOperation["variant.segments"] = exp.Variant.Segments
		}
		if len(segmentsStartEndTime) > 0 {
			gotAllMarkers := true
			for i, marker := range exp.Variant.Segments.Markers {
				val, ok := segmentsStartEndTime[marker.Id]
				if ok {
					exp.Variant.Segments.Markers[i].Stime = val.(map[string]int64)["start_time"]
					exp.Variant.Segments.Markers[i].Etime = val.(map[string]int64)["end_time"]
				} else {
					fmt.Println("not found for marker:", marker.Id)
					gotAllMarkers = false
				}
			}
			fmt.Println("got all markers:", gotAllMarkers)
			if gotAllMarkers {
				exp.Variant.Segments.UseMarkerVideo = false
			}
			updateOperation["variant.segments"] = exp.Variant.Segments
		}
		updateOperation["videos"] = Videos
		updateOperation["images"] = Images
		if exp.Scene != nil {
			exp.Scene.WindowRatio = WindowRatio
			updateOperation["scene"] = exp.Scene
		}
	}
	if len(WfResult.TaskResults) == 1 && strings.Contains(WfResult.TaskResults[0].TaskId, "stitchsegment_") {
		IsStitchWf = true
		updateOperation["status"] = exp.Status
	}
	if WfResult.Status == dtos.Failed {
		if !IsStitchWf {
			updateOperation["status"] = consts.Failed
			updateOperation["workflow_error"] = WfResult.WorkflowError
		}
	}
	if WfResult.Status == dtos.TimedOut {
		if !IsStitchWf {
			updateOperation["status"] = consts.TimedOut
			updateOperation["workflow_error"] = WfResult.WorkflowError
		}
	}
	if WfResult.Status == dtos.NoCredit {
		if !IsStitchWf {
			updateOperation["status"] = consts.Failed
			updateOperation["workflow_error"] = WfResult.WorkflowError
		}
	}
	if WfResult.Status == dtos.Cancelled {
		if !IsStitchWf {
			updateOperation["status"] = consts.Cancelled
			updateOperation["workflow_error"] = WfResult.WorkflowError
		}
	}
	if WfResult.Status == dtos.Completed {
		if !IsStitchWf {
			unsetOperation["workflow_error"] = 1
		}
	}

	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	var NewExp models.Experience
	result = ExpColl.FindOneAndUpdate(ctx, filter, updates, &opt)
	if err := result.Err(); err != nil {
		log.Printf("Error updating experience: %v", err)
		return err
	}
	if err := result.Decode(&NewExp); err != nil {
		log.Printf("Error decoding updated experience: %v", err)
		return err
	}

	if NewExp.EditedBy == nil {
		NewExp.EditedBy = NewExp.CreatedBy
	}

	if WfResult.Status != dtos.Completed {
		// go utils.CampaignPushNotif(&models.Campaign{
		// 	ID:        campaign.ID,
		// 	ShortCode: campaign.ShortCode,
		// 	Name:      campaign.Name,
		// 	CreatedBy: NewExp.EditedBy,
		// }, impl.lgr, "campaign_failed")

		go utils.PublishCampaignMail(&models.Campaign{
			ShortCode: campaign.ShortCode,
			Name:      campaign.Name,
			CreatedBy: NewExp.EditedBy,
			UpdatedAt: campaign.UpdatedAt,
			CreatedAt: campaign.UpdatedAt,
		}, impl.lgr, "", campaign.TrackType, "campaign_failed", "", 0, false)
	}

	if WfResult.Status == dtos.Completed {
		after := options.After
		opt := options.FindOneAndUpdateOptions{
			ReturnDocument: &after,
		}
		setCampaign := bson.M{"$set": updateCampaign}
		var update bool
		fmt.Println("publish", WfResult.Publish)
		var IsProcesseed *bool
		if WfResult.Publish {
			IsProcesseed, err = impl.GetCampaignExperiencesStatus(exp.CampaignID.Hex())
			if err != nil {
				log.Printf("Error checking campaign experiences status: %v", err)
				return err
			}
			if *IsProcesseed {
				update = true
				updateCampaign["updated_at"] = time.Now().UnixMilli()
				updateCampaign["publish"] = WfResult.Publish
			}
			if CreditConsume {
				update = true
				updateCampaign["golive_at"] = time.Now().UnixMilli()
				expiryWithUserName, userSvcErr := utils.GetClientCampaignExpiry(NewExp.CreatedBy.ID)
				if userSvcErr != nil {
					return errors.InternalServerError(userSvcErr.Error())
				}
				updateCampaign["expires_at"] = time.Duration(expiryWithUserName.Duration)
			}
		}
		if ogImage {
			update = true
			updateCampaign["updated_at"] = time.Now().UnixMilli()
			updateCampaign["scan.image_url"] = ogImageUrl
		}
		if update {
			result = CampaignColl.FindOneAndUpdate(ctx, query, setCampaign, &opt)
			if err = result.Err(); err != nil {
				log.Printf("Error updating campaign: %v", err)
				return err
			}
			if err = result.Decode(&campaign); err != nil {
				log.Printf("Error decoding updated campaign: %v", err)
				return err
			}
		}

		if WfResult.Publish && *IsProcesseed && CreditConsume {
			//Campaign publish push notification
			// go utils.CampaignPushNotif(&models.Campaign{
			// 	ID:        campaign.ID,
			// 	ShortCode: campaign.ShortCode,
			// 	Name:      campaign.Name,
			// 	CreatedBy: NewExp.EditedBy,
			// }, impl.lgr, "campaign_published")

			go utils.PublishCampaignMail(&models.Campaign{
				ShortCode: campaign.ShortCode,
				Name:      campaign.Name,
				CreatedBy: NewExp.EditedBy,
				CreatedAt: campaign.CreatedAt,
				ClientId:  campaign.ClientId,
			}, impl.lgr, ogImageUrl, campaign.TrackType, "campaign_published", ConsumeCreditRes.CreditType, ConsumeCreditRes.Balance, ConsumeCreditRes.Unlimited)
		} else if campaign.Publish || WfResult.Status == dtos.Completed {
			// go utils.CampaignPushNotif(&models.Campaign{
			// 	ID:        campaign.ID,
			// 	ShortCode: campaign.ShortCode,
			// 	Name:      campaign.Name,
			// 	CreatedBy: NewExp.EditedBy,
			// }, impl.lgr, "campaign_update")
		}
	}

	impl.redisClient.ExpireCampaignExperiences(campaign.ShortCode, false, false)
	categories, err := impl.GetCategoryByCampaignShortCodeDao(campaign.ShortCode)
	if err != nil {
		impl.lgr.Errorf("Error getting categories by campaign short code: %v", err)
	} else {
		for _, category := range categories {
			if err := impl.redisClient.ExpireCampaignExperiences(category.SiteCode, false, true); err != nil {
				impl.lgr.Errorf("Error expiring campaign experiences: %v", err)
			}
		}
	}

	return nil
}

// func (impl *ExperienceDaoImpl) ConsumeCredit(data *dtos.CampaignExperienceDto, CreditAllowanceID string) error {
// 	httpClient := httpClient.NewClient(impl.lgr)

// 	apiURL := fmt.Sprintf("%s/%s/%s", config.LoadConfig().BaseURL, consts.PaymentSvcRoutePrefix, "api/v1/credit/consume")
// 	fmt.Println("Credit consume API: ", apiURL)

// 	headers := map[string]string{
// 		"Authorization": config.LoadConfig().InterServiceToken,
// 	}

// 	consumeCreditRequest := &dtos.ConsumeCreditRequest{
// 		RefID:             data.Campaign.ShortCode,
// 		RefType:           "CAMPAIGN",
// 		RefName:           data.Campaign.Name,
// 		CreditAllowanceID: CreditAllowanceID,
// 	}

// 	_, err := httpClient.DoPost(apiURL, consumeCreditRequest, headers)
// 	if err != nil {
// 		if err.Error() == "failed to consume credit" {
// 			return err
// 		}
// 		fmt.Println("error consume credit request: ", err)
// 		return fmt.Errorf("%s", err.Error())
// 	}

// 	return nil
// }

func (impl *ExperienceDaoImpl) UpdateExperienceAssetsWithQrImageDao(WfResult dtos.WorkflowFinalPubResult) error {
	parts := strings.Split(WfResult.WorkflowId, "_")
	var exp models.Experience
	ExpObjId, err := primitive.ObjectIDFromHex(parts[1])
	if err != nil {
		log.Printf("Error converting workflow ID to ObjectID: %v", err)
		return err
	}

	filter := bson.M{"_id": ExpObjId}
	ExpColl := impl.db.Collection(consts.ExperienceCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result := ExpColl.FindOne(ctx, filter)
	err = result.Decode(&exp)
	if err != nil {
		log.Printf("Error decoding experience: %v", err)
		return err
	}

	var campaign models.Campaign
	query := bson.M{"_id": exp.CampaignID}
	CampaignColl := impl.db.Collection(consts.CampaignCollection)
	result = CampaignColl.FindOne(ctx, query)
	err = result.Decode(&campaign)
	if err != nil {
		log.Printf("Error decoding campaign: %v", err)
		return err
	}

	updateOperation := map[string]interface{}{}
	unsetOperation := map[string]interface{}{}
	updates := bson.M{
		"$set":   updateOperation,
		"$unset": unsetOperation,
	}
	var CreditConsume bool
	if WfResult.Status == dtos.Completed && WfResult.Publish {
		data, err := impl.GetCampaignExperiencesDao(campaign.ID.Hex())
		if err != nil {
			return err
		}
		if len(data.Experiences) == 0 {
			var UserID string
			if data.Experience.EditedBy != nil {
				UserID = data.Experience.EditedBy.ID
			} else if data.Experience.CreatedBy != nil {
				UserID = data.Experience.CreatedBy.ID
			}
			_, err = utils.ConsumeCredit(data, data.Experience.CreditAllowanceID.Hex(), impl.lgr, UserID)
			if err != nil {
				if err.Error() == "failed to consume credit" {
					WfResult.Status = dtos.NoCredit
					WfResult.WorkflowError = models.WorkflowError{
						ConsumerType: "credit",
						Msg:          "Server Error: failed to consume credit",
					}
				} else {
					WfResult.Status = dtos.NoCredit
					WfResult.WorkflowError = models.WorkflowError{
						Msg:          err.Error(),
						ConsumerType: "credit",
					}
				}
				AdjustEscrowCreditRequest := dtos.AdjustEscrowCreditsRequest{
					Reverse:           true,
					CreditAllowanceID: data.Experience.CreditAllowanceID.Hex(),
				}
				_, err = utils.AdjustEscrowCredits(AdjustEscrowCreditRequest, impl.lgr)
				if err != nil {
					impl.lgr.Infof("Error adjusting escrow credit: %v", err)
				}
				unsetOperation["credit_allowance_id"] = 1
			} else {
				CreditConsume = true
				impl.lgr.Infof("Credit consume successfully: %s", data.Campaign.ID)
			}
		}
	} else if WfResult.Publish {
		// adjust +1  escrow credit but before this
		// check if any experience has deducted credit. if none of them have deducted then only adjust.
		data, err := impl.GetCampaignExperiencesDao(campaign.ID.Hex())
		if err != nil {
			return err
		}
		if len(data.Experiences) == 0 {
			AdjustEscrowCreditRequest := dtos.AdjustEscrowCreditsRequest{
				Reverse:           true,
				CreditAllowanceID: data.Experience.CreditAllowanceID.Hex(),
			}
			_, err = utils.AdjustEscrowCredits(AdjustEscrowCreditRequest, impl.lgr)
			if err != nil {
				impl.lgr.Errorf("Error adjusting escrow credit: %v", err)
			}
		}
	}
	Images := []models.Image{}
	updateOperation["status"] = consts.Processed
	updateOperation["updated_at"] = time.Now().UnixMilli()
	if CreditConsume {
		updateOperation["credit_deduct"] = true
	}
	var ogImageUrl string
	updateCampaign := map[string]interface{}{}
	if WfResult.Status == dtos.Completed || WfResult.Status == dtos.NoCredit {
		for _, task := range WfResult.TaskResults {
			taskID := task.TaskId
			data := task.Payload
			switch {
			case taskID == "main_image":
				if data.OgImageWithQR != "" {
					ogImageUrl = data.OgImageWithQR
					Images = append(Images, models.Image{K: "original", V: data.OgImageWithQR})
					Images = append(Images, models.Image{K: "original_input", V: data.OgImageWithQR})
				}
				if data.CompressedImage != "" {
					Images = append(Images, models.Image{K: "compressed", V: data.CompressedImage})
				}
				if data.ColorCompressedImage != "" {
					updateCampaign["scan.compressed_image_url"] = data.ColorCompressedImage
					Images = append(Images, models.Image{K: "color_compressed", V: data.ColorCompressedImage})
				}
				if data.StdCompressedImage != "" {
					Images = append(Images, models.Image{K: "std_compressed", V: data.StdCompressedImage})
				}
				if data.FeatureImage != "" {
					Images = append(Images, models.Image{K: "feature_image", V: data.FeatureImage})
				}
				if data.SpawnCompressedImage != "" {
					Images = append(Images, models.Image{K: "compressed_spawn", V: data.SpawnCompressedImage})
				}
			}
		}
	}
	if WfResult.Status == dtos.Failed {
		updateOperation["status"] = consts.Failed
		updateOperation["workflow_error"] = WfResult.WorkflowError
	}
	if WfResult.Status == dtos.TimedOut {
		updateOperation["status"] = consts.TimedOut
		updateOperation["workflow_error"] = WfResult.WorkflowError
	}
	if WfResult.Status == dtos.NoCredit {
		updateOperation["status"] = consts.Failed
		updateOperation["workflow_error"] = WfResult.WorkflowError
	}
	if WfResult.Status == dtos.Cancelled {
		updateOperation["status"] = consts.Cancelled
		updateOperation["workflow_error"] = WfResult.WorkflowError
	}
	if WfResult.Status == dtos.Completed {
		unsetOperation["workflow_error"] = 1
	}
	updateOperation["images"] = Images

	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	var NewExp models.Experience
	result = ExpColl.FindOneAndUpdate(ctx, filter, updates, &opt)
	if err := result.Err(); err != nil {
		log.Printf("Error updating experience: %v", err)
		return err
	}
	if err := result.Decode(&NewExp); err != nil {
		log.Printf("Error decoding updated experience: %v", err)
		return err
	}

	if WfResult.Status == dtos.Completed {
		after := options.After
		opt := options.FindOneAndUpdateOptions{
			ReturnDocument: &after,
		}
		setCampaign := bson.M{"$set": updateCampaign}
		var update bool
		fmt.Println("publish", WfResult.Publish)
		var IsProcesseed *bool
		if WfResult.Publish {
			IsProcesseed, err = impl.GetCampaignExperiencesStatus(exp.CampaignID.Hex())
			if err != nil {
				log.Printf("Error checking campaign experiences status: %v", err)
				return err
			}
			if *IsProcesseed {
				update = true
				updateCampaign["updated_at"] = time.Now().UnixMilli()
				updateCampaign["publish"] = WfResult.Publish
			}
			if CreditConsume {
				update = true
				updateCampaign["golive_at"] = time.Now().UnixMilli()
				expiryWithUserName, userSvcErr := utils.GetClientCampaignExpiry(NewExp.CreatedBy.ID)
				if userSvcErr != nil {
					return errors.InternalServerError(userSvcErr.Error())
				}
				updateCampaign["expires_at"] = time.Duration(expiryWithUserName.Duration)
			}
		}

		update = true
		updateCampaign["updated_at"] = time.Now().UnixMilli()
		updateCampaign["scan.image_url"] = ogImageUrl

		if update {
			result = CampaignColl.FindOneAndUpdate(ctx, query, setCampaign, &opt)
			if err = result.Err(); err != nil {
				log.Printf("Error updating campaign: %v", err)
				return err
			}
			if err = result.Decode(&campaign); err != nil {
				log.Printf("Error decoding updated campaign: %v", err)
				return err
			}
		}

	}

	impl.redisClient.ExpireCampaignExperiences(campaign.ShortCode, false, false)
	categories, err := impl.GetCategoryByCampaignShortCodeDao(campaign.ShortCode)
	if err != nil {
		impl.lgr.Errorf("Error getting categories by campaign short code: %v", err)
	} else {
		for _, category := range categories {
			if err := impl.redisClient.ExpireCampaignExperiences(category.SiteCode, false, true); err != nil {
				impl.lgr.Errorf("Error expiring campaign experiences: %v", err)
			}
		}
	}
	return nil
}

func (impl *ExperienceDaoImpl) UpdateExperienceWorkflowData(ID primitive.ObjectID, WorkflowID string, StitchWfID string, TaskLenght int32) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{"_id": ID}
	coll := impl.db.Collection(consts.ExperienceCollection)
	updateExp := map[string]interface{}{}
	update := bson.M{"$set": updateExp}
	updateExp["total_task"] = TaskLenght
	updateExp["workflow_id"] = WorkflowID
	if StitchWfID != "" {
		updateExp["stitch_workflow_id"] = StitchWfID
	}
	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		impl.lgr.Infof("error updating workflow data: %v", err)
		return err
	}
	return nil
}

func (impl *ExperienceDaoImpl) ConsumeCreditAndPublishCampaign(CampaignID string, EditedBy *models.User, ExpId primitive.ObjectID) (*bool, error) {
	data, err := impl.GetCampaignExperiencesDao(CampaignID)
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var CreditConsume bool
	var CreditAllowanceID string
	var ConsumeCreditRes *dtos.ConsumeCreditResDto
	if len(data.Experiences) == 0 {
		if data.Experience == nil || data.Experience.TemplateDetails == nil || data.Experience.TemplateDetails["credit_type"] == nil {
			return nil, errors.InternalServerError("Cannot publish this experience, credit type not found")
		}
		AdjustEscrowCreditRequest := dtos.AdjustEscrowCreditsRequest{
			ClientId:   data.Campaign.ClientId.Hex(),
			CreditType: data.Experience.TemplateDetails["credit_type"].(string),
		}
		CreditAllowanceID, err = utils.AdjustEscrowCredits(AdjustEscrowCreditRequest, impl.lgr)
		if err != nil {
			log.Printf("Error adjusting escrow credits: %v", err)
			return nil, err
		}

		ConsumeCreditRes, err = utils.ConsumeCredit(data, CreditAllowanceID, impl.lgr, EditedBy.ID)
		if err != nil {
			AdjustEscrowCreditRequest := dtos.AdjustEscrowCreditsRequest{
				Reverse:           true,
				CreditAllowanceID: CreditAllowanceID,
			}
			_, err = utils.AdjustEscrowCredits(AdjustEscrowCreditRequest, impl.lgr)
			if err != nil {
				impl.lgr.Infof("Error adjusting escrow credit: %v", err)
			}
			impl.lgr.Infof("Not enough credit: %s", data.Campaign.ID)
			return nil, err
		} else {
			CreditConsume = true
			impl.lgr.Infof("Credit consume successfully: %s", data.Campaign.ID)
		}

		//update consume flag
		CreditAllowanceObjId, err := primitive.ObjectIDFromHex(CreditAllowanceID)
		if err != nil {
			impl.lgr.Infof("Error converting credit allowance ID to ObjectID: %v", err)
			return nil, err
		}
		filter := bson.M{"_id": data.Experience.ID}
		coll := impl.db.Collection(consts.ExperienceCollection)

		updateExp := map[string]interface{}{}
		set := bson.M{"$set": updateExp}
		updateExp["updated_at"] = time.Now().UnixMilli()
		updateExp["credit_allowance_id"] = CreditAllowanceObjId
		updateExp["credit_deduct"] = true
		_, err = coll.UpdateOne(ctx, filter, set)
		if err != nil {
			impl.lgr.Infof("error consuming credit: %v", err)
			return nil, err
		}
	}

	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	updateCampaign := map[string]interface{}{}
	setCampaign := bson.M{"$set": updateCampaign}
	updateCampaign["updated_at"] = time.Now().UnixMilli()
	updateCampaign["publish"] = true
	if CreditConsume {
		expiryWithUserName, userSvcErr := utils.GetClientCampaignExpiry(data.Experience.CreatedBy.ID)
		if userSvcErr != nil {
			return nil, errors.InternalServerError(userSvcErr.Error())
		}

		campaignExpiry := expiryWithUserName.Duration
		updateCampaign["expires_at"] = campaignExpiry
		updateCampaign["golive_at"] = time.Duration(time.Now().UnixMilli())
	}

	query := bson.M{"_id": data.Campaign.ID}
	CampaignColl := impl.db.Collection(consts.CampaignCollection)
	result := CampaignColl.FindOneAndUpdate(ctx, query, setCampaign, &opt)
	if err = result.Err(); err != nil {
		log.Printf("Error updating campaign: %v", err)
		return nil, err
	}
	if CreditConsume {
		go utils.CampaignPushNotif(&models.Campaign{
			ID:        data.Campaign.ID,
			ShortCode: data.Campaign.ShortCode,
			Name:      data.Campaign.Name,
			CreatedBy: EditedBy,
		}, impl.lgr, "campaign_published")

		//campaign publish email
		Exp, err := impl.GetExperienceByID(ExpId)
		if err != nil {
			log.Printf("Error updating campaign: %v", err)
			//! no return on error ??
		}
		var triggerImage string
		for _, img := range Exp.Images {
			if img.K == "original" {
				triggerImage = img.V
				break
			}
		}

		go utils.PublishCampaignMail(&models.Campaign{
			ShortCode: data.Campaign.ShortCode,
			Name:      data.Campaign.Name,
			CreatedBy: EditedBy,
			CreatedAt: data.Campaign.CreatedAt,
			ClientId:  data.Campaign.ClientId,
		}, impl.lgr, triggerImage, data.Campaign.TrackType, "campaign_published", ConsumeCreditRes.CreditType, ConsumeCreditRes.Balance, ConsumeCreditRes.Unlimited)
	} else {
		// go utils.CampaignPushNotif(&models.Campaign{
		// 	ID:        data.Campaign.ID,
		// 	ShortCode: data.Campaign.ShortCode,
		// 	Name:      data.Campaign.Name,
		// 	CreatedBy: EditedBy,
		// }, impl.lgr, "campaign_update")
	}

	return &CreditConsume, nil
}

func (impl *ExperienceDaoImpl) ResetExperienceDao(dto *dtos.ExperienceResetDto, editedBy *models.User) (*models.Experience, *errors.AppError) {
	experienceID, err := primitive.ObjectIDFromHex(dto.ExperienceID)
	if err != nil {
		return nil, errors.BadRequest("invalid experience id" + err.Error())
	}

	oldExp, err := impl.GetExperienceByID(experienceID)
	if err != nil {
		return nil, errors.BadRequest("failed to get experience: " + err.Error())
	}

	if oldExp.Status == consts.Processed {
		return nil, errors.BadRequest("experience already processed")
	}

	filter := bson.M{"_id": experienceID}

	if dto.UIElements == nil {
		dto.UIElements = &models.UIElements{}
	}

	var exp models.Experience
	exp.ID = experienceID
	exp.Canvas.IOS = 2100
	exp.Name = dto.Name
	exp.CampaignID = oldExp.CampaignID
	exp.Status = consts.Draft
	exp.Variant = *dto.Variant
	exp.IsActive = true
	exp.UIElements = dto.UIElements
	exp.CreatedBy = oldExp.CreatedBy
	exp.TemplateDetails = oldExp.TemplateDetails
	exp.TemplateCategory = dto.TemplateCategory
	if dto.Rewards == nil {
		exp.Rewards = models.Rewards{
			Enabled: false,
		}
	}
	if exp.Variant.ScaleAxis.X == 0 {
		exp.Variant.ScaleAxis.X = 1
	}
	if exp.Variant.ScaleAxis.Y == 0 {
		exp.Variant.ScaleAxis.Y = 1
	}
	if dto.EngagmentOptions != nil {
		exp.EngagmentOptions = dto.EngagmentOptions
	}
	exp.Scene = dto.Scene
	exp.Overlay = dto.Overlay
	exp.QrPanel = dto.QrPanel
	exp.CreatedAt = oldExp.CreatedAt
	exp.UpdatedAt = time.Duration(time.Now().UnixMilli())
	exp.EditedBy = editedBy
	exp.WorkflowID = oldExp.WorkflowID
	exp.Videos = []models.Video{}
	exp.Images = []models.Image{}
	exp.GLBs = []models.GLB{}
	if oldExp.CreditAllowanceID != primitive.NilObjectID {
		exp.CreditAllowanceID = oldExp.CreditAllowanceID
	}

	exp.CreditDeduct = oldExp.CreditDeduct

	coll := impl.db.Collection(consts.ExperienceCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	opts := options.Replace().SetUpsert(true)
	_, err = coll.ReplaceOne(ctx, filter, exp, opts)
	if err != nil {
		return nil, errors.InternalServerError("failed to upsert experience: " + err.Error())
	}

	var campaign models.Campaign
	query := bson.M{"_id": oldExp.CampaignID}
	coll = impl.db.Collection(consts.CampaignCollection)
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Duration(time.Now().UnixMilli()),
		},
	}
	result := coll.FindOneAndUpdate(ctx, query, update)
	if err := result.Err(); err != nil {
		return nil, errors.InternalServerError("failed to update campaign" + err.Error())
	}
	err = result.Decode(&campaign)
	if err != nil {
		return nil, errors.InternalServerError("failed to decode campaign" + err.Error())
	}

	return &exp, nil
}

func (impl *ExperienceDaoImpl) GetCategoryByCampaignShortCodeDao(campaignShortCode string) ([]models.Category, error) {
	coll := impl.db.Collection(consts.CategoryCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{"categories.campaigns": bson.M{"$in": []string{campaignShortCode}}, "is_active": true}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var categories []models.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

func (impl *ExperienceDaoImpl) GetExpByProductId(productId []string, clientId string) ([]string, error) {
	coll := impl.db.Collection(consts.ExperienceCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ClientObjId, err := primitive.ObjectIDFromHex(clientId)
	if err != nil {
		return nil, err
	}
	pipeline := []bson.M{
		{"$match": bson.M{"catalogue_details.product_id": bson.M{"$in": productId}, "is_active": true}},
		{"$lookup": bson.M{
			"from":         "campaigns",
			"localField":   "campaign_id",
			"foreignField": "_id",
			"as":           "campaign",
		}},
		{"$match": bson.M{"campaign.client_id": ClientObjId}},
		{"$group": bson.M{
			"_id":         nil,
			"product_ids": bson.M{"$addToSet": "$catalogue_details.product_id"},
		}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return []string{}, nil
	}

	productIds, ok := result[0]["product_ids"].(primitive.A)
	if !ok {
		return []string{}, nil
	}

	var ids []string
	for _, id := range productIds {
		if strId, ok := id.(string); ok {
			ids = append(ids, strId)
		}
	}
	return ids, nil
}

func (impl *ExperienceDaoImpl) UpdateRegenerateExperienceAssetsDao(WfResult dtos.WorkflowFinalPubResult) error {
	parts := strings.Split(WfResult.WorkflowId, "_")
	fmt.Println(parts)
	ExpObjId, err := primitive.ObjectIDFromHex(parts[1])
	if err != nil {
		log.Printf("Error converting workflow ID to ObjectID: %v", err)
		return err
	}
	filter := bson.M{"_id": ExpObjId}
	ExpColl := impl.db.Collection(consts.ExperienceCollection)

	updateOperation := map[string]interface{}{}
	updates := bson.M{
		"$set": updateOperation,
	}

	if WfResult.Status == dtos.Completed {
		for _, task := range WfResult.TaskResults {
			taskID := task.TaskId
			data := task.Payload
			switch {
			case taskID == "main_fal_low_resolution":
				if data.GenStudioOutput != nil && data.GenStudioOutput.Value != "" {
					updateOperation["video_generation.video_url"] = data.GenStudioOutput.Value
					updateOperation["video_generation.status"] = consts.Processed
				}
			}
		}
	}

	if WfResult.Status == dtos.Failed {
		updateOperation["video_generation.status"] = consts.Failed
	}

	if WfResult.Status == dtos.TimedOut {
		updateOperation["video_generation.status"] = consts.Failed
	}

	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result := ExpColl.FindOneAndUpdate(ctx, filter, updates, &opt)
	if err := result.Err(); err != nil {
		log.Printf("Error updating experience: %v", err)
		return err
	}
	return nil
}
