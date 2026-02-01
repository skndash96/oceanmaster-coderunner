package engine

import (
	"math"
	"slices"
    "fmt"
)

const (
	TotalTicks        = 1000
	SpawnEnergy       = 50.0
	VisionRadius      = 4
	BaseMovementCost  = 2.0
	BaseScrapCost     = 10
	SelfDestructRange = 1
	BasePadCoolDown   = 50
	BankDepositTime   = 100
	BankDepositRange  = 4
	ScoutRadius       = 4
	MAXBOTS           = 50
	BOARDWIDTH        = 20
	BOARDHEIGHT       = 20
	MAXALGAEHELD      = 5
)

const (
    PlayerOne = iota
    PlayerTwo
    Draw
)

func (engine *GameEngine) UpdateState(move PlayerMoves) {
	playerID := engine.currentPlayerID()

	for botID, spawnCmd := range move.Spawns {
		engine.spawnBot(spawnCmd, playerID, botID)
	}

	for botID, actionCmd := range move.Actions {
		engine.actionBot(botID, actionCmd)
	}
	engine.TickPermanentEntities()
	engine.Ticks++
}

func (engine *GameEngine) TickPermanentEntities() {
	for _, bank := range engine.Banks {
		if bank.DepositOccuring {
			if bank.DepositTicksLeft == 1 {
				engine.PermanentAlgae[bank.DepositOwner] += bank.DepositAmount
				bank.DepositAmount = 0
				bank.DepositOccuring = false
				bank.DepositOwner = -1
			}
			if bank.DepositTicksLeft > 0 {
				bank.DepositTicksLeft--
			}
		}
	}
	for _, EnergyPad := range engine.EnergyPads {
		if EnergyPad.TicksLeft > 0 {
			EnergyPad.TicksLeft--
		}
		if EnergyPad.TicksLeft == 0 {
			EnergyPad.Available = true
		}
	}
}

func (engine *GameEngine) CheckWinCondition() int {
	if engine.PermanentAlgae[PlayerOne] > engine.AlgaeCount/2 {
		engine.Winner = PlayerOne
	}
	if engine.PermanentAlgae[PlayerTwo] > engine.AlgaeCount/2 {
		engine.Winner = PlayerTwo
	}
	if engine.Ticks == 1000 {
		if engine.PermanentAlgae[PlayerOne] > engine.PermanentAlgae[PlayerTwo] {
			engine.Winner = PlayerOne
		}
		if engine.PermanentAlgae[PlayerOne] < engine.PermanentAlgae[PlayerTwo] {
			engine.Winner = PlayerTwo
		}
		if engine.PermanentAlgae[PlayerOne] == engine.PermanentAlgae[PlayerTwo] {
			engine.Winner = Draw //DRAW
		}
	}
	return engine.Winner

}

func (engine *GameEngine) currentPlayerID() int {
	if (engine.Ticks+1) % 2 == PlayerOne {
        return PlayerOne
    } else {
        return PlayerTwo
    }
}

func (engine *GameEngine) spawnBot(spawn SpawnCmd, playerID int, botID int) bool {
	if isValid, scrapCost := engine.validateSpawn(spawn); isValid {
		bot := Bot{
			ID:            botID,
			OwnerID:       playerID,
			Location:      spawn.Location,
			Energy:        SpawnEnergy,
			Scraps:        scrapCost,
			Abilities:     spawn.Abilities,
			VisionRadius:  VisionRadius,
			TraversalCost: engine.calculateTraversalCost(spawn.Abilities),
		}
		engine.AllBots[bot.ID] = &bot
		engine.Scraps[playerID] -= scrapCost
		return true
	} else {
		return false
	}
}

// BOT LOCATION IS MESSED UP. DIRECT USED SOMEWHERE POINT USED ELSEWHERE //fixed
func (engine *GameEngine) actionBot(botID int, action ActionCmd) {
	bot := engine.getBot(botID)
	if bot == nil {
		return
	}
	if validMove, energyCost := engine.validateMove(botID, action); validMove == true {
		bot.Energy -= energyCost

		if action.Direction != "NIL" {
			engine.moveBot(botID, action.Direction)
		}

		switch action.Action {
		case "HARVEST":
			engine.harvestAlgae(botID)
		case "SELFDESTRUCT":
			engine.selfDestructBot(botID)
		case "POISON":
			engine.poisonAlgae(botID)
		case "LOCKPICK":
			engine.startLockPick(botID)
		case "DEPOSIT":
			engine.startDeposit(botID)
		}
	}
}

