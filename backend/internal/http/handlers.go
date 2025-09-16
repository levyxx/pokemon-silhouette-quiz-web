package http

import (
	"encoding/json"
	"math/rand/v2"
	stdhttp "net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/levyxx/pokemon-silhouette-quiz/backend/internal/poke"
	"github.com/levyxx/pokemon-silhouette-quiz/backend/internal/quiz"
)

type Handlers struct {
	poke  *poke.Client
	store *quiz.Store
}

func NewHandlers(p *poke.Client, s *quiz.Store) *Handlers { return &Handlers{poke: p, store: s} }

func (h *Handlers) Register(r chi.Router) {
	r.Get("/health", func(w stdhttp.ResponseWriter, r *stdhttp.Request) { w.Write([]byte("ok")) })
	r.Post("/api/quiz/start", h.startQuiz)
	r.Post("/api/quiz/guess", h.guess)
	r.Post("/api/quiz/giveup", h.giveup)
	r.Get("/api/quiz/silhouette/{id}", h.silhouette)
}

type startRequest struct {
	Regions     []string `json:"regions"`
	AllowMega   bool     `json:"allowMega"`
	AllowPrimal bool     `json:"allowPrimal"`
}
type startResponse struct {
	SessionID string `json:"sessionId"`
}

func (h *Handlers) startQuiz(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req startRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, 400, err.Error())
		return
	}
	// collect candidate IDs from selected regions
	ids := make([]int, 0)
	selected := map[string]bool{}
	for _, rk := range req.Regions {
		selected[rk] = true
	}
	// if none selected treat as all
	allSelected := len(selected) == 0
	for _, rg := range poke.Regions {
		if allSelected || selected[rg.Key] {
			for id := rg.From; id <= rg.To; id++ {
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		httpError(w, 400, "no pokemon range selected")
		return
	}
	pick := ids[rand.IntN(len(ids))]
	p, err := h.poke.GetPokemon(pick)
	if err != nil {
		httpError(w, 502, err.Error())
		return
	}
	types := make([]string, 0, len(p.Types))
	for _, t := range p.Types {
		types = append(types, t.Type.Name)
	}
	regionKey := ""
	for _, rg := range poke.Regions {
		if rg.ContainsNationalID(pick) {
			regionKey = rg.Key
			break
		}
	}
	sess := quiz.NewSession(pick, p.Name, regionKey, types, req.AllowMega, req.AllowPrimal)
	h.store.Set(sess)
	writeJSON(w, startResponse{SessionID: sess.ID})
}

type guessRequest struct {
	SessionID string `json:"sessionId"`
	Answer    string `json:"answer"`
}
type guessResponse struct {
	Correct    bool `json:"correct"`
	Solved     bool `json:"solved"`
	RetryAfter int  `json:"retryAfter"`
}

func (h *Handlers) guess(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req guessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, 400, err.Error())
		return
	}
	sess, ok := h.store.Get(req.SessionID)
	if !ok {
		httpError(w, 404, "session not found")
		return
	}
	correct, err := sess.SubmitGuess(req.Answer)
	if err == quiz.ErrTooSoon {
		remaining := int((quiz.AllowedGuessInterval - time.Since(sess.LastGuessAt)).Seconds())
		writeJSON(w, guessResponse{Correct: false, Solved: false, RetryAfter: remaining})
		return
	}
	if err == quiz.ErrAlreadyFinished {
		writeJSON(w, guessResponse{Correct: false, Solved: true})
		return
	}
	writeJSON(w, guessResponse{Correct: correct, Solved: sess.Solved})
}

type giveupRequest struct {
	SessionID string `json:"sessionId"`
}
type resultResponse struct {
	PokemonID int      `json:"pokemonId"`
	Name      string   `json:"name"`
	Types     []string `json:"types"`
	Region    string   `json:"region"`
}

func (h *Handlers) giveup(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req giveupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, 400, err.Error())
		return
	}
	sess, ok := h.store.Get(req.SessionID)
	if !ok {
		httpError(w, 404, "session not found")
		return
	}
	sess.GiveUp()
	writeJSON(w, resultResponse{PokemonID: sess.PokemonID, Name: sess.PokemonName, Types: sess.Types, Region: sess.RegionKey})
}

func (h *Handlers) silhouette(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)
	img, err := h.poke.GetOfficialArtwork(id)
	if err != nil {
		httpError(w, 404, err.Error())
		return
	}
	data, err := poke.ToSilhouette(img)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(data)
}

func writeJSON(w stdhttp.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func httpError(w stdhttp.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
