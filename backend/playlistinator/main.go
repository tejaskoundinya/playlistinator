package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
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
	Track interface{} `json:"track"` // Can be either a single track or an array of tracks
	Attr  struct {
		User       string `json:"user"`
		Page       string `json:"page"`
		PerPage    string `json:"perPage"`
		TotalPages string `json:"totalPages"`
		Total      string `json:"total"`
	} `json:"@attr"`
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

// API response types
type GenerateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Count   int    `json:"count,omitempty"`
}

// Function to get recent tracks from Last.fm
func GetLastFmRecentTracks(apiKey string, user string, fromTimestamp int64) []LastFmTrack {
	var allTracks []LastFmTrack
	page := 1
	perPage := 1000 // Maximum allowed by Last.fm API
	totalPages := 1

	for page <= totalPages {
		fmt.Printf("Fetching Last.fm page %d of %d...\n", page, totalPages)
		url := fmt.Sprintf("http://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&user=%s&api_key=%s&format=json&page=%d&from=%d&limit=%d",
			user, apiKey, page, fromTimestamp, perPage)
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
			log.Printf("Error unmarshaling response: %v", err)
			log.Fatal(err)
		}

		// Update total pages from the API response
		if totalPagesStr := recentTracksResponse.RecentTracks.Attr.TotalPages; totalPagesStr != "" {
			if total, err := strconv.Atoi(totalPagesStr); err == nil {
				totalPages = total
			}
		}

		// Handle both single track and array of tracks
		switch tracks := recentTracksResponse.RecentTracks.Track.(type) {
		case []interface{}:
			// Multiple tracks case
			fmt.Printf("Found %d tracks on page %d\n", len(tracks), page)
			for _, track := range tracks {
				trackJSON, err := json.Marshal(track)
				if err != nil {
					log.Printf("Error marshaling track: %v", err)
					continue
				}
				var lastFmTrack LastFmTrack
				if err := json.Unmarshal(trackJSON, &lastFmTrack); err != nil {
					log.Printf("Error unmarshaling track: %v", err)
					continue
				}
				allTracks = append(allTracks, lastFmTrack)
			}
		case map[string]interface{}:
			// Single track case
			fmt.Printf("Found 1 track on page %d\n", page)
			trackJSON, err := json.Marshal(tracks)
			if err != nil {
				log.Printf("Error marshaling single track: %v", err)
				continue
			}
			var lastFmTrack LastFmTrack
			if err := json.Unmarshal(trackJSON, &lastFmTrack); err != nil {
				log.Printf("Error unmarshaling single track: %v", err)
				continue
			}
			allTracks = append(allTracks, lastFmTrack)
		default:
			log.Printf("Unexpected track type: %T", tracks)
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

// Function to get access token from Spotify
func GetSpotifyAccessToken() string {
	clientId := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	refreshToken := os.Getenv("SPOTIFY_REFRESH_TOKEN")

	url := "https://accounts.spotify.com/api/token"
	reqBody := fmt.Sprintf("grant_type=refresh_token&refresh_token=%s&client_id=%s&client_secret=%s", refreshToken, clientId, clientSecret)
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(reqBody)))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to get Spotify access token: %s", resp.Status)
	}

	var tokenResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		log.Fatal(err)
	}

	accessToken, ok := tokenResponse["access_token"].(string)
	if !ok {
		log.Fatal("Failed to parse access token from Spotify response")
	}

	return accessToken
}

// Function to get the playlist from Spotify with the given name
func GetSpotifyPlaylistId(accessToken string, playlistName string) string {
	url := "https://api.spotify.com/v1/me/playlists"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}

	// Look for the playlist with the given name
	for _, playlist := range result.Items {
		if playlist.Name == playlistName {
			return playlist.Id
		}
	}

	// If playlist not found, create it
	return CreateSpotifyPlaylist(accessToken, playlistName)
}

// Function to create a playlist in Spotify with the given name
func CreateSpotifyPlaylist(accessToken string, playlistName string) string {
	// First, get the user's ID
	url := "https://api.spotify.com/v1/me"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var userResult struct {
		Id string `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userResult); err != nil {
		log.Fatal(err)
	}

	// Create the playlist
	url = fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", userResult.Id)
	playlistData := struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Public      bool   `json:"public"`
	}{
		Name:        playlistName,
		Description: "Top 100 songs from the last 30 days",
		Public:      false,
	}

	jsonData, err := json.Marshal(playlistData)
	if err != nil {
		log.Fatal(err)
	}

	req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var result struct {
		Id string `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}

	return result.Id
}

/**
 * Function to get song URIs from Spotify
 * Songs are available from last.fm
 */
