// player.go
package routes

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

	// Tenta extrair ID do TMDB e buscar metadados
	metadata := ""
	log := a.log.Named("player")

	log.Info("Iniciando busca de metadados", zap.String("fileName", fileName))

	tmdbID, idStr := utils.ExtractTMDBID(fileName)
	log.Info("TMDB ID extraído", zap.Int("tmdbID", tmdbID), zap.String("idStr", idStr))

	if tmdbID > 0 && config.ValueOf.TMDBAPIKey != "" {
		log.Info("Buscando metadados no TMDB", zap.Int("tmdbID", tmdbID))
		
		if utils.IsEpisode(fileName) {
			log.Info("Detectado como episódio")
			season, episode := utils.ExtractSeasonEpisode(fileName)
			log.Info("Season e episode extraídos", zap.Int("season", season), zap.Int("episode", episode))

			show, err := utils.GetTVShowInfo(tmdbID, config.ValueOf.TMDBAPIKey)
			if err != nil {
				log.Error("Erro ao buscar série", zap.Error(err))
			} else {
				log.Info("Série encontrada", zap.String("showName", show.Name))
				
				ep, err := utils.GetEpisodeInfo(tmdbID, season, episode, config.ValueOf.TMDBAPIKey)
				if err != nil {
					log.Error("Erro ao buscar episódio", zap.Error(err))
				} else {
					log.Info("Episódio encontrado", zap.String("epName", ep.Name))
					metadata = utils.FormatEpisodeMetadata(show, ep, season, episode)
					log.Info("Metadados do episódio formatados")
				}
			}
		} else {
			log.Info("Detectado como filme")
			movie, err := utils.GetMovieInfo(tmdbID, config.ValueOf.TMDBAPIKey)
			if err != nil {
				log.Error("Erro ao buscar filme", zap.Error(err))
			} else {
				log.Info("Filme encontrado", zap.String("movieTitle", movie.Title))
				metadata = utils.FormatMovieMetadata(movie)
				log.Info("Metadados do filme formatados")
			}
		}
	} else {
		if tmdbID == 0 {
			log.Info("TMDB ID não encontrado no nome do arquivo")
		}
		if config.ValueOf.TMDBAPIKey == "" {
			log.Info("TMDB_API_KEY não configurada")
		}
	}

	log.Info("Renderizando player", zap.String("metadata", metadata))

	// Renderiza o template HTML, passando os novos dados
	c.HTML(http.StatusOK, "player.html", gin.H{
		"StreamURL":   streamURL,
		"DownloadURL": downloadURL,
		"FileName":    fileName,
		"FileSize":    fileSizeFormatted,
		"Metadata":    metadata,
	})
}
