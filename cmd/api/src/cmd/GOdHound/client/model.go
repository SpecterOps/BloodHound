package client

type FileUploadJob struct {
	ID               int64  `json:"id"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	UserID           string `json:"user_id"`
	UserEmailAddress string `json:"user_email_address"`
	Status           int    `json:"status"`
	StatusMessage    string `json:"status_message"`
	StartTime        string `json:"start_time"`
	EndTime          string `json:"end_time"`
	LastIngest       string `json:"last_ingest"`
	TotalFiles       int    `json:"total_files"`
	FailedFiles      int    `json:"failed_files"`
}
