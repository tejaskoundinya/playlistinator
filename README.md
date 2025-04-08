# Playlistinator

A web application that generates Spotify playlists based on your Last.fm listening history.

## Project Structure

- `frontend/`: Next.js frontend application
- `backend/`: Go backend application

## Development Setup

### Prerequisites

- Node.js and npm
- Go 1.16 or later
- Docker

### Environment Variables

This project uses environment variables for configuration. Never commit actual secrets to the repository.

#### Backend

Create a `.env` file in the `backend/playlistinator` directory with the following variables:

```
LASTFM_API_KEY=your_lastfm_api_key
LASTFM_USER=your_lastfm_username
SPOTIFY_CLIENT_ID=your_spotify_client_id
SPOTIFY_CLIENT_SECRET=your_spotify_client_secret
SPOTIFY_REFRESH_TOKEN=your_spotify_refresh_token
```

#### Frontend

Create a `.env` file in the `frontend` directory with the following variables:

```
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=Playlistinator
```

### Running Locally

1. Start the backend server:

```bash
cd backend/playlistinator
go run .
```

2. Start the frontend development server:

```bash
cd frontend
npm install
npm run dev
```

### Building Docker Images

To build the Docker images for the application:

```bash
# Build frontend image
cd frontend
docker build -t playlistinator-frontend:latest .
cd ..

# Build backend image
cd backend/playlistinator
docker build -t playlistinator-backend:latest .
cd ../..
```

## Features

- Fetches your Last.fm listening history
- Analyzes your most played tracks
- Creates a Spotify playlist with your top tracks
- Updates the playlist automatically
- Beautiful and responsive web interface

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 