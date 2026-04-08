// player.go
package routes

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

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

	if messageID == "" || hash == "" {
		c.String(http.StatusBadRequest, "Link inválido ou expirado.")
		return
	}

	// Constrói as URLs
	streamURL := fmt.Sprintf("/stream/%s?hash=%s", messageID, hash)
	downloadURL := fmt.Sprintf("%s&d=true", streamURL)

	// Pega os novos dados da query
	fileName := c.Query("filename")
	if fileName == "" {
		fileName = "Vídeo"
	}

	fileSizeStr := c.Query("filesize")
	fileSize, _ := strconv.ParseInt(fileSizeStr, 10, 64)
	fileSizeFormatted := formatBytes(fileSize)

	// Tenta extrair ID do TMDB e buscar metadados (com timeout)
	metadata := ""
	tmdbID, _ := utils.ExtractTMDBID(fileName)

	if tmdbID > 0 && config.ValueOf.TMDBAPIKey != "" {
		// Usa goroutine com timeout para evitar travar
		done := make(chan string, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					a.log.Sugar().Error("Erro ao buscar TMDB:", r)
					done <- ""
				}
			}()

			if utils.IsEpisode(fileName) {
				season, episode := utils.ExtractSeasonEpisode(fileName)
				if show, err := utils.GetTVShowInfo(tmdbID, config.ValueOf.TMDBAPIKey); err == nil {
					if ep, err := utils.GetEpisodeInfo(tmdbID, season, episode, config.ValueOf.TMDBAPIKey); err == nil {
						done <- utils.FormatEpisodeMetadata(show, ep, season, episode)
					} else {
						a.log.Sugar().Debug("Erro ao buscar episódio:", err)
						done <- ""
					}
				} else {
					a.log.Sugar().Debug("Erro ao buscar série:", err)
					done <- ""
				}
			} else {
				if movie, err := utils.GetMovieInfo(tmdbID, config.ValueOf.TMDBAPIKey); err == nil {
					done <- utils.FormatMovieMetadata(movie)
				} else {
					a.log.Sugar().Debug("Erro ao buscar filme:", err)
					done <- ""
				}
			}
		}()

		// Aguarda por 5 segundos no máximo
		select {
		case result := <-done:
			metadata = result
		case <-time.After(5 * time.Second):
			a.log.Sugar().Warn("Timeout ao buscar metadados TMDB")
			metadata = ""
		}
	}

	// Renderiza o template HTML, passando os novos dados
	c.HTML(http.StatusOK, "player.html", gin.H{
		"StreamURL":   streamURL,
		"DownloadURL": downloadURL,
		"FileName":    fileName,
		"FileSize":    fileSizeFormatted,
		"Metadata":    metadata,
	})
}
