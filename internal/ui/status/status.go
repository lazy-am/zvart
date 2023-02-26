package status

import "time"

type Status struct {
	defStatus    string
	tempStatus   string
	tempStTime   time.Time
	tempStDurSec int
	needCalc     bool
	Rebuild      chan bool

	lastNotes time.Time
	notesDur  int
}

func Build(def string) *Status {
	r := make(chan bool, 2)
	r <- true
	return &Status{defStatus: def,
		needCalc:  false,
		Rebuild:   r,
		lastNotes: time.Now()}
}

func (s *Status) Get() string {
	if s.needCalc && int(time.Now().Sub(s.tempStTime).Seconds()) < s.tempStDurSec {
		return s.tempStatus
	}
	s.needCalc = false
	return s.defStatus
}

func (s *Status) Set(st string, durSec int) {
	s.needCalc = true
	s.tempStatus = st
	s.tempStDurSec = durSec
	s.tempStTime = time.Now()
	s.Rebuild <- true
}

func (s *Status) Notes(notes <-chan string) {
	if (len(notes) > 0) && int(time.Now().Sub(s.lastNotes).Seconds()) < s.notesDur {
		st := <-notes
		s.Set(st, s.notesDur)
		s.lastNotes = time.Now()
	}
}
