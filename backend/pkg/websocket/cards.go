package websocket

type Card struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Kind        string `json:"kind"`
	Amount      int    `json:"amount"`
	Target      int    `json:"target"`
	HouseCost   int    `json:"houseCost"`
	HotelCost   int    `json:"hotelCost"`
}

var communityCards = []Card{
	{
		ID:          1,
		Description: "Rücke vor bis zum Start. Ziehe 200 ein.",
		Type:        "move",
		Target:      0,
	},
	{
		ID:          2,
		Description: "Die Bank hat dir versehentlich 200 zu viel überwiesen. Du erhältst 200.",
		Type:        "money",
		Amount:      200,
	},
	{
		ID:          3,
		Description: "Zahle deine Arztrechnung von 50.",
		Type:        "money",
		Amount:      -50,
	},
	{
		ID:          4,
		Description: "Du verkaufst Aktien und erhältst 50.",
		Type:        "money",
		Amount:      50,
	},
	{
		ID:          5,
		Description: "Du kommst aus dem Gefängnis frei. Diese Karte kannst du behalten.",
		Type:        "getOutOfJail",
	},
	{
		ID:          6,
		Description: "Gehe direkt ins Gefängnis. Gehe nicht über Start und ziehe kein Geld ein.",
		Type:        "jail",
	},
	{
		ID:          7,
		Description: "Du erhältst dein Urlaubsgeld von 100.",
		Type:        "money",
		Amount:      100,
	},
	{
		ID:          8,
		Description: "Du erhältst eine Steuerrückerstattung von 20.",
		Type:        "money",
		Amount:      20,
	},
	{
		ID:          9,
		Description: "Es ist dein Geburtstag! Jeder Spieler schenkt dir 10.",
		Type:        "collectEachPlayer",
		Amount:      10,
	},
	{
		ID:          10,
		Description: "Du erhältst 100 aus deiner Lebensversicherung.",
		Type:        "money",
		Amount:      100,
	},
	{
		ID:          11,
		Description: "Du musst deine Spitalrechnung bezahlen. Zahle 100.",
		Type:        "money",
		Amount:      -100,
	},
	{
		ID:          12,
		Description: "Du musst deine Schulgebühren bezahlen. Zahle 50.",
		Type:        "money",
		Amount:      -50,
	},
	{
		ID:          13,
		Description: "Du erhältst ein Beratungshonorar von 25.",
		Type:        "money",
		Amount:      25,
	},
	{
		ID:          14,
		Description: "Du musst Reparaturen an deinen Grundstücken durchführen. Zahle 40 pro Haus und 115 pro Hotel.",
		Type:        "repairs",
		HouseCost:   40,
		HotelCost:   115,
	},
	{
		ID:          15,
		Description: "Du hast den zweiten Preis bei einem Schönheitswettbewerb gewonnen. Du erhältst 10.",
		Type:        "money",
		Amount:      10,
	},
	{
		ID:          16,
		Description: "Du hast geerbt und erhältst 100.",
		Type:        "money",
		Amount:      100,
	},
}

