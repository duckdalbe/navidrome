package subsonic

import (
	"context"
	"net/http"
	"time"

	"github.com/deluan/navidrome/engine"
	"github.com/deluan/navidrome/log"
	"github.com/deluan/navidrome/model"
	"github.com/deluan/navidrome/server/subsonic/responses"
	"github.com/deluan/navidrome/utils"
)

type MediaAnnotationController struct {
	scrobbler engine.Scrobbler
	ratings   engine.Ratings
}

func NewMediaAnnotationController(scrobbler engine.Scrobbler, ratings engine.Ratings) *MediaAnnotationController {
	return &MediaAnnotationController{
		scrobbler: scrobbler,
		ratings:   ratings,
	}
}

func (c *MediaAnnotationController) SetRating(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	id, err := RequiredParamString(r, "id", "Required id parameter is missing")
	if err != nil {
		return nil, err
	}
	rating, err := RequiredParamInt(r, "rating", "Required rating parameter is missing")
	if err != nil {
		return nil, err
	}

	log.Debug(r, "Setting rating", "rating", rating, "id", id)
	err = c.ratings.SetRating(r.Context(), id, rating)

	switch {
	case err == model.ErrNotFound:
		log.Error(r, err)
		return nil, NewError(responses.ErrorDataNotFound, "ID not found")
	case err != nil:
		log.Error(r, err)
		return nil, NewError(responses.ErrorGeneric, "Internal Error")
	}

	return NewResponse(), nil
}

func (c *MediaAnnotationController) Star(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	ids := utils.ParamStrings(r, "id")
	albumIds := utils.ParamStrings(r, "albumId")
	artistIds := utils.ParamStrings(r, "artistId")
	if len(ids)+len(albumIds)+len(artistIds) == 0 {
		return nil, NewError(responses.ErrorMissingParameter, "Required id parameter is missing")
	}
	ids = append(ids, albumIds...)
	ids = append(ids, artistIds...)

	err := c.star(r.Context(), true, ids...)
	if err != nil {
		return nil, err
	}

	return NewResponse(), nil
}

func (c *MediaAnnotationController) star(ctx context.Context, starred bool, ids ...string) error {
	if len(ids) == 0 {
		return nil
	}
	log.Debug(ctx, "Changing starred", "ids", ids, "starred", starred)
	err := c.ratings.SetStar(ctx, starred, ids...)
	switch {
	case err == model.ErrNotFound:
		log.Error(ctx, err)
		return NewError(responses.ErrorDataNotFound, "ID not found")
	case err != nil:
		log.Error(ctx, err)
		return NewError(responses.ErrorGeneric, "Internal Error")
	}
	return nil
}

func (c *MediaAnnotationController) Unstar(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	ids := utils.ParamStrings(r, "id")
	albumIds := utils.ParamStrings(r, "albumId")
	artistIds := utils.ParamStrings(r, "artistId")
	if len(ids)+len(albumIds)+len(artistIds) == 0 {
		return nil, NewError(responses.ErrorMissingParameter, "Required id parameter is missing")
	}
	ids = append(ids, albumIds...)
	ids = append(ids, artistIds...)

	err := c.star(r.Context(), false, ids...)
	if err != nil {
		return nil, err
	}

	return NewResponse(), nil
}

func (c *MediaAnnotationController) Scrobble(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	ids, err := RequiredParamStrings(r, "id", "Required id parameter is missing")
	if err != nil {
		return nil, err
	}
	times := utils.ParamTimes(r, "time")
	if len(times) > 0 && len(times) != len(ids) {
		return nil, NewError(responses.ErrorGeneric, "Wrong number of timestamps: %d, should be %d", len(times), len(ids))
	}
	submission := utils.ParamBool(r, "submission", true)
	playerId := 1 // TODO Multiple players, based on playerName/username/clientIP(?)
	playerName := utils.ParamString(r, "c")
	username := utils.ParamString(r, "u")

	log.Debug(r, "Scrobbling tracks", "ids", ids, "times", times, "submission", submission)
	for i, id := range ids {
		var t time.Time
		if len(times) > 0 {
			t = times[i]
		} else {
			t = time.Now()
		}
		if submission {
			_, err := c.scrobbler.Register(r.Context(), playerId, id, t)
			if err != nil {
				log.Error(r, "Error scrobbling track", "id", id, err)
				continue
			}
		} else {
			_, err := c.scrobbler.NowPlaying(r.Context(), playerId, playerName, id, username)
			if err != nil {
				log.Error(r, "Error setting current song", "id", id, err)
				continue
			}
		}
	}
	return NewResponse(), nil
}
