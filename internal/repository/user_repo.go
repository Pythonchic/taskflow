// internal/repository/user_repo.go
package repository

import (
	"errors"
	"taskflow/internal/database"
	"taskflow/internal/models"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Создание пользователя
func (r *UserRepository) Create(user *models.User) error {
	return database.DB.Create(user).Error
}

// Получение пользователя по email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := database.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Проверка существования email
func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int64
	err := database.DB.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// Получение пользователя по ID
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := database.DB.First(&user, id).Error
	return &user, err
}

func (r *UserRepository) CreateIfNotExists(user *models.User) (bool, error) {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Проверяем внутри транзакции
		var count int64
		if err := tx.Model(&models.User{}).Where("email = ?", user.Email).Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			return errors.New("ErrEmailExists") // кастомная ошибка
		}

		return tx.Create(user).Error
	})

	if err == errors.New("ErrEmailExists") {
		return false, nil
	}
	return true, err
}

// SaveVerificationCode сохраняет код подтверждения для пользователя
func (r *UserRepository) SaveVerificationCode(email, code string) error {
	return database.DB.Model(&models.User{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"verify_code":  code,
			"code_expires": time.Now().Add(15 * time.Minute),
			"is_verified":  false,
		}).Error
}

// VerifyUser проверяет код и активирует пользователя
func (r *UserRepository) VerifyUser(email, code string) (bool, error) {
	var user models.User
	err := database.DB.Where("email = ? AND verify_code = ? AND code_expires > ?",
		email, code, time.Now()).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // код неверный или истёк
		}
		return false, err
	}

	// Активируем пользователя
	err = database.DB.Model(&user).Updates(map[string]interface{}{
		"is_verified":  true,
		"verify_code":  "",
		"code_expires": nil,
	}).Error

	return true, err
}

// IsVerified проверяет, подтверждён ли email
func (r *UserRepository) IsVerified(email string) (bool, error) {
	var user models.User
	err := database.DB.Select("is_verified").Where("email = ?", email).First(&user).Error
	if err != nil {
		return false, err
	}
	return user.IsVerified, nil
}

// ClearExpiredCodes очищает просроченные коды (можно запускать по расписанию)
func (r *UserRepository) ClearExpiredCodes() error {
	return database.DB.Model(&models.User{}).
		Where("code_expires < ? AND is_verified = ?", time.Now(), false).
		Updates(map[string]interface{}{
			"verify_code":  "",
			"code_expires": nil,
		}).Error
}

// Delete удаляет пользователя (для неверифицированных)
func (r *UserRepository) Delete(id uint) error {
	return database.DB.Delete(&models.User{}, id).Error
}
