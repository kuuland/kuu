package kuu

import (
	"errors"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"math"
	"mime/multipart"
	"regexp"
	"strconv"
	"time"
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

// timeLocationUTC defined the UTC time location.
var timeLocationUTC, _ = time.LoadLocation("UTC")
var shotDateReg, _ = regexp.Compile(`^\d{2}-\d{2}-\d{2}$`)
var dateDateReg, _ = regexp.Compile(`^\d{4}-\d{2}-\d{2}$`)
var numberReg, _ = regexp.Compile(`^\d+$`)

// timeToUTCTime provides a function to convert time to UTC time.
func timeToUTCTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), timeLocationUTC)
}

// timeToExcelTime provides a function to convert time to Excel time.
func TimeToExcelTime(t time.Time) float64 {
	// TODO in future this should probably also handle date1904 and like TimeFromExcelTime
	var excelTime float64
	var deltaDays int64
	excelTime = 0
	deltaDays = 290 * 364
	// check if UnixNano would be out of int64 range
	for t.Unix() > deltaDays*24*60*60 {
		// reduce by aprox. 290 years, which is max for int64 nanoseconds
		delta := time.Duration(deltaDays) * 24 * time.Hour
		excelTime = excelTime + float64(deltaDays)
		t = t.Add(-delta)
	}
	// finally add remainder of UnixNano to keep nano precision
	// and 25569 which is days between 1900 and 1970
	return excelTime + float64(t.UnixNano())/8.64e13 + 25569.0
}

// shiftJulianToNoon provides a function to process julian date to noon.
func shiftJulianToNoon(julianDays, julianFraction float64) (float64, float64) {
	switch {
	case -0.5 < julianFraction && julianFraction < 0.5:
		julianFraction += 0.5
	case julianFraction >= 0.5:
		julianDays++
		julianFraction -= 0.5
	case julianFraction <= -0.5:
		julianDays--
		julianFraction += 1.5
	}
	return julianDays, julianFraction
}

// fractionOfADay provides a function to return the integer values for hour,
// minutes, seconds and nanoseconds that comprised a given fraction of a day.
// values would round to 1 us.
func fractionOfADay(fraction float64) (hours, minutes, seconds, nanoseconds int) {

	const (
		c1us  = 1e3
		c1s   = 1e9
		c1day = 24 * 60 * 60 * c1s
	)

	frac := int64(c1day*fraction + c1us/2)
	nanoseconds = int((frac%c1s)/c1us) * c1us
	frac /= c1s
	seconds = int(frac % 60)
	frac /= 60
	minutes = int(frac % 60)
	hours = int(frac / 60)
	return
}

// julianDateToGregorianTime provides a function to convert julian date to
// gregorian time.
func julianDateToGregorianTime(part1, part2 float64) time.Time {
	part1I, part1F := math.Modf(part1)
	part2I, part2F := math.Modf(part2)
	julianDays := part1I + part2I
	julianFraction := part1F + part2F
	julianDays, julianFraction = shiftJulianToNoon(julianDays, julianFraction)
	day, month, year := doTheFliegelAndVanFlandernAlgorithm(int(julianDays))
	hours, minutes, seconds, nanoseconds := fractionOfADay(julianFraction)
	return time.Date(year, time.Month(month), day, hours, minutes, seconds, nanoseconds, time.UTC)
}

// doTheFliegelAndVanFlandernAlgorithm; By this point generations of
// programmers have repeated the algorithm sent to the editor of
// "Communications of the ACM" in 1968 (published in CACM, volume 11, number
// 10, October 1968, p.657). None of those programmers seems to have found it
// necessary to explain the constants or variable names set out by Henry F.
// Fliegel and Thomas C. Van Flandern.  Maybe one day I'll buy that jounal and
// expand an explanation here - that day is not today.
func doTheFliegelAndVanFlandernAlgorithm(jd int) (day, month, year int) {
	l := jd + 68569
	n := (4 * l) / 146097
	l = l - (146097*n+3)/4
	i := (4000 * (l + 1)) / 1461001
	l = l - (1461*i)/4 + 31
	j := (80 * l) / 2447
	d := l - (2447*j)/80
	l = j / 11
	m := j + 2 - (12 * l)
	y := 100*(n-49) + i + l
	return d, m, y
}

