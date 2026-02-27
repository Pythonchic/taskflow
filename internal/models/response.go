package models

// Общие ответы
type MessageResponse struct {
    Message string `json:"message"`
}

type ErrorResponse struct {
    Error string `json:"error"`
}

type RedirectResponse struct {
	Redirect string `json:"redirect"`
}
