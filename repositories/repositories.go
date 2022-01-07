package repositories

import (
	"github.com/muety/broilerplate/models"
	"time"
)

type IKeyValueRepository interface {
	GetAll() ([]*models.KeyStringValue, error)
	GetString(string) (*models.KeyStringValue, error)
	PutString(*models.KeyStringValue) error
	DeleteString(string) error
}

type IUserRepository interface {
	GetById(string) (*models.User, error)
	GetByIds([]string) ([]*models.User, error)
	GetByApiKey(string) (*models.User, error)
	GetByEmail(string) (*models.User, error)
	GetByResetToken(string) (*models.User, error)
	GetAll() ([]*models.User, error)
	GetByLoggedInAfter(time.Time) ([]*models.User, error)
	Count() (int64, error)
	InsertOrGet(*models.User) (*models.User, bool, error)
	Update(*models.User) (*models.User, error)
	UpdateField(*models.User, string, interface{}) (*models.User, error)
	Delete(*models.User) error
}
