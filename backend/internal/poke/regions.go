package poke

// Region represents a pokedex region/generation toggle
type Region struct {
	Key         string // e.g., "kanto"
	DisplayName string // 日本語表示
	Generation  int
	// ID range inclusive (national dex)
	From int
	To   int
}

var Regions = []Region{
	{Key: "kanto", DisplayName: "カントー", Generation: 1, From: 1, To: 151},
	{Key: "johto", DisplayName: "ジョウト", Generation: 2, From: 152, To: 251},
	{Key: "hoenn", DisplayName: "ホウエン", Generation: 3, From: 252, To: 386},
	{Key: "sinnoh", DisplayName: "シンオウ", Generation: 4, From: 387, To: 493},
	{Key: "unova", DisplayName: "イッシュ", Generation: 5, From: 494, To: 649},
	{Key: "kalos", DisplayName: "カロス", Generation: 6, From: 650, To: 721},
	{Key: "alola", DisplayName: "アローラ", Generation: 7, From: 722, To: 809},
	{Key: "galar", DisplayName: "ガラル", Generation: 8, From: 810, To: 905},
	{Key: "paldea", DisplayName: "パルデア", Generation: 9, From: 906, To: 1010},
}

// ContainsNationalID returns true if the id is inside region range
func (r Region) ContainsNationalID(id int) bool { return id >= r.From && id <= r.To }
