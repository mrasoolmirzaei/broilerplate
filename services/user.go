package services

import (
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/broilerplate/config"
	"github.com/muety/broilerplate/models"
	"github.com/muety/broilerplate/repositories"
	"github.com/muety/broilerplate/utils"
	"github.com/patrickmn/go-cache"
	uuid "github.com/satori/go.uuid"
	"time"
)

type UserService struct {
	config      *config.Config
	cache       *cache.Cache
	eventBus    *hub.Hub
	mailService IMailService
	repository  repositories.IUserRepository
}

func NewUserService(mailService IMailService, userRepo repositories.IUserRepository) *UserService {
	srv := &UserService{
		config:      config.Get(),
		eventBus:    config.EventBus(),
		cache:       cache.New(1*time.Hour, 2*time.Hour),
		mailService: mailService,
		repository:  userRepo,
	}

	return srv
}

func (srv *UserService) GetUserById(userId string) (*models.User, error) {
	if u, ok := srv.cache.Get(userId); ok {
		return u.(*models.User), nil
	}

	u, err := srv.repository.GetById(userId)
	if err != nil {
		return nil, err
	}

	srv.cache.Set(u.ID, u, cache.DefaultExpiration)
	return u, nil
}

func (srv *UserService) GetUserByKey(key string) (*models.User, error) {
	if u, ok := srv.cache.Get(key); ok {
		return u.(*models.User), nil
	}

	u, err := srv.repository.GetByApiKey(key)
	if err != nil {
		return nil, err
	}

	srv.cache.SetDefault(u.ID, u)
	return u, nil
}

func (srv *UserService) GetUserByEmail(email string) (*models.User, error) {
	return srv.repository.GetByEmail(email)
}

func (srv *UserService) GetUserByResetToken(resetToken string) (*models.User, error) {
	return srv.repository.GetByResetToken(resetToken)
}

func (srv *UserService) GetAll() ([]*models.User, error) {
	return srv.repository.GetAll()
}

func (srv *UserService) Count() (int64, error) {
	return srv.repository.Count()
}

func (srv *UserService) CreateOrGet(signup *models.Signup, isAdmin bool) (*models.User, bool, error) {
	u := &models.User{
		ID:       signup.Username,
		ApiKey:   uuid.NewV4().String(),
		Email:    signup.Email,
		Location: signup.Location,
		Password: signup.Password,
		IsAdmin:  isAdmin,
	}

	if hash, err := utils.HashBcrypt(u.Password, srv.config.Security.PasswordSalt); err != nil {
		return nil, false, err
	} else {
		u.Password = hash
	}

	return srv.repository.InsertOrGet(u)
}

func (srv *UserService) Update(user *models.User) (*models.User, error) {
	srv.cache.Flush()
	srv.notifyUpdate(user)
	return srv.repository.Update(user)
}

func (srv *UserService) ResetApiKey(user *models.User) (*models.User, error) {
	srv.cache.Flush()
	user.ApiKey = uuid.NewV4().String()
	return srv.Update(user)
}

func (srv *UserService) GenerateResetToken(user *models.User) (*models.User, error) {
	return srv.repository.UpdateField(user, "reset_token", uuid.NewV4())
}

func (srv *UserService) Delete(user *models.User) error {
	srv.cache.Flush()

	srv.notifyUpdate(user)

	return srv.repository.Delete(user)
}

func (srv *UserService) FlushCache() {
	srv.cache.Flush()
}

func (srv *UserService) notifyUpdate(user *models.User) {
	srv.eventBus.Publish(hub.Message{
		Name:   config.EventUserUpdate,
		Fields: map[string]interface{}{config.FieldPayload: user},
	})
}
