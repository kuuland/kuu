package kuu

import (
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"mime/multipart"
)

// ParseExcelFromFileHeader
func ParseExcelFromFileHeader(fh *multipart.FileHeader, index int, sheetName ...string) (rows [][]string, err error) {
	var (
		file multipart.File
		f    *excelize.File
		name string
	)

	if len(sheetName) > 0 && sheetName[0] != "" {
		name = sheetName[0]
	}

	// 解析Excel
	if file, err = fh.Open(); err != nil {
		return rows, err
	}
	defer ERROR(file.Close())
	if f, err = excelize.OpenReader(file); err != nil {
		return rows, err
	}
	// 选择工作表
	if name == "" && index > 0 {
		name = f.GetSheetName(index)
	}
	if name == "" {
		name = f.GetSheetName(f.GetActiveSheetIndex())
	}
	if name == "" {
		name = f.GetSheetName(1)
	}
	// 读取行
	rows, err = f.GetRows(name)
	return
}
