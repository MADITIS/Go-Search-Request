package tmdb

// not implemented

import (
	tmdbApi "github.com/cyruzin/golang-tmdb"
	config "github.com/maditis/search-go/src/config"
	internal "github.com/maditis/search-go/src/internal"
)

func Tmdb()(*tmdbApi.SearchMovies) {
	tmdbClient, err := tmdbApi.Init(config.EnvFields.TmdbAPI)
	internal.WarningLog(err, "Provide TMDB API KEY")
	movie,err := tmdbClient.GetSearchMovies("john wick", nil)
	return movie
}