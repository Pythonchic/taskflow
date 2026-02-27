// internal/handlers/auth.go
package handlers

import (
	"fmt"
	"net/http"
	"taskflow/internal/auth"
	"taskflow/internal/email"
	"taskflow/internal/models"
	"taskflow/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userRepo     *repository.UserRepository
	emailService *email.Service
	testEmail    string // 👈 просто строка, без лишних зависимостей
}

func NewAuthHandler(
	userRepo *repository.UserRepository,
	emailService *email.Service,
	testEmail string, // 👈 передаём только то что нужно
) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		emailService: emailService,
		testEmail:    testEmail,
	}
}

// POST /api/v1/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// ===== 1. ЕСЛИ ЭТО ТЕСТОВЫЙ EMAIL - УДАЛЯЕМ СТАРОГО ПОЛЬЗОВАТЕЛЯ =====
	if req.Email == h.testEmail {
		existing, _ := h.userRepo.GetByEmail(req.Email)
		if existing != nil {
			// Удаляем старого пользователя (даже если верифицирован!)
			if err := h.userRepo.Delete(existing.ID); err != nil {
				fmt.Printf("⚠️ Failed to delete test user: %v\n", err)
			} else {
				fmt.Println("🧹 Тестовый пользователь удалён для перерегистрации")
			}
		}
	} else {
		// ===== 2. ДЛЯ ОБЫЧНЫХ ПОЛЬЗОВАТЕЛЕЙ - ПРОВЕРЯЕМ ВЕРИФИКАЦИЮ =====
		existingUser, _ := h.userRepo.GetByEmail(req.Email)

		// Если пользователь уже есть и верифицирован - ошибка
		if existingUser != nil && existingUser.IsVerified {
			c.JSON(http.StatusConflict, models.ErrorResponse{Error: "Email already registered"})
			return
		}

		// Если пользователь есть, но не верифицирован - удаляем
		if existingUser != nil && !existingUser.IsVerified {
			if err := h.userRepo.Delete(existingUser.ID); err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to process registration"})
				return
			}
			fmt.Println("🗑️ Неверифицированный пользователь удалён")
		}
	}

	// ===== 3. СОЗДАЁМ НОВОГО ПОЛЬЗОВАТЕЛЯ =====
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to hash password"})
		return
	}

	// Генерируем код
	verificationCode := email.GenerateCode()

	user := &models.User{
		Email:       req.Email,
		Password:    hashedPassword,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		IsVerified:  false,
		VerifyCode:  verificationCode,
		CodeExpires: time.Now().Add(15 * time.Minute),
	}

	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create user"})
		return
	}

	// ===== 4. ОТПРАВЛЯЕМ КОД =====
	go func() {
		if err := h.emailService.SendVerificationCode(user.Email, verificationCode); err != nil {
			fmt.Printf("Failed to send verification email: %v\n", err)
		}
	}()

	// ===== 5. ОТВЕЧАЕМ =====
	c.JSON(http.StatusCreated, gin.H{
		"message":  "Registration successful. Please check your email for verification code.",
		"email":    user.Email,
		"redirect": "",
	})
}

// POST /api/v1/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid email or password"})
		return
	}

	// 👇 НОВАЯ ПРОВЕРКА
	if !user.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Email not verified",
			"email":   user.Email,
			"message": "Please verify your email first",
		})
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid email or password"})
		return
	}

	token, _ := auth.GenerateToken(user.ID, user.Email)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Login successful",
		"token":    token,
		"redirect": "/tasks",
		"user": gin.H{
			"id":        user.ID,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"email":     user.Email,
		},
	})
}

// POST /api/v1/verify
func (h *AuthHandler) Verify(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required,len=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Получаем пользователя
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid or expired code"})
		return
	}

	// Проверяем код
	verified, err := h.userRepo.VerifyUser(req.Email, req.Code)
	if err != nil {
		fmt.Printf("❌ DB error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Database error"})
		return
	}

	if !verified {
		fmt.Printf("❌ Code mismatch or expired\n")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid or expired code"})
		return
	}

	// Генерируем JWT
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	// Отправляем приветственное письмо (асинхронно)
	go func() {
		fullName := user.FirstName + " " + user.LastName
		if err := h.emailService.SendWelcomeEmail(user.Email, fullName); err != nil {
			fmt.Printf("Failed to send welcome email: %v\n", err)
		}
	}()

	// Успех - логиним пользователя
	c.JSON(http.StatusOK, gin.H{
		"message":  "Email verified successfully",
		"token":    token,
		"redirect": "/tasks",
		"user": gin.H{
			"id":        user.ID,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"email":     user.Email,
		},
	})
}

// POST /api/v1/resend-code
func (h *AuthHandler) ResendCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil || user == nil {
		// Не говорим, что пользователь не найден (безопасность)
		c.JSON(http.StatusOK, gin.H{"message": "If email exists, code will be sent"})
		return
	}

	if user.IsVerified {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Email already verified"})
		return
	}

	// Генерируем новый код
	newCode := email.GenerateCode()

	// Сохраняем в БД
	err = h.userRepo.SaveVerificationCode(user.Email, newCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate code"})
		return
	}

	// Отправляем
	go h.emailService.SendVerificationCode(user.Email, newCode)

	c.JSON(http.StatusOK, gin.H{"message": "Code sent successfully"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Очищаем cookie
	c.SetCookie(
		"token",
		"",
		-1, // maxAge = -1 → удалить
		"/",
		"",
		false, // secure (в dev false)
		true,  // httpOnly
	)

	// Отвечаем
	c.JSON(http.StatusOK, gin.H{
		"message":  "Logged out successfully",
		"redirect": "/login",
	})
}
