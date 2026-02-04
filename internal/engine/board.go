package engine

import (
	"math/rand"
)

// Stores all information availlable in a game
type GameEngine struct {
	Ticks          int
	BotIDSeed      [2]int
	MaxBots        int
	Grid           [20][20]Tile
	AllBots        map[int]*Bot  //Map of Bot structures, key is its ID
	Scraps         [2]int        // 0 -> player A, 1 -> player B
	Banks          map[int]*Bank //key is bankID
	EnergyPads     map[int]*Pad
	PermanentAlgae [2]int
	Winner         int
	AlgaeCount     int
	Walls          []Point // Added again alongside grid for redundancy and speed
	gl             *GameLogger
}

type Bot struct {
	ID            int      `json:"id"`
	OwnerID       int      `json:"owner_id"` // 0 = Player A, 1 = Player B
	Location      Point    `json:"location"`
	Energy        float64  `json:"energy"`
	Scraps        int      `json:"scraps"`
	Abilities     []string `json:"abilities"`
	VisionRadius  int      `json:"vision_radius"`
	AlgaeHeld     int      `json:"algae_held"`
	TraversalCost float64  `json:"traversal_cost"`
	Status        string   `json:"status"`
}

var CostDB = map[string]int{
	"HARVEST":      10,
	"SCOUT":        10,
	"SELFDESTRUCT": 5,
	"LOCKPICK":     5,
	"SPEEDBOOST":   10,
	"POISON":       5,
	"SHIELD":       5,
}

var EnergyDB = map[string]EnergyCost{
	"HARVEST":      {0, 1},
	"SCOUT":        {1.5, 0}, //Pulse mechanic needs be discussed
	"SELFDESTRUCT": {0.5, 0},
	"SPEEDBOOST":   {1, 0},
	"POISON":       {0.5, 2},
	"LOCKPICK":     {1.5, 0},
	"SHIELD":       {0.25, 0},
	"DEPOSIT":      {0, 1},
	"MOVE":         {0, 0},
}

type EnergyCost struct {
	Traversal float64
	Ability   float64
}

type Tile struct {
	HasAlgae bool
	IsPoison bool
	IsWall   bool
}

type Bank struct {
	ID                int   `json:"id"`
	Location          Point `json:"location"`
	DepositOccuring   bool  `json:"deposit_occuring"`
	DepositAmount     int   `json:"deposit_amount"`
	DepositOwner      int   `json:"deposit_owner"`
	BankOwner         int   `json:"bank_owner"` //0 = player A, 1 = player B
	DepositTicksLeft  int   `json:"deposit_ticks_left"`
	LockPickOccuring  bool  `json:"lockpick_occuring"`
	LockPickTicksLeft int   `json:"lockpick_ticks_left"`
	LockPickBotID     int   `json:"lockpick_botid"`
}

type Pad struct {
	ID        int   `json:"id"`
	Location  Point `json:"location"`
	Available bool  `json:"available"`
	TicksLeft int   `json:"ticks_left"`
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type PermanentEntities struct {
	Banks      map[int]Bank `json:"banks"`
	EnergyPads map[int]Pad  `json:"energy_pads"`
	Walls      []Point      `json:"walls"`
}

// Starts empty game engine instance
func InitGameEngine(gl *GameLogger) *GameEngine {
	ge := &GameEngine{
		Ticks:     1,
		BotIDSeed: [2]int{100, 200},
		MaxBots:   50,
		Grid:      [20][20]Tile{},
		Scraps:    [2]int{},

		AllBots:    make(map[int]*Bot),
		Banks:      make(map[int]*Bank),
		EnergyPads: make(map[int]*Pad),
		Winner:     -1,

		gl: gl,
	}
	ge.Scraps[PlayerOne] = 100
	ge.Scraps[PlayerTwo] = 100

	ge.initBanks()
	ge.initPads()
	ge.generateBoard()
	return ge
}

func (ge *GameEngine) initBanks() {
	ge.Banks[1] = initBank(1, 4, 4, PlayerOne)
	ge.Banks[2] = initBank(2, 15, 4, PlayerTwo)
	ge.Banks[3] = initBank(3, 4, 15, PlayerOne)
	ge.Banks[4] = initBank(4, 15, 15, PlayerTwo)
}

// Need to update Bank structure to have ownership of Banks(can't deposit in enemy bank)
func initBank(id int, x int, y int, playerID int) *Bank {
	return &Bank{
		ID:               id,
		Location:         Point{x, y},
		DepositOccuring:  false,
		DepositAmount:    0,
		BankOwner:        playerID,
		DepositOwner:     -1,
		DepositTicksLeft: 0,
	}
}

func (ge *GameEngine) initPads() {
	ge.EnergyPads[0] = initPad(1, 9, 8)
	ge.EnergyPads[1] = initPad(2, 10, 11)
}

func initPad(id int, x int, y int) *Pad {
	return &Pad{
		ID:        id,
		Location:  Point{x, y},
		Available: true,
		TicksLeft: 0,
	}
}

//overhead is negligible due to just 400 tiles. need to choose random tiles if the board size is increased

func (ge *GameEngine) generateBoard() {
	for x := range BOARDWIDTH {
		for y := range BOARDHEIGHT {
			roll := rand.Float64()
			if ((x == 6 || x == 13) && (y < 6 && y > 2 || y > 13 && y < 17)) || ((y == 6 || y == 13) && (x < 6 && x > 2 || x > 13 && x < 17)) {
				ge.Grid[x][y].IsWall = true
				ge.Walls = append(ge.Walls, Point{x, y})

			} else if roll < 0.15 {
				ge.Grid[x][y].HasAlgae = true
				ge.AlgaeCount++
			} else if roll < 0.20 { //5% chance of poison
				ge.Grid[x][y].HasAlgae = true
				ge.Grid[x][y].IsPoison = true
			}
		}
	}
}
