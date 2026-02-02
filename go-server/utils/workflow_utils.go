package utils

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/homingos/campaign-svc/config"
	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/campaign-svc/models"
	"github.com/homingos/campaign-svc/types/consts"
	"github.com/nrednav/cuid2"
)

func GenerateShortCode(fingerprint string) (string, error) {
	generate, err := cuid2.Init(
		cuid2.WithRandomFunc(rand.Float64),
		cuid2.WithLength(6),
		cuid2.WithFingerprint(fingerprint),
	)
	if err != nil {
		return "", err
	}
	return generate(), nil
}

func GetImageFromURL(url string) (string, error) {
	url = strings.Split(url, "?")[0]
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error getting image from URL: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	filename := fmt.Sprintf("%s%s", cuid2.Generate(), ".jpg")
	bucketName := config.LoadConfig().GCP.GCS_BUCKET
	objectName := GcpUploadKey(filename, "image")

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create storage client: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	object := bucket.Object(objectName)
	wc := object.NewWriter(ctx)
	_, err = wc.Write(body)
	if err != nil {
		return "", fmt.Errorf("error writing image to bucket: %v", err)
	}
	err = wc.Close()
	if err != nil {
		return "", fmt.Errorf("error closing writer: %v", err)
	}

	return GetGCPURL(objectName), nil
}

func GenerateUniqueID() string {
	return cuid2.Generate()
}

func GcpUploadKey(filename string, Type string) string {
	prefixes := strings.Split(filename, ".")
	extension := prefixes[len(prefixes)-1]
	id := GenerateUniqueID()
	var prefix string
	if strings.Contains(Type, "image") {
		prefix = "original/images/"
	} else {
		prefix = "original/videos/"
	}
	return fmt.Sprintf("%s%s%s%s", prefix, id, ".", extension)
}

func GetGCPURL(objectName string) string {
	return fmt.Sprintf("%s%s/%s", consts.GCPStoragePrefix, config.LoadConfig().GCP.GCS_BUCKET, objectName)
}

func createImageRagAITask(mediaProcess *models.MediaProcess, workflowID string) *dtos.Task {
	ogImage, _ := GetExperienceImageUrls(mediaProcess.Experience)
	if ogImage == "" {
		return nil
	}
	workflowInput := dtos.WorkflowInput{
		ShortCode:                mediaProcess.ShortCode,
		ImageVectorLLMProductJob: mediaProcess.ImageVectorLLMProductJob,
	}
	taskID := CreateMainTaskID("image_vector_llm")
	subject := GetSubjectForHandler("image_vector_llm")
	task := dtos.Task{
		WorkflowId: workflowID,
		Id:         taskID,
		Body:       workflowInput,
		Subject:    subject,
	}
	return &task
}

func createAlphaVideoTask(mediaProcess *models.MediaProcess, workflowID string) *dtos.Task {
	workflowInput := dtos.WorkflowInput{
		AlphaVideoJob: &models.AlphaVideoJob{},
	}
	taskID := CreateMainTaskID("alpha_video")
	subject := GetSubjectForHandler("alpha_video")
	task := dtos.Task{
		WorkflowId: workflowID,
		Id:         taskID,
		Body:       workflowInput,
		Subject:    subject,
	}
	return &task
}

func createFalVideoTask(mediaProcess *models.MediaProcess, workflowID string) *dtos.Task {
	workflowInput := dtos.WorkflowInput{
		GenStudioJob: mediaProcess.GenStudiJob,
	}
	taskID := CreateMainTaskID("fal")
	subject := GetGenStudioSubjectForHandler("video")
	task := dtos.Task{
		WorkflowId: workflowID,
		Id:         taskID,
		Body:       workflowInput,
		Subject:    subject,
	}
	return &task
}

