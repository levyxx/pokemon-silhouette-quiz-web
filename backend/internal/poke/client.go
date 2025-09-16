package poke

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"net/http"
	"sync"
	"time"
)

const baseURL = "https://pokeapi.co/api/v2"

type Client struct {
	http *http.Client
	ttl  time.Duration

	mu        sync.RWMutex
	pokemonCh map[int]*cacheEntry[Pokemon]
	spriteCh  map[int]*cacheEntry[image.Image]
	speciesCh map[int]*cacheEntry[Species]
}

type cacheEntry[T any] struct {
	v   T
	exp time.Time
}

type Pokemon struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Types   []PType `json:"types"`
	Sprites Sprites `json:"sprites"`
}

type PType struct {
	Slot int `json:"slot"`
	Type struct {
		Name string `json:"name"`
	} `json:"type"`
}

type Sprites struct {
	Other struct {
		OfficialArtwork struct {
			FrontDefault string `json:"front_default"`
		} `json:"official-artwork"`
	} `json:"other"`
}

func NewClient(ttl time.Duration) *Client {
	return &Client{http: &http.Client{Timeout: 15 * time.Second}, ttl: ttl, pokemonCh: make(map[int]*cacheEntry[Pokemon]), spriteCh: make(map[int]*cacheEntry[image.Image]), speciesCh: make(map[int]*cacheEntry[Species])}
}

func (c *Client) GetPokemon(id int) (Pokemon, error) {
	c.mu.RLock()
	if ce, ok := c.pokemonCh[id]; ok && time.Now().Before(ce.exp) {
		c.mu.RUnlock()
		return ce.v, nil
	}
	c.mu.RUnlock()
	url := fmt.Sprintf("%s/pokemon/%d", baseURL, id)
	resp, err := c.http.Get(url)
	if err != nil {
		return Pokemon{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return Pokemon{}, fmt.Errorf("pokeapi status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Pokemon{}, err
	}
	var p Pokemon
	if err := json.Unmarshal(data, &p); err != nil {
		return Pokemon{}, err
	}
	c.mu.Lock()
	c.pokemonCh[id] = &cacheEntry[Pokemon]{v: p, exp: time.Now().Add(c.ttl)}
	c.mu.Unlock()
	return p, nil
}

func (c *Client) GetOfficialArtwork(id int) (image.Image, error) {
	c.mu.RLock()
	if ce, ok := c.spriteCh[id]; ok && time.Now().Before(ce.exp) {
		c.mu.RUnlock()
		return ce.v, nil
	}
	c.mu.RUnlock()
	p, err := c.GetPokemon(id)
	if err != nil {
		return nil, err
	}
	art := p.Sprites.Other.OfficialArtwork.FrontDefault
	if art == "" {
		return nil, fmt.Errorf("no artwork")
	}
	resp, err := c.http.Get(art)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.spriteCh[id] = &cacheEntry[image.Image]{v: img, exp: time.Now().Add(c.ttl)}
	c.mu.Unlock()
	return img, nil
}

// Species (subset) for name localization
type Species struct {
	Names []struct {
		Language struct {
			Name string `json:"name"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
}

func (c *Client) GetSpecies(id int) (Species, error) {
	c.mu.RLock()
	if ce, ok := c.speciesCh[id]; ok && time.Now().Before(ce.exp) {
		c.mu.RUnlock()
		return ce.v, nil
	}
	c.mu.RUnlock()
	url := fmt.Sprintf("%s/pokemon-species/%d", baseURL, id)
	resp, err := c.http.Get(url)
	if err != nil {
		return Species{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return Species{}, fmt.Errorf("species status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Species{}, err
	}
	var sp Species
	if err := json.Unmarshal(data, &sp); err != nil {
		return Species{}, err
	}
	c.mu.Lock()
	c.speciesCh[id] = &cacheEntry[Species]{v: sp, exp: time.Now().Add(c.ttl)}
	c.mu.Unlock()
	return sp, nil
}

// GetJapaneseName returns a Japanese display name (prefers ja-Hrkt then ja)
func (c *Client) GetJapaneseName(id int) (string, error) {
	sp, err := c.GetSpecies(id)
	if err != nil {
		return "", err
	}
	var ja string
	for _, n := range sp.Names {
		if n.Language.Name == "ja-Hrkt" {
			return n.Name, nil
		}
		if n.Language.Name == "ja" {
			ja = n.Name
		}
	}
	if ja != "" {
		return ja, nil
	}
	return "", fmt.Errorf("japanese name not found")
}
