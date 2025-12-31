package dto

type CreateUserRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name"  binding:"required,min=2,max=80"`
}
