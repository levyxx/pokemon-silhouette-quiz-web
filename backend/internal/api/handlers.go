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
	r.Get("/api/quiz/silhouette/{sessionId}", h.silhouetteBySession)
	r.Get("/api/quiz/artwork/{sessionId}", h.artworkBySession)
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

	baseIDs := make([]int, 0)
	selected := map[string]bool{}
	for _, rk := range req.Regions {
		selected[rk] = true
	}

	allSelected := len(selected) == 0
	for _, rg := range poke.Regions {
		if allSelected || selected[rg.Key] {
			for id := rg.From; id <= rg.To; id++ {
				baseIDs = append(baseIDs, id)
			}
		}
	}

	if len(baseIDs) == 0 {
		httpError(w, 400, "no pokemon range selected")
		return
	}

	// Collect candidate forms (store as struct with PokemonID + display overrides)
	type cand struct {
		id    int
		name  string
		jp    string
		types []string
	}
	candidates := make([]cand, 0, len(baseIDs))
	for _, id := range baseIDs {
		p, err := h.poke.GetPokemon(id)
		if err != nil {
			continue
		}
		// base
		jpName, _ := h.poke.GetJapaneseName(id)
		types := make([]string, 0, len(p.Types))
		for _, t := range p.Types {
			types = append(types, t.Type.Name)
		}
		candidates = append(candidates, cand{id: id, name: p.Name, jp: jpName, types: types})

		// optional: forms (mega/primal) if allowed (and later filtered by region rules)
		if req.AllowMega || req.AllowPrimal {
			sp, err := h.poke.GetSpecies(id)
			if err != nil {
				continue
			}
			for _, v := range sp.Varieties {
				// Skip default (base) already added
				if v.IsDefault {
					continue
				}

				n := v.Pokemon.Name // e.g. "charizard-mega-x"
				lower := strings.ToLower(n)
				isMega := strings.Contains(lower, "mega")
				isPrimal := strings.Contains(lower, "primal") || strings.Contains(lower, "groudon-primal") || strings.Contains(lower, "kyogre-primal")
				if (isMega && !req.AllowMega) || (isPrimal && !req.AllowPrimal) {
					continue
				}
				// Regional form filter (alola/galar/hisui/paldea) -> これらは選択された地方に含まれていない場合除外
				// 名前に "-alola", "-galar", "-hisui", "-paldea" 等を含む場合に該当
				regionalTag := ""
				if strings.Contains(lower, "-alola") {
					regionalTag = "alola"
				}
				if strings.Contains(lower, "-galar") {
					regionalTag = "galar"
				}
				if strings.Contains(lower, "-hisui") {
					regionalTag = "hisui"
				} // Hisui -> 現在 Regions に hisui は無いが将来的拡張考慮
				if strings.Contains(lower, "-paldea") {
					regionalTag = "paldea"
				}
				if regionalTag != "" {
					// Regions に存在しないタグ (hisui) は、現在選択対象にない限り除外（hisui 未サポートのためデフォ除外）
					if !allSelected { // 全選択であれば残す
						if _, ok := selected[regionalTag]; !ok {
							continue
						}
					}
				}
				// Extract id from pokemon URL (ends with /pokemon/{id}/)
				// URL pattern: https://pokeapi.co/api/v2/pokemon/{id}/
				parts := strings.Split(strings.TrimSuffix(v.Pokemon.URL, "/"), "/")
				if len(parts) < 1 {
					continue
				}
				idStr := parts[len(parts)-1]
				formID, err := strconv.Atoi(idStr)
				if err != nil {
					continue
				}
				fp, err := h.poke.GetPokemon(formID)
				if err != nil {
					continue
				}
				fTypes := make([]string, 0, len(fp.Types))
				for _, t := range fp.Types {
					fTypes = append(fTypes, t.Type.Name)
				}
				// Japanese name for form: fallback to base JP if specific not provided (species names are species-level)
				// For now we use base species JP so AcceptAnswers include both base JP and base EN; form-specific english kept.
				candidates = append(candidates, cand{id: formID, name: fp.Name, jp: jpName, types: fTypes})
			}
		}
	}
	if len(candidates) == 0 {
		httpError(w, 500, "no candidates available")
		return
	}
	picked := candidates[rand.IntN(len(candidates))]
	// region determination uses national dex id (base id mapping). For forms, fallback to underlying species base id logic (approx: if larger id, still find by baseIDs inclusion).
	regionKey := ""
	for _, rg := range poke.Regions {
		if rg.ContainsNationalID(picked.id) {
			regionKey = rg.Key
			break
		}
	}

	sess := quiz.NewSession(picked.id, picked.name, regionKey, picked.types, req.AllowMega, req.AllowPrimal)
	if picked.jp != "" {
		sess.DisplayName = picked.jp
		sess.AcceptAnswers = append(sess.AcceptAnswers, picked.jp)
	}
	// Accept base english name if form (strip suffix after last '-')
	if strings.Contains(picked.name, "-") {
		base := picked.name
		if idx := strings.Index(base, "-mega"); idx > 0 {
			base = base[:idx]
		}
		if idx := strings.Index(base, "-primal"); idx > 0 {
			base = base[:idx]
		}
		if idx := strings.Index(base, "-gmax"); idx > 0 {
			base = base[:idx]
		}

		if base != picked.name {
			sess.AcceptAnswers = append(sess.AcceptAnswers, base)
		}
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
