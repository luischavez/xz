// Copyright 2014-2022 Ulrich Kunitz. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lzma

// states defines the overall State count
const states = 12

// State maintains the full State of the operation encoding or decoding
// process.
type State struct {
	rep         [4]uint32
	isMatch     [states << maxPosBits]prob
	isRepG0Long [states << maxPosBits]prob
	isRep       [states]prob
	isRepG0     [states]prob
	isRepG1     [states]prob
	isRepG2     [states]prob
	litCodec    literalCodec
	lenCodec    lengthCodec
	repLenCodec lengthCodec
	distCodec   distCodec
	state       uint32
	posBitMask  uint32
	Properties  Properties
}

// initProbSlice initializes a slice of probabilities.
func initProbSlice(p []prob) {
	for i := range p {
		p[i] = probInit
	}
}

// Reset sets all state information to the original values.
func (s *State) Reset() {
	p := s.Properties
	*s = State{
		Properties: p,
		// dict:       s.dict,
		posBitMask: (uint32(1) << uint(p.PB)) - 1,
	}
	initProbSlice(s.isMatch[:])
	initProbSlice(s.isRep[:])
	initProbSlice(s.isRepG0[:])
	initProbSlice(s.isRepG1[:])
	initProbSlice(s.isRepG2[:])
	initProbSlice(s.isRepG0Long[:])
	s.litCodec.init(p.LC, p.LP)
	s.lenCodec.init()
	s.repLenCodec.init()
	s.distCodec.init()
}

// NewState creates a new state from the give Properties.
func NewState(p Properties) *State {
	s := &State{Properties: p}
	s.Reset()
	return s
}

// deepcopy initializes s as a deep copy of the source.
func (s *State) deepcopy(src *State) {
	if s == src {
		return
	}
	s.rep = src.rep
	s.isMatch = src.isMatch
	s.isRepG0Long = src.isRepG0Long
	s.isRep = src.isRep
	s.isRepG0 = src.isRepG0
	s.isRepG1 = src.isRepG1
	s.isRepG2 = src.isRepG2
	s.litCodec.deepcopy(&src.litCodec)
	s.lenCodec.deepcopy(&src.lenCodec)
	s.repLenCodec.deepcopy(&src.repLenCodec)
	s.distCodec.deepcopy(&src.distCodec)
	s.state = src.state
	s.posBitMask = src.posBitMask
	s.Properties = src.Properties
}

// cloneState creates a new clone of the give State.
func cloneState(src *State) *State {
	s := new(State)
	s.deepcopy(src)
	return s
}

// updateStateLiteral updates the State for a literal.
func (s *State) updateStateLiteral() {
	switch {
	case s.state < 4:
		s.state = 0
		return
	case s.state < 10:
		s.state -= 3
		return
	}
	s.state -= 6
}

// updateStateMatch updates the State for a match.
func (s *State) updateStateMatch() {
	if s.state < 7 {
		s.state = 7
	} else {
		s.state = 10
	}
}

// updateStateRep updates the State for a repetition.
func (s *State) updateStateRep() {
	if s.state < 7 {
		s.state = 8
	} else {
		s.state = 11
	}
}

// updateStateShortRep updates the State for a short repetition.
func (s *State) updateStateShortRep() {
	if s.state < 7 {
		s.state = 9
	} else {
		s.state = 11
	}
}

// states computes the states of the operation codec.
func (s *State) states(dictHead int64) (state1, state2, posState uint32) {
	state1 = s.state
	posState = uint32(dictHead) & s.posBitMask
	state2 = (s.state << maxPosBits) | posState
	return
}

// litState computes the literal State.
func (s *State) litState(prev byte, dictHead int64) uint32 {
	lp, lc := uint(s.Properties.LP), uint(s.Properties.LC)
	litState := ((uint32(dictHead) & ((1 << lp) - 1)) << lc) |
		(uint32(prev) >> (8 - lc))
	return litState
}
