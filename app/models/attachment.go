	package models

	import (
		"time"
		"github.com/google/uuid"
	)

	type Attachment struct {
		ID           uuid.UUID `bson:"_id" json:"id"`
		FileName     string    `bson:"fileName" json:"file_name"`
		FileURL      string    `bson:"fileUrl" json:"file_url"`
		FileType     string    `bson:"fileType" json:"file_type"`
		FileSize     int64     `bson:"fileSize" json:"file_size"`
		UploadedAt   time.Time `bson:"uploadedAt" json:"uploaded_at"`
	}