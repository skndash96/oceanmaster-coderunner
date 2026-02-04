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
    BOARDWIDTH        = 20
    BOARDHEIGHT       = 20
    MAXALGAEHELD      = 5
    LockPickTime      = 20
)

const (
    PlayerOne = iota
    PlayerTwo
    Draw
)

func (engine *GameEngine) UpdateState(move PlayerMoves) {
    playerID := engine.currentPlayerID()

    for botID, spawnCmd := range move.Spawns {
    		// TODO: critical, check botID <= engine.BotIDSeed[playerID] + engine.MaxBots[playerID]
      	if playerID == PlayerTwo {
       			// correct?
       			spawnCmd.Location.X = 19
       	}
        engine.spawnBot(spawnCmd, playerID, botID)
    }

    for botID, actionCmd := range move.Actions {
    		if playerID == PlayerTwo {
						if actionCmd.Direction == "NORTH" {
							actionCmd.Direction = "SOUTH"
						} else if actionCmd.Direction == "SOUTH" {
							actionCmd.Direction = "NORTH"
						}
        }
        engine.actionBot(botID, actionCmd)
    }
    engine.TickPermanentEntities()
    engine.CheckWinCondition()
    engine.Ticks++
}

func (engine *GameEngine) TickPermanentEntities() {
    for _, bank := range engine.Banks {
        if bank.LockPickOccuring {
            if bank.LockPickTicksLeft == 0 {
                bot := engine.getBot(bank.LockPickBotID)
                engine.gl.Log(GameLogDebug, fmt.Sprintf("Deposit at bankID= %d has been stolen", bank.ID))
                bank.DepositOwner = bot.OwnerID
                bank.LockPickOccuring = false
                bank.LockPickBotID = -1
            }
            if bank.LockPickTicksLeft > 0 {
                if isNearBank, _ := engine.isNearBank(bank.LockPickBotID); isNearBank{
                    bank.LockPickTicksLeft--
                } else {
                    engine.gl.Log(GameLogDebug, fmt.Sprintf("LockPick at bankID=%d has been stopped", bank.ID))
                    bank.LockPickTicksLeft = 0
                    bank.LockPickOccuring = false
                    bank.LockPickBotID = -1
                }
            }
        }
        if bank.DepositOccuring {
            if bank.DepositTicksLeft == 0 {
                engine.PermanentAlgae[bank.DepositOwner] += bank.DepositAmount
                engine.gl.Log(GameLogDebug, fmt.Sprintf("%d Deposited to Player %d at Bank %d", bank.DepositAmount, bank.DepositOwner, bank.ID))
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
            if EnergyPad.TicksLeft == 1 {
            		engine.gl.Log(GameLogDebug, fmt.Sprintf("Energy Pad %d replenished", EnergyPad.ID))
            }
            EnergyPad.TicksLeft--
        }
        if EnergyPad.TicksLeft == 0 {
            EnergyPad.Available = true
        }
    }
}

func (engine *GameEngine) CheckWinCondition() int {
    if engine.PermanentAlgae[PlayerOne] > engine.AlgaeCount/2 {
        engine.gl.Log(GameLogDebug,"Player one has won")
        engine.Winner = PlayerOne
    }
    if engine.PermanentAlgae[PlayerTwo] > engine.AlgaeCount/2 {
        engine.gl.Log(GameLogDebug,"Player two has won")
        engine.Winner = PlayerTwo
    }
    if engine.Ticks >= 1000 {
        if engine.PermanentAlgae[PlayerOne] > engine.PermanentAlgae[PlayerTwo] {
            engine.gl.Log(GameLogDebug,"Player one has won")
            engine.Winner = PlayerOne
        }
        if engine.PermanentAlgae[PlayerOne] < engine.PermanentAlgae[PlayerTwo] {
            engine.gl.Log(GameLogDebug,"Player two has won")
            engine.Winner = PlayerTwo
        }
        if engine.PermanentAlgae[PlayerOne] == engine.PermanentAlgae[PlayerTwo] {
            engine.gl.Log(GameLogDebug,"Game ended in draw")
            engine.Winner = Draw
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
    if isValid, scrapCost := engine.validateSpawn(spawn, botID); isValid {
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
        engine.gl.Log(GameLogWarn, fmt.Sprintf("Cannot to spawn BotID=%d", botID))
        return false
    }
}

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
    } else {
        engine.gl.Log(GameLogWarn, fmt.Sprintf("Cannot to perform action for BotID=%d", botID))
    }
}

func incrementLocation(loc Point, direction string) (Point, bool) {
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
    if point.X < 0 || point.Y < 0 || point.X > 19 || point.Y > 19 {
        return point, false
    }
    return point, true
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
    isOutOfBounds := false
    if newLocation.X < 0 {
        isOutOfBounds = true
        newLocation.X = 0
    }
    if newLocation.X > BOARDWIDTH-1 {
        isOutOfBounds = true
        newLocation.X = BOARDWIDTH-1
    }
    if newLocation.Y < 0 {
        isOutOfBounds = true
        newLocation.Y = 0
    }
    if newLocation.Y > BOARDHEIGHT-1 {
        isOutOfBounds = true
        newLocation.Y = BOARDHEIGHT-1
    }
    if isOutOfBounds {
        engine.gl.Log(GameLogWarn, "Attempted to move out of bounds", botID)
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

func (engine *GameEngine) validateSpawn(spawn SpawnCmd, botID int) (bool, int) {
    scrapCost := 0
    playerID := engine.currentPlayerID()
    bot := engine.getBot(botID)
    if bot != nil {
        engine.gl.Log(GameLogWarn, fmt.Sprintf("BotID=%d already exists", botID))
        return false, scrapCost
    }

    if engine.LocationOccupied(spawn.Location) {
        engine.gl.Log(GameLogWarn, fmt.Sprintf("BotID=%d attempted spawn at occupied location", botID))
        return false, scrapCost
    }

    for _, ability := range spawn.Abilities {
        scrapCost += CostDB[ability]
    }

    if scrapCost > engine.Scraps[playerID] {
        engine.gl.Log(GameLogWarn, fmt.Sprintf("BotID=%d does not have enough scraps to spawn", botID))
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
    if bot == nil {
        engine.gl.Log(GameLogWarn, fmt.Sprintf("Invalid BotID %d", botID))
        return false, energyCost
    }
    if bot.OwnerID != playerID {
        engine.gl.Log(GameLogWarn, fmt.Sprintf("Player %d attempted to control invalid Bot %d", playerID, botID))
        return false, energyCost
    }

    if move.Direction != "NULL" {
        point, ok := incrementLocation(bot.Location, move.Direction)
        if !ok {
            engine.gl.Log(GameLogWarn, fmt.Sprintf("botID=%d attempted to move out of bounds at (%d %d)", botID, point.X, point.Y))
            return false, energyCost
        }
        if engine.LocationOccupied(point) {
            engine.gl.Log(GameLogWarn, fmt.Sprintf("botID=%d attempted to move at Occupied Location at (%d %d)", botID, point.X, point.Y))
            return false, energyCost
        }
        if engine.isWall(point){
            engine.gl.Log(GameLogWarn, fmt.Sprintf("botID=%d attempted to move to a wall at (%d %d)", botID, point.X, point.Y))
            return false, energyCost
        }
        energyCost += bot.TraversalCost
    }
    if move.Action != "MOVE"{
        if !engine.hasAbility(botID, move.Action) {
            engine.gl.Log(GameLogWarn, fmt.Sprintf("botID=%d does not have ability=%s", botID, move.Action))
            return false, energyCost
        }
    }

    energyCost += EnergyDB[move.Action].Ability

    if energyCost > bot.Energy {
        engine.gl.Log(GameLogWarn, fmt.Sprintf("botID=%d does not have enough energy for ability=%s", botID, move.Action))
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

func (engine *GameEngine) harvestAlgae(botID int) {
    bot := engine.getBot(botID)
    if bot.AlgaeHeld > MAXALGAEHELD {
        return
    }
    if engine.isAlgae(bot.Location) {
        if engine.isPoison(bot.Location) {
            engine.gl.Log(GameLogDebug, fmt.Sprintf("botID=%d has harvested a poisonous algae", botID))
            engine.KillBot(botID)
        }
        engine.Grid[bot.Location.X][bot.Location.Y].HasAlgae = false
        engine.Grid[bot.Location.X][bot.Location.Y].IsPoison = false
        bot.AlgaeHeld += 1
        bot.Energy -= EnergyDB["HARVEST"].Ability
    } else {
        engine.gl.Log(GameLogWarn, fmt.Sprintf("botID=%d attempted to harvest empty location", botID))
    }
}

func (engine *GameEngine) poisonAlgae(botID int) {
    bot := engine.getBot(botID)
    if engine.isAlgae(bot.Location) {
        engine.Grid[bot.Location.X][bot.Location.Y].IsPoison = true
        engine.AlgaeCount--
        engine.gl.Log(GameLogDebug, fmt.Sprintf("botID=%d has poisoned algae at (%d %d)", botID, bot.Location.X, bot.Location.Y))
    } else {
        engine.gl.Log(GameLogWarn, fmt.Sprintf("botID=%d attempted to poison empty location", botID))
    }
    bot.Energy -= EnergyDB["POISON"].Ability
}

func (engine *GameEngine) startLockPick(botID int) {
    bot := engine.getBot(botID)
    if NearBank, bankID := engine.isNearBank(botID); NearBank {
        engine.Banks[bankID].LockPickOccuring = true
        engine.Banks[bankID].LockPickTicksLeft = LockPickTime
        engine.Banks[bankID].LockPickBotID = botID
        bot.Energy -= EnergyDB["LOCKPICK"].Ability
    } else {
        engine.gl.Log(GameLogWarn, fmt.Sprintf("botID=%d attempted to LockPick too far from a bank", botID))
    }
}

func (engine *GameEngine) startDeposit(botID int) {
    bot := engine.getBot(botID)
    playerID := engine.currentPlayerID()
    if isNearBank, bankID := engine.isNearBank(botID); isNearBank {
        bank := engine.Banks[bankID]
        if bank.BankOwner == playerID && bank.DepositOccuring == false && bot.AlgaeHeld > 0 {
            fmt.Printf("Entered here")
            bank.DepositOwner = playerID
            bank.DepositTicksLeft = BankDepositTime
            bank.DepositOccuring = true
            bank.DepositAmount = bot.AlgaeHeld
            bot.AlgaeHeld = 0
            bot.Energy -= EnergyDB["DEPOSIT"].Ability
            engine.gl.Log(GameLogWarn, fmt.Sprintf("botID=%d attempted to deposit at bank already undergoing a deposit", botID))
        } else {
            engine.gl.Log(GameLogWarn, fmt.Sprintf("botID=%d attempted to deposit at bank not owned by them", botID))
        }
    } else {
        engine.gl.Log(GameLogWarn, fmt.Sprintf("botID=%d attempted to deposit too far from a bank", botID))
    }
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

func (engine *GameEngine) isWall(loc Point) bool {
    return engine.Grid[loc.X][loc.Y].IsWall
}

func (engine *GameEngine) getBot(botID int) *Bot {
    if bot, ok := engine.AllBots[botID]; ok {
        return bot
    }
    // engine.gl.Log(GameLogWarn, fmt.Sprintf("botID=%d does not exist", botID))
    return nil
}

func (engine *GameEngine) hasAbility(botID int, targetAbility string) bool {
    bot := engine.getBot(botID)
    if targetAbility == "DEPOSIT" {
        targetAbility = "HARVEST" //deposit automatically comes with harvest
    }
    return slices.Contains(bot.Abilities, targetAbility)
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
