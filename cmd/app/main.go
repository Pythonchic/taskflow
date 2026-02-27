// cmd/app/main.go
package main

import (
	"taskflow/internal/config"
	"taskflow/internal/email"
	"taskflow/internal/server"

	prettyprint "taskflow/pkg/pretty_print"
)

func main() {
	cfg := config.Load()

	prettyprint.Init(prettyprint.Config{
		DebugMode: cfg.Debug,
	})

	// Создаём email сервис
	emailService := email.NewService(
		cfg.Email.ResendAPIKey,
		cfg.Email.FromEmail,
		cfg.Email.TestEmail,
	)

	// Передаём его в сервер
	srv := server.New(cfg, emailService)

	if err := srv.Setup(); err != nil {
		prettyprint.Fatal("Failed to setup server: %v", err)
	}

	if err := srv.Run(); err != nil {
		prettyprint.Fatal("Server error: %v", err)
	}
}
