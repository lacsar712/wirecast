package model

type DryState string

const (
	DryIdle    DryState = "idle"
	DryHeating DryState = "heating"
	DryHold    DryState = "hold"
	DryCool    DryState = "cool"
)
