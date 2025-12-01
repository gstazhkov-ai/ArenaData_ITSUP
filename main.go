package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Обновленная структура с новыми метриками
type MetricsData struct {
	OpenTickets string `json:"open_tickets"`
	SLA         string `json:"sla"`
	CSAT        string `json:"csat"`
	AvgTime     string `json:"avg_time"`
	FCR         string `json:"fcr"`       // New: First Contact Resolution
	Incidents   string `json:"incidents"` // New: % of Incidents vs Requests
	LastUpdated string `json:"last_updated"`
}

var (
	metricsFile = "metrics.json"
	fileMutex   sync.Mutex
)

func main() {
	r := gin.Default()
	r.MaxMultipartMemory = 50 << 20 // 50 MB
	r.LoadHTMLGlob("templates/*")
	r.Static("/uploads", "./uploads")

	// --- PUBLIC ROUTES ---
	
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "IT Support Roadmap"})
	})

	r.GET("/api/metrics", func(c *gin.Context) {
		data := loadMetrics()
		c.JSON(http.StatusOK, data)
	})

	r.GET("/api/files/:okrId", func(c *gin.Context) {
		okrId := c.Param("okrId")
		uploadPath := filepath.Join("uploads", okrId)
		var files []map[string]string
		entries, err := os.ReadDir(uploadPath)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() {
					files = append(files, map[string]string{
						"name": e.Name(),
						"url":  fmt.Sprintf("/uploads/%s/%s", okrId, e.Name()),
					})
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"files": files})
	})

	// --- ADMIN ROUTES ---
	// Login: admin / secret
	admin := r.Group("/admin", gin.BasicAuth(gin.Accounts{"admin": "secret"}))

	admin.GET("/check", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "admin_verified"})
	})

	admin.POST("/upload/:okrId", func(c *gin.Context) {
		okrId := c.Param("okrId")
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не получен"})
			return
		}
		uploadPath := filepath.Join("uploads", okrId)
		if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
			os.MkdirAll(uploadPath, os.ModePerm)
		}
		dst := filepath.Join(uploadPath, filepath.Base(file.Filename))
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "uploaded"})
	})

	admin.DELETE("/delete/:okrId/:filename", func(c *gin.Context) {
		okrId := c.Param("okrId")
		filename := c.Param("filename")
		cleanFilename := filepath.Base(filename)
		err := os.Remove(filepath.Join("uploads", okrId, cleanFilename))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	admin.POST("/metrics", func(c *gin.Context) {
		var newMetrics MetricsData
		if err := c.ShouldBindJSON(&newMetrics); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad JSON"})
			return
		}
		newMetrics.LastUpdated = time.Now().Format("02.01.2006")
		saveMetrics(newMetrics)
		c.JSON(http.StatusOK, newMetrics)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

func loadMetrics() MetricsData {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	defaultMetrics := MetricsData{
		OpenTickets: "0", SLA: "100%", CSAT: "5.0", AvgTime: "1ч",
		FCR: "85%", Incidents: "15%", LastUpdated: "-",
	}

	data, err := os.ReadFile(metricsFile)
	if err != nil {
		return defaultMetrics
	}

	var m MetricsData
	if err := json.Unmarshal(data, &m); err != nil {
		return defaultMetrics
	}
	return m
}

func saveMetrics(m MetricsData) {
	fileMutex.Lock()
	defer fileMutex.Unlock()
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(metricsFile, data, 0644)
}