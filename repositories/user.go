package repositories

import (
	"errors"
	"github.com/muety/broilerplate/models"
	"gorm.io/gorm"
	"time"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetById(userId string) (*models.User, error) {
	u := &models.User{}
	if err := r.db.Where(&models.User{ID: userId}).First(u).Error; err != nil {
		return u, err
	}
	return u, nil
}

func (r *UserRepository) GetByIds(userIds []string) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.
		Model(&models.User{}).
		Where("id in ?", userIds).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetByApiKey(key string) (*models.User, error) {
	if key == "" {
		return nil, errors.New("invalid input")
	}
	u := &models.User{}
	if err := r.db.Where(&models.User{ApiKey: key}).First(u).Error; err != nil {
		return u, err
	}
	return u, nil
}

func (r *UserRepository) GetByResetToken(resetToken string) (*models.User, error) {
	if resetToken == "" {
		return nil, errors.New("invalid input")
	}
	u := &models.User{}
	if err := r.db.Where(&models.User{ResetToken: resetToken}).First(u).Error; err != nil {
		return u, err
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, errors.New("invalid input")
	}
	u := &models.User{}
	if err := r.db.Where(&models.User{Email: email}).First(u).Error; err != nil {
		return u, err
	}
	return u, nil
}

func (r *UserRepository) GetAll() ([]*models.User, error) {
	var users []*models.User
	if err := r.db.
		Where(&models.User{}).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetByLoggedInAfter(t time.Time) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.
		Where("last_logged_in_at >= ?", t.Local()).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Count() (int64, error) {
	var count int64
	if err := r.db.
		Model(&models.User{}).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *UserRepository) InsertOrGet(user *models.User) (*models.User, bool, error) {
	if u, err := r.GetById(user.ID); err == nil && u != nil && u.ID != "" {
		return u, false, nil
	}

	result := r.db.Create(user)
	if err := result.Error; err != nil {
		return nil, false, err
	}

	return user, true, nil
}

func (r *UserRepository) Update(user *models.User) (*models.User, error) {
	updateMap := map[string]interface{}{
		"api_key":           user.ApiKey,
		"password":          user.Password,
		"email":             user.Email,
		"last_logged_in_at": user.LastLoggedInAt,
		"reset_token":       user.ResetToken,
		"location":          user.Location,
	}

	result := r.db.Model(user).Updates(updateMap)
	if err := result.Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) UpdateField(user *models.User, key string, value interface{}) (*models.User, error) {
	result := r.db.Model(user).Update(key, value)
	if err := result.Error; err != nil {
		return nil, err
	}

	if result.RowsAffected != 1 {
		return nil, errors.New("nothing updated")
	}

	return user, nil
}

func (r *UserRepository) Delete(user *models.User) error {
	return r.db.Delete(user).Error
}
