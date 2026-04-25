package upload

// Description 定义视频描述参数。
type Description struct {
	Description *string `form:"description" json:"description"`
}

// Tags 定义视频标签参数。
type Tags struct {
	Tags *string `form:"tags" json:"tags"`
}

// PrivateStatus 定义视频可见性状态参数。
type PrivateStatus struct {
	PrivateStatus *float64 `form:"private_status" json:"private_status" binding:"required,numeric"`
}

// UploadID 定义上传任务 ID 参数。
type UploadID struct {
	UploadID *string `form:"upload_id" json:"upload_id" binding:"required"`
}

// OptionalUploadID 定义可选的上传任务 ID 参数。
type OptionalUploadID struct {
	UploadID *string `form:"upload_id" json:"upload_id"`
}

// FileName 定义上传文件名参数。
type FileName struct {
	FileName *string `form:"file_name" json:"file_name" binding:"required"`
}

// FileSize 定义上传文件大小参数。
type FileSize struct {
	FileSize *float64 `form:"file_size" json:"file_size" binding:"required,numeric"`
}

// ChunkSize 定义分片大小参数。
type ChunkSize struct {
	ChunkSize *float64 `form:"chunk_size" json:"chunk_size" binding:"required,numeric"`
}

// TotalChunks 定义总分片数参数。
type TotalChunks struct {
	TotalChunks *float64 `form:"total_chunks" json:"total_chunks" binding:"required,numeric"`
}

// ChunkIndex 定义当前分片索引参数。
type ChunkIndex struct {
	ChunkIndex *float64 `form:"chunk_index" json:"chunk_index" binding:"required,numeric"`
}

// ChunkHash 定义分片摘要参数。
type ChunkHash struct {
	ChunkHash *string `form:"chunk_hash" json:"chunk_hash" binding:"required"`
}

// ContentType 定义上传内容类型参数。
type ContentType struct {
	ContentType *string `form:"content_type" json:"content_type" binding:"required"`
}

// TaskID 定义后台上传处理任务 ID 参数。
type TaskID struct {
	TaskID *string `form:"task_id" json:"task_id" binding:"required"`
}
