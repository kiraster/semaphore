package inventory

import (
	"encoding/json"
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

func GetXLSXInventory(w http.ResponseWriter, r *http.Request) {
	var files []InventoryFile
	
	// 从配置读取路径，如果配置不存在则使用默认值
	ansiblePath := "/etc/semaphore/inventory_ansible.xlsx"
	nornirPath := "/etc/semaphore/inventory_nornir.xlsx"

	if util.Config.Inventory != nil {
		if util.Config.Inventory.AnsiblePath != "" {
			ansiblePath = util.Config.Inventory.AnsiblePath
		}
		if util.Config.Inventory.NornirPath != "" {
			nornirPath = util.Config.Inventory.NornirPath
		}
	}

	if ansibleFile := readInventoryFile(ansiblePath, "ansible"); ansibleFile.Name != "" {
		files = append(files, ansibleFile)
	}

	if nornirFile := readInventoryFile(nornirPath, "nornir"); nornirFile.Name != "" {
		files = append(files, nornirFile)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(files)
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