func createExperienceWorkflow(tasks []dtos.Task, publish bool, workflowID string, optionalDependencies ...map[string][]string) dtos.Workflow {
	taskMap := make(map[string]dtos.Task, len(tasks))
	dependencyMap := make(map[string][]string)
	if len(optionalDependencies) > 0 {
		for _, dependencies := range optionalDependencies {
			for key, value := range dependencies {
				dependencyMap[key] = value
			}
		}
	}
	for _, task := range tasks {
		taskMap[task.Id] = task
	}

	workflow := dtos.Workflow{
		ID:           workflowID,
		ReplySubject: consts.WorkflowCompleteSubject,
		Tasks:        taskMap,
		Publish:      publish,
		Dependencies: dependencyMap,
	}

	return workflow
}

func CreateExperienceWorkflow(Task []dtos.Task, Publish bool, workflowID string, optionalDependencies ...map[string][]string) dtos.Workflow {

	taskMap := make(map[string]dtos.Task, len(Task))
	dependencyMap := make(map[string][]string)
	if len(optionalDependencies) > 0 {
		for _, dependencies := range optionalDependencies {
			for key, value := range dependencies {
				dependencyMap[key] = value
			}
		}
	}
	for _, task := range Task {
		taskMap[task.Id] = task
	}

	workflow := dtos.Workflow{
		ID:           workflowID,
		ReplySubject: consts.WorkflowCompleteSubject,
		Tasks:        taskMap,
		Publish:      Publish,
		Dependencies: dependencyMap,
	}

	return workflow
}

// return original image, spawn image
func GetExperienceImageUrls(Experience models.Experience) (string, string) {
	var OgImage, SpawnImage string
	for _, img := range Experience.Images {
		if img.K == "original" {
			OgImage = img.V
		} else if img.K == "spawn" {
			SpawnImage = img.V
		}
	}
	return OgImage, SpawnImage
}

func GetExperienceGreenScreenImageUrl(Experience models.Experience) string {
	var GreenScreenImage string
	for _, img := range Experience.Images {
		if img.K == "original_green_screen" {
			GreenScreenImage = img.V
		}
	}
	return GreenScreenImage
}

// return original video, spawn mask
func GetExperienceVideoUrls(Experience models.Experience) (string, string) {
	var OgVideo, MaskVideo string
	for _, vid := range Experience.Videos {
		if vid.K == "original" {
			OgVideo = vid.V
		} else if vid.K == "mask" {
			MaskVideo = vid.V
		}
	}
	return OgVideo, MaskVideo
}

// func GetSegmentVideoUrls(Experience models.Experience) ([]string, []string) {
// 	var SegmentVideos []string
// 	var SegmentMaskVideos []string
// 	for _, marker := range Experience.Variant.Segments.Markers {
// 		SegmentVideos = append(SegmentVideos, marker.Videos.Original)
// 		if marker.Videos.Mask != "" {
// 			SegmentMaskVideos = append(SegmentMaskVideos, marker.Videos.Mask)
// 		}
// 	}
// 	return SegmentVideos, SegmentMaskVideos
// }

func CreateImageWithQRTask(Experience models.Experience, ShortCode string, Req dtos.CampaignBulkCreateDto, WorkflowID string, NewExpID string, QRCoordinates []dtos.QRCoordinates, QRImageURl string) *dtos.Task {
	OgImage, SpawnImage := GetExperienceImageUrls(Experience)
	if OgImage == "" && SpawnImage == "" {
		return nil
	}
	WorkflowInput := dtos.WorkflowInput{
		URL:           OgImage,
		QRGenerate:    true,
		Variant:       Experience.Variant,
		QRCode:        Experience.QrCode,
		Template:      Experience.TemplateDetails,
		ShortCode:     ShortCode,
		ExperienceID:  NewExpID,
		Publish:       true,
		QRCoordinates: QRCoordinates,
		QRImageURl:    QRImageURl,
	}

	if Req.QrID != nil {
		WorkflowInput.QrID = *Req.QrID
	}
	if Req.QrBGColor != nil {
		WorkflowInput.QrBGColor = *Req.QrBGColor
	}
	if Req.QrTextColor != nil {
		WorkflowInput.QrTextColor = *Req.QrTextColor
	}

	TaskID := CreateMainTaskID("image")
	Subject := GetSubjectForHandler("image_qr")
	Task := dtos.Task{
		WorkflowId: WorkflowID,
		Id:         TaskID,
		Body:       WorkflowInput,
		Subject:    Subject,
	}
	return &Task
}

