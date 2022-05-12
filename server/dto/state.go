package dto

type State struct {
	InEvent     bool
	SavedStatus string
}

func DefaultState() *State {
	return &State{
		InEvent:     false,
		SavedStatus: "",
	}
}
