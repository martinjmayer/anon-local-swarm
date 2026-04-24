module als/orchestrator

go 1.23

require (
	als/obs v0.0.0
	github.com/google/uuid v1.6.0
	gopkg.in/yaml.v3 v3.0.1
	modernc.org/sqlite v1.34.4
)

replace als/obs => ../obs
