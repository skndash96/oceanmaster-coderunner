// this contains code that comes in and goes out to sandbox
// objective is to not leak things like playerID as 0/1 to the sandbox. why? elegance.
package engine

// Sandbox sends this
type PlayerMoves struct {
	Tick    int               `json:"tick"`
	Spawns  map[int]SpawnCmd  `json:"spawns"`
	Actions map[int]ActionCmd `json:"actions"`
}

type SpawnCmd struct {
	Abilities []string `json:"abilities"`
	Location  Point    `json:"location"`
}

type ActionCmd struct {
	Action    string `json:"action"`
	Direction string `json:"direction"`
}

// Logged to Game Log
type GameViewDTO struct {
	Tick              int               `json:"tick"`
	Scraps            [2]int            `json:"scraps"`
	AllBots           map[int]Bot       `json:"bots"`
	Algae             [2]int            `json:"algae_count"`
	BotCount          int               `json:"bot_count"`
	MaxBots           int               `json:"max_bots"`
	Width             int               `json:"width"`
	Height            int               `json:"height"`
	PermanentEntities PermanentEntities `json:"permanent_entities"`
	AlgaeMap          []VisibleAlgaeDTO `json:"algae"`
}

func (engine *GameEngine) getGameView() GameViewDTO {
	allBots := make(map[int]Bot, 0)
	for _, bot := range engine.AllBots {
		allBots[bot.ID] = *bot
	}

	Banks := make(map[int]Bank, 0)
	for _, bank := range engine.Banks {
		Banks[bank.ID] = *bank
	}

	Pads := make(map[int]Pad, 0)
	for _, pad := range engine.EnergyPads {
		Pads[pad.ID] = *pad
	}

	visibleAlgae := make([]VisibleAlgaeDTO, 0)
	for x := range BOARDWIDTH {
		for y := range BOARDHEIGHT {
			tile := engine.Grid[x][y]
			if tile.HasAlgae {
				poisonStatus := ""
				if tile.IsPoison {
					poisonStatus = "TRUE"
				} else {
					poisonStatus = "FALSE"
				}

				algae := VisibleAlgaeDTO{
					Location: Point{x, y},
					IsPoison: poisonStatus,
				}
				visibleAlgae = append(visibleAlgae, algae)
			}
		}
	}

	return GameViewDTO{
		Tick:     engine.Ticks,
		Scraps:   [2]int{engine.Scraps[PlayerOne], engine.Scraps[PlayerTwo]},
		Algae:    [2]int{engine.PermanentAlgae[PlayerOne], engine.PermanentAlgae[PlayerTwo]},
		BotCount: len(engine.AllBots),
		MaxBots:  engine.MaxBots,
		Width:    BOARDWIDTH,
		Height:   BOARDHEIGHT,
		AllBots:  allBots,
		PermanentEntities: PermanentEntities{
			Banks:      Banks,
			EnergyPads: Pads,
			Walls:      engine.Walls,
		},
		AlgaeMap: visibleAlgae,
	}
}

// Sent out to sandbox
type PlayerViewDTO struct {
	Tick              int                  `json:"tick"`
	Scraps            int                  `json:"scraps"` //e.g value of Scraps variable will be set to value of scraps in json
	Algae             int                  `json:"algae"`
	BotIDSeed         int                  `json:"bot_id_seed"`
	MaxBots           int                  `json:"max_bots"`
	Width             int                  `json:"width"`
	Height            int                  `json:"height"`
	Bots              map[int]PlayerBotDTO `json:"bots"` // THINK: this only refers to the bots player's bots right?
	VisibleEntities   VisibleEntitiesDTO   `json:"visible_entities"`
	PermanentEntities PermanentEntitiesDTO `json:"permanent_entities"`
}

type VisibleAlgaeDTO struct {
	Location Point  `json:"location"`
	IsPoison string `json:"is_poison"`
}

type PlayerBotDTO struct {
	ID            int      `json:"id"`
	Location      Point    `json:"location"`
	Energy        float64  `json:"energy"`
	Scraps        int      `json:"scraps"`
	Abilities     []string `json:"abilities"`
	AlgaeHeld     int      `json:"algae_held"`
	TraversalCost float64  `json:"traversal_cost"`
	Status        string   `json:"status"`
	VisionRadius  int      `json:"vision_radius"`
}

type EnemyBotDTO struct {
	ID        int      `json:"id"`
	Location  Point    `json:"location"`
	Scraps    int      `json:"scraps"`
	Abilities []string `json:"abilities"`
}
type BankDTO struct {
	ID                int   `json:"id"`
	Location          Point `json:"location"`
	DepositOccuring   bool  `json:"deposit_occuring"`
	DepositAmount     int   `json:"deposit_amount"`
	IsDepositOwner    bool  `json:"is_deposit_owner"`
	IsBankOwner       bool  `json:"is_bank_owner"`
	DepositTicksLeft  int   `json:"deposit_ticks_left"`
	LockPickOccuring  bool  `json:"lockpick_occuring"`
	LockPickTicksLeft int   `json:"lockpick_ticks_left"`
	LockPickBotID     int   `json:"lockpick_botid"`
}

type PadDTO Pad

