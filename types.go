package main

type Mod struct {
	Name     string    `json:"name"`            // mod name
	GUID     string    `json:"guid"`            // mod GUID
	Url      string    `json:"url"`             // mod URL
	Version  string    `json:"current_version"` // mod version
	Releases []Release `json:"releases"`        // mod releases
}

type Release struct {
	Version     string     `json:"version"`   // release version
	ChannelName string     `json:"channel"`   // release channel
	Notes       string     `json:"notes"`     // release notes
	Downloads   []Download `json:"downloads"` // release downloads
}

type Download struct {
	Name   string `json:"name"`   // download name
	Type   string `json:"type"`   // download type (ZIP or DLL)
	Url    string `json:"url"`    // download url
	Sha256 string `json:"sha256"` // download sha256 for verification
}
