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
}
