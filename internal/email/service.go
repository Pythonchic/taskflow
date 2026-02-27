package email

import (
	"fmt"

	"github.com/resendlabs/resend-go"
)

type Service struct {
    client     *resend.Client
    from       string
    TestEmail  string
}

func NewService(apiKey, from, testEmail string) *Service {
    client := resend.NewClient(apiKey)
    return &Service{
        client:    client,
        from:      from,
        TestEmail: testEmail,
    }
}

// SendVerificationCode отправляет 6-значный код подтверждения
func (s *Service) SendVerificationCode(to, code string) error {
	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				.container {
					font-family: Arial, sans-serif;
					max-width: 600px;
					margin: 0 auto;
					padding: 20px;
					background-color: #f9f9f9;
					border-radius: 10px;
				}
				.header {
					text-align: center;
					color: #333;
				}
				.code {
					font-size: 48px;
					font-weight: bold;
					text-align: center;
					letter-spacing: 10px;
					color: #667eea;
					padding: 20px;
					background: white;
					border-radius: 10px;
					margin: 20px 0;
				}
				.footer {
					text-align: center;
					color: #999;
					font-size: 12px;
					margin-top: 20px;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<h1 class="header">Подтверждение email</h1>
				<p>Здравствуйте!</p>
				<p>Для завершения регистрации введите следующий код:</p>
				<div class="code">%s</div>
				<p>Код действителен в течение 15 минут.</p>
				<p>Если вы не регистрировались, просто проигнорируйте это письмо.</p>
			</div>
		</body>
		</html>
	`, code)

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: "Код подтверждения",
		Html:    html,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendWelcomeEmail отправляет приветственное письмо после подтверждения
func (s *Service) SendWelcomeEmail(to, name string) error {
	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				.container {
					font-family: Arial, sans-serif;
					max-width: 600px;
					margin: 0 auto;
					padding: 20px;
					background-color: #f9f9f9;
					border-radius: 10px;
				}
				.header {
					text-align: center;
					color: #667eea;
				}
				.message {
					font-size: 16px;
					line-height: 1.6;
					color: #333;
				}
				.button {
					display: inline-block;
					padding: 12px 24px;
					background-color: #667eea;
					color: white;
					text-decoration: none;
					border-radius: 5px;
					margin: 20px 0;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<h1 class="header">Добро пожаловать в TaskFlow!</h1>
				<div class="message">
					<p>Здравствуйте, %s!</p>
					<p>Ваш email успешно подтверждён. Теперь вы можете:</p>
					<ul>
						<li>Создавать задачи</li>
						<li>Отслеживать прогресс</li>
						<li>Организовывать свои дела</li>
					</ul>
					<p>Нажмите кнопку ниже, чтобы перейти к задачам:</p>
					<div style="text-align: center;">
						<a href="http://localhost:8080/tasks" class="button">Перейти к задачам</a>
					</div>
				</div>
			</div>
		</body>
		</html>
	`, name)

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: "Добро пожаловать в TaskFlow!",
		Html:    html,
	}

	_, err := s.client.Emails.Send(params)
	return err
}
