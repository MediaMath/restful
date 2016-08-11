package restful

import "time"

//Stats is an interface for reporting statistics
type Stats interface {
	Incr(string)
	TimingPeriod(string, time.Time, time.Time)
}

type nilstats struct{}

//NilStats throws stats on the floor
var NilStats = &nilstats{}

//Incr will increment the stat named
func (s *nilstats) Incr(string) {}

//TimingPeriod will report a timing for the stat named and the start/end provided
func (s *nilstats) TimingPeriod(string, time.Time, time.Time) {}
