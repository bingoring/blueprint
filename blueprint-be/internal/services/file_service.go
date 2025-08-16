package services

import (
	"crypto/rand"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// FileService 파일 업로드 및 관리 서비스
type FileService struct {
	uploadPath string
	baseURL    string
}

// NewFileService 생성자
func NewFileService(uploadPath, baseURL string) *FileService {
	// 업로드 디렉토리 생성
	os.MkdirAll(uploadPath, 0755)
	
	return &FileService{
		uploadPath: uploadPath,
		baseURL:    baseURL,
	}
}

// UploadFile 파일 업로드
func (s *FileService) UploadFile(file multipart.File, header *multipart.FileHeader, category string) (string, error) {
	// 파일 확장자 추출
	ext := filepath.Ext(header.Filename)
	
	// 고유한 파일명 생성
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	filename := fmt.Sprintf("%x_%d%s", randBytes, time.Now().Unix(), ext)
	
	// 카테고리별 디렉토리 생성
	categoryPath := filepath.Join(s.uploadPath, category)
	os.MkdirAll(categoryPath, 0755)
	
	// 파일 경로
	filePath := filepath.Join(categoryPath, filename)
	
	// 파일 저장
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("파일 생성 실패: %w", err)
	}
	defer dst.Close()
	
	// 파일 내용 복사
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("파일 저장 실패: %w", err)
	}
	
	// 접근 가능한 URL 반환
	fileURL := fmt.Sprintf("%s/%s/%s", s.baseURL, category, filename)
	return fileURL, nil
}

// DeleteFile 파일 삭제
func (s *FileService) DeleteFile(filePath string) error {
	return os.Remove(filePath)
}

// GetFileInfo 파일 정보 조회
func (s *FileService) GetFileInfo(filePath string) (os.FileInfo, error) {
	return os.Stat(filePath)
}