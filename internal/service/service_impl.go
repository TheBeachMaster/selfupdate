package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"slices"
	"strings"

	"com.thebeachmaster/selfupdate/internal/pkg/github"
	"com.thebeachmaster/selfupdate/internal/pkg/updater"
	"com.thebeachmaster/selfupdate/internal/pkg/version"
	githubSDK "github.com/google/go-github/v74/github"
)

type serviceHandler struct {
	updateService  updater.Updater
	githubClient   github.GithubClient
	versionService version.Version
	cmdArgs        string // The commandline args that the app was started with
}

func NewServiceHandler(_cmdArgs string, opts ...string) ServiceHandler {
	var _githubClient github.GithubClient
	if len(opts) > 0 {
		// We only care about the first option - for now
		_githubClient = github.NewGithubClient(github.WithAccessToken(opts[0]))
	} else {
		_githubClient = github.NewGithubClient()
	}
	return &serviceHandler{
		githubClient:   _githubClient,
		updateService:  updater.NewUpdater(),
		versionService: version.NewVersionService(),
		cmdArgs:        _cmdArgs,
	}
}

// CheckAppVersionHandler implements ServiceHandler.
func (s *serviceHandler) CheckAppVersionHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_req := &github.ReleaseRequest{}
		if err := json.NewDecoder(request.Body).Decode(_req); err != nil {
			log.Printf("ERROR:unable to parse user input due to %s", err.Error())
			http.Error(writer, "invalid request data", http.StatusBadRequest)
			return
		}

		_res, err := s.githubClient.GetLatestReleaseAsset(request.Context(), _req)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}

		if len(_res) == 0 {
			http.Error(writer, "no releases found", http.StatusNotFound)
			return
		}

		// Get the checksum file {selfupdate_2025.8.11_checksums.txt}
		n := slices.IndexFunc(_res, func(re *githubSDK.ReleaseAsset) bool {
			return strings.Contains(*re.Name, "checksum")
		})

		// log.Println(_res)

		if n < 0 {
			log.Println("ERROR: no checksum files")
			http.Error(writer, "no checksum file", http.StatusInternalServerError)
			return
		}

		// Check if version is latest

		_checkFileData := _res[n]

		/// split
		chkParts := strings.Split(*_checkFileData.Name, "_")
		if len(chkParts) < 3 {
			log.Printf("ERROR: found %d in %v\n", len(chkParts), chkParts)
			http.Error(writer, "no valid checksum file with version", http.StatusNotFound)
			return
		}
		/// is remote version latest?
		_isLatest := s.versionService.CompareVersions(chkParts[1])
		if !_isLatest {
			log.Printf("ERROR: local version %s newer than remote version %s\n", version.CurrentVersion, chkParts[1])
			http.Error(writer, "no new versions", http.StatusNotFound)
			return
		}
		/// get the version appropriate for this OS
		_archEntries := map[string]string{
			"arm64": "arm64",
			"amd64": "x86_64",
		}
		_o := slices.IndexFunc(_res, func(re *githubSDK.ReleaseAsset) bool {
			if re.BrowserDownloadURL != nil {
				_downloadURL := strings.ToLower(*re.BrowserDownloadURL)
				_currArch, ok := _archEntries[runtime.GOARCH]
				if !ok {
					return false
				}
				return strings.Contains(_downloadURL, fmt.Sprintf("_%s_%s", runtime.GOOS, _currArch))
			}
			return false
		})
		if _o < 0 {
			log.Println("ERROR: no dowloadable files")
			http.Error(writer, "no files available for download file", http.StatusInternalServerError)
			return
		}

		_data := &UpdateCheckResponse{
			Version:     chkParts[1],
			DownloadURL: *_res[_o].BrowserDownloadURL,
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(writer).Encode(_data)

	}
}

// UpdateAppHandler implements ServiceHandler.
func (s *serviceHandler) UpdateAppHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_req := &UpgradeAppVersionRequest{}
		if err := json.NewDecoder(request.Body).Decode(_req); err != nil {
			log.Printf("ERROR:unable to parse user input due to %s\n", err.Error())
			http.Error(writer, "invalid request data", http.StatusBadRequest)
			return
		}
		// Make sure the URL is valid
		if _, err := url.Parse(_req.DownloadURL); err != nil {
			if err := json.NewDecoder(request.Body).Decode(_req); err != nil {
				log.Printf("ERROR:unable to parse URL due to %s\n", err.Error())
				http.Error(writer, "invalid download URL", http.StatusBadRequest)
				return
			}
		}

		// Download
		_downloadPath, err := s.updateService.DownloadFile(_req.DownloadURL)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		// Extract
		_binFilePath, err := s.updateService.ExtractFile(_downloadPath)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		// Update
		go func(m_binFilePath string) {
			if err := s.updateService.UpdateBinary(m_binFilePath, s.cmdArgs); err != nil {
				log.Println(err.Error())
			}
		}(_binFilePath)
		//
		// Bye!!
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(writer, "update in progress\n")
	}
}
