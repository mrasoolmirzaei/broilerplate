package services

import (
	"github.com/muety/broilerplate/models"
)

type IKeyValueService interface {
	GetString(string) (*models.KeyStringValue, error)
	MustGetString(string) *models.KeyStringValue
	PutString(*models.KeyStringValue) error
	DeleteString(string) error
}

type IMailService interface {
	SendPasswordReset(*models.User, string) error
}

type IUserService interface {
	GetUserById(string) (*models.User, error)
	GetUserByKey(string) (*models.User, error)
	GetUserByEmail(string) (*models.User, error)
	GetUserByResetToken(string) (*models.User, error)
	GetAll() ([]*models.User, error)
	Count() (int64, error)
	CreateOrGet(*models.Signup, bool) (*models.User, bool, error)
	Update(*models.User) (*models.User, error)
	Delete(*models.User) error
	ResetApiKey(*models.User) (*models.User, error)
	GenerateResetToken(*models.User) (*models.User, error)
	FlushCache()
}
