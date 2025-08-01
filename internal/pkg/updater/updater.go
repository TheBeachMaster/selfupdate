package updater

type Updater interface {
	DownloadFile(fileURL string) (filePath string, err error)
	ExtractFile(filePath string) (extractedPath string, err error)
	UpdateBinary(binaryPath string, flArgs string) error
}
