package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type TMDBMovie struct {
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
	Overview    string `json:"overview"`
	PosterPath  string `json:"poster_path"`
	VoteAverage float64 `json:"vote_average"`
}

type TMDBTVShow struct {
	Name        string `json:"name"`
	FirstAirDate string `json:"first_air_date"`
	Overview    string `json:"overview"`
	PosterPath  string `json:"poster_path"`
	VoteAverage float64 `json:"vote_average"`
}

type TMDBSeason struct {
	Episodes []TMDBEpisode `json:"episodes"`
}

type TMDBEpisode struct {
	Name        string `json:"name"`
	EpisodeNumber int `json:"episode_number"`
	Overview    string `json:"overview"`
	StillPath   string `json:"still_path"`
	VoteAverage float64 `json:"vote_average"`
	AirDate     string `json:"air_date"`
}

// ExtractTMDBID extrai o ID do TMDB do nome do arquivo
// Formatos suportados: [tmdbid-12345] ou tmdbid-12345
func ExtractTMDBID(filename string) (int, string) {
	re := regexp.MustCompile(`\[?tmdbid-(\d+)\]?`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) > 1 {
		id, err := strconv.Atoi(matches[1])
		if err == nil {
			return id, matches[0]
		}
	}
	return 0, ""
}

// IsEpisode detecta se é um episódio de série (S##E##)
func IsEpisode(filename string) bool {
	re := regexp.MustCompile(`S\d{2}E\d{2}`)
	return re.MatchString(filename)
}

// ExtractSeasonEpisode extrai temporada e episódio
func ExtractSeasonEpisode(filename string) (season, episode int) {
	re := regexp.MustCompile(`S(\d{2})E(\d{2})`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) > 2 {
		s, _ := strconv.Atoi(matches[1])
		e, _ := strconv.Atoi(matches[2])
		return s, e
	}
	return 0, 0
}

// GetMovieInfo busca informações do filme no TMDB
func GetMovieInfo(tmdbID int, apiKey string) (*TMDBMovie, error) {
	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d?api_key=%s&language=pt-BR", tmdbID, apiKey)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao buscar filme: status %d", resp.StatusCode)
	}

	var movie TMDBMovie
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return nil, err
	}

	return &movie, nil
}

// GetTVShowInfo busca informações da série no TMDB
func GetTVShowInfo(tmdbID int, apiKey string) (*TMDBTVShow, error) {
	url := fmt.Sprintf("https://api.themoviedb.org/3/tv/%d?api_key=%s&language=pt-BR", tmdbID, apiKey)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao buscar série: status %d", resp.StatusCode)
	}

	var show TMDBTVShow
	if err := json.NewDecoder(resp.Body).Decode(&show); err != nil {
		return nil, err
	}

	return &show, nil
}

// GetEpisodeInfo busca informações do episódio específico
func GetEpisodeInfo(tmdbID int, season int, episode int, apiKey string) (*TMDBEpisode, error) {
	url := fmt.Sprintf("https://api.themoviedb.org/3/tv/%d/season/%d/episode/%d?api_key=%s&language=pt-BR", tmdbID, season, episode, apiKey)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao buscar episódio: status %d", resp.StatusCode)
	}

	var ep TMDBEpisode
	if err := json.NewDecoder(resp.Body).Decode(&ep); err != nil {
		return nil, err
	}

	return &ep, nil
}

// FormatMovieMetadata formata metadados do filme para exibição
func FormatMovieMetadata(movie *TMDBMovie) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📽️ %s\n", movie.Title))
	if movie.ReleaseDate != "" {
		year := movie.ReleaseDate[:4]
		sb.WriteString(fmt.Sprintf("📅 %s\n", year))
	}
	if movie.VoteAverage > 0 {
		sb.WriteString(fmt.Sprintf("⭐ %.1f/10\n", movie.VoteAverage))
	}
	if movie.Overview != "" {
		overview := movie.Overview
		if len(overview) > 150 {
			overview = overview[:150] + "..."
		}
		sb.WriteString(fmt.Sprintf("📝 %s", overview))
	}
	return sb.String()
}

// FormatEpisodeMetadata formata metadados do episódio para exibição
func FormatEpisodeMetadata(show *TMDBTVShow, episode *TMDBEpisode, season, ep int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📺 %s\n", show.Name))
	sb.WriteString(fmt.Sprintf("🎬 S%02dE%02d - %s\n", season, ep, episode.Name))
	if episode.AirDate != "" {
		sb.WriteString(fmt.Sprintf("📅 %s\n", episode.AirDate))
	}
	if episode.VoteAverage > 0 {
		sb.WriteString(fmt.Sprintf("⭐ %.1f/10\n", episode.VoteAverage))
	}
	if episode.Overview != "" {
		overview := episode.Overview
		if len(overview) > 150 {
			overview = overview[:150] + "..."
		}
		sb.WriteString(fmt.Sprintf("📝 %s", overview))
	}
	return sb.String()
}
