package upload

type Description struct {
	Description *string `form:"description" json:"description"`
}

type Tags struct {
	Tags *string `form:"tags" json:"tags"`
}

type PrivateStatus struct {
	PrivateStatus *float64 `form:"private_status" json:"private_status" binding:"required,numeric"`
}

type UploadID struct {
	UploadID *string `form:"upload_id" json:"upload_id" binding:"required"`
}

type FileName struct {
	FileName *string `form:"file_name" json:"file_name" binding:"required"`
}

type FileSize struct {
	FileSize *float64 `form:"file_size" json:"file_size" binding:"required,numeric"`
}

type ChunkSize struct {
	ChunkSize *float64 `form:"chunk_size" json:"chunk_size" binding:"required,numeric"`
}

type TotalChunks struct {
	TotalChunks *float64 `form:"total_chunks" json:"total_chunks" binding:"required,numeric"`
}

type ChunkIndex struct {
	ChunkIndex *float64 `form:"chunk_index" json:"chunk_index" binding:"required,numeric"`
}

type ChunkHash struct {
	ChunkHash *string `form:"chunk_hash" json:"chunk_hash" binding:"required"`
}

type ContentType struct {
	ContentType *string `form:"content_type" json:"content_type" binding:"required"`
}