func incrementLocation(loc Point, direction string) Point {
	point := loc
	switch direction {
	case "NORTH":
		point.Y++
	case "SOUTH":
		point.Y--
	case "EAST":
		point.X++
	case "WEST":
		point.X--
	}
	return point
}

func (engine *GameEngine) moveBot(botID int, direction string) {
	bot := engine.getBot(botID)
	newLocation := bot.Location
	if engine.hasAbility(botID, "SPEEDBOOST") {
		switch direction {
		case "NORTH":
			newLocation.Y += 2
		case "SOUTH":
			newLocation.Y -= 2
		case "EAST":
			newLocation.X += 2
		case "WEST":
			newLocation.X -= 2
		}

	} else {
		switch direction {
		case "NORTH":
			newLocation.Y++
		case "SOUTH":
			newLocation.Y--
		case "EAST":
			newLocation.X++
		case "WEST":
			newLocation.X--
		}
	}
	if newLocation.X < 0 {
		newLocation.X = 0
	}
	if newLocation.X > MAXWIDTH-1 {
		newLocation.X = MAXWIDTH-1
	}
	if newLocation.Y < 0 {
		newLocation.Y = 0
	}
	if newLocation.Y > MAXHEIGHT-1 {
		newLocation.Y = MAXHEIGHT-1
	}
	bot.Location = newLocation
	engine.energyPadCheck(botID)
}

func (engine *GameEngine) energyPadCheck(botID int) {
	bot := engine.getBot(botID)
	if OnEnergyPad, padID := engine.isOnEnergyPad(botID); OnEnergyPad {
		pad := engine.EnergyPads[padID]
		if pad.Available {
			pad.Available = false
			pad.TicksLeft = engine.getPadCoolDown()
			bot.Energy = float64(SpawnEnergy)
		}
	}
}

func (engine *GameEngine) getPadCoolDown() int {
	if engine.Ticks < TotalTicks*3/10 {
		return BasePadCoolDown
	}
	if engine.Ticks < TotalTicks*5/10 {
		return BasePadCoolDown * 5 / 10
	}
	if engine.Ticks < TotalTicks*7/10 {
		return BasePadCoolDown * 1 / 4
	}
	return BasePadCoolDown * 2 / 10
}

func (engine *GameEngine) selfDestructBot(botID int) {
	bot := engine.getBot(botID)
	for _, botB := range engine.AllBots {
		if math.Abs(float64(bot.Location.X-botB.Location.X)) <= SelfDestructRange && math.Abs(float64(bot.Location.Y-botB.Location.Y)) <= SelfDestructRange {
			if engine.hasAbility(botB.ID, "SHIELD") {
				engine.removeShield(botB.ID)
			} else {
				engine.KillBot(botB.ID)
			}
		}
	}
	engine.KillBot(bot.ID)
}

func (engine *GameEngine) KillBot(botID int) {
	delete(engine.AllBots, botID)
}

func (engine *GameEngine) removeShield(botID int) {
	bot := engine.getBot(botID)
	newAbilities := make([]string, 0, len(bot.Abilities)-1)
	for _, ability := range bot.Abilities {
		if ability != "SHIELD" {
			newAbilities = append(newAbilities, ability)
		}
	}
	bot.TraversalCost -= EnergyDB["SHIELD"].Traversal
	bot.Abilities = newAbilities
}

func (engine *GameEngine) validateSpawn(spawn SpawnCmd) (bool, int) {
	scrapCost := 0
	playerID := engine.currentPlayerID()

	if engine.LocationOccupied(spawn.Location) {
		return false, 0
	}

	for _, ability := range spawn.Abilities {
		scrapCost += CostDB[ability]
	}

	if scrapCost > engine.Scraps[playerID] {
		return false, scrapCost
	}
	return true, scrapCost

}

func (engine *GameEngine) LocationOccupied(point Point) bool {
	// TODO: What about other factors like banks ?
	for _, bot := range engine.AllBots {
		if point == bot.Location {
			return true
		}
	}
	return false
}

