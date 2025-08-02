package service

type UpdateCheckResponse struct {
	DownloadURL string `json:"url"`
	Version     string `json:"version"`
}

type UpgradeAppVersionRequest struct {
	DownloadURL string `json:"url"`
}
