// - **Flow**:
//  1. Handler receives request → Parses JSON body
//  2. Handler calls authService → Passes credentials
//  3. AuthService validates → Hashes password (register) or compares hash (login)
//  4. AuthService calls userRepo → Database operations
//  5. Returns user → Handler generates JWT access token (15 min) + Refresh token (7 days) → Returns both to client
package services

import (
	"database/sql"
	"errors"
	apperrors "github.com/zahrashariati/task-manager/internal/errors"
	"github.com/zahrashariati/task-manager/internal/models"
	"github.com/zahrashariati/task-manager/internal/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo UserRepositoryInterface
	rtkRepo  RTKRepositoryInterface
}

func NewAuthService(userRepo UserRepositoryInterface, rtkRepo RTKRepositoryInterface) *AuthService {
	return &AuthService{userRepo: userRepo, rtkRepo: rtkRepo}
}

func (s *AuthService) Register(user *models.User) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, apperrors.Wrap(err, "error hashing password")
	}
	user.Password = string(hashedPassword)
	id, err := s.userRepo.CreateUser(user)
	if err != nil {
		return 0, apperrors.Wrap(err, "error creating user")
	}
	return id, nil
}

func (s *AuthService) Login(user *models.User) (*models.User, error) {
	dbUser, err := s.userRepo.GetByUsername(user.Username)
	if err == sql.ErrNoRows {
		return nil, apperrors.Wrap(errors.New("user not found"), "user not found")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, "error getting user by username")
	}
	if dbUser == nil {
		return nil, apperrors.Wrap(errors.New("user not found"), "user not found")
	}
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err != nil {
		return nil, apperrors.Wrap(err, "error comparing password")
	}
	return dbUser, nil
}

func (s *AuthService) GetUserByID(id int) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err == sql.ErrNoRows {
		return nil, apperrors.Wrap(errors.New("user not found"), "user not found")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, "error getting user by id")
	}
	return user, nil
}

func (s *AuthService) UpdateUser(id int, user *models.User) error {
	err := s.userRepo.UpdateUser(id, user)
	if err != nil {
		return apperrors.Wrap(err, "error updating user")
	}
	return nil
}

func (s *AuthService) DeleteUser(id int) error {
	err := s.userRepo.DeleteUser(id)
	if err != nil {
		return apperrors.Wrap(err, "error deleting user")
	}
	return nil
}

func (s *AuthService) CreateRefreshToken(user_id int) (string, error) {
	// Revoke all existing refresh tokens for this user (token rotation)
	// This ensures only one active refresh token per user at a time
	err := s.rtkRepo.RevokeAllRefreshTokensForUser(user_id)
	if err != nil {
		// Log but don't fail - continue with creating new token
		// This prevents issues if there are no existing tokens
	}

	token := utils.GenerateRandomToken()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	_, err = s.rtkRepo.CreateRefreshToken(user_id, token, expiresAt)
	if err != nil {
		return "", apperrors.Wrap(err, "error creating refresh token")
	}
	return token, nil
}

func (s *AuthService) RevokeRefreshToken(user_id int, token string) error {
	err := s.rtkRepo.RevokeRefreshToken(token)
	if err != nil {
		return apperrors.Wrap(err, "error revoking refresh token")
	}
	return nil
}

func (s *AuthService) ValidateRefreshToken(token string) (int, error) {
	refreshToken, err := s.rtkRepo.GetRefreshTokenByToken(token)
	if err != nil {
		return 0, apperrors.Wrap(err, "error getting refresh token")
	}
	if refreshToken == nil {
		return 0, apperrors.Wrap(errors.New("refresh token not found"), "refresh token not found")
	}
	if refreshToken.Revoked {
		return 0, apperrors.Wrap(errors.New("refresh token revoked"), "refresh token revoked")
	}
	if time.Now().After(refreshToken.ExpiresAt) {
		return 0, apperrors.Wrap(errors.New("refresh token expired"), "refresh token expired")
	}
	return refreshToken.UserID, nil
}

// RotateRefreshToken validates the old refresh token, revokes it, and creates a new one
// This implements refresh token rotation for better security
func (s *AuthService) RotateRefreshToken(oldToken string) (int, string, error) {
	// Validate the old refresh token
	userID, err := s.ValidateRefreshToken(oldToken)
	if err != nil {
		return 0, "", err
	}

	// Revoke the old refresh token
	err = s.rtkRepo.RevokeRefreshToken(oldToken)
	if err != nil {
		return 0, "", apperrors.Wrap(err, "error revoking old refresh token")
	}

	// Create a new refresh token
	newToken, err := s.CreateRefreshToken(userID)
	if err != nil {
		return 0, "", apperrors.Wrap(err, "error creating new refresh token")
	}

	return userID, newToken, nil
}
