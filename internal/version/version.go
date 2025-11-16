package version

// These variables can be overridden at build time using -ldflags, e.g.:
//   go build -ldflags "-X github.com/HrodWolfS/snipster/internal/version.Version=v0.1.0 -X github.com/HrodWolfS/snipster/internal/version.Commit=abcdef -X github.com/HrodWolfS/snipster/internal/version.Date=2025-11-16"
var (
    // Version is the semantic version of the build (e.g. v0.1.0). Defaults to dev.
    Version = "dev"
    // Commit is the git commit hash for this build.
    Commit  = ""
    // Date is the build date in ISO-8601.
    Date    = ""
)
