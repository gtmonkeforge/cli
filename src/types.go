package main

type Mod struct {
	Name     string    `json:"name"`            // mod name
	GUID     string    `json:"guid"`            // mod GUID
	Url      string    `json:"url"`             // mod URL
	Version  string    `json:"current_version"` // mod version
	Releases []Release `json:"releases"`        // mod releases
	Channels []Channel `json:"channels"`
}

type Releases struct {
	Items []Release `json:"items"`
}

type Release struct {
	Version     string     `json:"version"`   // release version
	ChannelName string     `json:"channel"`   // release channel
	Notes       string     `json:"notes"`     // release notes
	Downloads   []Download `json:"downloads"` // release downloads
}

type Download struct {
	Name   string `json:"name"`   // download name
	Type   string `json:"kind"`   // download type (ZIP or DLL)
	Url    string `json:"url"`    // download url
	Sha256 string `json:"sha256"` // download sha256 for verification
	Size   int64  `json:"size"`   // download size in bytes
}

type Channel struct {
	Name    string `json:"name"`
	Version string `json:"current_version"`
}
