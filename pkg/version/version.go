package version

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func Info() string {
	return Version + " (" + Commit + ") built at " + BuildTime
}
