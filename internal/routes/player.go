// player.go
package routes

import (
	"fmt"
	"math"
	"net/http"
	"strconv" // <-- ADICIONAR

	"github.com/gin-gonic/gin"
)

// LoadPlayerRoutes permanece o mesmo
func (a *allRoutes) LoadPlayerRoutes(r *Route) {
	r.Engine.GET("/player/:messageID", a.handlePlayerPage)
	a.log.Info("Loaded Player routes")
}

// formatBytes (Nova função auxiliar)
// Converte bytes em um formato legível (KB, MB, GB)
func formatBytes(b int64) string {
	if b == 0 {
		return "0 Bytes"
	}
	const unit = 1024
	suffixes := []string{"Bytes", "KB", "MB", "GB", "TB", "PB"}
	i := math.Floor(math.Log(float64(b)) / math.Log(unit))
	return fmt.Sprintf("%.2f %s", float64(b)/math.Pow(unit, i), suffixes[int(i)])
}

// handlePlayerPage (Atualizado para lidar com os novos dados)
func (a *allRoutes) handlePlayerPage(c *gin.Context) {
	messageID := c.Param("messageID")
	hash := c.Query("hash")
	// mimeType := c.Query("mime") // ◄◄◄ REMOVA ESTA LINHA

	if messageID == "" || hash == "" {
		c.String(http.StatusBadRequest, "Link inválido ou expirado.")
		return
	}

	// if mimeType == "" { // ◄◄◄ REMOVA ESTA
	// 	mimeType = "video/mp4" // ◄◄◄ REMOVA ESTA
	// } // ◄◄◄ REMOVA ESTA

	// ▼▼▼ INÍCIO DAS MODIFICAÇÕES ▼▼▼

	// Constrói as URLs
	streamURL := fmt.Sprintf("/stream/%s?hash=%s", messageID, hash)
	downloadURL := fmt.Sprintf("%s&d=true", streamURL) // URL de download direto

	// Pega os novos dados da query
	fileName := c.Query("filename") // Gin já decodifica o URL
	if fileName == "" {
		fileName = "Vídeo" // Nome padrão
	}

	fileSizeStr := c.Query("filesize")
	fileSize, _ := strconv.ParseInt(fileSizeStr, 10, 64)
	fileSizeFormatted := formatBytes(fileSize) // Formata os bytes

	// Renderiza o template HTML, passando os novos dados
	c.HTML(http.StatusOK, "player.html", gin.H{
		"StreamURL":   streamURL,
		"DownloadURL": downloadURL,
		// "MimeType":    mimeType, // ◄◄◄ REMOVA ESTA LINHA
		"FileName":    fileName,
		"FileSize":    fileSizeFormatted,
	})
	// ▲▲▲ FIM DAS MODIFICAÇÕES ▲▲▲
}
