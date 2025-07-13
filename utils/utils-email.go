package utils

import (

    "RAAS/core/config"

    "gopkg.in/mail.v2"
    "github.com/google/uuid"
     // make sure this is correctly imported based on your structure
)

type EmailConfig struct {
    Host     string
    Port     int
    Username string
    Password string
    From     string
    UseTLS   bool
}

func SendEmail(cfg EmailConfig, to, subject, body string) error {
    m := mail.NewMessage()
    m.SetHeader("From", cfg.From)
    m.SetHeader("To", to)
    m.SetHeader("Subject", subject)
    m.SetBody("text/html", body)

    d := mail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
    d.TLSConfig = nil // Optional: add custom TLS settings if needed

    return d.DialAndSend(m)
}

func GenerateVerificationToken() string {
    return uuid.New().String()
}

func GetEmailConfig() EmailConfig {
    return EmailConfig{
        Host:     config.Cfg.Cloud.EmailHost,
        Port:     config.Cfg.Cloud.EmailPort,
        Username: config.Cfg.Cloud.EmailHostUser,
        Password: config.Cfg.Cloud.EmailHostPassword,
        From:     config.Cfg.Cloud.DefaultFromEmail,
        UseTLS:   config.Cfg.Cloud.EmailUseTLS,
    }
}