func CreateImageTask(mediaProcess *models.MediaProcess, WorkflowID string) *dtos.Task {
	OgImage, SpawnImage := GetExperienceImageUrls(mediaProcess.Experience)
	if OgImage == "" && SpawnImage == "" {
		return nil
	}
	WorkflowInput := dtos.WorkflowInput{
		URL:                 OgImage,
		SpawnImage:          SpawnImage,
		Variant:             mediaProcess.Experience.Variant,
		QRCode:              mediaProcess.Experience.QrCode,
		Template:            mediaProcess.Experience.TemplateDetails,
		ShortCode:           mediaProcess.ShortCode,
		ExperienceID:        mediaProcess.Experience.ID.Hex(),
		Publish:             mediaProcess.Publish,
		ScanUrl:             mediaProcess.ScanUrl,
		GenerateGreenScreen: mediaProcess.GenerateGreenScreen,
	}
	TaskID := CreateMainTaskID("image")
	Subject := GetSubjectForHandler("image")
	Task := dtos.Task{
		WorkflowId: WorkflowID,
		Id:         TaskID,
		Body:       WorkflowInput,
		Subject:    Subject,
	}
	return &Task
}

func CreateImageRagAITask(mediaProcess *models.MediaProcess, WorkflowID string) *dtos.Task {
	OgImage, _ := GetExperienceImageUrls(mediaProcess.Experience)
	if OgImage == "" {
		return nil
	}
	WorkflowInput := dtos.WorkflowInput{
		ShortCode:                mediaProcess.ShortCode,
		ImageVectorLLMProductJob: mediaProcess.ImageVectorLLMProductJob,
	}
	TaskID := CreateMainTaskID("image_vector_llm")
	Subject := GetSubjectForHandler("image_vector_llm")
	Task := dtos.Task{
		WorkflowId: WorkflowID,
		Id:         TaskID,
		Body:       WorkflowInput,
		Subject:    Subject,
	}
	return &Task
}

func CreateAlphaVideoTask(mediaProcess *models.MediaProcess, WorkflowID string) *dtos.Task {
	WorkflowInput := dtos.WorkflowInput{
		AlphaVideoJob: &models.AlphaVideoJob{},
	}
	TaskID := CreateMainTaskID("alpha_video")
	Subject := GetSubjectForHandler("alpha_video")
	Task := dtos.Task{
		WorkflowId: WorkflowID,
		Id:         TaskID,
		Body:       WorkflowInput,
		Subject:    Subject,
	}
	return &Task
}

func CreateFalVideoTask(mediaProcess *models.MediaProcess, WorkflowID string) *dtos.Task {
	WorkflowInput := dtos.WorkflowInput{
		GenStudioJob: mediaProcess.GenStudiJob,
	}
	TaskID := CreateMainTaskID("fal")
	if mediaProcess.GenStudiJob.LowResolution {
		TaskID = CreateMainTaskID("fal_low_resolution")
	}
	Subject := GetGenStudioSubjectForHandler("video")
	Task := dtos.Task{
		WorkflowId: WorkflowID,
		Id:         TaskID,
		Body:       WorkflowInput,
		Subject:    Subject,
	}
	return &Task
}

