package models

type CreateTaskReq struct {
	Title       string `json:"title" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=1000"`
}

type UpdateTaskReq struct {
    Title       *string `json:"title,omitempty" binding:"omitempty,min=3,max=200"`
    Description *string `json:"description,omitempty"`
    Completed   *bool   `json:"completed,omitempty"`
}

