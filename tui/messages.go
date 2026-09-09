package tui

type FileSelectedMsg struct {
	FileName      string
	TotalFileSize uint64
}

type StatsUpdateMsg struct {
	Progress        float32
	Speed           float64
	BytesDownloaded uint64
	ActivePeers     int
	TotalPeers      int
}

type DownloadCompleteMsg struct{}
