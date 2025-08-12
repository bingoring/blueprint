package handlers

import (
	"blueprint-module/pkg/queue"
	"blueprint-worker/internal/config"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
)

type EmailHandler struct {
	config *config.Config
	auth   smtp.Auth
}

func NewEmailHandler(cfg *config.Config) *EmailHandler {
	// SMTP 인증 설정
	auth := smtp.PlainAuth("",
		cfg.Email.SMTPUsername,
		cfg.Email.SMTPPassword,
		cfg.Email.SMTPHost,
	)

	return &EmailHandler{
		config: cfg,
		auth:   auth,
	}
}

func (h *EmailHandler) StartEmailWorker() error {
	log.Println("📧 Email worker started")

	return queue.ConsumeJobs("email_queue", "email_workers", "email_worker_1", h.handleEmailJob)
}

func (h *EmailHandler) handleEmailJob(jobData map[string]interface{}) error {
	jobType, ok := jobData["type"].(string)
	if !ok {
		return fmt.Errorf("missing job type")
	}

	switch jobType {
	case "send_email":
		return h.sendEmail(jobData)
	case "magic_link":
		return h.sendMagicLinkEmail(jobData)
	default:
		return fmt.Errorf("unknown email job type: %s", jobType)
	}
}

func (h *EmailHandler) sendEmail(jobData map[string]interface{}) error {
	// 필수 필드 추출
	to, ok := jobData["to"].(string)
	if !ok {
		return fmt.Errorf("missing email recipient")
	}

	template, ok := jobData["template"].(string)
	if !ok {
		return fmt.Errorf("missing email template")
	}

	data, ok := jobData["data"].(map[string]interface{})
	if !ok {
		data = make(map[string]interface{})
	}

	// 템플릿에 따른 이메일 내용 생성
	subject, body, err := h.generateEmailContent(template, data)
	if err != nil {
		return fmt.Errorf("failed to generate email content: %w", err)
	}

	// 이메일 전송
	if err := h.sendSMTP(to, subject, body); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", to, err)
	}

	log.Printf("✅ Email sent successfully to %s (template: %s)", to, template)
	return nil
}

func (h *EmailHandler) generateEmailContent(template string, data map[string]interface{}) (string, string, error) {
	switch template {
	case "email_verification":
		code, ok := data["code"].(string)
		if !ok {
			return "", "", fmt.Errorf("missing verification code")
		}
		username, _ := data["username"].(string)

		subject := "[Blueprint] 이메일 인증 코드"
		body := fmt.Sprintf(`
안녕하세요 %s님,

Blueprint 이메일 인증 코드입니다:

인증 코드: %s

이 코드는 15분간 유효합니다.
본인이 요청하지 않은 경우 이 메일을 무시해주세요.

감사합니다.
Blueprint 팀
`, username, code)

		return subject, body, nil

	case "work_email_verification":
		code, ok := data["code"].(string)
		if !ok {
			return "", "", fmt.Errorf("missing verification code")
		}
		company, _ := data["company"].(string)

		subject := "[Blueprint] 직장 이메일 인증"
		body := fmt.Sprintf(`
안녕하세요,

%s 소속 확인을 위한 인증 코드입니다:

인증 코드: %s

이 코드는 15분간 유효합니다.
본인이 요청하지 않은 경우 이 메일을 무시해주세요.

감사합니다.
Blueprint 팀
`, company, code)

		return subject, body, nil

	default:
		return "", "", fmt.Errorf("unknown email template: %s", template)
	}
}

func (h *EmailHandler) sendSMTP(to, subject, body string) error {
	// 이메일 메시지 구성
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body))

	// SMTP 서버 주소
	addr := h.config.Email.SMTPHost + ":" + h.config.Email.SMTPPort

	// Gmail 같은 경우 TLS 필요
	if h.config.Email.SMTPHost == "smtp.gmail.com" {
		return h.sendSMTPWithTLS(addr, to, msg)
	}

	// 일반 SMTP 전송
	return smtp.SendMail(addr, h.auth, h.config.Email.FromEmail, []string{to}, msg)
}

