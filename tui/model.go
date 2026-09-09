package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/bubbles/progress"
)

type model struct {
	fileName        string
	progress        float32
	speed           float64
	activePeers     int
	totalPeers      int
	bytesDownloaded uint64
	totalSize       uint64
	isDone          bool

	progressBar progress.Model
}

func InitialModel() model {
	return model{
		fileName: "",
		progress: 0,
		isDone:   false,

		progressBar: progress.New(progress.WithDefaultGradient()),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}
