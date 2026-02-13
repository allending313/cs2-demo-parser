package cs2demoparser

import "embed"

//go:embed all:web/dist
var WebFS embed.FS

//go:embed all:assets/maps
var MapsFS embed.FS
