package services

import (
	"github.com/muety/broilerplate/config"
	"github.com/muety/broilerplate/models"
	"github.com/muety/broilerplate/repositories"
)

type KeyValueService struct {
	config     *config.Config
	repository repositories.IKeyValueRepository
}

func NewKeyValueService(keyValueRepo repositories.IKeyValueRepository) *KeyValueService {
	return &KeyValueService{
		config:     config.Get(),
		repository: keyValueRepo,
	}
}

func (srv *KeyValueService) GetString(key string) (*models.KeyStringValue, error) {
	return srv.repository.GetString(key)
}

func (srv *KeyValueService) MustGetString(key string) *models.KeyStringValue {
	kv, err := srv.repository.GetString(key)
	if err != nil {
		return &models.KeyStringValue{
			Key:   key,
			Value: "",
		}
	}
	return kv
}

func (srv *KeyValueService) PutString(kv *models.KeyStringValue) error {
	return srv.repository.PutString(kv)
}

func (srv *KeyValueService) DeleteString(key string) error {
	return srv.repository.DeleteString(key)
}