func GetSpotifySongUris(accessToken string, songs []LastFmTrack) []string {
	var songUris []string

	for _, song := range songs {
		url := fmt.Sprintf("https://api.spotify.com/v1/search?q=track:%s artist:%s&type=track", song.Name, song.Artist.Name)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var getTrackResponse SpotifyGetTrackResponse
		err = json.Unmarshal(body, &getTrackResponse)
		if err != nil {
			log.Fatal(err)
		}

		if len(getTrackResponse.Tracks.Items) > 0 {
			songUris = append(songUris, getTrackResponse.Tracks.Items[0].Uri)
		}
	}

	return songUris
}

/**
 * Function to add songs to a Spotify playlist
 * All songs are removed from the playlist before adding the new songs
 */
func AddSpotifySongs(accessToken string, playlistId string, songUris []string) {
	// Remove all songs from the playlist
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistId)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Add the new songs to the playlist
	url = fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistId)
	songUrisJson, err := json.Marshal(SpotifyAddItemRequest{songUris})
	if err != nil {
		log.Fatal(err)
	}
	req, err = http.NewRequest("POST", url, io.NopCloser(bytes.NewReader(songUrisJson)))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}

// Function to add songs to a Spotify playlist
func AddSongsToPlaylist(accessToken string, playlistId string, songUris []string) {
	// First, get all existing tracks in the playlist
	fmt.Println("Getting existing tracks from playlist...")
	var allExistingTracks []struct {
		Track struct {
			Uri string `json:"uri"`
		} `json:"track"`
	}

	client := &http.Client{}
	nextURL := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistId)
	for nextURL != "" {
		req, err := http.NewRequest("GET", nextURL, nil)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		var result struct {
			Items []struct {
				Track struct {
					Uri string `json:"uri"`
				} `json:"track"`
			} `json:"items"`
			Next string `json:"next"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			log.Fatal(err)
		}
		resp.Body.Close()

		allExistingTracks = append(allExistingTracks, result.Items...)
		nextURL = result.Next
	}

	// If there are existing tracks, remove them
	if len(allExistingTracks) > 0 {
		fmt.Printf("Removing %d existing tracks from playlist...\n", len(allExistingTracks))
		var tracksToRemove []struct {
			Uri string `json:"uri"`
		}
		for _, item := range allExistingTracks {
			tracksToRemove = append(tracksToRemove, struct {
				Uri string `json:"uri"`
			}{Uri: item.Track.Uri})
		}

		// Remove tracks in batches of 100 (Spotify API limit)
		for i := 0; i < len(tracksToRemove); i += 100 {
			end := i + 100
			if end > len(tracksToRemove) {
				end = len(tracksToRemove)
			}
			batch := tracksToRemove[i:end]

			url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistId)
			jsonData, err := json.Marshal(struct {
				Tracks []struct {
					Uri string `json:"uri"`
				} `json:"tracks"`
			}{Tracks: batch})
			if err != nil {
				log.Fatal(err)
			}

			req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonData))
			if err != nil {
				log.Fatal(err)
			}

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			resp.Body.Close()
		}
		fmt.Println("Successfully cleared playlist")
	}

	// Add the new songs to the playlist
	fmt.Println("Adding new tracks to playlist...")
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistId)

	// Spotify has a limit of 100 songs per request, so we need to batch the requests
	for i := 0; i < len(songUris); i += 100 {
		end := i + 100
		if end > len(songUris) {
			end = len(songUris)
		}

		batch := songUris[i:end]
		uris := struct {
			Uris []string `json:"uris"`
		}{
			Uris: batch,
		}

		jsonData, err := json.Marshal(uris)
		if err != nil {
			log.Fatal(err)
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			log.Printf("Failed to add songs batch %d-%d: %s", i, end, resp.Status)
		}
	}
}

// Function to search for a song on Spotify and get its URI
func SearchSpotifySong(accessToken string, track LastFmTrack) (string, error) {
	query := fmt.Sprintf("track:%s artist:%s", track.Name, track.Artist.Name)
	url := fmt.Sprintf("https://api.spotify.com/v1/search?q=%s&type=track&limit=1", url.QueryEscape(query))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Tracks struct {
			Items []struct {
				Uri string `json:"uri"`
			} `json:"items"`
		} `json:"tracks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Tracks.Items) == 0 {
		return "", fmt.Errorf("no matching track found")
	}

	return result.Tracks.Items[0].Uri, nil
}

