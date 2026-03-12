package version

type Info struct {
	Version string
	Commit  string
	Date    string
}

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
