package logdna

import "errors"

const (
	LogDNAIngestionAPI = "https://logs.logdna.com/logs/ingest"
	DefaultBufferSize  = 100
)

var (
	ErrorAppIsRequired      = errors.New("app is required")
	ErrorAPIKeyIsRequired   = errors.New("API key is required")
	ErrorHostnameIsRequired = errors.New("hostname is required")
)
