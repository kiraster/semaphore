package inventory

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
	log "github.com/sirupsen/logrus"
	"github.com/semaphoreui/semaphore/util" 
)

type SheetData struct {
	Name    string              `json:"name"`
	Headers []string            `json:"headers"`
	Rows    []map[string]string `json:"rows"`
}

type InventoryFile struct {
	Name   string     `json:"name"`
	Source string     `json:"source"`
	Sheets []SheetData `json:"sheets"`
}

// 添加密码脱敏函数
func maskPassword(password string) string {
    return "******"
}

func GetXLSXInventory(w http.ResponseWriter, r *http.Request) {
	var files []InventoryFile
	
	// 1. 从环境变量获取路径列表（逗号分隔）
	envPaths := os.Getenv("SEMAPHORE_INVENTORY_PATHS")
	var paths []string
	
	if envPaths != "" {
		for _, path := range strings.Split(envPaths, ",") {
			path = strings.TrimSpace(path)
			if path != "" {
				paths = append(paths, path)
			}
		}
	}
	
	// 2. 如果配置文件存在，覆盖路径列表
	if util.Config.Inventory != nil && len(util.Config.Inventory.Paths) > 0 {
		paths = util.Config.Inventory.Paths
	}
	
	// 3. 如果都没有配置，返回错误
	if len(paths) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "没有配置XLSX清单文件路径",
			"hint":  "请设置环境变量 SEMAPHORE_INVENTORY_PATHS",
		})
		return
	}
	
	// 4. 遍历所有路径，收集 xlsx 文件（带环境信息）
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			log.Warnf("Path not found: %s", path)
			continue
		}
		
		if info.IsDir() {
			// 获取环境名称（从路径提取）
			envName := strings.ToUpper(filepath.Base(path))
			
			// 扫描目录下所有 xlsx 文件
			dirFiles, err := filepath.Glob(filepath.Join(path, "*.xlsx"))
			if err != nil {
				log.WithError(err).Warnf("Failed to scan directory: %s", path)
				continue
			}
			
			for _, filePath := range dirFiles {
				fileName := filepath.Base(filePath)
				
				// 判断来源
				source := "CUSTOM"
				if strings.Contains(strings.ToLower(fileName), "ansible") {
					source = "ANSIBLE"
				} else if strings.Contains(strings.ToLower(fileName), "nornir") {
					source = "NORNIR"
				}
				
				// 构建显示名称：SOURCE(ENV)
				displayName := fmt.Sprintf("%s(%s)", source, envName)
				
				// 读取文件内容
				inventoryFile := readInventoryFileWithName(filePath, displayName, strings.ToLower(source))
				if inventoryFile.Name != "" {
					files = append(files, inventoryFile)
				}
			}
		} else {
			// 是文件，直接处理
			if strings.HasSuffix(strings.ToLower(path), ".xlsx") {
				fileName := filepath.Base(path)
				dirName := filepath.Base(filepath.Dir(path))
				
				// 判断来源
				source := "CUSTOM"
				if strings.Contains(strings.ToLower(fileName), "ansible") {
					source = "ANSIBLE"
				} else if strings.Contains(strings.ToLower(fileName), "nornir") {
					source = "NORNIR"
				}
				
				// 构建显示名称：SOURCE(ENV)
				envName := "LOCAL"
				if isEnvDirectory(dirName) {
					envName = strings.ToUpper(dirName)
				}
				displayName := fmt.Sprintf("%s(%s)", source, envName)
				
				inventoryFile := readInventoryFileWithName(path, displayName, strings.ToLower(source))
				if inventoryFile.Name != "" {
					files = append(files, inventoryFile)
				}
			}
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(files)
}

// readInventoryFileWithName 读取文件并使用指定的显示名称
func readInventoryFileWithName(filePath, displayName, source string) InventoryFile {
	result := InventoryFile{
		Name:   displayName,
		Source: source,
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Warnf("Inventory file not found: %s", filePath)
		return result
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.WithError(err).Warnf("Failed to open file: %s", filePath)
		return result
	}
	defer f.Close()

	sheets := f.GetSheetList()
	for _, sheetName := range sheets {
		rows, _ := f.GetRows(sheetName)
		if len(rows) == 0 {
			continue
		}

		headers := rows[0]
		dataRows := []map[string]string{}

		for _, row := range rows[1:] {
			if isEmptyRow(row) {
				continue
			}
			rowMap := make(map[string]string)
			for i, header := range headers {
				if i < len(row) {
					rowMap[strings.TrimSpace(header)] = strings.TrimSpace(row[i])
				}
			}
			dataRows = append(dataRows, rowMap)
		}

		result.Sheets = append(result.Sheets, SheetData{
			Name:    sheetName,
			Headers: headers,
			Rows:    dataRows,
		})
	}

	// 扩展脱敏字段列表
    passwordFields := []string{"ansible_password", "password", "ssh_password", "winrm_password", "secret", "pass"}
    
    // 对所有密码字段脱敏
    for sheetIdx := range result.Sheets {
        for rowIdx := range result.Sheets[sheetIdx].Rows {
            for _, field := range passwordFields {
                if _, exists := result.Sheets[sheetIdx].Rows[rowIdx][field]; exists {
                    result.Sheets[sheetIdx].Rows[rowIdx][field] = maskPassword("")
                }
            }
        }
    }
    
	return result
}

// isEnvDirectory 判断目录名是否是环境名称
func isEnvDirectory(name string) bool {
	envNames := []string{"dev", "prod", "test", "development", "production", "staging"}
	for _, env := range envNames {
		if strings.EqualFold(name, env) {
			return true
		}
	}
	return false
}


func readInventoryFile(filePath, source string) InventoryFile {
	result := InventoryFile{
		Name:   filepath.Base(filePath),
		Source: source,
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Warnf("Inventory file not found: %s", filePath)
		return result
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.WithError(err).Warnf("Failed to open file: %s", filePath)
		return result
	}
	defer f.Close()

	sheets := f.GetSheetList()
	for _, sheetName := range sheets {
		rows, _ := f.GetRows(sheetName)
		if len(rows) == 0 {
			continue
		}

		headers := rows[0]
		dataRows := []map[string]string{}

		for _, row := range rows[1:] {
			if isEmptyRow(row) {
				continue
			}
			rowMap := make(map[string]string)
			for i, header := range headers {
				if i < len(row) {
					rowMap[strings.TrimSpace(header)] = strings.TrimSpace(row[i])
				}
			}
			dataRows = append(dataRows, rowMap)
		}

		result.Sheets = append(result.Sheets, SheetData{
			Name:    sheetName,
			Headers: headers,
			Rows:    dataRows,
		})
	}

	return result
}

func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}