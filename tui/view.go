package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	appBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	fileStyle = lipgloss.NewStyle().
			Bold(true)

	statsBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#5d5d5d")).
			Padding(1).
			Italic(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04bb78")).
			Bold(true)
)

func formatBytes(bytes uint64) string {
	const (
		MB = 1024 * 1024
		GB = 1024 * 1024 * 1024
	)

	if bytes >= GB {
		valueInGB := float64(bytes) / float64(GB)
		return fmt.Sprintf("%.2f GB", valueInGB)
	}

	valueInMB := float64(bytes) / float64(MB)
	return fmt.Sprintf("%.2f MB", valueInMB)
}

func (m model) View() tea.View {
	title := titleStyle.Render("GO-BITTORRENT CLIENT")
	fileInfo := fmt.Sprintf("File: %s", fileStyle.Render(m.fileName))
	help := "Press 'q' to quit."

	var mainContent string

	if m.isDone {
		successMsg := successStyle.Render("DOWNLOAD SUCCESSFUL!")

		mainContent = lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			fileInfo,
			successMsg,
			help,
		)
	} else {
		progressBar := "Progress:\n" + m.progressBar.ViewAs(float64(m.progress)/100.0)

		downloadedStr := formatBytes(m.bytesDownloaded)
		totalSizeStr := formatBytes(m.totalSize)
		statsText := fmt.Sprintf("Speed: %.2f MB/s\n%s / %s downloaded\nPeers: %v connected (out of %v)", m.speed, downloadedStr, totalSizeStr, m.activePeers, m.totalPeers)

		statsBox := statsBoxStyle.Render(statsText)

		mainContent = lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			fileInfo,
			"",
			progressBar,
			statsBox,
			help,
		)
	}

	return tea.NewView(appBoxStyle.Render(mainContent))
}
