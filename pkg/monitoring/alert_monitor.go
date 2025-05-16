// Package monitoring fornece funcionalidades para monitoramento do sistema
package monitoring

import (
	"fmt"
	"mms_api/pkg/logger"

	"gopkg.in/mail.v2"
)

// AlertMonitor é uma interface para envio de alertas
type AlertMonitor interface {
	SendAlert(alertType string, message string)
}

// alertMonitorImpl é a implementação concreta do AlertMonitor
type alertMonitorImpl struct {
	config AlertConfig
	logger logger.Logger
}

// AlertConfig contém as configurações para o sistema de alertas
type AlertConfig struct {
	Enabled bool
	Email   EmailConfig
}

// EmailConfig contém as configurações para envio de email
type EmailConfig struct {
	Enabled      bool
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	ToEmails     []string
}

// NewAlertMonitor cria uma nova instância do monitor de alertas
func NewAlertMonitor(config AlertConfig, logger logger.Logger) AlertMonitor {
	return &alertMonitorImpl{
		config: config,
		logger: logger,
	}
}

// sendEmail envia um alerta por email
func (m *alertMonitorImpl) sendEmail(alertType string, message string) error {
	if !m.config.Email.Enabled {
		return nil
	}

	msg := mail.NewMessage()

	// Configurar email
	msg.SetHeader("From", m.config.Email.FromEmail)
	msg.SetHeader("To", m.config.Email.ToEmails...)
	msg.SetHeader("Subject", fmt.Sprintf("Alerta: %s", alertType))
	msg.SetBody("text/plain", message)

	// Criar cliente SMTP
	d := mail.NewDialer(
		m.config.Email.SMTPHost,
		m.config.Email.SMTPPort,
		m.config.Email.SMTPUsername,
		m.config.Email.SMTPPassword,
	)

	// Enviar email
	if err := d.DialAndSend(msg); err != nil {
		m.logger.Error("Erro ao enviar email", err, "tipo", alertType)
		return fmt.Errorf("erro ao enviar email: %v", err)
	}

	m.logger.Info("Email enviado com sucesso", "tipo", alertType)
	return nil
}

// SendAlert envia uma mensagem de alerta
func (m *alertMonitorImpl) SendAlert(alertType string, message string) {
	if !m.config.Enabled {
		return
	}

	// Registrar o alerta
	m.logger.Info("Alerta", "tipo", alertType, "mensagem", message)

	// Tentar enviar por email
	if err := m.sendEmail(alertType, message); err != nil {
		m.logger.Error("Falha no envio do alerta por email", err)
	}
}