func (h *EmailHandler) sendSMTPWithTLS(addr, to string, msg []byte) error {
	// TLS 연결 설정
	tlsConfig := &tls.Config{
		ServerName: h.config.Email.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, h.config.Email.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Quit()

	// SMTP 인증
	if err := client.Auth(h.auth); err != nil {
		return err
	}

	// 송신자 설정
	if err := client.Mail(h.config.Email.FromEmail); err != nil {
		return err
	}

	// 수신자 설정
	if err := client.Rcpt(to); err != nil {
		return err
	}

	// 메시지 전송
	w, err := client.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = w.Write(msg)
	return err
}

// sendMagicLinkEmail 매직링크 이메일 전송
func (h *EmailHandler) sendMagicLinkEmail(jobData map[string]interface{}) error {
	// 필수 필드 추출
	email, ok := jobData["email"].(string)
	if !ok {
		return fmt.Errorf("missing email address")
	}

	code, ok := jobData["code"].(string)
	if !ok {
		return fmt.Errorf("missing verification code")
	}

	// 이메일 내용 생성 (Polymarket 스타일)
	subject := "Log in to Blueprint"
	body := h.generateMagicLinkHTML(email, code)

	return h.sendHTMLEmail(email, subject, body)
}

// generateMagicLinkHTML Polymarket 스타일 매직링크 HTML 생성
func (h *EmailHandler) generateMagicLinkHTML(email, code string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Log in to Blueprint</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background-color: #f8f9fa;
            color: #333;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #ffffff;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #1890ff 0%%, #9333ea 100%%);
            padding: 40px 30px;
            text-align: center;
        }
        .logo {
            font-size: 48px;
            margin-bottom: 16px;
        }
        .header h1 {
            margin: 0;
            color: white;
            font-size: 28px;
            font-weight: 600;
        }
        .content {
            padding: 40px 30px;
            text-align: center;
        }
        .message {
            font-size: 16px;
            line-height: 1.6;
            color: #666;
            margin-bottom: 32px;
        }
        .email-address {
            font-size: 18px;
            font-weight: 600;
            color: #1890ff;
            margin: 16px 0;
        }
        .security-section {
            background-color: #f8f9fa;
            border-radius: 12px;
            padding: 24px;
            margin: 32px 0;
            border: 2px dashed #e0e0e0;
        }
        .security-title {
            font-size: 14px;
            color: #666;
            margin-bottom: 12px;
        }
        .code {
            font-size: 32px;
            font-weight: bold;
            color: #333;
            letter-spacing: 4px;
            margin: 12px 0;
            font-family: 'Courier New', monospace;
        }
        .login-button {
            display: inline-block;
            background: linear-gradient(135deg, #1890ff 0%%, #9333ea 100%%);
            color: white;
            text-decoration: none;
            padding: 16px 32px;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            margin: 24px 0;
            box-shadow: 0 4px 12px rgba(24, 144, 255, 0.3);
        }
        .expiry-info {
            font-size: 14px;
            color: #666;
            margin: 24px 0;
        }
        .footer {
            background-color: #f8f9fa;
            padding: 24px 30px;
            text-align: center;
            font-size: 12px;
            color: #999;
        }
        .powered-by {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 8px;
            margin-top: 16px;
        }
        .magic-logo {
            width: 16px;
            height: 16px;
            opacity: 0.6;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">🔮</div>
            <h1>Blueprint</h1>
        </div>

        <div class="content">
            <div class="message">
                <strong>Check your email</strong><br>
                Log in using the magic link sent to
            </div>

            <div class="email-address">%s</div>

            <div class="security-section">
                <div class="security-title">Then enter this security code:</div>
                <div class="code">%s</div>
            </div>

            <a href="#" class="login-button">Log in to Blueprint</a>

            <div class="expiry-info">
                <strong>Button not showing?</strong> <a href="#" style="color: #1890ff;">Click here</a><br><br>
                Confirming this request will securely log you in using<br>
                <strong>%s</strong><br><br>
                This login was requested using <strong>Chrome, Mac OS X</strong> at <strong>%s</strong>
            </div>
        </div>

        <div class="footer">
            - Blueprint Team

            <div class="powered-by">
                <span>Secured By</span>
                <div style="display: flex; align-items: center; gap: 4px; opacity: 0.6;">
                    <span style="font-weight: 600;">⚡</span>
                    <span>Magic</span>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`, email, code, email, "August 10, 2025")
}

// sendHTMLEmail HTML 이메일 전송
func (h *EmailHandler) sendHTMLEmail(to, subject, htmlBody string) error {
	from := h.config.Email.FromEmail

	// MIME 헤더 설정
	msg := []byte(fmt.Sprintf(`From: Blueprint <%s>
To: %s
Subject: %s
MIME-Version: 1.0
Content-Type: text/html; charset=UTF-8

%s`, from, to, subject, htmlBody))

	// SMTP 클라이언트 생성 및 전송
	client, err := smtp.Dial(fmt.Sprintf("%s:%s", h.config.Email.SMTPHost, h.config.Email.SMTPPort))
	if err != nil {
		return err
	}
	defer client.Close()

	// TLS 시작
	tlsConfig := &tls.Config{
		ServerName: h.config.Email.SMTPHost,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return err
	}

	// 인증
	if err = client.Auth(h.auth); err != nil {
		return err
	}

	// 송신자 설정
	if err = client.Mail(from); err != nil {
		return err
	}

	// 수신자 설정
	if err = client.Rcpt(to); err != nil {
		return err
	}

	// 메시지 전송
	w, err := client.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = w.Write(msg)
	return err
}
