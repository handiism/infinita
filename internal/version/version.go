// Package version holds the application version string,
// injected at build time via -ldflags.
package version

// Version is set at build time via -ldflags.
// Default value is used when building without version injection.
var Version = "dev"
