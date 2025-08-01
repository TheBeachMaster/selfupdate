package updater

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/cavaliergopher/grab/v3"
	archiver "github.com/mholt/archives"
)

type updater struct {
}

func NewUpdater() Updater {
	return &updater{}
}

// DownloadFile implements Updater.
func (u *updater) DownloadFile(fileURL string) (filePath string, err error) {
	parsedPath, err := url.Parse(fileURL)
	if err != nil {
		log.Printf("ERROR: unable to parse url due to %s\n", err.Error())
		return "", fmt.Errorf("unable to parse url")
	}

	m_parsedPath := path.Base(parsedPath.Path)
	tmpDst, err := os.MkdirTemp("", "downloads")
	if err != nil {
		log.Printf("ERROR: unable to create temporary directory due to %s \n", err.Error())
		return "", fmt.Errorf("could not create temporary directory")
	}
	fileDestination := path.Join(tmpDst, m_parsedPath)

	client := grab.NewClient()
	client.UserAgent = "selfupdater-" + runtime.GOOS
	req, err := grab.NewRequest(fileDestination, fileURL)
	if err != nil {
		log.Printf("ERROR: could initialize transfer request due to %s \n", err.Error())
		return "", fmt.Errorf("could initialize transfer request")
	}

	// start download
	log.Printf("downloading %s...\n", req.URL())
	resp := client.Do(req)
	log.Printf("  %s\n", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			log.Printf("  transferred %d / %d bytes (%.2f%%)\n", resp.BytesComplete(), resp.Size(), 100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		log.Printf("ERROR: download failed due to %s\n", err.Error())
		return "", fmt.Errorf("download failed")
	}

	log.Printf("download saved to %s \n", resp.Filename)
	return resp.Filename, nil
}

// ExtractFile implements Updater.
func (u *updater) ExtractFile(downloadFilePath string) (extractedPath string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fileReader, err := os.Open(downloadFilePath)
	if err != nil {
		log.Printf("ERROR: could not open file due to %s\n", err.Error())
		return "", fmt.Errorf("unable to open file")
	}
	defer fileReader.Close() //nolint:all

	// Get the archive type
	fileFormat, readerInput, err := archiver.Identify(ctx, downloadFilePath, fileReader)
	if err != nil {
		log.Printf("ERROR: unable to identify file format due to %s\n", err.Error())
		return "", fmt.Errorf("unknown archive format")
	}

	_dirPath := filepath.Dir(downloadFilePath)

	var binaryPath string

	extractorFunc := func(_ctx context.Context, extrFile archiver.FileInfo) error {
		extractionPath := path.Join(_dirPath, extrFile.Name())

		fileInput, err := extrFile.Open()
		if err != nil {
			log.Printf("ERROR: unable to open file %s for reading due to %s\n", extrFile.Name(), err.Error())
			return fmt.Errorf("could not open file")
		}
		defer fileInput.Close() //nolint:all

		destinationDir := path.Dir(extractionPath)
		if err := os.MkdirAll(destinationDir, 0777); err != nil {
			log.Printf("ERROR: unable to create directory %s due to %s\n", destinationDir, err.Error())
			return fmt.Errorf("unable to create directory")
		}

		fileOut, err := os.Create(extractionPath)
		if err != nil {
			log.Printf("ERROR: unable to create %s for reading due to %s\n", extractionPath, err.Error())
			return fmt.Errorf("could not create file")
		}
		defer fileOut.Close() //nolint:all

		sz, err := io.Copy(fileOut, fileInput)
		if err != nil {
			log.Printf("ERROR: unable to copy files due to %s\n", err.Error())
			return fmt.Errorf("could not copy bytes from buffer")
		}

		log.Printf("copied %d bytes\n", sz)

		// We now need to confirm if the file is there

		destRead, err := os.Open(destinationDir)
		if err != nil {
			log.Printf("ERROR: unable to open directory %s for reading due to %s\n", destinationDir, err.Error())
			return fmt.Errorf("could not open directory")
		}
		defer destRead.Close() //nolint:all

		// Bad
		m_destRead, err := destRead.ReadDir(0)
		if err != nil {
			log.Printf("ERROR: unable to traverse extracted module directory %s for reading due to %s\n", destinationDir, err.Error())
			return fmt.Errorf("could not traverse module directory")
		}

		// Bad
		binaryPath = path.Join(destinationDir, m_destRead[0].Name())

		log.Printf("binary extracted to %s\n", binaryPath)

		return nil
	}

	if extractor, ok := fileFormat.(archiver.Extractor); ok {
		if err := extractor.Extract(ctx, readerInput, extractorFunc); err != nil {
			log.Printf("ERROR: unable to extract files due to %s\n", err.Error())
			return "", fmt.Errorf("unable to extract files")
		}
	}

	if err := os.Remove(downloadFilePath); err != nil {
		log.Printf("WARN: unable to delete temporary path %s due to %s\n", downloadFilePath, err.Error())
	}

	return binaryPath, nil
}

// UpdateBinary implements Updater.
func (u *updater) UpdateBinary(binaryPath string, flArgs string) error {
	// Wait for shutdown (maybe????)
	log.Println("INFO: starting service...")
	if err := os.Chdir(filepath.Dir(binaryPath)); err != nil {
		log.Printf("unable to switch to %s due to %s\n", binaryPath, err.Error())
		return fmt.Errorf("unable to switch to path")
	}
	_cmd := exec.Command(binaryPath, flArgs)

	if err := _cmd.Start(); err != nil {
		log.Printf("ERROR: unable to start app due to %s\n", err.Error())
		return fmt.Errorf("app start failed")
	}

	// Release resources
	if err := _cmd.Wait(); err != nil {
		log.Printf("ERROR: unable to release resource due to to %s\n", err.Error())
		return fmt.Errorf("unable to release resources")
	}

	// Do we have an actual process running???
	if _cmd.Process == nil {
		log.Println("ERROR: app failed to start")
		return fmt.Errorf("app did not start")
	}

	log.Println("INFO: app started")

	// Gracefully exit
	os.Exit(0)

	// We will never get here
	return nil
}
