package services

import (
	"errors"
	"fmt"

	"user-service/internal/models"
	"user-service/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// CreateUser создает нового пользователя
func (s *UserService) CreateUser(req *models.UserCreateRequest) (*models.UserResponse, error) {
	exists, err := s.userRepo.IsEmailExists(req.Email)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки email: %w", err)
	}
	if exists {
		return nil, errors.New("пользователь с таким email уже существует")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	role := req.Role
	if role == "" {
		role = string(models.RoleUser)
	}

	if !models.UserRole(role).IsValid() {
		return nil, errors.New("недопустимая роль пользователя")
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     role,
		IsActive: true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// Login авторизует пользователя
func (s *UserService) Login(req *models.UserLoginRequest) (*models.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("неверный пароль")
	}

	if !user.IsActive {
		return nil, errors.New("пользователь деактивирован")
	}

	response := &models.LoginResponse{
		User:  user.ToResponse(),
		Token: "dummy_token",
	}

	return response, nil
}

// GetUsers получает список пользователей
func (s *UserService) GetUsers(page, limit int, role, active string) ([]models.UserResponse, int64, error) {
	users, total, err := s.userRepo.GetUsersWithPagination(page, limit, role, active)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения пользователей: %w", err)
	}

	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, total, nil
}

// GetUser получает пользователя по ID
func (s *UserService) GetUser(id uint) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// UpdateUser обновляет пользователя
func (s *UserService) UpdateUser(id uint, req *models.UserUpdateRequest) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		if req.Email != user.Email {
			exists, err := s.userRepo.IsEmailExists(req.Email)
			if err != nil {
				return nil, fmt.Errorf("ошибка проверки email: %w", err)
			}
			if exists {
				return nil, errors.New("пользователь с таким email уже существует")
			}
			user.Email = req.Email
		}
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("ошибка хеширования пароля: %w", err)
		}
		user.Password = string(hashedPassword)
	}
	if req.Role != "" {
		if !models.UserRole(req.Role).IsValid() {
			return nil, errors.New("недопустимая роль пользователя")
		}
		user.Role = req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("ошибка обновления пользователя: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// DeleteUser удаляет пользователя
func (s *UserService) DeleteUser(id uint) error {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("пользователь не найден")
		}
		return fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	if err := s.userRepo.Delete(id); err != nil {
		return fmt.Errorf("ошибка удаления пользователя: %w", err)
	}

	return nil
}

// ChangePassword меняет пароль пользователя
func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("пользователь не найден")
		}
		return fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("неверный текущий пароль")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("ошибка обновления пароля: %w", err)
	}

	return nil
}

// SearchUsers ищет пользователей
func (s *UserService) SearchUsers(query string) ([]models.UserResponse, error) {
	users, err := s.userRepo.SearchUsers(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска пользователей: %w", err)
	}

	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, nil
}

// GetUsersByRole получает пользователей по роли
func (s *UserService) GetUsersByRole(role string) ([]models.UserResponse, error) {
	if !models.UserRole(role).IsValid() {
		return nil, errors.New("недопустимая роль пользователя")
	}

	users, err := s.userRepo.GetUsersByRole(role)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователей по роли: %w", err)
	}

	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, nil
}

// ActivateUser активирует пользователя
func (s *UserService) ActivateUser(id uint) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	user.IsActive = true
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("ошибка активации пользователя: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// DeactivateUser деактивирует пользователя
func (s *UserService) DeactivateUser(id uint) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	user.IsActive = false
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("ошибка деактивации пользователя: %w", err)
	}

	response := user.ToResponse()
	return &response, nil
}
