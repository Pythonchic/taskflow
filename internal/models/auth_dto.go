package models

// REQUESTS
type RegisterReq struct {
	FirstName string `json:"firstName" binding:"max=25"`
	LastName  string `json:"lastName" binding:"max=25"`
	Email     string `json:"email" binding:"required,email,max=50"`
	Password  string `json:"password" binding:"required,min=6,max=30"`
}

type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RESPONSES
type AuthResponse struct {
	Message  string       `json:"message"`
	Token    string       `json:"token"`
	Redirect string       `json:"redirect"`
	User     UserResponse `json:"user"`
}

type UserResponse struct {
	ID        uint    `json:"id"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
	Email     string  `json:"email"`
}
