// api/reports/reports.go
package reports

import (
	"archive/zip"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const reportDir = "/etc/semaphore/export_report"

// FileInfo represents a file in the report directory
type FileInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
}

// GetReportFiles returns a list of files in the report directory
func GetReportFiles(w http.ResponseWriter, r *http.Request) {
	pathParam := r.URL.Query().Get("path")

	if strings.Contains(pathParam, "..") || strings.Contains(pathParam, "\\") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	dirPath := filepath.Join(reportDir, pathParam)

	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.WithError(err).Error("Failed to read report directory")
		http.Error(w, "Failed to read report directory", http.StatusInternalServerError)
		return
	}

	var fileList []FileInfo
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		fileList = append(fileList, FileInfo{
			Name:    file.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
			IsDir:   file.IsDir(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(fileList)
}

// DownloadReportFile serves a single file for download
func DownloadReportFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	if strings.Contains(filename, "..") || strings.Contains(filename, "\\") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(reportDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		http.Error(w, "Not a file", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(filePath))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	http.ServeFile(w, r, filePath)
}

// DownloadReportFilesAsZip serves multiple files as a zip archive
func DownloadReportFilesAsZip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["project_id"]
	
	// Parse the JSON body containing the list of files to download
	var requestBody struct {
		Files []string `json:"files"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if len(requestBody.Files) == 0 {
		http.Error(w, "No files specified", http.StatusBadRequest)
		return
	}
	
	// Create a temporary file for the zip archive
	tempFile, err := os.CreateTemp("", "reports_*.zip")
	if err != nil {
		log.WithError(err).Error("Failed to create temp file")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	tempFilePath := tempFile.Name()
	defer os.Remove(tempFilePath) // Clean up when done
	
	// Create a zip writer
	zipWriter := zip.NewWriter(tempFile)
	
	// Add each file/directory to the zip archive
	for _, filename := range requestBody.Files {
		// Security check
		if strings.Contains(filename, "..") || strings.Contains(filename, "\\") {
			continue
		}
		
		filePath := filepath.Join(reportDir, filename)
		
		// Check if file exists
		fileInfo, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			continue
		}
		
		// If it's a directory, recursively add all files
		if fileInfo.IsDir() {
			if err := addDirectoryToZip(zipWriter, reportDir, filename); err != nil {
				log.WithError(err).Errorf("Failed to add directory %s to zip", filename)
			}
			continue
		}
		
		// Open the file
		file, err := os.Open(filePath)
		if err != nil {
			continue
		}
		
		// Create a zip entry
		zipEntry, err := zipWriter.Create(filename)
		if err != nil {
			file.Close()
			continue
		}
		
		// Copy file content to zip entry
		_, err = io.Copy(zipEntry, file)
		file.Close()
		if err != nil {
			continue
		}
	}
	
	// Close the zip writer
	if err := zipWriter.Close(); err != nil {
		log.WithError(err).Error("Failed to close zip writer")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	// Close the temp file
	tempFile.Close()
	
	// Read the zip file and send to client
	zipFile, err := os.Open(tempFilePath)
	if err != nil {
		log.WithError(err).Error("Failed to open zip file")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer zipFile.Close()
	
	// Set response headers
	timestamp := time.Now().Format("20060102_150405")
	w.Header().Set("Content-Disposition", "attachment; filename=reports_"+projectID+"_"+timestamp+".zip")
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	
	// Send the file
	_, err = io.Copy(w, zipFile)
	if err != nil {
		log.WithError(err).Error("Failed to send zip file")
	}
}

// addDirectoryToZip recursively adds all files in a directory to the zip archive
func addDirectoryToZip(zipWriter *zip.Writer, baseDir, dirPath string) error {
	dirFullPath := filepath.Join(baseDir, dirPath)
	
	files, err := os.ReadDir(dirFullPath)
	if err != nil {
		return err
	}
	
	for _, file := range files {
		fileFullPath := filepath.Join(dirFullPath, file.Name())
		fileRelPath := filepath.Join(dirPath, file.Name())
		
		if file.IsDir() {
			if err := addDirectoryToZip(zipWriter, baseDir, fileRelPath); err != nil {
				return err
			}
			continue
		}
		
		f, err := os.Open(fileFullPath)
		if err != nil {
			return err
		}
		
		zipEntry, err := zipWriter.Create(fileRelPath)
		if err != nil {
			f.Close()
			return err
		}
		
		_, err = io.Copy(zipEntry, f)
		f.Close()
		if err != nil {
			return err
		}
	}
	
	return nil
}