func CreateScanImageTask(mediaProcess *models.MediaProcess, WorkflowID string) *dtos.Task {
	if mediaProcess.ScanUrl == "" {
		return nil
	}
	WorkflowInput := dtos.WorkflowInput{
		ShortCode: mediaProcess.ShortCode,
		ScanUrl:   mediaProcess.ScanUrl,
	}
	TaskID := CreateMainTaskID("scan_image")
	Subject := GetSubjectForHandler("image")
	Task := dtos.Task{
		WorkflowId: WorkflowID,
		Id:         TaskID,
		Body:       WorkflowInput,
		Subject:    Subject,
	}
	return &Task

}

func CreateOverlayTask(mediaProcess *models.MediaProcess, WorkflowID string) *dtos.Task {
	OgImage, _ := GetExperienceImageUrls(mediaProcess.Experience)
	WorkflowInput := dtos.WorkflowInput{
		URL:          OgImage,
		Overlay:      *mediaProcess.Experience.Overlay,
		ShortCode:    mediaProcess.ShortCode,
		ExperienceID: mediaProcess.Experience.ID.Hex(),
		Publish:      mediaProcess.Publish,
	}
	TaskID := CreateMainTaskID("overlay")
	Subject := GetSubjectForHandler("overlay")
	Task := dtos.Task{
		WorkflowId: WorkflowID,
		Id:         TaskID,
		Body:       WorkflowInput,
		Subject:    Subject,
	}
	return &Task
}

func CreateVideoTask(mediaProcess *models.MediaProcess, WorkflowID string) *dtos.Task {
	VideoUrl, MaskVideoUrl := GetExperienceVideoUrls(mediaProcess.Experience)
	if VideoUrl == "" {
		return nil
	}
	WorkflowInput := dtos.WorkflowInput{
		URL:          VideoUrl,
		MaskURL:      MaskVideoUrl,
		Variant:      mediaProcess.Experience.Variant,
		QRCode:       mediaProcess.Experience.QrCode,
		ShortCode:    mediaProcess.ShortCode,
		ExperienceID: mediaProcess.Experience.ID.Hex(),
		Publish:      mediaProcess.Publish,
	}
	TaskID := CreateMainTaskID("video")
	Subject := GetSubjectForHandler("video")
	Task := dtos.Task{
		WorkflowId: WorkflowID,
		Id:         TaskID,
		Body:       WorkflowInput,
		Subject:    Subject,
	}
	return &Task
}

func CreateVideoTaskWithDependencies(mediaProcess *models.MediaProcess, WorkflowID string) *dtos.Task {

	WorkflowInput := dtos.WorkflowInput{
		Variant:      mediaProcess.Experience.Variant,
		QRCode:       mediaProcess.Experience.QrCode,
		ShortCode:    mediaProcess.ShortCode,
		ExperienceID: mediaProcess.Experience.ID.Hex(),
		Publish:      mediaProcess.Publish,
	}
	TaskID := CreateMainTaskID("video")
	Subject := GetSubjectForHandler("video")
	Task := dtos.Task{
		WorkflowId: WorkflowID,
		Id:         TaskID,
		Body:       WorkflowInput,
		Subject:    Subject,
	}
	return &Task
}