func (engine *GameEngine) validateMove(botID int, move ActionCmd) (bool, float64) {
    playerID := engine.currentPlayerID()
	bot := engine.getBot(botID)
	energyCost := 0.0
    if bot.OwnerID != playerID {
        return false, energyCost
    }

	if move.Direction != "NULL" {
		if engine.LocationOccupied(incrementLocation(bot.Location, move.Direction)) {
			return false, energyCost
		}
		energyCost += bot.TraversalCost
	}
	if !engine.hasAbility(botID, move.Action) {
		return false, energyCost
	}
	energyCost += EnergyDB[move.Action].Ability

	if energyCost > bot.Energy {
		return false, energyCost
	}

	return true, energyCost
}

func (engine *GameEngine) calculateTraversalCost(Abilities []string) float64 {
	energyCost := float64(BaseMovementCost)

	for _, ability := range Abilities {
		energyCost += EnergyDB[ability].Traversal
	}
	return energyCost
}

/*func (engine *GameEngine) resolveCollisions(moves PlayerMoves) { //Is not needed anymore as the first move will be processed first now.

}*/

func (engine *GameEngine) harvestAlgae(botID int) {
	bot := engine.getBot(botID)
	if bot.AlgaeHeld > MAXALGAEHELD {
		return
	}
	if engine.isAlgae(bot.Location) {
		if engine.isPoison(bot.Location) {
			engine.KillBot(botID)
		}
		engine.Grid[bot.Location.X][bot.Location.Y].HasAlgae = false
		engine.Grid[bot.Location.X][bot.Location.Y].IsPoison = false
		bot.AlgaeHeld += 1
		bot.Energy -= EnergyDB["HARVEST"].Ability
	}
}

func (engine *GameEngine) poisonAlgae(botID int) {
	bot := engine.getBot(botID)
	if engine.isAlgae(bot.Location) {
		engine.Grid[bot.Location.X][bot.Location.Y].IsPoison = true
		engine.AlgaeCount--
	}
	bot.Energy -= EnergyDB["POISON"].Ability
}

func (engine *GameEngine) startLockPick(botID int) {
	bot := engine.getBot(botID)
	if NearBank, bankID := engine.isNearBank(botID); NearBank {
		engine.Banks[bankID].DepositOwner = bot.OwnerID
	}
	bot.Energy -= EnergyDB["LOCKPICK"].Ability
}

func (engine *GameEngine) startDeposit(botID int) {
	bot := engine.getBot(botID)
	playerID := engine.currentPlayerID()
	if isNearBank, bankID := engine.isNearBank(botID); isNearBank {
		bank := engine.Banks[bankID]
        fmt.Printf("PlayerID -> %d BankOwnerID -> %d", playerID, bank.BankOwner)
		if bank.BankOwner == playerID && bank.DepositOccuring == false {
            fmt.Printf("Entered here")
			bank.DepositOwner = playerID
			bank.DepositTicksLeft = BankDepositTime
			bank.DepositOccuring = true
			bank.DepositAmount = bot.AlgaeHeld
			bot.AlgaeHeld = 0
		}
	}
	bot.Energy -= EnergyDB["DEPOSIT"].Ability
}

func (engine *GameEngine) isNearBank(botID int) (bool, int) {
	bot := engine.getBot(botID)
	for bankID, bank := range engine.Banks {
		if math.Abs(float64(bot.Location.X-bank.Location.X)) <= BankDepositRange && math.Abs(float64(bot.Location.Y-bank.Location.Y)) <= BankDepositRange {
			return true, bankID
		}
	}
	return false, -1
}

func (engine *GameEngine) isOnEnergyPad(botID int) (bool, int) {
	bot := engine.getBot(botID)
	for EnergyPadID, EnergyPad := range engine.EnergyPads {
		if EnergyPad.Location.X == bot.Location.X && EnergyPad.Location.Y == bot.Location.Y {
			return true, EnergyPadID
		}
	}
	return false, -1
}

func (engine *GameEngine) isPoison(loc Point) bool {
	return engine.Grid[loc.X][loc.Y].IsPoison
}

func (engine *GameEngine) isAlgae(loc Point) bool {
	return engine.Grid[loc.X][loc.Y].HasAlgae
}

func (engine *GameEngine) getBot(botID int) *Bot {
	if bot, ok := engine.AllBots[botID]; ok {
		return bot
	}
	return nil
}

func (engine *GameEngine) hasAbility(botID int, targetAbility string) bool {
	bot := engine.getBot(botID)
	if targetAbility == "DEPOSIT" {
		targetAbility = "HARVEST" //deposit automatically comes with harvest
	}
	return slices.Contains(bot.Abilities, targetAbility)
}

