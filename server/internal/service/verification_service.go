package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"crypto/tls"

	"github.com/xianlin-network/sso-platform/server/internal/model"
	"github.com/xianlin-network/sso-platform/server/internal/pkg/security"
	"github.com/xianlin-network/sso-platform/server/internal/repository"
)

const (
	verificationPurposeRegister = "register"
	verificationCodeLifetime    = 10 * time.Minute
)

type VerificationService struct {
	store *repository.Store
}

func NewVerificationService(store *repository.Store) *VerificationService {
	return &VerificationService{store: store}
}

func (service *VerificationService) SendRegistrationCode(ctx context.Context, settings PlatformSettings, email string, captchaToken string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") {
		return ErrInvalidInput
	}
	if err := service.verifyCAPTCHA(ctx, settings, captchaToken); err != nil {
		return err
	}
	existing, err := service.store.FindUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrConflict
	}
	code, err := generateVerificationCode()
	if err != nil {
		return err
	}
	if err := service.store.DeleteEmailVerificationCodes(ctx, email, verificationPurposeRegister); err != nil {
		return err
	}
	verification := &model.EmailVerificationCode{
		ID:        security.NewID(),
		Email:     email,
		Purpose:   verificationPurposeRegister,
		CodeHash:  security.HashToken(code),
		ExpiresAt: time.Now().UTC().Add(verificationCodeLifetime),
	}
	if err := service.store.CreateEmailVerificationCode(ctx, verification); err != nil {
		return err
	}
	return sendSMTPMail(settings, email, "XLNetAccount 注册验证码", registrationCodeBody(code))
}

func (service *VerificationService) VerifyRegistrationCode(ctx context.Context, email string, code string) error {
	verification, err := service.store.FindLatestEmailVerificationCode(ctx, strings.TrimSpace(strings.ToLower(email)), verificationPurposeRegister)
	if err != nil {
		return err
	}
	if verification == nil || verification.ConsumedAt != nil || verification.ExpiresAt.Before(time.Now().UTC()) {
		return ErrInvalidGrant
	}
	if security.HashToken(strings.TrimSpace(code)) != verification.CodeHash {
		return ErrInvalidGrant
	}
	now := time.Now().UTC()
	verification.ConsumedAt = &now
	return service.store.SaveEmailVerificationCode(ctx, verification)
}

func (service *VerificationService) verifyCAPTCHA(ctx context.Context, settings PlatformSettings, token string) error {
	endpoint := strings.TrimSpace(settings.CAPAPIEndpoint)
	secret := strings.TrimSpace(settings.CAPSecretKey)
	if endpoint == "" || secret == "" {
		return nil
	}
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("%w: 请先完成人机验证", ErrInvalidInput)
	}
	body, err := json.Marshal(map[string]string{
		"secret":   secret,
		"response": strings.TrimSpace(token),
	})
	if err != nil {
		return err
	}
	verifyURL := strings.TrimRight(endpoint, "/") + "/siteverify"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, verifyURL, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("%w: 人机验证配置无效", ErrInvalidInput)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("%w: 人机验证请求失败", ErrInvalidInput)
	}
	defer response.Body.Close()
	var payload struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return fmt.Errorf("%w: 人机验证响应无效", ErrInvalidInput)
	}
	if !payload.Success {
		return fmt.Errorf("%w: 人机验证失败", ErrInvalidInput)
	}
	return nil
}

func sendSMTPMail(settings PlatformSettings, to string, subject string, body string) error {
	host := strings.TrimSpace(settings.SMTPHost)
	user := strings.TrimSpace(settings.SMTPUser)
	password := strings.TrimSpace(settings.SMTPPassword)
	port := strings.TrimSpace(settings.SMTPPort)
	if host == "" || user == "" || password == "" || port == "" {
		return fmt.Errorf("%w: 请先在设置页完善邮件服务配置", ErrInvalidInput)
	}
	address := net.JoinHostPort(host, port)
	message := buildSMTPMessage(user, to, subject, body)
	auth := smtp.PlainAuth("", user, password, host)
	if settings.SMTPTLS && port == "465" {
		return sendImplicitTLSMail(address, host, user, auth, to, message)
	}
	client, err := smtp.Dial(address)
	if err != nil {
		return fmt.Errorf("%w: 连接邮件服务器失败", ErrInvalidInput)
	}
	defer client.Close()
	if settings.SMTPTLS {
		if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return fmt.Errorf("%w: 启用邮件 TLS 失败", ErrInvalidInput)
		}
	}
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("%w: 邮件服务认证失败", ErrInvalidInput)
	}
	if err := client.Mail(user); err != nil {
		return fmt.Errorf("%w: 设置发件人失败", ErrInvalidInput)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("%w: 设置收件人失败", ErrInvalidInput)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("%w: 写入邮件内容失败", ErrInvalidInput)
	}
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("%w: 发送邮件失败", ErrInvalidInput)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("%w: 发送邮件失败", ErrInvalidInput)
	}
	return client.Quit()
}

func sendImplicitTLSMail(address string, host string, from string, auth smtp.Auth, to string, message string) error {
	conn, err := tls.Dial("tcp", address, &tls.Config{ServerName: host})
	if err != nil {
		return fmt.Errorf("%w: 连接邮件服务器失败", ErrInvalidInput)
	}
	defer conn.Close()
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("%w: 建立邮件客户端失败", ErrInvalidInput)
	}
	defer client.Close()
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("%w: 邮件服务认证失败", ErrInvalidInput)
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("%w: 设置发件人失败", ErrInvalidInput)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("%w: 设置收件人失败", ErrInvalidInput)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("%w: 写入邮件内容失败", ErrInvalidInput)
	}
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("%w: 发送邮件失败", ErrInvalidInput)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("%w: 发送邮件失败", ErrInvalidInput)
	}
	return client.Quit()
}

func buildSMTPMessage(from string, to string, subject string, body string) string {
	return strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")
}

func registrationCodeBody(code string) string {
	return "您的 XLNetAccount 注册验证码是：" + code + "\n\n验证码 10 分钟内有效，请勿泄露给他人。"
}

func generateVerificationCode() (string, error) {
	max := big.NewInt(1000000)
	number, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", number.Int64()), nil
}
