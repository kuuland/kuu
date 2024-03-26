package kuu

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/valyala/fasthttp"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/guregu/null.v4"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Time tools

func DiffDays(start, end time.Time) []time.Time {
	var dates []time.Time
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d)
	}
	return dates
}

func DiffDaysString(start, end time.Time) []string {
	return lo.Map[time.Time, string](DiffDays(start, end), func(item time.Time, index int) string {
		return item.Format(time.DateOnly)
	})
}

// Parse tools

func ParseNullFloat(s string) null.Float {
	if v, err := strconv.ParseFloat(s, 64); err != nil {
		return null.NewFloat(0, false)
	} else {
		return null.FloatFrom(v)
	}
}

func ParseNullInt(s string) null.Int {
	if v, err := strconv.Atoi(s); err != nil {
		return null.NewInt(0, false)
	} else {
		return null.IntFrom(int64(v))
	}
}

func ParseNullString(s, defaultValue string) null.String {
	s = strings.TrimSpace(s)
	if s == "" {
		return null.NewString(defaultValue, true)
	} else {
		return null.StringFrom(s)
	}
}

func ParseNullBool(s string) null.Bool {
	if v, err := strconv.ParseBool(s); err != nil {
		return null.NewBool(false, false)
	} else {
		return null.BoolFrom(v)
	}
}

// UUID tools

func NewUUID(upper bool, hyphen bool) string {
	s := uuid.NewString()
	if upper {
		s = strings.ToUpper(s)
	}
	if !hyphen {
		s = strings.ReplaceAll(s, "-", "")
	}
	return s
}

func NewUUIDToken() string {
	return NewUUID(true, false)
}

// JSON tools

