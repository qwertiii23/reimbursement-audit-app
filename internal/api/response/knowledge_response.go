package response

type KnowledgeFileResponse struct {
	ID            string `json:"id"`
	FileName      string `json:"file_name"`
	FilePath      string `json:"file_path"`
	FileType      string `json:"file_type"`
	FileSize      int64  `json:"file_size"`
	Category      string `json:"category"`
	Description   string `json:"description"`
	UploadedBy    string `json:"uploaded_by"`
	UploaderName  string `json:"uploader_name"`
	DownloadCount int    `json:"download_count"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type KnowledgeFilesListResponse struct {
	List  []KnowledgeFileResponse `json:"list"`
	Total int64                `json:"total"`
}
