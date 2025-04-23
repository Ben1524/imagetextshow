package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

// 图片元数据结构（新增 categoryId）
type ImageMetadata struct {
	PatientID  string `json:"patientId"`  // 病人ID
	CategoryID string `json:"categoryId"` // 类别ID（1-7）
	ImageID    string `json:"imageId"`    // 图片ID（如 img_001）
	ImageData  string `json:"imageData"`  // 图片Base64数据
	FileName   string `json:"fileName"`   // 原始文件名
}

type Images struct {
	Url        string `json:"url"`        // 图片URL
	CategoryID string `json:"categoryId"` // 类别ID
	PatientId  string `json:"patientId"`  // 病人ID
	ImageId    string `json:"imageId"`    // 图片ID
}

type PatientSelection struct {
	TextDescription string   `json:"textDescription"` // 文本描述
	Images          []Images `json:"images"`          // 图片列表
}

// 图片文件名正则（假设格式：img_(\d+)\.(jpg|png|gif)）
var imageRegex = regexp.MustCompile(`img_(\d+)\.(jpg|png|gif)$`)

// 允许的类别目录（1-7）
var allowedCategories = map[string]bool{
	"category_1": true,
	"category_2": true,
	"category_3": true,
	"category_4": true,
	"category_5": true,
	"category_6": true,
	"category_7": true,
}

// 图片根目录
const imageRoot = "./patient_images"

// WebSocket 升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // 允许跨域
}

// 客户端订阅的病人ID
var clientSubscriptions = make(map[*websocket.Conn]string)
var subMutex sync.Mutex

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/ws", handleWebSocket)

	go watchDirectoryChanges()
	log.Println("服务器启动，监听 :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket 升级失败:", err)
		return
	}
	defer func() {
		subMutex.Lock()
		delete(clientSubscriptions, conn)
		subMutex.Unlock()
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("读取消息失败:", err)
			break
		}

		patientSelection := &PatientSelection{}
		if err := json.Unmarshal(msg, patientSelection); err != nil {

		} else {
			id := patientSelection.Images[0].PatientId
			// 保存到json文件中
			path := fmt.Sprintf("patient_%s_selection.json", id)
			jsonFile, err := os.Create(path)
			if err != nil {
				log.Println("创建文件失败:", err)
				return
			}
			defer jsonFile.Close()
			encoder := json.NewEncoder(jsonFile)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(patientSelection); err != nil {
				log.Println("写入文件失败:", err)
				return
			}
			log.Println("JSON 文件已保存:", "patient_selection.json")
			continue
		}

		patientID := string(msg)
		log.Printf("订阅病人: %s", patientID)

		subMutex.Lock()
		clientSubscriptions[conn] = patientID
		subMutex.Unlock()

		// 首次加载该病人的所有图片
		go sendAllImages(conn, patientID)
	}
}

// 发送病人所有类别和图片
func sendAllImages(conn *websocket.Conn, patientID string) {
	patientDir := filepath.Join(imageRoot, patientID)
	if !isDirExists(patientDir) {
		log.Printf("病人目录不存在: %s", patientDir)
		return
	}

	// 遍历所有类别目录
	filepath.WalkDir(patientDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		// 解析类别ID和图片ID
		categoryID, imageMeta, err := parseImagePath(path, patientID)
		if err != nil {
			log.Printf("解析失败: %s, 错误: %v", path, err)
			return nil
		}

		// 发送带类别信息的图片数据
		if err := conn.WriteJSON(ImageMetadata{
			PatientID:  patientID,
			CategoryID: categoryID,
			ImageID:    imageMeta.ImageID,
			ImageData:  imageMeta.ImageData,
			FileName:   imageMeta.FileName,
		}); err != nil {
			log.Println("发送失败:", err)
			return filepath.SkipDir
		}

		time.Sleep(10 * time.Millisecond)
		return nil
	})
}

// 解析图片路径获取元数据
func parseImagePath(fullPath, patientID string) (string, *ImageMetadata, error) {
	// 路径结构: patient_images/{patientID}/category_{n}/img_{id}.ext
	parts := strings.Split(fullPath, "\\")
	if len(parts) < 4 {
		return "", nil, fmt.Errorf("路径格式错误")
	}

	categoryDir := parts[len(parts)-2]
	if !allowedCategories[categoryDir] {
		return "", nil, fmt.Errorf("无效类别: %s", categoryDir)
	}
	categoryID := strings.Split(categoryDir, "_")[1] // 提取类别ID（如 category_1 -> 1）

	filename := parts[len(parts)-1]
	matches := imageRegex.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return "", nil, fmt.Errorf("图片文件名格式错误")
	}

	imgData, err := os.ReadFile(fullPath)
	if err != nil {
		return "", nil, err
	}

	return categoryID, &ImageMetadata{
		ImageID:   matches[1],
		FileName:  filename,
		ImageData: base64.StdEncoding.EncodeToString(imgData),
	}, nil
}

// 监控目录变化
func watchDirectoryChanges() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("监控器创建失败:", err)
	}
	defer watcher.Close()

	// 监控所有病人目录
	filepath.WalkDir(imageRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		// 只监控类别目录（子目录层级为2: patientId/categoryId）
		if strings.Count(path, "/") == 2 { // patient_images/patientId/categoryId
			return watcher.Add(path)
		}
		return nil
	})

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					// 解析新图片路径
					patientID, categoryID, imageMeta, err := parseNewImagePath(event.Name)
					if err != nil {
						log.Printf("新图片解析失败: %s, 错误: %v", event.Name, err)
						continue
					}

					// 推送给订阅该病人的客户端
					sendNewImage(patientID, categoryID, imageMeta)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("监控错误:", err)
			}
		}
	}()

	<-make(chan bool)
}

// 解析新图片完整路径
func parseNewImagePath(fullPath string) (string, string, *ImageMetadata, error) {
	// 路径示例: ./patient_images/123/category_2/img_003.jpg
	parts := strings.Split(fullPath, "/")
	if len(parts) < 4 || parts[0] != "." || parts[1] != "patient_images" {
		return "", "", nil, fmt.Errorf("无效路径")
	}

	patientID := parts[2]
	categoryDir := parts[3]
	filename := parts[4]

	categoryID := strings.Split(categoryDir, "_")[1]
	imgMeta, err := parseImageFileName(filename)
	if err != nil {
		return "", "", nil, err
	}

	imgData, err := os.ReadFile(fullPath)
	if err != nil {
		return "", "", nil, err
	}
	imgMeta.ImageData = base64.StdEncoding.EncodeToString(imgData)

	return patientID, categoryID, imgMeta, nil
}

// 解析图片文件名
func parseImageFileName(filename string) (*ImageMetadata, error) {
	matches := imageRegex.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return nil, fmt.Errorf("文件名格式错误，需为 img_{id}.ext")
	}
	return &ImageMetadata{
		ImageID:  matches[1],
		FileName: filename,
	}, nil
}

// 发送新图片给订阅病人
func sendNewImage(patientID, categoryID string, meta *ImageMetadata) {
	payload := ImageMetadata{
		PatientID:  patientID,
		CategoryID: categoryID,
		ImageID:    meta.ImageID,
		ImageData:  meta.ImageData,
		FileName:   meta.FileName,
	}

	subMutex.Lock()
	defer subMutex.Unlock()

	for conn, subID := range clientSubscriptions {
		if subID == patientID {
			if err := conn.WriteJSON(payload); err != nil {
				log.Println("客户端连接已断开，移除订阅")
				delete(clientSubscriptions, conn)
			}
		}
	}
}

// 检查目录是否存在
func isDirExists(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && fileInfo.IsDir()
}
