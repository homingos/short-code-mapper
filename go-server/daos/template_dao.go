package dao

import (
	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/campaign-svc/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TemplateDao interface {
	CreateTemplateDao(createTemplateDto dtos.CreateTemplateDto) (string, error)
	GetTemplateByIdDao(ID primitive.ObjectID) (*models.Template, error)
	GetAllTemplateDao(filters dtos.TemplateFiltersDto) ([]*dtos.AllTemplatesWithTrackType, error)
	UpdateTemplateDao(updateTemplateDto dtos.UpdateTemplateDto) (string, error)
	GetAllRawTemplatesDao() ([]models.Template, error)
	GetTemplateCredit() ([]map[string]interface{}, error)
	GetAlphaTemplateDao() (map[string]interface{}, error)
}
