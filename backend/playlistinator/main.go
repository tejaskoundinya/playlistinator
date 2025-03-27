package main

func main() {
	// Last.fm types
	type LastFmArtist struct {
		Name string `json:"#text"`
	}

	type LastFmAlbum struct {
		Name string `json:"#text"`
	}

	type LastFmTrack struct {
		Artist LastFmArtist `json:"artist"`
		Album  LastFmAlbum  `json:"album"`
		Name   string       `json:"name"`
	}

	type LastFmRecentTracks struct {
		Track []LastFmTrack `json:"track"`
	}

	type LastFmRecentTracksResponse struct {
		RecentTracks LastFmRecentTracks `json:"recenttracks"`
	}

	// Spotify types
	type SpotifySongTrack struct {
		Uri  string `json:"uri"`
		Id   string `json:"id"`
		Name string `json:"name"`
	}

	type SpotifySongItem struct {
		Track    SpotifySongTrack `json:"track"`
		PlayedAt string           `json:"played_at"`
	}

	type SpotifyCursors struct {
		After  string `json:"after"`
		Before string `json:"before"`
	}

	type SpotifyRecentSongResponse struct {
		Items   []SpotifySongItem `json:"items"`
		Next    string            `json:"next"`
		Cursors SpotifyCursors    `json:"cursors"`
		Limit   int               `json:"limit"`
	}

	type SpotifyAddItemRequest struct {
		Uris []string
	}

	type SpotifySongsResponse struct {
		Songs     []string
		Timestamp []int
		Before    string
	}

	type SpotifyGetTrackTracks struct {
		Items []SpotifySongTrack `json:"items"`
	}

	type SpotifyGetTrackResponse struct {
		Tracks SpotifyGetTrackTracks `json:"tracks"`
	}
}
