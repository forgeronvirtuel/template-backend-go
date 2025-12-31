package buildinfo

// Version is meant to be overridden at build time via -ldflags.
// Example:
//
//	go build -ldflags "-X template-backend-go/internal/buildinfo.Version=1.2.3" ./...
var Version = "dev"
