package request

type UploadFileRequest struct {
	FileName     string `json:"file_name" binding:"required"`
	FilePath     string `json:"file_path" binding:"required"`
	FileType     string `json:"file_type" binding:"required"`
	FileSize     int64  `json:"file_size"`
	Category     string `json:"category"`
	Description  string `json:"description"`
	UploadedBy   string `json:"uploaded_by" binding:"required"`
	UploaderName string `json:"uploader_name" binding:"required"`
}

type UpdateFileRequest struct {
	ID          string `json:"id" binding:"required"`
	FileName    string `json:"file_name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type GetFilesRequest struct {
	Category string `form:"category"`
	Status   string `form:"status"`
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"page_size" binding:"required,min=1"`
}

type DeleteFileRequest struct {
	ID string `uri:"id" binding:"required"`
}
