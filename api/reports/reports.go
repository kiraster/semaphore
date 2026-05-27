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
	files, err := os.ReadDir(reportDir)
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
	// 使用 json.NewEncoder 或类似方式输出 JSON
	_ = json.NewEncoder(w).Encode(fileList)
}

// DownloadReportFile serves a file for download
func DownloadReportFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// 安全检查：防止路径遍历攻击
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(reportDir, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "application/octet-stream")

	// 读取并返回文件内容
	http.ServeFile(w, r, filePath)
}
