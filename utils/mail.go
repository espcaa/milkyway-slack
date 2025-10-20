package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
)

func SendOtpEmail(toEmail, otpLink string) error {
	smtpHost := os.Getenv("SMTP_HOST") // e.g., "smtp.purelymail.com"
	smtpPort := os.Getenv("SMTP_PORT") // e.g., "465"
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP credentials are not set")
	}

	// Construct the email message
	subject := "Your magic Milkyway/Slack link!"
	body := fmt.Sprintf("Hi! You recently requested a linking/unlinking of your Milkyway/Slack accounts. Click on the following link to proceed:\n\n%s", otpLink)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s\r\n", smtpUser, toEmail, subject, body))

	// Connect to the SMTP server using TLS
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false, // change to true if using self-signed certs
		ServerName:         smtpHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set sender and recipient
	if err := client.Mail(smtpUser); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err := client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send email data
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	if _, err := wc.Write(msg); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return nil
}
