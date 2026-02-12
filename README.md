# cs2-demo-parser

Parse CS2 demo files and view them in a 2D playback viewer.

Upload a `.dem` file and watch the match replay with player positions, movement, and round progression.

## Requirements

- Go 1.25+
- Node.js 18+

## Local Setup

1. Clone the repo
```bash
git clone https://github.com/allending313/cs2-demo-parser.git
cd cs2-demo-parser
```

2. Install Go dependencies
```bash
go mod download
```

3. Install and build the web viewer
```bash
cd web
npm install
npm run build
cd ..
```

4. Run the server
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:3001`.


## Project Structure

- `cmd/server` - HTTP server for uploading demos and serving the viewer
- `internal/parser` - Demo parsing logic using demoinfocs-golang
- `web/` - React viewer application
- `assets/maps` - CS2 map radar images
- `data/` - Uploaded demos and parsed match data

## Credits

Demo parsing powered by [markus-wa/demoinfocs-golang](https://github.com/markus-wa/demoinfocs-golang)
