package repo

import (
	"time"
	"database/sql"
	"task_manager/internal/models"
	apperrors "task_manager/internal/errors"
)

type RTKRepository struct {
	db *sql.DB
}

func NewRTKRepository(db *sql.DB) *RTKRepository {
	return &RTKRepository{db: db}
}

func (r *RTKRepository) CreateRefreshToken(userID int, token string, expiresAt time.Time) (int, error) {
	query := "INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3) RETURNING id"
	var id int
	err := r.db.QueryRow(query, userID, token, expiresAt).Scan(&id)
	if err != nil {
		return id, apperrors.Wrap(err, "error creating refresh token")
	}
	return id, nil
}

func (r *RTKRepository) GetRefreshTokenByToken(token string) (*models.RTK, error) {
	query := "Select id, user_id, token, expires_at, created_at, revoked From refresh_tokens WHERE token = $1"
	var rtk models.RTK
	err := r.db.QueryRow(query, token).Scan(&rtk.ID, &rtk.UserID, &rtk.Token, &rtk.ExpiresAt, &rtk.CreatedAt, &rtk.Revoked)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.Wrap(err, "refresh token not found")
		}
		return nil, apperrors.Wrap(err, "error getting refresh token by token")
	}
	return &rtk, nil

}

func (r *RTKRepository) RevokeRefreshToken(token string) error {
	query := "UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1"
	result, err := r.db.Exec(query, token)
	if err != nil {
		return apperrors.Wrap(err, "error revoking refresh token")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "error getting rows affected")
	}
	if rowsAffected == 0 {
		return apperrors.Wrap(sql.ErrNoRows, "refresh token not found")
	}
	return nil
}

func (r *RTKRepository) GetRefreshTokenByUserID(userID int) (*models.RTK, error) {
	query := "SELECT id, user_id, token, expires_at, created_at, revoked FROM refresh_tokens WHERE user_id = $1"
	var rtk models.RTK
	err := r.db.QueryRow(query, userID).Scan(&rtk.ID, &rtk.UserID, &rtk.Token, &rtk.ExpiresAt, &rtk.CreatedAt, &rtk.Revoked) 
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.Wrap(err, "refresh token not found")
		}
		return nil, apperrors.Wrap(err, "error getting refresh token by user_id")
	}
	return &rtk, nil
}

// RevokeAllRefreshTokensForUser revokes all active refresh tokens for a user
func (r *RTKRepository) RevokeAllRefreshTokensForUser(userID int) error {
	query := "UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1 AND revoked = FALSE"
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return apperrors.Wrap(err, "error revoking all refresh tokens for user")
	}
	return nil
}