func CreateParallaxTask(scene *models.Scene, WorkflowID string, variant models.Variant) ([]dtos.Task, int32) {
	Tasks := []dtos.Task{}
	var TaskLenght int32
	for i := range scene.Parallax {
		if scene.Parallax[i].Mask != nil && scene.Parallax[i].Mask.URL != "" {
			WorkflowInput := dtos.WorkflowInput{
				URL:          scene.Parallax[i].Mask.URL,
				ExperienceID: scene.Parallax[i].ID.Hex(),
			}
			TaskID := CreateParallaxMaskTaskID(scene.Parallax[i].ID.Hex(), string(dtos.TypeImage))
			Subject := GetSubjectForHandler(dtos.TypeImage)
			Task := dtos.Task{
				WorkflowId: WorkflowID,
				Id:         TaskID,
				Body:       WorkflowInput,
				Subject:    Subject,
			}
			TaskLenght += 1
			Tasks = append(Tasks, Task)
		}
		for j := range scene.Parallax[i].Planes {
			Plane := scene.Parallax[i].Planes[j]
			if *Plane.Type == 0 {
				WorkflowInput := dtos.WorkflowInput{
					URL:          Plane.URL,
					ExperienceID: Plane.ID.Hex(),
					Variant: models.Variant{
						TrackType: "PARALLAX", //dont remove this, need it for parallax image processing
					},
				}
				TaskID := CreateParallaxPlaneTaskID(scene.Parallax[i].ID.Hex(), Plane.ID.Hex(), string(dtos.TypeImage))
				Subject := GetSubjectForHandler(dtos.TypeImage)
				Task := dtos.Task{
					WorkflowId: WorkflowID,
					Id:         TaskID,
					Body:       WorkflowInput,
					Subject:    Subject,
				}
				TaskLenght += 1
				Tasks = append(Tasks, Task)
			} else if *Plane.Type == 1 || *Plane.Type == 2 {
				WorkflowInput := dtos.WorkflowInput{
					URL:          Plane.URL,
					MaskURL:      Plane.Mask,
					ExperienceID: Plane.ID.Hex(),
					Variant:      variant,
				}
				TaskID := CreateParallaxPlaneTaskID(scene.Parallax[i].ID.Hex(), Plane.ID.Hex(), string(dtos.TypeVideo))
				Subject := GetSubjectForHandler(dtos.TypeVideo)
				Task := dtos.Task{
					WorkflowId: WorkflowID,
					Id:         TaskID,
					Body:       WorkflowInput,
					Subject:    Subject,
				}
				TaskLenght += consts.VideoTaskLength
				Tasks = append(Tasks, Task)
			}
		}

	}
	return Tasks, TaskLenght
}

func GetAllSegmentTask(Experience models.Experience, WorkflowID string, StitchWorkflowID string) ([]dtos.Task, []dtos.Task, int32) {
	if Experience.Variant.Segments == nil || len(Experience.Variant.Segments.Markers) == 0 {
		return nil, nil, 0
	}
	SegmentInfo := dtos.SegmentInfo{
		ImageInfo: []dtos.UpdatedVariantInfo{},
		VideoInfo: []dtos.UpdatedVariantInfo{},
		VideoUrls: []dtos.SegmentVideo{},
	}
	for _, marker := range Experience.Variant.Segments.Markers {
		if marker.Videos.Original != "" {
			SegmentInfo.VideoInfo = append(SegmentInfo.VideoInfo, dtos.UpdatedVariantInfo{
				MarkerId: marker.Id,
				VideoURL: marker.Videos.Original,
				MaskURL:  marker.Videos.Mask,
				Type:     "VIDEO",
			})

			SegmentInfo.VideoUrls = append(SegmentInfo.VideoUrls, dtos.SegmentVideo{
				MarkedID:    marker.Id,
				OriginalURL: marker.Videos.Original,
				MaskURL:     marker.Videos.Mask,
			})
		}
	}

	for _, btn := range Experience.Variant.Buttons {
		if btn.Type == "image" && btn.AssetUrl != "" {
			SegmentInfo.ImageInfo = append(SegmentInfo.ImageInfo, dtos.UpdatedVariantInfo{
				MarkerId: btn.MarkerId,
				AssetURL: btn.AssetUrl,
				Type:     "IMAGE",
			})
		}
	}
	if len(SegmentInfo.ImageInfo) == 0 && len(SegmentInfo.VideoInfo) == 0 {
		return nil, nil, 0
	}
	SegmentInfo.ProcessStitchVideo = true //setting true because 1st time processing
	Tasks, StitchTasks, TaskLenght := CreateSegmentTask(&SegmentInfo, WorkflowID, StitchWorkflowID, Experience.ID.Hex())
	return Tasks, StitchTasks, TaskLenght
}