// timeFromExcelTime provides a function to convert an excelTime
// representation (stored as a floating point number) to a time.Time.
func TimeFromExcelTime(excelTime float64, date1904 bool) time.Time {
	const MDD int64 = 106750 // Max time.Duration Days, aprox. 290 years
	var date time.Time
	var intPart = int64(excelTime)
	// Excel uses Julian dates prior to March 1st 1900, and Gregorian
	// thereafter.
	if intPart <= 61 {
		const OFFSET1900 = 15018.0
		const OFFSET1904 = 16480.0
		const MJD0 float64 = 2400000.5
		var date time.Time
		if date1904 {
			date = julianDateToGregorianTime(MJD0, excelTime+OFFSET1904)
		} else {
			date = julianDateToGregorianTime(MJD0, excelTime+OFFSET1900)
		}
		return date
	}
	var floatPart = excelTime - float64(intPart)
	var dayNanoSeconds float64 = 24 * 60 * 60 * 1000 * 1000 * 1000
	if date1904 {
		date = time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		date = time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	}

	// Duration is limited to aprox. 290 years
	for intPart > MDD {
		durationDays := time.Duration(MDD) * time.Hour * 24
		date = date.Add(durationDays)
		intPart = intPart - MDD
	}
	durationDays := time.Duration(intPart) * time.Hour * 24
	durationPart := time.Duration(dayNanoSeconds * floatPart)
	return date.Add(durationDays).Add(durationPart)
}

func ParseExcelDate(text string) (*time.Time, error) {
	if shotDateReg.MatchString(text) {
		t, err := time.Parse("01-02-06", text)
		if err != nil {
			return nil, err
		}
		return &t, nil
	}
	if dateDateReg.MatchString(text) {
		t, err := time.Parse("2006-01-02", text)
		if err != nil {
			return nil, err
		}
		return &t, nil
	}
	if numberReg.MatchString(text) {
		float, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return nil, err
		}
		t := TimeFromExcelTime(float, false)
		return &t, nil
	}
	return nil, errors.New("date format error")
}

type Header struct {
	Label string // header column name
	Field string // fieldname
	Index int    // index
}

type Headers []Header

// get excel column name，support 26*26=676 columns。  e.g: A,B,C,D,E org AA, AB
func (h *Header) GetExcelCol() string {
	r := h.Index / 26
	// 第一轮，只有单字母坐标
	if r == 0 {
		return string(rune(65 + h.Index%65))
	}
	// 从第二轮起，坐标为双字母
	i := h.Index - r*26
	return string(rune(65+r-1)) + string(rune(65+i%65))
}

// get excel axis e.g: A1, A2, A4
func (h *Header) GetExcelAxis(rowIndex int) string {
	return h.GetExcelCol() + strconv.Itoa(rowIndex+1)
}

// get filename by index
func (cols Headers) GetField(index int) string {
	for _, col := range cols {
		if col.Index == index {
			return col.Field
		}
	}
	return ""
}

// get index by fieldname
func (cols Headers) GetIndex(field string) int {
	for _, col := range cols {
		if col.Field == field {
			return col.Index
		}
	}
	return -1
}

// get column by fieldname
func (cols Headers) GetColByField(field string) *Header {
	for _, col := range cols {
		if col.Field == field {
			return &col
		}
	}
	return nil
}

// get column by index
func (cols Headers) GetColByIndex(index int) *Header {
	for _, col := range cols {
		if col.Index == index {
			return &col
		}
	}
	return nil
}

// only set header label
func SetHeader(file *excelize.File, sheet string, headers Headers) {
	for _, header := range headers {
		column := header.GetExcelAxis(0)
		file.SetCellValue(sheet, column, header.Label)
	}
}

// set header and data row
func SetData(file *excelize.File, sheet string, headers Headers, list []map[string]interface{}) {
	// 设置表头
	SetHeader(file, sheet, headers)
	// 设置数据
	for i, row := range list {
		for _, header := range headers {
			column := header.GetExcelAxis(i + 1)
			file.SetCellValue(sheet, column, row[header.Field])
		}
	}
}

// read data
func ReadData(file *excelize.File, sheet string, headers Headers) ([]map[string]string, error) {
	rows, err := file.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	var list []map[string]string
	for rowIndex, row := range rows {
		// 默认为第一行是表头， 跳过
		if rowIndex == 0 {
			continue
		}
		m := map[string]string{}
		for i, v := range row {
			col := headers.GetColByIndex(i)
			if col == nil {
				continue
			}
			m[col.Field] = v
		}
		list = append(list, m)
	}
	return list, nil
}