var chanceCards = []Card{
	{
		ID:          1,
		Description: "Rücke vor bis zum Start. Ziehe 200 ein.",
		Type:        "move",
		Target:      0,
	},
	{
		ID:          2,
		Description: "Rücke vor bis nach Genf.",
		Type:        "move",
		Target:      39,
	},
	{
		ID:          3,
		Description: "Rücke vor bis nach Paris.",
		Type:        "move",
		Target:      21,
	},
	{
		ID:          4,
		Description: "Rücke vor bis zum nächsten Versorgungswerk. Falls es jemandem gehört, zahle das Zehnfache des Würfelwurfs.",
		Type:        "nearestUtility",
	},
	{
		ID:          5,
		Description: "Rücke vor bis zum nächsten Flughafen. Zahle dem Besitzer das Doppelte der normalen Miete.",
		Type:        "nearestAirportDouble",
	},
	{
		ID:          6,
		Description: "Rücke vor bis zum nächsten Flughafen. Zahle dem Besitzer das Doppelte der normalen Miete.",
		Type:        "nearestAirportDouble",
	},
	{
		ID:          7,
		Description: "Die Bank zahlt dir eine Dividende von 50.",
		Type:        "money",
		Amount:      50,
	},
	{
		ID:          8,
		Description: "Du kommst aus dem Gefängnis frei. Diese Karte kannst du behalten.",
		Type:        "getOutOfJail",
	},
	{
		ID:          9,
		Description: "Gehe drei Felder zurück.",
		Type:        "moveBack",
		Amount:      3,
	},
	{
		ID:          10,
		Description: "Gehe direkt ins Gefängnis. Gehe nicht über Start und ziehe kein Geld ein.",
		Type:        "jail",
	},
	{
		ID:          11,
		Description: "Mache Reparaturen an deinen Grundstücken. Zahle 25 pro Haus und 100 pro Hotel.",
		Type:        "repairs",
		HouseCost:   25,
		HotelCost:   100,
	},
	{
		ID:          12,
		Description: "Du wurdest wegen zu schnellen Fahrens gebüsst. Zahle 15.",
		Type:        "money",
		Amount:      -15,
	},
	{
		ID:          13,
		Description: "Du wirst zum Verwaltungsratsvorsitzenden ernannt. Zahle jedem Spieler 50.",
		Type:        "payEachPlayer",
		Amount:      50,
	},
	{
		ID:          14,
		Description: "Dein Baudarlehen wird fällig. Du erhältst 150.",
		Type:        "money",
		Amount:      150,
	},
	{
		ID:          15,
		Description: "Rücke vor bis zum Flughafen Zürich.",
		Type:        "move",
		Target:      5,
	},
	{
		ID:          16,
		Description: "Rücke vor bis zum nächsten Versorgungswerk. Falls es jemandem gehört, zahle das Zehnfache des Würfelwurfs.",
		Type:        "nearestUtility",
	},
}

func (pool *Pool) executeCard(client *Client, card Card, player *Player, t *tile) {
	switch card.Type {
	case "money":
		player.Money += card.Amount

	case "move":
		player.Position = card.Target

	case "moveBack":
		player.Position -= card.Amount

		if player.Position < 0 {
			player.Position += len(defaultTiles)
		}

	case "nearestUtility":
		found := false

		for _, p := range defaultTiles {
			if p.Type == "utility" && p.ID > player.Position {
				player.Position = p.ID
				found = true
				break
			}
		}

		if !found {
			for _, p := range defaultTiles {
				if p.Type == "utility" {
					player.Position = p.ID
					break
				}
			}
		}

	case "nearestAirportDouble":
		found := false

		for _, p := range defaultTiles {
			if p.Type == "airport" && p.ID > player.Position {
				player.Position = p.ID
				found = true
				break
			}
		}

		if !found {
			for _, p := range defaultTiles {
				if p.Type == "airport" {
					player.Position = p.ID
					break
				}
			}
		}

	case "jail":
		player.Position = 10

	case "getOutOfJail":
		player.GetOutOfJailCards++

	case "payEachPlayer":
		for _, p := range pool.GameState.Players {
			if p.ID != player.ID {
				p.Money += card.Amount
				player.Money -= card.Amount
			}
		}

	case "collectEachPlayer":
		for _, p := range pool.GameState.Players {
			if p.ID != player.ID {
				p.Money -= card.Amount
				player.Money += card.Amount
			}
		}

	case "repairs":
		houses, hotels := pool.countBuildings(player)

		total := houses*card.HouseCost + hotels*card.HotelCost

		player.Money -= total

	}
}

func (pool *Pool) countBuildings(player *Player) (int, int) {
	houses := 0
	hotels := 0

	for _, p := range pool.GameState.Tiles {
		if p.OwnerID == nil || p.House == nil {
			continue
		}

		if *p.OwnerID != player.ID {
			continue
		}

		if *p.House == 5 {
			hotels++
		} else {
			houses += *p.House
		}
	}

	return houses, hotels
}

func (pool *Pool) handleGetOutOfJail(client *Client, player *Player) {

	if player.GetOutOfJailCards <= 0 {
		return
	}

	if player.Position != 10 {
		return
	}

	player.GetOutOfJailCards--
	player.InJail = false
}