type PermanentEntitiesDTO struct {
	Banks      map[int]BankDTO `json:"banks"`
	EnergyPads map[int]PadDTO  `json:"energy_pads"`
	Walls      []Point         `json:"walls"`
}

type VisibleEntitiesDTO struct {
	Enemies []EnemyBotDTO     `json:"enemies"`
	Algae   []VisibleAlgaeDTO `json:"algae"`
}

func (engine *GameEngine) GetPlayerView(playerID int) PlayerViewDTO {
	playerBots := make(map[int]PlayerBotDTO)

	for _, bot := range engine.AllBots {
		if bot.OwnerID == playerID {
			playerBots[bot.ID] = PlayerBotDTO{
				ID:            bot.ID,
				Location:      bot.Location,
				Energy:        bot.Energy,
				Scraps:        bot.Scraps,
				Abilities:     bot.Abilities,
				VisionRadius:  bot.VisionRadius,
				AlgaeHeld:     bot.AlgaeHeld,
				TraversalCost: bot.TraversalCost,
				Status:        bot.Status,
			}
		}
	}

	// ---- Banks ----
	banks := make(map[int]BankDTO)
	for _, bank := range engine.Banks {
		banks[bank.ID] = BankDTO{
			ID:                bank.ID,
			Location:          bank.Location,
			DepositOccuring:   bank.DepositOccuring,
			DepositAmount:     bank.DepositAmount,
			IsDepositOwner:    bank.DepositOwner == playerID,
			IsBankOwner:       bank.BankOwner == playerID,
			DepositTicksLeft:  bank.DepositTicksLeft,
			LockPickOccuring:  bank.LockPickOccuring,
			LockPickTicksLeft: bank.LockPickTicksLeft,
			LockPickBotID:     bank.LockPickBotID,
		}
	}

	// ---- Energy pads ----
	pads := make(map[int]PadDTO)
	for _, pad := range engine.EnergyPads {
		pads[pad.ID] = PadDTO(*pad) // PadDTO is alias of Pad
	}

	// ---- Assemble final view ----
	return PlayerViewDTO{
		Tick:      engine.Ticks,
		BotIDSeed: engine.BotIDSeed[playerID],
		Scraps:    engine.Scraps[playerID],
		Algae:     engine.PermanentAlgae[playerID],
		MaxBots:   engine.MaxBots,
		Width:     BOARDWIDTH,
		Height:    BOARDHEIGHT,

		Bots: playerBots,

		VisibleEntities: engine.calculateVisibleEntities(playerID),

		PermanentEntities: PermanentEntitiesDTO{
			Banks:      banks,
			EnergyPads: pads,
			Walls:      engine.Walls,
		},
	}
}

func (engine *GameEngine) calculateVisibleEntities(playerID int) VisibleEntitiesDTO {
	visibleEnemies := make([]EnemyBotDTO, 0)
	visibleAlgae := make([]VisibleAlgaeDTO, 0)

	//    canSee := [20][20]bool{}
	canScout := [BOARDWIDTH][BOARDHEIGHT]bool{}

	for _, bot := range engine.AllBots {
		if bot.OwnerID == playerID {
			isScout := false
			for _, ability := range bot.Abilities {
				if ability == "SCOUT" {
					isScout = true
					break
				}

				minX := max(0, bot.Location.X-VisionRadius)
				maxX := min(BOARDWIDTH-1, bot.Location.X+VisionRadius)
				minY := max(0, bot.Location.Y-VisionRadius)
				maxY := min(BOARDHEIGHT-1, bot.Location.Y+VisionRadius)

				for x := minX; x <= maxX; x++ {
					for y := minY; y <= maxY; y++ {
						dist := manhattanDist(bot.Location.X, bot.Location.Y, x, y)

						// if dist <= VisionRadius {
						// canSee[x][y] = true // NO MORE VISION

						if isScout && dist <= ScoutRadius {
							canScout[x][y] = true
						}
						//}
					}
				}
			}
		}
	}

	// calculate all enemies in visible region
	for _, otherBot := range engine.AllBots {
		if otherBot.OwnerID != playerID {
			//            if canSee[otherBot.Location.X][otherBot.Location.Y] {
			enemy := EnemyBotDTO{
				ID:        otherBot.ID,
				Location:  Point{X: otherBot.Location.X, Y: otherBot.Location.Y},
				Scraps:    otherBot.Scraps,
				Abilities: otherBot.Abilities,
			}
			visibleEnemies = append(visibleEnemies, enemy)
			//            }
		}
	}

	// map of all algae in the region
	for x := range BOARDWIDTH {
		for y := range BOARDHEIGHT {
			tile := engine.Grid[x][y]
			if tile.HasAlgae /*&& canSee[x][y]*/ {
				poisonStatus := "UNKNOWN"

				if canScout[x][y] {
					if tile.IsPoison {
						poisonStatus = "TRUE"
					} else {
						poisonStatus = "FALSE"
					}
				}

				algae := VisibleAlgaeDTO{
					Location: Point{x, y},
					IsPoison: poisonStatus,
				}
				visibleAlgae = append(visibleAlgae, algae)
			}
		}
	}

	return VisibleEntitiesDTO{
		Enemies: visibleEnemies,
		Algae:   visibleAlgae,
	}
}
