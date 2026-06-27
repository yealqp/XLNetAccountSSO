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

	"github.com/XianLinNet/XLNetAccount/internal/model"
	"github.com/XianLinNet/XLNetAccount/internal/pkg/security"
	"github.com/XianLinNet/XLNetAccount/internal/repository"
)

const (
	verificationPurposeRegister        = "register"
	verificationPurposeProfilePassword = "profile_password"
	verificationCodeLifetime           = 10 * time.Minute
	verificationResendInterval         = 60 * time.Second
)

type VerificationService struct {
	store *repository.Store
}

func NewVerificationService(store *repository.Store) *VerificationService {
	return &VerificationService{store: store}
}

func (service *VerificationService) SendRegistrationCode(ctx context.Context, settings PlatformSettings, email string, captchaToken string) error {
	if !settings.AllowRegistration {
		return ErrForbidden
	}
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") {
		return fmt.Errorf("%w: 请输入有效邮箱", ErrInvalidInput)
	}
	existing, err := service.store.FindUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrConflict
	}
	return service.sendCode(ctx, email, verificationPurposeRegister, settings, captchaToken, true, "XLNetAccount 注册验证码", registrationCodeBody)
}

func (service *VerificationService) SendProfilePasswordCode(ctx context.Context, settings PlatformSettings, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") {
		return fmt.Errorf("%w: 当前账号未配置有效邮箱", ErrInvalidInput)
	}
	return service.sendCode(ctx, email, verificationPurposeProfilePassword, settings, "", false, "XLNetAccount 密码修改验证码", profilePasswordCodeBody)
}

func (service *VerificationService) VerifyProfilePasswordCode(ctx context.Context, email string, code string) error {
	return service.verifyCode(ctx, email, verificationPurposeProfilePassword, code)
}

func (service *VerificationService) sendCode(ctx context.Context, email string, purpose string, settings PlatformSettings, captchaToken string, requireCaptcha bool, subject string, bodyBuilder func(string) string) error {
	if requireCaptcha {
		if err := service.verifyCAPTCHA(ctx, settings, captchaToken); err != nil {
			return err
		}
	}
	latestCode, err := service.store.FindLatestEmailVerificationCode(ctx, email, purpose)
	if err != nil {
		return err
	}
	if latestCode != nil {
		remaining := time.Until(latestCode.CreatedAt.Add(verificationResendInterval))
		if remaining > 0 {
			remainingSeconds := int((remaining + time.Second - 1) / time.Second)
			return fmt.Errorf("%w: 同一邮箱验证码发送间隔为60秒，请在 %d 秒后重试", ErrInvalidInput, remainingSeconds)
		}
	}
	code, err := generateVerificationCode()
	if err != nil {
		return err
	}
	if err := service.store.DeleteEmailVerificationCodes(ctx, email, purpose); err != nil {
		return err
	}
	verification := &model.EmailVerificationCode{
		ID:        security.NewID(),
		Email:     email,
		Purpose:   purpose,
		CodeHash:  security.HashToken(code),
		ExpiresAt: time.Now().UTC().Add(verificationCodeLifetime),
	}
	if err := service.store.CreateEmailVerificationCode(ctx, verification); err != nil {
		return err
	}
	return sendSMTPMail(settings, email, subject, bodyBuilder(code))
}

func (service *VerificationService) VerifyRegistrationCode(ctx context.Context, email string, code string) error {
	return service.verifyCode(ctx, email, verificationPurposeRegister, code)
}

func (service *VerificationService) verifyCode(ctx context.Context, email string, purpose string, code string) error {
	verification, err := service.store.FindLatestEmailVerificationCode(ctx, strings.TrimSpace(strings.ToLower(email)), purpose)
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
	siteKey := strings.Trim(strings.TrimSpace(settings.CAPSiteKey), "/")
	secret := strings.TrimSpace(settings.CAPSecretKey)
	if endpoint == "" || siteKey == "" || secret == "" {
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
	verifyURL := strings.TrimRight(endpoint, "/") + "/" + siteKey + "/siteverify"
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
		Success bool     `json:"success"`
		Error   string   `json:"error"`
		Errors  []string `json:"error-codes"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return fmt.Errorf("%w: 人机验证响应无效", ErrInvalidInput)
	}
	if !payload.Success {
		reason := strings.TrimSpace(payload.Error)
		if reason == "" && len(payload.Errors) > 0 {
			reason = strings.TrimSpace(payload.Errors[0])
		}
		if reason != "" {
			return fmt.Errorf("%w: 人机验证失败（%s）", ErrInvalidInput, reason)
		}
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

func profilePasswordCodeBody(code string) string {
	return "您的 XLNetAccount 密码修改验证码是：" + code + "\n\n验证码 10 分钟内有效，请勿泄露给他人。"
}

func generateVerificationCode() (string, error) {
	max := big.NewInt(1000000)
	number, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", number.Int64()), nil
}
