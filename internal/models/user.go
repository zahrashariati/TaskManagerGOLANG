// - **User struct**: Used in repository (database), service (business logic), handlers (HTTP)
// - **LoginRequest**: Used in `internal/handlers/auth_handler.go` → `Login()` handler
// - **RegisterRequest**: Used in `internal/handlers/auth_handler.go` → `Register()` handler
// - **AuthResponse**: Returned by Login and Register handlers to client
package models

 //User struct: Represents a user in the database
type User struct {
	ID int `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"-" db:"password"`
	Email string `json:"email" db:"email"`
}

//LoginRequest struct: Represents a login request
type LoginRequest struct {
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

//RegisterRequest struct: Represents a register request
type RegisterRequest struct {
	Username string `json:"username" db:"username"`
	Email string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

//AuthResponse struct: Represents a response after login (with tokens)
type AuthResponse struct {
	Token string `json:"token" db:"token"`
	RefreshToken string `json:"refresh_token" db:"refresh_token"`
	User User `json:"user" db:"user"`
}

//RegisterResponse struct: Represents a response after registration (no tokens)
type RegisterResponse struct {
	Message string `json:"message"`
	User User `json:"user"`
}