package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	redirectURI = "http://localhost:8080/callback"
	scope       = "playlist-modify-public playlist-modify-private playlist-read-private playlist-read-collaborative"
)

func StartAuthServer() {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	// Serve the login page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authURL := fmt.Sprintf("https://accounts.spotify.com/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=%s",
			clientID,
			redirectURI,
			scope)
		fmt.Fprintf(w, `<a href="%s">Login with Spotify</a>`, authURL)
	})

	// Handle the callback
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No code received", http.StatusBadRequest)
			return
		}

		// Exchange the code for tokens
		tokenURL := "https://accounts.spotify.com/api/token"
		data := fmt.Sprintf("grant_type=authorization_code&code=%s&redirect_uri=%s&client_id=%s&client_secret=%s",
			code, redirectURI, clientID, clientSecret)

		req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var result struct {
			AccessToken  string `json:"access_token"`
			TokenType    string `json:"token_type"`
			ExpiresIn    int    `json:"expires_in"`
			RefreshToken string `json:"refresh_token"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the .env file with the refresh token
		envContent := fmt.Sprintf(`LASTFM_API_KEY=%s
LASTFM_USER=%s
SPOTIFY_CLIENT_ID=%s
SPOTIFY_CLIENT_SECRET=%s
SPOTIFY_REFRESH_TOKEN=%s
`, os.Getenv("LASTFM_API_KEY"),
			os.Getenv("LASTFM_USER"),
			os.Getenv("SPOTIFY_CLIENT_ID"),
			os.Getenv("SPOTIFY_CLIENT_SECRET"),
			result.RefreshToken)

		if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Authentication successful! Your refresh token has been saved to .env file. You can now close this window.")
	})

	fmt.Println("Starting auth server on http://localhost:8080")
	fmt.Println("Please visit http://localhost:8080 in your browser to authenticate with Spotify")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
