# cs2-demo-parser

Parse CS2 demo files and view them in an interactive 2D playback viewer.

Upload a `.dem` file and watch the match replay with interpolated player movement, view cones, grenade trajectories and effects, a live killfeed, and team panels with health status.

![r4](https://github.com/user-attachments/assets/7a521d28-cb0c-434f-8350-86d6d0776ebb)

## Tech Stack

- **Backend:** Go 1.25+
- **Frontend:** React + TypeScript, Vite, TailwindCSS
- **Rendering:** HTML5 Canvas

## Local Setup

1. Clone the repo
```bash
git clone https://github.com/allending313/cs2-demo-parser.git
cd cs2-demo-parser
```

2. Start the Go server
```bash
go run cmd/server/main.go
```

3. Start the web viewer
```bash
cd web
npm install
npm run dev
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Space` | Play / Pause |
| `←` | Skip back 5 seconds |
| `→` | Skip forward 5 seconds |
| `↑` | Previous round |
| `↓` | Next round |
| `.` | Cycle playback speed (0.5x / 1x / 2x / 4x) |

## Project Structure

- `cmd/server` - HTTP server for uploading demos and serving the viewer
- `internal/parser` - Demo parsing logic using demoinfocs-golang
- `web/` - React viewer application
- `assets/maps` - CS2 map radar images
- `data/` - Uploaded demos and parsed match data

## Credits

Demo parsing powered by [markus-wa/demoinfocs-golang](https://github.com/markus-wa/demoinfocs-golang)
