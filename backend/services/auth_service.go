package services

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"security-solution/config"
	"security-solution/models"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// GenerateJWT is used when a new token needs to be generated
// for an authenticated user, including controlled impersonation.
func (s *AuthService) GenerateJWT(user *models.User) (string, error) {
	return s.generateJWT(user)
}

func (s *AuthService) Register(user *models.User) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(user.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	user.Password = string(hashedPassword)

	if err := config.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(email, password string) (string, *models.User, error) {
	var user models.User

	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("invalid credentials")
		}

		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	token, err := s.generateJWT(&user)
	if err != nil {
		return "", nil, err
	}

	return token, &user, nil
}

func (s *AuthService) generateJWT(user *models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(secret))
}

func (s *AuthService) ValidateToken(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(secret), nil
	})
}

func (s *AuthService) GetUserByID(id string) (*models.User, error) {
	var user models.User

	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