func (engine *GameEngine) getState(playerID int) PlayerView {
	playerBots := make(map[int]Bot, 0)
	for _, bot := range engine.AllBots {
		if bot.OwnerID == playerID {
			playerBots[bot.ID] = *bot
		}
	}

	Banks := make(map[int]Bank, 0)
	for _, bank := range engine.Banks {
		Banks[bank.ID] = *bank
	}

	Pads := make(map[int]Pad, 0)
	for _, pad := range engine.EnergyPads {
		Pads[pad.ID] = *pad
	}

	return PlayerView{
		Tick:            engine.Ticks,
		Scraps:          engine.Scraps[playerID],
		Algae:           engine.PermanentAlgae[playerID],
		BotCount:        len(playerBots),
		MaxBots:         MAXBOTS,
		Width:           BOARDWIDTH,
		Height:          BOARDHEIGHT,
		Bots:            playerBots,
		VisibleEntities: engine.calculateVisibleEntities(playerID),
		PermanentEntities: PermanentEntities{
			Banks:      Banks,
			EnergyPads: Pads,
		},
	}
}

func (engine *GameEngine) getGameView() GameView {
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
	return GameView{
		Tick:     engine.Ticks,
		Scraps:   [2]int{engine.Scraps[PlayerOne], engine.Scraps[PlayerTwo]},
		Algae:    [2]int{engine.PermanentAlgae[PlayerOne], engine.PermanentAlgae[PlayerTwo]},
		BotCount: len(engine.AllBots),
		MaxBots:  MAXBOTS,
		Width:    BOARDWIDTH,
		Height:   BOARDHEIGHT,
		AllBots:  allBots,
		PermanentEntities: PermanentEntities{
			Banks:      Banks,
			EnergyPads: Pads,
		},
		AlgaeMap: engine.getAlgaeMap(),
	}
}

func (engine *GameEngine) getAlgaeMap() []VisibleAlgae {
	visibleAlgae := make([]VisibleAlgae, 0)
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

				algae := VisibleAlgae{
					Location: Point{x, y},
					IsPoison: poisonStatus,
				}
				visibleAlgae = append(visibleAlgae, algae)
			}
		}
	}
	return visibleAlgae
}

func (engine *GameEngine) calculateVisibleEntities(playerID int) VisibleEntities {
	visibleEnemies := make([]EnemyBot, 0)
	visibleAlgae := make([]VisibleAlgae, 0)

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
				minX := max(0, bot.Location.X-VisionRadius) //discard out of bounds coordinate
				maxX := min(BOARDWIDTH-1, bot.Location.X+VisionRadius)
				minY := max(0, bot.Location.Y-VisionRadius)
				maxY := min(BOARDHEIGHT-1, bot.Location.Y+VisionRadius)

				for x := minX; x <= maxX; x++ {
					for y := minY; y <= maxY; y++ {
						dist := manhattanDist(bot.Location.X, bot.Location.Y, x, y)

						// if dist <= VisionRadius {
						// canSee[x][y] = true // NO MORE VISION

						if isScout && dist <= ScoutRadius { //whats scout radis?
							canScout[x][y] = true
						}
						//}
					}
				}
			}
		}
	}
	//calculate all enemies in visible region
	for _, otherBot := range engine.AllBots {
		if otherBot.OwnerID != playerID {
			//            if canSee[otherBot.Location.X][otherBot.Location.Y] { //Give full vision to everyone
			enemy := EnemyBot{
				ID:        otherBot.ID,
				Location:  Point{X: otherBot.Location.X, Y: otherBot.Location.Y},
				Scraps:    otherBot.Scraps,
				Abilities: otherBot.Abilities,
			}
			visibleEnemies = append(visibleEnemies, enemy)
			//            }
		}
	}
	//map of all algae in the region
	for x := 0; x < BOARDWIDTH; x++ {
		for y := 0; y < BOARDHEIGHT; y++ {
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
				algae := VisibleAlgae{
					Location: Point{x, y},
					IsPoison: poisonStatus,
				}
				visibleAlgae = append(visibleAlgae, algae)
			}
		}
	}
	return VisibleEntities{
		Enemies: visibleEnemies,
		Algae:   visibleAlgae,
	}
}

func manhattanDist(x1, y1, x2, y2 int) int {
	return absDiffInt(x1, x2) + absDiffInt(y1, y2)
}

func absDiffInt(x, y int) int {
	if x < y {
		return y - x
	}
	return x - y
}
