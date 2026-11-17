package websocket

// The board: tile types and the 40 tiles.

type board struct {
	size  int
	Tiles []tile `json:"tiles"`
}

func NewBoard() *board {
	return &board{
		size:  40,
		Tiles: make([]tile, 40),
	}
}

type tile struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`              // "start", "property", "airport", "utility", "tax", "chance", "community", "jail", "vacation", "goToJail"
	Country    string  `json:"country,omitempty"` // only for properties
	Price      int     `json:"price,omitempty"`
	BaseTax    *int    `json:"baseTax,omitempty"`
	OwnerID    *string `json:"ownerId,omitempty"`
	Tax        *int    `json:"tax,omitempty"`
	Level      *int    `json:"level,omitempty"`
	House      *int    `json:"house,omitempty"`
	HousePrice *int    `json:"housePrice,omitempty"`
}

// newTiles returns a fresh copy of the board for one room, so ownership in one
// room doesn't leak into others.
func newTiles() []tile {
	return append([]tile(nil), defaultTiles...)
}

// defaultTiles is the board, in order. Fields: ID, Name, Type, Country, Price, OwnerID, Tax, Level.
// Never change it directly; each room works on its own copy (Pool.Tiles).

func intPtr(i int) *int {
	return &i
}

	var defaultTiles = []tile{
		{0, "Start", "start", "", 0, nil, nil, nil, nil, nil, nil},
		{1, "Athen", "property", "Griechenland", 60, nil, nil, nil, nil, nil, intPtr(50)},
		{2, "Gemeinschaftskarte", "community", "", 0, nil, nil, nil, nil, nil, nil},
		{3, "Thessaloniki", "property", "Griechenland", 60, nil, nil, nil, nil, nil, intPtr(50)},
		{4, "Einkommenssteuer", "tax", "", 0, nil, nil, nil, nil, nil, nil},
		{5, "Flughafen Zürich", "airport", "", 200, nil, nil, nil, nil, nil, nil},
		{6, "Rom", "property", "Italien", 100, nil, nil, nil, nil, nil, intPtr(50)},
		{7, "Ereigniskarte", "chance", "", 0, nil, nil, nil, nil, nil, nil},
		{8, "Mailand", "property", "Italien", 100, nil, nil, nil, nil, nil, intPtr(50)},
		{9, "Venedig", "property", "Italien", 120, nil, nil, nil, nil, nil, intPtr(50)},

		{10, "Gefängnis", "jail", "", 0, nil, nil, nil, nil, nil, nil},
		{11, "Madrid", "property", "Spanien", 140, nil, nil, nil, nil, nil, intPtr(100)},
		{12, "Elektrizitätswerk", "utility", "", 150, nil, nil, nil, nil, nil, nil},
		{13, "Barcelona", "property", "Spanien", 140, nil, nil, nil, nil, nil, intPtr(100)},
		{14, "Sevilla", "property", "Spanien", 160, nil, nil, nil, nil, nil, intPtr(100)},
		{15, "Flughafen Frankfurt", "airport", "", 200, nil, nil, nil, nil, nil, nil},
		{16, "Berlin", "property", "Deutschland", 180, nil, nil, nil, nil, nil, intPtr(100)},
		{17, "Gemeinschaftskarte", "community", "", 0, nil, nil, nil, nil, nil, nil},
		{18, "München", "property", "Deutschland", 180, nil, nil, nil, nil, nil, intPtr(100)},
		{19, "Hamburg", "property", "Deutschland", 200, nil, nil, nil, nil, nil, intPtr(100)},

		{20, "Ferien", "vacation", "", 0, nil, nil, nil, nil, nil, nil},
		{21, "Paris", "property", "Frankreich", 220, nil, nil, nil, nil, nil, intPtr(150)},
		{22, "Ereigniskarte", "chance", "", 0, nil, nil, nil, nil, nil, nil},
		{23, "Lyon", "property", "Frankreich", 220, nil, nil, nil, nil, nil, intPtr(150)},
		{24, "Marseille", "property", "Frankreich", 240, nil, nil, nil, nil, nil, intPtr(150)},
		{25, "Flughafen Paris", "airport", "", 200, nil, nil, nil, nil, nil, nil},
		{26, "London", "property", "Grossbritannien", 260, nil, nil, nil, nil, nil, intPtr(150)},
		{27, "Manchester", "property", "Grossbritannien", 260, nil, nil, nil, nil, nil, intPtr(150)},
		{28, "Wasserwerk", "utility", "", 150, nil, nil, nil, nil, nil, nil},
		{29, "Liverpool", "property", "Grossbritannien", 280, nil, nil, nil, nil, nil, intPtr(150)},

		{30, "Geh ins Gefängnis", "goToJail", "", 0, nil, nil, nil, nil, nil, nil},
		{31, "Tokio", "property", "Japan", 300, nil, nil, nil, nil, nil, intPtr(200)},
		{32, "Osaka", "property", "Japan", 300, nil, nil, nil, nil, nil, intPtr(200)},
		{33, "Gemeinschaftskarte", "community", "", 0, nil, nil, nil, nil, nil, nil},
		{34, "Kyoto", "property", "Japan", 320, nil, nil, nil, nil, nil, intPtr(200)},
		{35, "Flughafen New York", "airport", "", 200, nil, nil, nil, nil, nil, nil},
		{36, "Ereigniskarte", "chance", "", 0, nil, nil, nil, nil, nil, nil},
		{37, "Zürich", "property", "Schweiz", 350, nil, nil, nil, nil, nil, intPtr(200)},
		{38, "Luxussteuer", "tax", "", 0, nil, nil, nil, nil, nil, nil},
		{39, "Genf", "property", "Schweiz", 400, nil, nil, nil, nil, nil, intPtr(200)},
	}
