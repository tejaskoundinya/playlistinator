package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/joho/godotenv"
)

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

// Function to get recent tracks from Last.fm
func GetLastFmRecentTracks(apiKey string, user string, fromTimestamp int64) []LastFmTrack {
	var allTracks []LastFmTrack
	page := 1

	for {
		url := fmt.Sprintf("http://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&user=%s&api_key=%s&format=json&page=%d&from=%d", user, apiKey, page, fromTimestamp)
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var recentTracksResponse LastFmRecentTracksResponse
		err = json.Unmarshal(body, &recentTracksResponse)
		if err != nil {
			log.Fatal(err)
		}

		// Append tracks from the current page
		allTracks = append(allTracks, recentTracksResponse.RecentTracks.Track...)

		// Check if there are more pages
		if len(recentTracksResponse.RecentTracks.Track) == 0 {
			break
		}

		page++
	}

	return allTracks
}

// Type to store the track and count of the track
type TrackCount struct {
	Track LastFmTrack
	Count int
}

// Function to get a count of each track in the last.fm recent tracks
func GetLastFmTrackCounts(tracks []LastFmTrack) []TrackCount {
	trackCounts := make(map[LastFmTrack]int)

	for _, track := range tracks {
		trackCounts[track]++
	}

	// Convert the map to a slice of TrackCount objects
	var trackCountSlice []TrackCount
	for track, count := range trackCounts {
		trackCountSlice = append(trackCountSlice, TrackCount{track, count})
	}

	return trackCountSlice
}

func GetLastFmSongs() {
	// Get recent tracks from Last.fm
	lastFmApiKey := os.Getenv("LASTFM_API_KEY")
	lastFmUser := os.Getenv("LASTFM_USER")
	lastFmFromTimestamp := time.Now().AddDate(0, 0, -30).Unix()
	lastFmRecentTracks := GetLastFmRecentTracks(lastFmApiKey, lastFmUser, lastFmFromTimestamp)
	lastFmRecentTrackCounts := GetLastFmTrackCounts(lastFmRecentTracks)

	// Sort the track counts by descending order of count
	sort.Slice(lastFmRecentTrackCounts, func(i, j int) bool {
		return lastFmRecentTrackCounts[i].Count > lastFmRecentTrackCounts[j].Count
	})

	// Print the track name and count of each track ordered by descending count
	for _, trackCount := range lastFmRecentTrackCounts {
		fmt.Printf("%s - %s - %d\n", trackCount.Track.Artist.Name, trackCount.Track.Name, trackCount.Count)
	}
}

func main() {
	// Load environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get the most played songs from Last.fm
	GetLastFmSongs()
}
