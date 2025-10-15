package repository

import (
	"user-service/internal/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByEmail получает пользователя по email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmailAndPassword получает пользователя по email и паролю
func (r *UserRepository) GetByEmailAndPassword(email, password string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ? AND password = ?", email, password).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID получает пользователя по ID
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create создает нового пользователя
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// Update обновляет пользователя
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete удаляет пользователя
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// GetUsers получает пользователей с фильтрацией
func (r *UserRepository) GetUsers(role, active string) ([]models.User, error) {
	query := r.db

	if role != "" {
		query = query.Where("role = ?", role)
	}

	if active != "" {
		if active == "true" {
			query = query.Where("is_active = ?", true)
		} else if active == "false" {
			query = query.Where("is_active = ?", false)
		}
	}

	var users []models.User
	err := query.Find(&users).Error
	return users, err
}

// GetActiveUsers получает только активных пользователей
func (r *UserRepository) GetActiveUsers() ([]models.User, error) {
	var users []models.User
	err := r.db.Where("is_active = ?", true).Find(&users).Error
	return users, err
}

// GetUsersByRole получает пользователей по роли
func (r *UserRepository) GetUsersByRole(role string) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("role = ?", role).Find(&users).Error
	return users, err
}

// CountByRole возвращает количество пользователей по роли
func (r *UserRepository) CountByRole(role string) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("role = ?", role).Count(&count).Error
	return count, err
}

// CountActive возвращает количество активных пользователей
func (r *UserRepository) CountActive() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

// SearchUsers ищет пользователей по имени или email
func (r *UserRepository) SearchUsers(query string) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("name ILIKE ? OR email ILIKE ?", "%"+query+"%", "%"+query+"%").Find(&users).Error
	return users, err
}

// GetUsersWithPagination получает пользователей с пагинацией
func (r *UserRepository) GetUsersWithPagination(page, limit int, role, active string) ([]models.User, int64, error) {
	db := r.db

	if role != "" {
		db = db.Where("role = ?", role)
	}
	if active != "" {
		if active == "true" {
			db = db.Where("is_active = ?", true)
		} else if active == "false" {
			db = db.Where("is_active = ?", false)
		}
	}

	var total int64
	if err := db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []models.User
	offset := (page - 1) * limit
	err := db.Offset(offset).Limit(limit).Find(&users).Error

	return users, total, err
}

// IsEmailExists проверяет существование email
func (r *UserRepository) IsEmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
