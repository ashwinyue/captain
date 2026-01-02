package service

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

type EmailTestRequest struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	SMTPUseTLS   bool   `json:"smtp_use_tls"`
	IMAPHost     string `json:"imap_host"`
	IMAPPort     int    `json:"imap_port"`
	IMAPUsername string `json:"imap_username"`
	IMAPPassword string `json:"imap_password"`
	IMAPUseSSL   bool   `json:"imap_use_ssl"`
}

type EmailTestResult struct {
	SMTPStatus     string `json:"smtp_status"`
	SMTPMessage    string `json:"smtp_message"`
	IMAPStatus     string `json:"imap_status"`
	IMAPMessage    string `json:"imap_message"`
	OverallSuccess bool   `json:"overall_success"`
}

func (s *EmailService) TestConnection(req *EmailTestRequest) *EmailTestResult {
	result := &EmailTestResult{}

	// Test SMTP
	smtpResult := s.testSMTP(req.SMTPHost, req.SMTPPort, req.SMTPUsername, req.SMTPPassword, req.SMTPUseTLS)
	result.SMTPStatus = smtpResult["status"]
	result.SMTPMessage = smtpResult["message"]

	// Test IMAP
	imapResult := s.testIMAP(req.IMAPHost, req.IMAPPort, req.IMAPUsername, req.IMAPPassword, req.IMAPUseSSL)
	result.IMAPStatus = imapResult["status"]
	result.IMAPMessage = imapResult["message"]

	result.OverallSuccess = result.SMTPStatus == "success" && result.IMAPStatus == "success"
	return result
}

func (s *EmailService) testSMTP(host string, port int, username, password string, useTLS bool) map[string]string {
	addr := fmt.Sprintf("%s:%d", host, port)
	timeout := 30 * time.Second

	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return map[string]string{"status": "failed", "message": fmt.Sprintf("Connection failed: %v", err)}
	}
	defer conn.Close()

	if useTLS {
		tlsConfig := &tls.Config{ServerName: host}
		conn = tls.Client(conn, tlsConfig)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return map[string]string{"status": "failed", "message": fmt.Sprintf("SMTP client error: %v", err)}
	}
	defer client.Close()

	if useTLS && port != 465 {
		if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return map[string]string{"status": "failed", "message": fmt.Sprintf("STARTTLS failed: %v", err)}
		}
	}

	auth := smtp.PlainAuth("", username, password, host)
	if err := client.Auth(auth); err != nil {
		return map[string]string{"status": "failed", "message": fmt.Sprintf("Authentication failed: %v", err)}
	}

	return map[string]string{"status": "success", "message": fmt.Sprintf("SMTP connection successful to %s:%d", host, port)}
}

func (s *EmailService) testIMAP(host string, port int, username, password string, useSSL bool) map[string]string {
	addr := fmt.Sprintf("%s:%d", host, port)
	timeout := 30 * time.Second

	var conn net.Conn
	var err error

	if useSSL {
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", addr, &tls.Config{ServerName: host})
	} else {
		conn, err = net.DialTimeout("tcp", addr, timeout)
	}

	if err != nil {
		return map[string]string{"status": "failed", "message": fmt.Sprintf("Connection failed: %v", err)}
	}
	defer conn.Close()

	// Read greeting
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(timeout))
	_, err = conn.Read(buf)
	if err != nil {
		return map[string]string{"status": "failed", "message": fmt.Sprintf("Failed to read greeting: %v", err)}
	}

	// Send LOGIN command
	loginCmd := fmt.Sprintf("a1 LOGIN %s %s\r\n", username, password)
	_, err = conn.Write([]byte(loginCmd))
	if err != nil {
		return map[string]string{"status": "failed", "message": fmt.Sprintf("Failed to send login: %v", err)}
	}

	// Read response
	conn.SetReadDeadline(time.Now().Add(timeout))
	n, err := conn.Read(buf)
	if err != nil {
		return map[string]string{"status": "failed", "message": fmt.Sprintf("Failed to read response: %v", err)}
	}

	response := string(buf[:n])
	if len(response) > 2 && response[:2] == "a1" {
		if len(response) > 5 && response[3:5] == "OK" {
			return map[string]string{"status": "success", "message": fmt.Sprintf("IMAP connection successful to %s:%d", host, port)}
		}
	}

	return map[string]string{"status": "failed", "message": "IMAP authentication failed"}
}
