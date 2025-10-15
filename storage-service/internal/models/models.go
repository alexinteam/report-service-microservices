package models

import (
	"time"

	"gorm.io/gorm"
)

type File struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Path        string         `json:"path" gorm:"not null"`
	Size        int64          `json:"size"`
	MimeType    string         `json:"mime_type"`
	Hash        string         `json:"hash"` // MD5 хеш файла
	Description string         `json:"description"`
	IsPublic    bool           `json:"is_public" gorm:"default:false"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (File) TableName() string {
	return "files"
}

type FileUploadRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

type FileUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

type FileResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	MimeType    string    `json:"mime_type"`
	Hash        string    `json:"hash"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (f *File) ToResponse() FileResponse {
	return FileResponse{
		ID:          f.ID,
		Name:        f.Name,
		Path:        f.Path,
		Size:        f.Size,
		MimeType:    f.MimeType,
		Hash:        f.Hash,
		Description: f.Description,
		IsPublic:    f.IsPublic,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
}

type FilesResponse struct {
	Files []FileResponse `json:"files"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

type FileUploadResponse struct {
	File    FileResponse `json:"file"`
	Message string       `json:"message"`
}

type FileDownloadResponse struct {
	File    FileResponse `json:"file"`
	Content []byte       `json:"content"`
}

type StorageStatsResponse struct {
	TotalFiles   int64 `json:"total_files"`
	TotalSize    int64 `json:"total_size"`
	PublicFiles  int64 `json:"public_files"`
	PrivateFiles int64 `json:"private_files"`
	AverageSize  int64 `json:"average_size"`
}
