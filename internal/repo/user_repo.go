package repo

import (
	"database/sql"
	"task_manager/internal/models"
	apperrors "task_manager/internal/errors"
)

type UserRepository struct {
	db *sql.DB //holds db connection- pointer to sql.DB struct
	secretKey string //holds secret key for JWT
}
//constructor function - returns pointer to UserRepository struct		
func NewUserRepository(db *sql.DB, secretKey string) *UserRepository {
	return &UserRepository{db: db, secretKey: secretKey}
}

func (r *UserRepository) CreateUser(user *models.User) (int, error) {
	query := "INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id"
	var id int
	err := r.db.QueryRow(query, user.Username, user.Email, user.Password).Scan(&id)
	if err != nil {
		return 0, apperrors.Wrap(err, "error creating user")
	}
	return id, nil
}		

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := "SELECT id, username, email, password FROM users WHERE username = $1"
	var user models.User
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.Wrap(err, "user not found")
		}
		return nil, apperrors.Wrap(err, "error getting user by username")
	}
	return &user, nil
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	query := "SELECT id, username, email FROM users WHERE id = $1"
	var user models.User
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.Wrap(err, "user not found")
		}
		return nil, apperrors.Wrap(err, "error getting user by id")
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(id int, user *models.User) error {
	query := "UPDATE users SET username = $1, email = $2, password = $3 WHERE id = $4"
	result, err := r.db.Exec(query, user.Username, user.Email, user.Password, id)
	if err != nil {
		return apperrors.Wrap(err, "error updating user")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "error getting rows affected")
	}
	if rowsAffected == 0 {
		return apperrors.Wrap(sql.ErrNoRows, "user not found")
	}
	return nil
}

func (r *UserRepository) DeleteUser(id int) error {
	query := "DELETE FROM users WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return apperrors.Wrap(err, "error deleting user")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "error getting rows affected")
	}
	if rowsAffected == 0 {
		return apperrors.Wrap(sql.ErrNoRows, "user not found")
	}
	return nil
}