package quiz

import (
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	mrand "math/rand/v2"
	"time"
)

var ErrTooSoon = errors.New("guess too soon")
var ErrAlreadyFinished = errors.New("quiz already finished")

// AllowedGuessInterval defines throttle duration
const AllowedGuessInterval = 5 * time.Second

// RandomID picks one id from provided slice
func RandomID(ids []int) int {
	if len(ids) == 0 {
		return 0
	}
	return ids[mrand.IntN(len(ids))]
}

func NewSession(pokemonID int, name string, regionKey string, types []string, allowMega, allowPrimal bool) *Session {
	sid := make([]byte, 8)
	_, _ = crand.Read(sid)
	return &Session{
		ID:          hex.EncodeToString(sid),
		PokemonID:   pokemonID,
		PokemonName: name,
		RegionKey:   regionKey,
		Types:       types,
		StartedAt:   time.Now(),
		LastGuessAt: time.Time{},
		AllowMega:   allowMega,
		AllowPrimal: allowPrimal,
	}
}

// CanGuess enforces 5s interval
func (s *Session) CanGuess() bool {
	if s.Solved || s.GaveUp {
		return false
	}
	if s.LastGuessAt.IsZero() {
		return true
	}
	return time.Since(s.LastGuessAt) >= AllowedGuessInterval
}

// SubmitGuess update state
func (s *Session) SubmitGuess(answer string) (correct bool, err error) {
	if s.Solved || s.GaveUp {
		return false, ErrAlreadyFinished
	}
	if !s.CanGuess() {
		return false, ErrTooSoon
	}
	s.LastGuessAt = time.Now()
	normalized := normalize(answer)
	// base english name
	if normalized == normalize(s.PokemonName) {
		s.Solved = true
		return true, nil
	}
	// display japanese
	if s.DisplayName != "" && normalized == normalize(s.DisplayName) {
		s.Solved = true
		return true, nil
	}
	// additional accepted answers
	for _, a := range s.AcceptAnswers {
		if normalized == normalize(a) {
			s.Solved = true
			return true, nil
		}
	}
	return false, nil
}

func (s *Session) GiveUp() { s.GaveUp = true }

func normalize(s string) string {
	// simple lower ascii; could expand (e.g. remove hyphens)
	b := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '-' || r == ' ' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			r = r - 'A' + 'a'
		}
		b = append(b, r)
	}
	return string(b)
}
