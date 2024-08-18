package usecase

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	UserType string `json:"user_type"`
}

type CreateUserResponse struct {
	UserId int64 `json:"id"`
}

type LoginRequest struct {
	ID       int    `json:"id"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type LoginResponseError struct {
}