// Function to handle CORS
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// API handler for generating playlists
func handleGeneratePlaylist(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get Last.fm API credentials
	lastFmApiKey := os.Getenv("LASTFM_API_KEY")
	lastFmUser := os.Getenv("LASTFM_USER")

	// Check for Spotify refresh token
	spotifyRefreshToken := os.Getenv("SPOTIFY_REFRESH_TOKEN")
	if spotifyRefreshToken == "" {
		json.NewEncoder(w).Encode(GenerateResponse{
			Success: false,
			Message: "SPOTIFY_REFRESH_TOKEN not found in .env file. Please run with -auth flag to authenticate with Spotify",
		})
		return
	}

	// Get Last.fm tracks from the last 30 days
	fromTimestamp := time.Now().AddDate(0, 0, -30).Unix()
	tracks := GetLastFmRecentTracks(lastFmApiKey, lastFmUser, fromTimestamp)

	// Get track counts and sort by count
	trackCounts := GetLastFmTrackCounts(tracks)
	sort.Slice(trackCounts, func(i, j int) bool {
		return trackCounts[i].Count > trackCounts[j].Count
	})

	// Create playlist description with top 10 tracks and their play counts
	var description strings.Builder
	description.WriteString("Top 100 songs from the last 30 days. Top 10 most played:\n")
	for i := 0; i < 10 && i < len(trackCounts); i++ {
		description.WriteString(fmt.Sprintf("%d. %s - %s (%d plays)\n",
			i+1,
			trackCounts[i].Track.Artist.Name,
			trackCounts[i].Track.Name,
			trackCounts[i].Count))
	}

	// Get Spotify access token
	accessToken := GetSpotifyAccessToken()

	// Get or create the playlist
	playlistName := "TK - Hot 100"
	playlistId := GetSpotifyPlaylistId(accessToken, playlistName)

	// Update playlist description
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s", playlistId)
	jsonData, err := json.Marshal(struct {
		Description string `json:"description"`
	}{
		Description: description.String(),
	})
	if err != nil {
		json.NewEncoder(w).Encode(GenerateResponse{
			Success: false,
			Message: "Failed to marshal playlist description: " + err.Error(),
		})
		return
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		json.NewEncoder(w).Encode(GenerateResponse{
			Success: false,
			Message: "Failed to create request: " + err.Error(),
		})
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		json.NewEncoder(w).Encode(GenerateResponse{
			Success: false,
			Message: "Failed to update playlist description: " + err.Error(),
		})
		return
	}
	resp.Body.Close()

	// Get Spotify URIs for the top 100 songs
	var songUris []string
	for i, trackCount := range trackCounts {
		if i >= 100 {
			break
		}

		uri, err := SearchSpotifySong(accessToken, trackCount.Track)
		if err != nil {
			log.Printf("Could not find Spotify URI for %s - %s: %v",
				trackCount.Track.Artist.Name, trackCount.Track.Name, err)
			continue
		}
		songUris = append(songUris, uri)
	}

	// Add songs to the playlist
	AddSongsToPlaylist(accessToken, playlistId, songUris)

	json.NewEncoder(w).Encode(GenerateResponse{
		Success: true,
		Message: fmt.Sprintf("Success! Added %d songs to playlist '%s'", len(songUris), playlistName),
		Count:   len(songUris),
	})
}

// Function to start the authentication server for Spotify
func StartSpotifyAuthServer() {
	clientId := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")

	if clientId == "" || clientSecret == "" || redirectURI == "" {
		log.Fatal("SPOTIFY_CLIENT_ID, SPOTIFY_CLIENT_SECRET, and SPOTIFY_REDIRECT_URI must be set in .env file")
	}

	// Create a channel to receive the authorization code
	authCodeChan := make(chan string)

	// Set up the HTTP server to handle the callback
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No code received", http.StatusBadRequest)
			return
		}
		authCodeChan <- code
		w.Write([]byte("Authentication successful! You can close this window."))
	})

	// Start the server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go server.ListenAndServe()

	// Construct the authorization URL
	authURL := fmt.Sprintf(
		"https://accounts.spotify.com/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=playlist-modify-public playlist-modify-private",
		clientId,
		url.QueryEscape(redirectURI),
	)

	fmt.Println("Please visit this URL to authorize the application:")
	fmt.Println(authURL)

	// Wait for the authorization code
	code := <-authCodeChan

	// Exchange the authorization code for a refresh token
	tokenURL := "https://accounts.spotify.com/api/token"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var result struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}

	// Update the .env file with the refresh token
	envFile, err := os.ReadFile(".env")
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(envFile), "\n")
	var newLines []string
	found := false
	for _, line := range lines {
		if strings.HasPrefix(line, "SPOTIFY_REFRESH_TOKEN=") {
			newLines = append(newLines, fmt.Sprintf("SPOTIFY_REFRESH_TOKEN=%s", result.RefreshToken))
			found = true
		} else {
			newLines = append(newLines, line)
		}
	}
	if !found {
		newLines = append(newLines, fmt.Sprintf("SPOTIFY_REFRESH_TOKEN=%s", result.RefreshToken))
	}

	err = os.WriteFile(".env", []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully saved refresh token to .env file")
	server.Close()
}

