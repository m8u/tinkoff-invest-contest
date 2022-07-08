package bots

type Bot interface {
	Serve()
	TogglePause()
	IsPaused() bool
	IsStarted() bool
	Remove()
}
