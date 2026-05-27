// api/reports/reports.go
package reports

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const reportDir = "/opt/semaphore/playbooks/nornir/export_report/"

// FileInfo represents a file in the report directory
type FileInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
}

// GetReportFiles returns a list of files in the report directory
func GetReportFiles(w http.ResponseWriter, r *http.Request) {
	// 获取路径参数
	pathParam := r.URL.Query().Get("path")

	// 安全检查：防止路径遍历
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

// DownloadReportFile serves a file for download
func DownloadReportFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// 安全检查：防止路径遍历攻击
	if strings.Contains(filename, "..") || strings.Contains(filename, "\\") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(reportDir, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 检查是否为文件（不是目录）
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		http.Error(w, "Not a file", http.StatusBadRequest)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(filePath))
	w.Header().Set("Content-Type", "application/octet-stream")

	// 读取并返回文件内容
	http.ServeFile(w, r, filePath)
}