func JSONStringify(v any, format ...bool) string {
	var (
		b   []byte
		err error
	)
	if len(format) > 0 && format[0] {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	if err == nil {
		return string(b)
	}
	return ""
}

func JSONParse(s string, v any) error {
	return json.Unmarshal([]byte(s), v)
}

// Encode tools

func EncodeURIComponent(str string) string {
	r := url.QueryEscape(str)
	r = strings.Replace(r, "+", "%20", -1)
	return r
}

func DecodeURIComponent(str string) string {
	if r, err := url.QueryUnescape(str); err == nil {
		return r
	}
	return str
}

// Random tools

func NewRandomNumber(length int) string {
	if length <= 0 {
		return ""
	}

	// 设置随机数生成器的种子
	rand.Seed(time.Now().UnixNano())

	// 第一位不能为0
	firstDigit := rand.Intn(9) + 1 // 1-9 之间的随机数

	// 生成剩余位数的随机数
	randomNumber := fmt.Sprintf("%d", firstDigit)
	for i := 1; i < length; i++ {
		randomNumber += fmt.Sprintf("%d", rand.Intn(10)) // 0-9 之间的随机数
	}

	return randomNumber
}

// Markdown tools

func MarkdownFindAndReplaceURLs(result string, keepText bool) (string, []string) {
	// 定义正则表达式模式，用于匹配标签内容
	regexPattern := `\[(.*?)\]\((.*?)\)`
	// 编译正则表达式
	re := regexp.MustCompile(regexPattern)
	// 查找所有匹配项
	matches := re.FindAllStringSubmatch(result, -1)
	var urls []string
	// 遍历所有匹配项
	for _, match := range matches {
		text := strings.TrimSpace(match[1])    // 获取匹配到的链接文本
		content := strings.TrimSpace(match[2]) // 获取匹配到的链接内容
		if content != "" {
			urls = append(urls, content)
		}
		// 进行文本替换
		if keepText && text != "" {
			result = strings.Replace(result, match[0], text, 1)
		} else {
			result = strings.Replace(result, match[0], "", 1)
		}
	}
	return result, urls
}

// File tools

type FileKind string

const (
	FileKindText  FileKind = "TEXT"
	FileKindImage FileKind = "IMAGE"
	FileKindVideo FileKind = "VIDEO"
	FileKindAudio FileKind = "AUDIO"
	FileKindFile  FileKind = "FILE"
)

type File struct {
	rawFilePath string
	ThumbID     null.String `kuu:"缩略图ID"`
	ThumbPath   null.String `kuu:"缩略图"`
	ThumbURL    null.String `kuu:"缩略图"`
	Duration    null.Float  `kuu:"音/视频时长（秒）"`

	UUID     string      `kuu:"文件唯一ID"`
	FileKind FileKind    `kuu:"文件类型"`
	MineType string      `kuu:"文件Mine-Type"` // 如image/jpeg
	Name     string      `kuu:"文件名称"`
	Path     string      `kuu:"存储路径"`
	Ext      string      `kuu:"文件扩展名"`
	URL      string      `kuu:"下载路径"`
	Size     null.Int    `kuu:"文件大小"`
	Sort     null.Int    `kuu:"排序值"`
	MD5Sum   null.String `kuu:"MD5校验码"`
	Width    null.Int    `kuu:"宽度"`
	Height   null.Int    `kuu:"高度"`

	Clazz     null.String `kuu:"文件分类"`
	OwnerID   null.Int    `kuu:"所属数据ID"`
	OwnerType null.String `kuu:"所属数据类型"`
}

func (f *File) SetMD5Sum() *File {
	v, _ := GetFileMD5Sum(f.rawFilePath)
	f.MD5Sum = null.StringFrom(v)
	return f
}

type FilepathObject struct {
	Raw           string
	FileNameNoExt string
	FullFileName  string
	FileExt       string
	FileDir       string
}

func PathExists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func EnsureDir(p string) error {
	exists, err := PathExists(p)
	if err != nil {
		return err
	}
	if !exists {
		return os.MkdirAll(p, os.ModePerm)
	}
	return nil
}

func ParseFilepath(fileURL string) (fo *FilepathObject) {
	fo = new(FilepathObject)
	fileURL = strings.TrimSpace(fileURL)
	if fileURL == "" {
		return
	}
	// 获取纯文件名（不含后缀）
	fileName := filepath.Base(fileURL)
	fileNameNoExt := fileName[:len(fileName)-len(filepath.Ext(fileName))]

	// 获取完整文件名（含后缀）
	fullFileName := filepath.Base(fileURL)

	// 获取文件后缀
	fileExt := strings.ToLower(filepath.Ext(fileURL))

	// 获取文件路径
	fileDir := filepath.Dir(fileURL)
	return &FilepathObject{
		Raw:           fileURL,
		FileNameNoExt: fileNameNoExt,
		FullFileName:  fullFileName,
		FileExt:       fileExt,
		FileDir:       fileDir,
	}
}

func SaveRemoteFileToLocal(pathBase, urlBase, fileName string, keepNameToPath bool, resp *http.Response) (*File, error) {
	ext := filepath.Ext(fileName)
	fileNameNoExt := fileName[:len(fileName)-len(ext)]
	filePathName := fileName
	if !keepNameToPath {
		filePathName = NewUUID(false, false) + ext
	}
	if err := EnsureDir(pathBase); err != nil {
		return nil, err
	}
	filePath := path.Join(pathBase, filePathName)
	if ff, err := os.Create(filePath); err != nil {
		return nil, err
	} else {
		written, err := io.Copy(ff, resp.Body)
		if err != nil {
			return nil, err
		}
		if err = resp.Body.Close(); err != nil {
			return nil, err
		}
		mtype, err := mimetype.DetectFile(filePath)
		if err != nil {
			return nil, err
		}
		mtypeStr := mtype.String()
		fileKind := FileKindFile
		if strings.HasPrefix(mtypeStr, "text/plain") {
			fileKind = FileKindText
		} else if strings.HasPrefix(mtypeStr, "image") {
			fileKind = FileKindImage
		} else if strings.HasPrefix(mtypeStr, "video") {
			fileKind = FileKindVideo
		} else if strings.HasPrefix(mtypeStr, "audio") {
			fileKind = FileKindAudio
		}
		var fileUrl string
		if strings.HasSuffix(urlBase, "/") {
			fileUrl = urlBase + filePathName
		} else {
			fileUrl = urlBase + "/" + filePathName
		}
		file := File{
			rawFilePath: filePath,
			UUID:        NewUUIDToken(),
			FileKind:    fileKind,
			MineType:    mtypeStr,
			Name:        fileNameNoExt,
			Ext:         ext,
			Path:        filePath,
			URL:         fileUrl,
			Size:        null.IntFrom(written),
			Sort:        null.IntFrom(100),
		}
		return &file, ff.Close()
	}
}

func SaveUploadFileToLocal(fh *multipart.FileHeader, subDirs []string, useRawName bool) (*File, error) {
	dirs := []string{os.Getenv("UPLOAD_PATH_BASE")}
	if len(subDirs) > 0 {
		dirs = append(dirs, subDirs...)
	}
	dirs = append(dirs, time.Now().Format("20060102"))
	pathBase := path.Join(dirs...)
	fileName := fh.Filename
	ext := filepath.Ext(fileName)
	if !useRawName {
		fileName = NewUUID(false, false) + ext
	}
	if err := EnsureDir(pathBase); err != nil {
		return nil, err
	}
	filePath := path.Join(pathBase, fileName)
	if err := fasthttp.SaveMultipartFile(fh, filePath); err != nil {
		return nil, err
	}
	urlPath := []string{os.Getenv("UPLOAD_URL_BASE")}
	urlPath = append(urlPath, dirs[1:]...)
	urlPath = append(urlPath, fileName)
	fileUrl := path.Join(urlPath...)

	mtype, err := mimetype.DetectFile(filePath)
	if err != nil {
		return nil, err
	}
	mtypeStr := mtype.String()
	fileKind := FileKindFile
	if strings.HasPrefix(mtypeStr, "text/plain") {
		fileKind = FileKindText
	} else if strings.HasPrefix(mtypeStr, "image") {
		fileKind = FileKindImage
	} else if strings.HasPrefix(mtypeStr, "video") {
		fileKind = FileKindVideo
	} else if strings.HasPrefix(mtypeStr, "audio") {
		fileKind = FileKindAudio
	}
	fileNameNoExt := fileName[:len(fileName)-len(ext)]
	file := File{
		rawFilePath: filePath,
		UUID:        NewUUIDToken(),
		FileKind:    fileKind,
		MineType:    mtypeStr,
		Name:        fileNameNoExt,
		Ext:         ext,
		Path:        filePath,
		URL:         fileUrl,
		Size:        null.IntFrom(fh.Size),
		Sort:        null.IntFrom(100),
	}
	return &file, nil

}

func GetFileMD5Sum(filePath string) (string, error) {
	// 打开文件
	fd, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	// 使用 md5 包创建一个新的 MD5 散列
	hash := md5.New()

	// 将文件内容写入哈希
	if _, err := io.Copy(hash, fd); err != nil {
		return "", err
	}

	// 计算 MD5 校验和
	hashInBytes := hash.Sum(nil)

	// 将字节转换为十六进制字符串
	md5Checksum := hex.EncodeToString(hashInBytes)

	return md5Checksum, nil
}

func IsImage(filePath string) bool {
	ext := filepath.Ext(filePath)
	fileType := map[string]string{
		".jpg":  "image/jpeg",
		".jp2":  "image/jp2",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".cr2":  "image/x-canon-cr2",
		".tif":  "image/tiff",
		".bmp":  "image/bmp",
		".jxr":  "image/vnd.ms-photo",
		".psd":  "image/vnd.adobe.photoshop",
		".ico":  "image/vnd.microsoft.icon",
		".heif": "image/heif",
		".dwg":  "image/vnd.dwg",
		".exr":  "image/x-exr",
		".avif": "image/avif",
	}[ext]
	return fileType != ""
}

func IsVideo(filePath string) bool {
	ext := filepath.Ext(filePath)
	fileType := map[string]string{
		".mp4":  "video/mp4",
		".m4v":  "video/x-m4v",
		".mkv":  "video/x-matroska",
		".webm": "video/webm",
		".mov":  "video/quicktime",
		".avi":  "video/x-msvideo",
		".wmv":  "video/x-ms-wmv",
		".mpg":  "video/mpeg",
		".flv":  "video/x-flv",
		".3gp":  "video/3gpp",
	}[ext]
	return fileType != ""
}

func IsAudio(filePath string) bool {
	ext := filepath.Ext(filePath)
	fileType := map[string]string{
		".mid":  "audio/midi",
		".mp3":  "audio/mpeg",
		".m4a":  "audio/mp4",
		".ogg":  "audio/ogg",
		".flac": "audio/x-flac",
		".wav":  "audio/x-wav",
		".amr":  "audio/amr",
		".aac":  "audio/aac",
		".aiff": "audio/x-aiff",
	}[ext]
	return fileType != ""
}

// querystring tools

var nameRegex = regexp.MustCompile(`\A[\[\]]*([^\[\]]+)\]*`)
var objectRegex1 = regexp.MustCompile(`^\[\]\[([^\[\]]+)\]$`)
var objectRegex2 = regexp.MustCompile(`^\[\](.+)$`)

func QueryStringify(hash Map) (string, error) {
	return buildNestedQuery(hash, "")
}

func QueryParse(qs string) (Map, error) {
	components := strings.Split(qs, "&")
	params := Map{}

	for _, c := range components {

		tuple := strings.Split(c, "=")
		for i, item := range tuple {
			if unesc, err := url.QueryUnescape(item); err == nil {
				tuple[i] = unesc
			}
		}

		key := ""

		if len(tuple) > 0 {
			key = tuple[0]
		}

		value := any(nil)

		if len(tuple) > 1 {
			value = tuple[1]
		}

		if err := normalizeParams(params, key, value); err != nil {
			return nil, err
		}
	}

	return params, nil
}

func normalizeParams(params Map, key string, value any) error {
	after := ""

	if pos := nameRegex.FindIndex([]byte(key)); len(pos) == 2 {
		after = key[pos[1]:]
	}

	matches := nameRegex.FindStringSubmatch(key)
	if len(matches) < 2 {
		return nil
	}

	k := matches[1]
	if after == "" {
		params[k] = value
		return nil
	}

	if after == "[]" {
		ival, ok := params[k]

		if !ok {
			params[k] = []any{value}
			return nil
		}

		array, ok := ival.([]any)

		if !ok {
			return fmt.Errorf("Expected type '[]any' for key '%s', but got '%T'", k, ival)
		}

		params[k] = append(array, value)
		return nil
	}

	object1Matches := objectRegex1.FindStringSubmatch(after)
	object2Matches := objectRegex2.FindStringSubmatch(after)

	if len(object1Matches) > 1 || len(object2Matches) > 1 {
		childKey := ""

		if len(object1Matches) > 1 {
			childKey = object1Matches[1]
		} else if len(object2Matches) > 1 {
			childKey = object2Matches[1]
		}

		if childKey != "" {
			ival, ok := params[k]

			if !ok {
				params[k] = []any{}
				ival = params[k]
			}

			array, ok := ival.([]any)

			if !ok {
				return fmt.Errorf("Expected type '[]any' for key '%s', but got '%T'", k, ival)
			}

			if length := len(array); length > 0 {
				if hash, ok := array[length-1].(Map); ok {
					if _, ok := hash[childKey]; !ok {
						normalizeParams(hash, childKey, value)
						return nil
					}
				}
			}

			newHash := Map{}
			normalizeParams(newHash, childKey, value)
			params[k] = append(array, newHash)

			return nil
		}
	}

	ival, ok := params[k]

	if !ok {
		params[k] = Map{}
		ival = params[k]
	}

	hash, ok := ival.(Map)

	if !ok {
		return fmt.Errorf("Expected type 'map[string]any' for key '%s', but got '%T'", k, ival)
	}

	if err := normalizeParams(hash, after, value); err != nil {
		return err
	}

	return nil
}

func buildNestedQuery(value any, prefix string) (string, error) {
	components := ""

	switch vv := value.(type) {
	case []any:
		for i, v := range vv {
			component, err := buildNestedQuery(v, prefix+"[]")

			if err != nil {
				return "", err
			}

			components += component

			if i < len(vv)-1 {
				components += "&"
			}
		}

	case map[string]any:
		length := len(vv)

		for k, v := range vv {
			childPrefix := ""

			if prefix != "" {
				childPrefix = prefix + "[" + url.QueryEscape(k) + "]"
			} else {
				childPrefix = url.QueryEscape(k)
			}

			component, err := buildNestedQuery(v, childPrefix)

			if err != nil {
				return "", err
			}

			components += component
			length -= 1

			if length > 0 {
				components += "&"
			}
		}

	case string:
		if prefix == "" {
			return "", fmt.Errorf("value must be a map[string]any")
		}

		components += prefix + "=" + url.QueryEscape(vv)

	default:
		components += prefix
	}

	return components, nil
}

// snowflake tools

var snowflakeNode *snowflake.Node

func init() {
	var err error
	snowflakeNode, err = snowflake.NewNode(rand.Int63n(1024))
	if err != nil {
		panic(fmt.Errorf("snowflake init failed: %w", err))
	}
}

func NextSnowflakeID() snowflake.ID {
	return snowflakeNode.Generate()
}

func NextSnowflakeIntID() int64 {
	return NextSnowflakeID().Int64()
}

func NextSnowflakeStringID() string {
	return NextSnowflakeID().String()
}

// bcrypt tools

func GenPassword(inputPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(inputPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	outputPassword := string(hash)
	return outputPassword, err
}

func CheckPassword(inputPassword, targetPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(targetPassword), []byte(inputPassword))
	return err == nil
}

func MD5Str(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