func CreateMainTaskID(Type string) string {
	return fmt.Sprintf("main_%s", Type)
}

func CreateParallaxPlaneTaskID(ParallaxId string, PlaneId string, Type string) string {
	return fmt.Sprintf("parallaxId_%s_planeId_%s_%s", ParallaxId, PlaneId, Type)
}

func CreateParallaxMaskTaskID(ParallaxId string, Type string) string {
	return fmt.Sprintf("parallaxId_%s_mask_%s", ParallaxId, Type)
}

func CreateSegmentTaskID(MarkerId string, Type string) string {
	return fmt.Sprintf("markerId_%s_%s", MarkerId, Type)
}

func CreateStitchSegmentTaskID(ExpID string, Type string) string {
	return fmt.Sprintf("stitchsegment_%s_%s", ExpID, Type)
}

func GetSubjectForHandler(ProcessType dtos.ProcessType) string {
	return fmt.Sprintf("MEDIAPROCESSOR.%s.process", ProcessType)
}

func GetGenStudioSubjectForHandler(ProcessType dtos.ProcessType) string {
	return fmt.Sprintf("GENSTUDIO.%s.process", ProcessType)
}

func CreateSegmentTask(SegmentInfo *dtos.SegmentInfo, WorkflowID string, StitchWorkflowID string, ExpId string) ([]dtos.Task, []dtos.Task, int32) {
	Tasks := []dtos.Task{}
	StitchingTask := []dtos.Task{}
	var TaskLenght int32
	IsAlpha := false
	for _, image := range SegmentInfo.ImageInfo {
		WorkflowInput := dtos.WorkflowInput{
			URL:          image.AssetURL,
			ExperienceID: ExpId,
		}
		TaskId := CreateSegmentTaskID(image.MarkerId, "image")
		Subject := GetSubjectForHandler(dtos.TypeImage)
		Task := dtos.Task{
			WorkflowId: WorkflowID,
			Id:         TaskId,
			Body:       WorkflowInput,
			Subject:    Subject,
		}
		TaskLenght += 1
		Tasks = append(Tasks, Task)
	}
	for _, video := range SegmentInfo.VideoInfo {
		WorkflowInput := dtos.WorkflowInput{
			URL:          video.VideoURL,
			ExperienceID: ExpId,
		}
		if video.MaskURL != "" {
			IsAlpha = true
			WorkflowInput.MaskURL = video.MaskURL
			WorkflowInput.Variant = models.Variant{
				IsAlpha: &IsAlpha,
				Class:   3,
			}
		}
		TaskId := CreateSegmentTaskID(video.MarkerId, "video")
		Subject := GetSubjectForHandler(dtos.TypeVideo)
		Task := dtos.Task{
			WorkflowId: WorkflowID,
			Id:         TaskId,
			Body:       WorkflowInput,
			Subject:    Subject,
		}
		TaskLenght += consts.VideoTaskLength
		Tasks = append(Tasks, Task)
	}

	if len(SegmentInfo.VideoUrls) > 0 && SegmentInfo.ProcessStitchVideo {
		WorkflowInput := dtos.WorkflowInput{
			Stitch:       true,
			ExperienceID: ExpId,
			Variant: models.Variant{
				Class:     3,
				TrackType: "CARD",
				IsAlpha:   &IsAlpha,
			},
			Segments: SegmentInfo.VideoUrls,
		}
		TaskId := CreateStitchSegmentTaskID(ExpId, "video")
		Subject := GetSubjectForHandler(dtos.TypeVideo)
		Task := dtos.Task{
			WorkflowId: StitchWorkflowID,
			Id:         TaskId,
			Body:       WorkflowInput,
			Subject:    Subject,
		}
		StitchingTask = append(StitchingTask, Task)
	}

	return Tasks, StitchingTask, TaskLenght
}
