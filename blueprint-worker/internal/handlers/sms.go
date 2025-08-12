package handlers

import (
	"blueprint-module/pkg/queue"
	"blueprint-worker/internal/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type SMSHandler struct {
	config *config.Config
}

type AligoSMSRequest struct {
	Key      string `json:"key"`
	UserID   string `json:"user_id"`
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"msg"`
	TestMode string `json:"testmode_yn,omitempty"`
}

type AligoSMSResponse struct {
	ResultCode string `json:"result_code"`
	Message    string `json:"message"`
	MsgID      string `json:"msg_id,omitempty"`
}

func NewSMSHandler(cfg *config.Config) *SMSHandler {
	return &SMSHandler{
		config: cfg,
	}
}

func (h *SMSHandler) StartSMSWorker() error {
	log.Println("📱 SMS worker started")

	return queue.ConsumeJobs("sms_queue", "sms_workers", "sms_worker_1", h.handleSMSJob)
}

func (h *SMSHandler) handleSMSJob(jobData map[string]interface{}) error {
	jobType, ok := jobData["type"].(string)
	if !ok {
		return fmt.Errorf("missing job type")
	}

	switch jobType {
	case "send_sms":
		return h.sendSMS(jobData)
	default:
		return fmt.Errorf("unknown SMS job type: %s", jobType)
	}
}

func (h *SMSHandler) sendSMS(jobData map[string]interface{}) error {
	// 필수 필드 추출
	to, ok := jobData["to"].(string)
	if !ok {
		return fmt.Errorf("missing SMS recipient")
	}

	message, ok := jobData["message"].(string)
	if !ok {
		return fmt.Errorf("missing SMS message")
	}

	// 프로바이더에 따른 SMS 전송
	switch h.config.SMS.Provider {
	case "aligo":
		return h.sendAligoSMS(to, message)
	case "twilio":
		return h.sendTwilioSMS(to, message)
	default:
		return fmt.Errorf("unsupported SMS provider: %s", h.config.SMS.Provider)
	}
}

func (h *SMSHandler) sendAligoSMS(to, message string) error {
	// Aligo SMS API 호출
	apiURL := "https://apis.aligo.in/send/"

	// 요청 데이터 준비
	data := url.Values{}
	data.Set("key", h.config.SMS.APIKey)
	data.Set("user_id", h.config.SMS.APISecret) // Aligo에서는 API Secret이 user_id 역할
	data.Set("sender", h.config.SMS.FromNumber)
	data.Set("receiver", to)
	data.Set("msg", message)
	data.Set("testmode_yn", "N") // 실제 발송

	// HTTP 요청 생성
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// HTTP 클라이언트로 요청 전송
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS request: %w", err)
	}
	defer resp.Body.Close()

	// 응답 읽기
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// 응답 파싱
	var aligoResp AligoSMSResponse
	if err := json.Unmarshal(body, &aligoResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// 결과 확인
	if aligoResp.ResultCode != "1" {
		return fmt.Errorf("SMS sending failed: %s", aligoResp.Message)
	}

	log.Printf("✅ SMS sent successfully to %s (msg_id: %s)", to, aligoResp.MsgID)
	return nil
}

func (h *SMSHandler) sendTwilioSMS(to, message string) error {
	// Twilio SMS API 구현
	// 실제 환경에서는 Twilio Go SDK 사용 권장

	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", h.config.SMS.APISecret)

	// 요청 데이터 준비
	data := url.Values{}
	data.Set("From", h.config.SMS.FromNumber)
	data.Set("To", to)
	data.Set("Body", message)

	// HTTP 요청 생성
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create Twilio request: %w", err)
	}

	// Basic Auth 설정
	req.SetBasicAuth(h.config.SMS.APISecret, h.config.SMS.APIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// HTTP 클라이언트로 요청 전송
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Twilio SMS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Twilio SMS failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("✅ Twilio SMS sent successfully to %s", to)
	return nil
}

// 추가: 휴대폰 번호 형식 검증
func (h *SMSHandler) validatePhoneNumber(phoneNumber string) error {
	// 한국 휴대폰 번호 형식 검증 (010-XXXX-XXXX 또는 01012345678)
	// 실제 환경에서는 더 정교한 검증 로직 구현 필요

	// 하이픈 제거
	cleaned := strings.ReplaceAll(phoneNumber, "-", "")

	// 길이 확인 (11자리)
	if len(cleaned) != 11 {
		return fmt.Errorf("invalid phone number length: %s", phoneNumber)
	}

	// 010으로 시작하는지 확인
	if !strings.HasPrefix(cleaned, "010") {
		return fmt.Errorf("phone number must start with 010: %s", phoneNumber)
	}

	return nil
}