func main() {
	// Parse command line flags
	authMode := flag.Bool("auth", false, "Run in authentication mode to get Spotify refresh token")
	serverMode := flag.Bool("server", false, "Run in server mode to provide API endpoints")
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	if *authMode {
		StartSpotifyAuthServer()
		return
	}

	if *serverMode {
		// Set up HTTP server
		mux := http.NewServeMux()
		mux.HandleFunc("/api/generate", handleGeneratePlaylist)

		// Add CORS middleware
		handler := enableCORS(mux)

		// Start server
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		log.Printf("Starting server on port %s...", port)
		log.Fatal(http.ListenAndServe(":"+port, handler))
		return
	}

	// Get Last.fm API credentials
	lastFmApiKey := os.Getenv("LASTFM_API_KEY")
	lastFmUser := os.Getenv("LASTFM_USER")

	// Check for Spotify refresh token
	spotifyRefreshToken := os.Getenv("SPOTIFY_REFRESH_TOKEN")
	if spotifyRefreshToken == "" {
		log.Fatal("SPOTIFY_REFRESH_TOKEN not found in .env file. Please run with -auth flag to authenticate with Spotify")
	}

	fmt.Println("Step 1: Fetching Last.fm tracks from the last 30 days...")
	// Get Last.fm tracks from the last 30 days
	fromTimestamp := time.Now().AddDate(0, 0, -30).Unix()
	tracks := GetLastFmRecentTracks(lastFmApiKey, lastFmUser, fromTimestamp)
	fmt.Printf("Found %d tracks from Last.fm\n", len(tracks))

	fmt.Println("\nStep 2: Counting and sorting tracks by play count...")
	// Get track counts and sort by count
	trackCounts := GetLastFmTrackCounts(tracks)
	sort.Slice(trackCounts, func(i, j int) bool {
		return trackCounts[i].Count > trackCounts[j].Count
	})
	fmt.Printf("Found %d unique tracks\n", len(trackCounts))

	// Create playlist description with top 10 tracks and their play counts
	var description strings.Builder
	description.WriteString("Top 100 songs from the last 30 days. Top 10 most played:\n")
	for i := 0; i < 10 && i < len(trackCounts); i++ {
		description.WriteString(fmt.Sprintf("%d. %s - %s (%d plays)\n",
			i+1,
			trackCounts[i].Track.Artist.Name,
			trackCounts[i].Track.Name,
			trackCounts[i].Count))
	}

	fmt.Println("\nStep 3: Getting Spotify access token...")
	// Get Spotify access token
	accessToken := GetSpotifyAccessToken()
	fmt.Println("Successfully obtained Spotify access token")

	fmt.Println("\nStep 4: Getting or creating Spotify playlist...")
	// Get or create the playlist
	playlistName := "TK - Hot 100"
	playlistId := GetSpotifyPlaylistId(accessToken, playlistName)
	fmt.Printf("Using playlist: %s (ID: %s)\n", playlistName, playlistId)

	// Update playlist description
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s", playlistId)
	jsonData, err := json.Marshal(struct {
		Description string `json:"description"`
	}{
		Description: description.String(),
	})
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	fmt.Println("\nStep 5: Searching for songs on Spotify...")
	// Get Spotify URIs for the top 100 songs
	var songUris []string
	for i, trackCount := range trackCounts {
		if i >= 100 {
			break
		}

		fmt.Printf("Searching for %d/100: %s - %s (%d plays)\n",
			i+1,
			trackCount.Track.Artist.Name,
			trackCount.Track.Name,
			trackCount.Count)
		uri, err := SearchSpotifySong(accessToken, trackCount.Track)
		if err != nil {
			log.Printf("Could not find Spotify URI for %s - %s: %v",
				trackCount.Track.Artist.Name, trackCount.Track.Name, err)
			continue
		}
		songUris = append(songUris, uri)
	}
	fmt.Printf("Found Spotify URIs for %d songs\n", len(songUris))

	fmt.Println("\nStep 6: Adding songs to playlist...")
	// Add songs to the playlist
	AddSongsToPlaylist(accessToken, playlistId, songUris)

	fmt.Printf("\nSuccess! Added %d songs to playlist '%s'\n", len(songUris), playlistName)
}
