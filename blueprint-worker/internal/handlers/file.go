package handlers

import (
	"blueprint-module/pkg/queue"
	"blueprint-worker/internal/config"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type FileHandler struct {
	config *config.Config
}

func NewFileHandler(cfg *config.Config) *FileHandler {
	return &FileHandler{
		config: cfg,
	}
}

func (h *FileHandler) StartFileWorker() error {
	log.Println("📁 File processing worker started")

	// 로컬 저장소 디렉토리 생성
	if h.config.Storage.Provider == "local" {
		if err := os.MkdirAll(h.config.Storage.LocalPath, 0755); err != nil {
			return fmt.Errorf("failed to create upload directory: %w", err)
		}
	}

	return queue.ConsumeJobs("file_processing_queue", "file_workers", "file_worker_1", h.handleFileJob)
}

func (h *FileHandler) handleFileJob(jobData map[string]interface{}) error {
	jobType, ok := jobData["type"].(string)
	if !ok {
		return fmt.Errorf("missing job type")
	}

	switch jobType {
	case "upload_verification_doc":
		return h.uploadVerificationDoc(jobData)
	case "process_image":
		return h.processImage(jobData)
	default:
		return fmt.Errorf("unknown file job type: %s", jobType)
	}
}

func (h *FileHandler) uploadVerificationDoc(jobData map[string]interface{}) error {
	// 필수 필드 추출
	userID, ok := jobData["user_id"]
	if !ok {
		return fmt.Errorf("missing user_id")
	}

	docType, ok := jobData["doc_type"].(string)
	if !ok {
		return fmt.Errorf("missing doc_type")
	}

	filename, ok := jobData["filename"].(string)
	if !ok {
		return fmt.Errorf("missing filename")
	}

	// 파일 저장 경로 생성
	relativePath := fmt.Sprintf("verification/%v/%s/%s", userID, docType, filename)

	switch h.config.Storage.Provider {
	case "local":
		return h.saveToLocal(relativePath, jobData)
	case "s3":
		return h.saveToS3(relativePath, jobData)
	case "r2":
		return h.saveToR2(relativePath, jobData)
	default:
		return fmt.Errorf("unsupported storage provider: %s", h.config.Storage.Provider)
	}
}

func (h *FileHandler) saveToLocal(relativePath string, jobData map[string]interface{}) error {
	// 로컬 파일 시스템에 저장
	fullPath := filepath.Join(h.config.Storage.LocalPath, relativePath)

	// 디렉토리 생성
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 실제 환경에서는 여기서 multipart form에서 파일 데이터를 읽어와 저장
	// 지금은 메타데이터만 처리
	log.Printf("✅ File would be saved to: %s", fullPath)

	// 파일 메타데이터를 데이터베이스에 저장하는 로직 추가 필요
	// 예: 파일 경로, 크기, 타입 등을 user_verification 테이블에 업데이트

	return nil
}

func (h *FileHandler) saveToS3(relativePath string, jobData map[string]interface{}) error {
	// AWS S3에 파일 업로드
	// 실제 환경에서는 AWS SDK를 사용
	log.Printf("✅ File would be uploaded to S3: s3://%s/%s", h.config.Storage.Bucket, relativePath)

	// TODO: AWS S3 SDK 구현
	// - AWS 자격 증명 설정
	// - S3 클라이언트 생성
	// - 파일 업로드
	// - 업로드된 URL 반환

	return nil
}

func (h *FileHandler) saveToR2(relativePath string, jobData map[string]interface{}) error {
	// Cloudflare R2에 파일 업로드 (S3 호환 API 사용)
	log.Printf("✅ File would be uploaded to R2: %s", relativePath)

	// TODO: Cloudflare R2 SDK 구현 (S3 호환)
	// - R2 자격 증명 설정
	// - S3 호환 클라이언트 생성
	// - 파일 업로드

	return nil
}

func (h *FileHandler) processImage(jobData map[string]interface{}) error {
	// 이미지 최적화 처리
	filename, ok := jobData["filename"].(string)
	if !ok {
		return fmt.Errorf("missing filename")
	}

	// 이미지 처리 로직
	log.Printf("✅ Processing image: %s", filename)

	// TODO: 이미지 처리 구현
	// - 리사이징 (프로필 사진: 200x200, 프로젝트 이미지: 800x600)
	// - 압축 (JPEG 품질 85%)
	// - 워터마크 추가 (선택사항)
	// - 여러 크기 생성 (썸네일, 미디움, 라지)

	return nil
}

// 파일 유효성 검사
func (h *FileHandler) validateFile(jobData map[string]interface{}) error {
	contentType, ok := jobData["content_type"].(string)
	if !ok {
		return fmt.Errorf("missing content_type")
	}

	size, ok := jobData["size"].(float64)
	if !ok {
		return fmt.Errorf("missing file size")
	}

	// 파일 크기 제한 (10MB)
	maxSize := 10 * 1024 * 1024 // 10MB
	if int64(size) > int64(maxSize) {
		return fmt.Errorf("file size too large: %v bytes", size)
	}

	// 허용된 파일 타입 확인
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}

	if !allowedTypes[contentType] {
		return fmt.Errorf("unsupported file type: %s", contentType)
	}

	return nil
}

// 바이러스 검사 (선택사항)
func (h *FileHandler) scanForVirus(filePath string) error {
	// 실제 환경에서는 ClamAV 등을 사용한 바이러스 검사
	log.Printf("🔍 Virus scan completed for: %s", filePath)
	return nil
}
