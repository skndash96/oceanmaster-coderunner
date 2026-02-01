package engine

import (
    "math/rand"
)

// Stores all information availlable in a game
type GameEngine struct {
    Ticks          int
    Grid           [20][20]Tile
    AllBots        map[int]*Bot  //Map of Bot structures, key is its ID
    Scraps         [2]int        // 0 -> player A, 1 -> player B
    Banks          map[int]*Bank //key is bankID
    EnergyPads     map[int]*Pad
    PermanentAlgae [2]int
    Winner         int
    AlgaeCount     int

    gl *GameLogger
}

// NEED TO FIX PLAYERID AS EITHER NUMBER OR STRING
// NEED TO FIX THE USE OF X AND Y SOMEWHERE AND Point ELSEWHERE
type Bot struct {
    ID      int
    OwnerID int // 0 = Player A, 1 = Player B
    //    X, Y          int
    Location      Point
    Energy        float64
    Scraps        int
    Abilities     []string
    VisionRadius  int //Can be removed as vision is mapwide now
    AlgaeHeld     int
    TraversalCost float64
}

// adding a constant movement cost to bot value. to not need to calculate it on every move

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
    "HARVEST":      EnergyCost{0, 1},
    "SCOUT":        EnergyCost{1.5, 0}, //Pulse mechanic needs be discussed
    "SELFDESTRUCT": EnergyCost{0.5, 0},
    "SPEEDBOOST":   EnergyCost{1, 0},
    "POISON":       EnergyCost{0.5, 2},
    "LOCKPICK":     EnergyCost{1.5, 0},
    "SHIELD":       EnergyCost{0.25, 0},
    "DEPOSIT":      EnergyCost{0, 1},
}

type EnergyCost struct {
    Traversal float64
    Ability   float64
}

type Tile struct {
    HasAlgae bool
    IsPoison bool
}

type Bank struct {
    ID       int `json:"id"`
    Location Point `json:"location"`
    //    X, Y             int
    DepositOccuring  bool `json:"deposit_occuring"`
    DepositAmount    int `json:"deposit_amount"`
    DepositOwner     int `json:"deposit_owner"`
    BankOwner        int `json:"bank_owner"` //0 = player A, 1 = player B
    DepositTicksLeft int `json:"deposit_ticks_left"`
}

type Pad struct {
    ID       int `json:"id"`
    Location Point `json:"location"`
    // X, Y       int
    Available bool `json:"available"`
    TicksLeft int `json:"ticks_left"`
}

type Point struct {
    X int `json:"x"`
    Y int `json:"y"`
}

type GameView struct {
    Tick              int               `json:"tick"`
    Scraps            [2]int            `json:"scraps"`
    AllBots           map[int]Bot       `json:"bots"`
    Algae             [2]int            `json:"algae_count"`
    BotCount          int               `json:"bot_count"`
    MaxBots           int               `json:"max_bots"`
    Width             int               `json:"width"`
    Height            int               `json:"height"`
    PermanentEntities PermanentEntities `json:"permanent_entities"`
    AlgaeMap          []VisibleAlgae    `json:"algae"`
}

type PlayerView struct {
    Tick              int               `json:"tick"`   //json tag
    Scraps            int               `json:"scraps"` //e.g value of Scraps variable will be set to value of scraps in json
    Algae             int               `json:"algae"`
    BotCount          int               `json:"bot_count"`
    MaxBots           int               `json:"max_bots"`
    Width             int               `json:"width"`
    Height            int               `json:"height"`
    Bots              map[int]Bot       `json:"bots"`
    VisibleEntities   VisibleEntities   `json:"visible_entities"`
    PermanentEntities PermanentEntities `json:"permanent_entities"`
}

type VisibleEntities struct {
    Enemies []EnemyBot     `json:"enemies"`
    Algae   []VisibleAlgae `json:"algae"`
}

type EnemyBot struct {
    ID        int      `json:"id"`
    Location  Point    `json:"location"`
    Scraps    int      `json:"scraps"`
    Abilities []string `json:"abilities"`
}

type VisibleAlgae struct {
    Location Point `json:"location"`
    // X        int    `json:"x"`
    // Y        int    `json:"y"`
    IsPoison string `json:"is_poison"`
}

type PermanentEntities struct {
    Banks      map[int]Bank `json:"banks"`
    EnergyPads map[int]Pad  `json:"energy_pads"`
}

type PlayerMoves struct {
    Tick    int               `json:"tick"`
    Spawns  map[int]SpawnCmd  `json:"spawns"`
    Actions map[int]ActionCmd `json:"actions"`
}

type SpawnCmd struct {
    Abilities []string `json:"template"`
    Location  Point    `json:"loc"`
}

type ActionCmd struct {
    Action    string `json:"action"`
    Direction string `json:"direction"`
}

const (
    PlayerOne = iota
    PlayerTwo
)

// Starts empty game engine instance
func InitGameEngine(gl *GameLogger) *GameEngine {
    ge := &GameEngine{
        Ticks:  1,
        Grid:   [20][20]Tile{},
        Scraps: [2]int{},

        AllBots:    make(map[int]*Bot),
        Banks:      make(map[int]*Bank),
        EnergyPads: make(map[int]*Pad),
        Winner:     -1,

        gl: gl,
    }
    ge.Scraps[0] = 100
    ge.Scraps[1] = 100

    ge.initBanks()
    ge.initPads()
    ge.generateAlgae()
    return ge
}

func (ge *GameEngine) initBanks() {

    ge.Banks[1] = initBank(1, 4, 4, 0)
    ge.Banks[2] = initBank(2, 14, 4, 1)
    ge.Banks[3] = initBank(3, 4, 14, 0)
    ge.Banks[4] = initBank(4, 14, 14, 1)
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
    ge.EnergyPads[0] = initPad(1, 9, 6)
    ge.EnergyPads[1] = initPad(2, 9, 13)
}

func initPad(id int, x int, y int) *Pad {
    return &Pad{
        ID:        id,
        Location:  Point{x, y},
        Available: true,
        TicksLeft: BasePadCoolDown,
    }
}

//overhead is negligible due to just 400 tiles. need to choose random tiles if the board size is increased

func (ge *GameEngine) generateAlgae() {
    for x := range BOARDWIDTH {
        for y := range BOARDHEIGHT {
            roll := rand.Float64()
            if roll < 0.15 {
                ge.Grid[x][y].HasAlgae = true
                ge.AlgaeCount++
            } else if roll < 0.20 { //5% chance of poison
                ge.Grid[x][y].HasAlgae = true
                ge.Grid[x][y].IsPoison = true
            }
        }
    }
}
