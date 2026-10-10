package tray

type Update struct {
	AvailableVersion chan string
	ShowDialog       chan struct{}
}
