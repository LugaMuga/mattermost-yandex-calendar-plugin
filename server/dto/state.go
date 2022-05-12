package dto

type State struct {
	CurrentEvent *Event
}

func DefaultState() *State {
	return &State{
		CurrentEvent: nil,
	}
}
