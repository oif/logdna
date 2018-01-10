package logdna

type Config struct {
	// required
	App      string
	APIKey   string
	Hostname string

	// optional
	Mac        string
	IP         string
	Tags       []string
	BufferSize uint32
}

func (c Config) Validate() error {
	if c.App == "" {
		return ErrorAppIsRequired
	}
	if c.APIKey == "" {
		return ErrorAPIKeyIsRequired
	}
	if c.Hostname == "" {
		return ErrorHostnameIsRequired
	}
	return nil
}
