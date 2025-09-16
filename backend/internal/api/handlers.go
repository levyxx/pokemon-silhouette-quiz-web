package api

import (
	"encoding/json"
	"image/png"
	"math/rand/v2"
	stdhttp "net/http"
	"strconv"
	"strings"
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
	r.Get("/api/quiz/silhouette/{id}", h.silhouette) // legacy
	r.Get("/api/quiz/silhouette/session/{sessionId}", h.silhouetteBySession)
	r.Get("/api/quiz/artwork/session/{sessionId}", h.artworkBySession)
	r.Get("/api/quiz/hint/{sessionId}", h.hintBySession)
	r.Get("/api/quiz/search", h.search)
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
	ids := make([]int, 0)
	selected := map[string]bool{}
	for _, rk := range req.Regions {
		selected[rk] = true
	}
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
	jpName, _ := h.poke.GetJapaneseName(pick)
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
	if jpName != "" {
		sess.DisplayName = jpName
		sess.AcceptAnswers = append(sess.AcceptAnswers, jpName)
	}
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
	name := sess.PokemonName
	if sess.DisplayName != "" {
		name = sess.DisplayName
	}
	writeJSON(w, resultResponse{PokemonID: sess.PokemonID, Name: name, Types: sess.Types, Region: sess.RegionKey})
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

func (h *Handlers) silhouetteBySession(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	sid := chi.URLParam(r, "sessionId")
	sess, ok := h.store.Get(sid)
	if !ok {
		httpError(w, 404, "session not found")
		return
	}
	img, err := h.poke.GetOfficialArtwork(sess.PokemonID)
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

// artworkBySession returns the original official artwork PNG (color) after quiz finished (solved or gave up)
func (h *Handlers) artworkBySession(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	sid := chi.URLParam(r, "sessionId")
	sess, ok := h.store.Get(sid)
	if !ok {
		httpError(w, 404, "session not found")
		return
	}
	if !sess.Solved && !sess.GaveUp { // forbid early reveal
		httpError(w, 403, "not revealed yet")
		return
	}
	img, err := h.poke.GetOfficialArtwork(sess.PokemonID)
	if err != nil {
		httpError(w, 404, err.Error())
		return
	}
	// re-encode as PNG (original could already be PNG; we just encode our image.Image)
	// Simple approach: use standard library png encoder.
	w.Header().Set("Content-Type", "image/png")
	// encode
	// Avoid importing image/png at top? bring here (already imported via underscore in client elsewhere) need explicit import.
	// We'll add png import at top of file.
	if err := png.Encode(w, img); err != nil {
		httpError(w, 500, err.Error())
		return
	}
}

type hintResponse struct {
	Types       []string `json:"types"`
	Region      string   `json:"region"`
	FirstLetter string   `json:"firstLetter"`
}

var typeJP = map[string]string{
	"normal": "ノーマル", "fire": "ほのお", "water": "みず", "grass": "くさ", "electric": "でんき", "ice": "こおり", "fighting": "かくとう", "poison": "どく", "ground": "じめん", "flying": "ひこう", "psychic": "エスパー", "bug": "むし", "rock": "いわ", "ghost": "ゴースト", "dragon": "ドラゴン", "dark": "あく", "steel": "はがね", "fairy": "フェアリー"}

func regionJP(key string) string {
	for _, r := range poke.Regions {
		if r.Key == key {
			return r.DisplayName
		}
	}
	return key
}

func (h *Handlers) hintBySession(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	sid := chi.URLParam(r, "sessionId")
	sess, ok := h.store.Get(sid)
	if !ok {
		httpError(w, 404, "session not found")
		return
	}
	// first letter: prefer display name (Japanese) else english
	name := sess.DisplayName
	if name == "" {
		name = sess.PokemonName
	}
	first := ""
	for _, r := range name {
		first = string(r)
		break
	}
	tJP := make([]string, 0, len(sess.Types))
	for _, t := range sess.Types {
		if v, ok := typeJP[t]; ok {
			tJP = append(tJP, v)
		} else {
			tJP = append(tJP, t)
		}
	}
	writeJSON(w, hintResponse{Types: tJP, Region: regionJP(sess.RegionKey), FirstLetter: first})
}

// search returns list of candidate names (Japanese if available else English)
func (h *Handlers) search(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	prefix := strings.ToLower(r.URL.Query().Get("prefix"))
	if prefix == "" {
		writeJSON(w, []string{})
		return
	}
	// naive: iterate first 1010 national dex (could cache list)
	limit := 1010
	out := make([]string, 0, 50)
	for id := 1; id <= limit; id++ {
		p, err := h.poke.GetPokemon(id)
		if err != nil {
			continue
		}
		jp, _ := h.poke.GetJapaneseName(id)
		name := p.Name
		if jp != "" {
			name = jp
		}
		if strings.HasPrefix(strings.ToLower(name), prefix) {
			out = append(out, name)
			if len(out) >= 50 {
				break
			}
		}
	}
	writeJSON(w, out)
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
