package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendEmail gửi email thông qua SMTP Server (VD: Gmail)
func SendEmail(to []string, subject string, body string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST") // smtp.gmail.com
	smtpPort := os.Getenv("SMTP_PORT") // 587

	if from == "" || password == "" {
		return fmt.Errorf("chưa cấu hình SMTP_EMAIL hoặc SMTP_PASSWORD trong file .env")
	}

	// Cấu hình Header của Email (Hỗ trợ HTML)
	msg := []byte("From: SENT SOC System <" + from + ">\r\n" +
		"To: " + to[0] + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
		body)

	// Cấu hình Authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Thực hiện gửi mail
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		return fmt.Errorf("lỗi khi gửi email: %v", err)
	}
	return nil
}
