package engine

const (
    SpawnEnergy = 50
    VisionRadius = 4
    BaseMovementCost = 2
    BaseScrapCost = 10
    SelfDestructRange = 1
)

func (engine *GameEngine) updateState(move PlayerMoves){
    engine.Tick++
    playerID := currentPlayerID()
    if invalid != nil { 
        return
    }
    for _, spawnCmd := range move.Spawns {
        engine.spawnBot(spawnCmd, playerID)
    }

    for botID, actionCmd := range move.Actions {
        engine.botAction(botID, actionCmd)
    }
}

func (engine *GameEngine) currentPlayerID(){
    return engine.Tick%2
}

func (engine *GameEngine) spawnBot(spawn SpawnCmd, playerID String) (bool SpawnedBot){
    if isValid, scrapCost := validateSpawn(spawn); isValid == true {
        bot := Bot {
            ID:             spawn.ID,
            OwnerID:        playerID,
            Location:       spawn.Location,
            Energy:         SpawnEnergy,
            Scraps:         scrapCost,
            Abilities:      spawn.Abilities,
            VisionRadius:   VisionRadius,
            TraversalCost:  calculateTraversalCost(spawn.Abilities)
        }
        engine.AllBots[bot.ID] = bot
        return true
    }
    else {
        return false
    }
}
//BOT LOCATION IS MESSED UP. DIRECT USED SOMEWHERE POINT USED ELSEWHERE
func (engine *GameEngine) actionBot(botID int, action ActionCmd){
    bot := engine.getBot(botID)
    if validMove, energyCost := engine.validateMove(botID, action); validMove == true {
        bot.Energy -= energyCost

        if direction != nil{
            engine.moveBot(botID, action.Direction)
        }

        if bot.Location

        switch action.Action {
            case "HARVEST":
                if engine.isAlgae(incrementLocation(bot.Location, action.Direction)){
                    engine.harvestAlgae(bot.OwnerID, incrementlocation(bot.Location, action.Direction))
                }
            case "SELFDESTRUCT":
                engine.selfDestructBot(botID)
            case "POISON":
                if engine.isAlgae(incrementLocation(bot.Location, action.Direction)){
                    engine.poisonAlgae(incrementLocation(bot.Location, action.Direction))
                }
            case "LOCKPICK":
                if engine.isNearBank(botID) {
                    engine.startLockPick(botID)
                }
        }
    }
}

func incrementLocation(loc Point, direction String) Point{
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

func (engine *GameEngine) moveBot(botID int, direction String)
    bot := getBot(botID)
    if hasAbility(botID, "SPEEDBOOST") {
        switch action.Direction {
            case "NORTH":
                bot.Y += 2
            case "SOUTH":
                bot.Y -= 2
            case "EAST":
                bot.X += 2
          case "WEST":
                bot.X -= 2
        } 

    }
    else {
        switch action.Direction {
            case "NORTH":
                bot.Y++
            case "SOUTH":
                bot.Y--
            case "EAST":
                bot.X++
          case "WEST":
                bot.X--
        } 
    }
}

//needs to be implemented
func (engine *GameEngine) selfDestructBot(botID int, engine *GameEngine){
    bot := getBot(botID)
    for botB := range engine.AllBots {
        if (math.Abs(bot.X-botB.X) <= selfDestructRange && math.Abs(bot.Y-botB.Y) <= selfDestructRange){
            if engine.hasAbility(botID, "SHIELD"){
                engine.removeShield(botID)
            }
            else {
                engine.KillBot(botB.ID)
            }
        } 
    }
    engine.KillBot(bot)
}

func (engine *GameEngine) KillBot(botID int){
    delete(engine.Allbots, botID)
}

func (engine *GameEngine) removeShield(botID int){
    bot := getBot(botID)
    for index, ability := range bot.Abilities {
        if ability == "SHIELD" {
            bot.Abilities = append(bot.Abilities[:i], bot.Abilities[i+1:]...)
        }
    }
    bot.TraversalCost -= EnergyDB["SHIELD"].Traversal
}

func (engine *GameEngine) validateSpawn(spawn SpawnCmd) (validSpawn bool, scrapCost int){
    scrapCost := 0
    playerID := currentPlayerID()

    if locationOccupied(spawn.Location) {
        return false, 0
    }

    for ability := range spawn.Abilities {
        scrapCost += costDB[ability]
    }

    if (scrapCost > engine.Scraps[playerID]){
        return false, scrapCost
    }
    return true, scrapCost

}

func (engine *GameEngine) locationOccupied(point Point) (occupied bool){
    for bot := range engine.Allbots{
        if point == Point {bot.X, bot.Y}{
            return false
        }
    }
    return true
}

func (engine *GameEngine) validateMove(botID int,move ActionCmd) (validMove bool, energyCost int){
    bot := getBot(botID)
    energyCost := 0

    if (ActionCmd.Direction != "NULL"){
        if locationOccupied(incrementLocation(bot.Location, ActionCmd.Direction)){
            return false, energyCost
            }
        energyCost += bot.TraversalCost
    }
    energyCost += EnergyDB[move.Action].Action

    if (energyCost > bot.Energy){
            return false, energyCost
    }

    return true, energyCost
}

func (engine *GameEngine) calculateTraversalCost(Abilities []String) (TraversalCost int){
    energyCost := BaseMovementCost

    for ability := range Abilities {
        energyCost += EnergyDB[ability].Traversal
    }
    return EnergyCost
}

func (engine *GameEngine) resolveCollisions(moves PlayerMoves){ //Is not needed anymore as the first move will be processed first now.

}

func (engine *GameEngine) startLockPick (botID int){

}

func (engine *GameEngine) isNearBank(botID int) (isNearBank bool,bankID int){
    bot := getBot(botID)
    for bankID, bank := range engine.Banks {
        if (math.Abs(bot.X-botB.X) <= BankDepositRange && math.Abs(bot.Y-botB.Y) <= BankDepositRange){
            return true, bankID
        }
    }
    return false, 0

}
func (engine *GameEngine) isOnEnergyPad(botID int) (isOnEnergyPad bool,bankID int){
    bot := getBot(botID)
    for EnergyPadsID, EnergyPad := range engine.EnergyPads {
        if (EnergyPad.X == bot.X && EnergyPad.Y == bot.Y){
            return true, EnergyPadId
        }
    }
    return false, 0
}


func (engine *GameEngine) isPoison(loc Point) bool{
    return engine.Grid[loc.X][loc.Y].isPoison
}

func (engine *GameEngine) isAlgae(loc Point) bool{
    return engine.Grid[loc.X][loc.Y].isAlgae
}

function (engine *GameEngine) getBot(botID int) *Bot{
    if bot, ok := engine.AllBots[botID]; ok {
        return bot;
    }   
}

func (engine *GameEngine) getState(playerID int) PlayerView {

}

func (engine *GameEngine) hasAbility(botID int, targetAbility String) bool {
    bot := getBot(botID)
    for _, ability := range bot.Abilities {
        if ability == targetAbility {
            return true
        }
    }
    return false
}

func (engine *GameEngine) calculateVisibleEntities(playerID int) VisibleEntities {
    visibleEnemies := make([]EnemyBot, 0)
    visibleAlgae := make([]VisibleAlgae, 0)
    
    //map of all cells that are visible or visible to a scout
    canSee := [20][20]bool{}
    canScout := [20][20]bool{}
    for _, bot := range engine.AllBots {
        if bot.OwnerID == playerID {
            isScout := false
            for ability := range bot.Abilities {
                if ability == "SCOUT" {
                    isScout = true
                    break
                }
                minX := max(0, bot.x-VisionRadius) //discard out of bounds coordinate
                maxX := min(19, bot.x+VisionRadius)
                minY := max(0, bot.y-VisionRadius)
                maxY := min(19, bot.y+VisionRadius)
                
                for x := minX; x <= maxX; x++ {
                    for y := minY; y <= maxY; y++ {
                        dist := manhattanDist(bot.x, bot.y, x, y) //abs(x1-x2)+abs(y1-y2)
                    
                    if dist <= VisionRadius {
                        canSee[x][y] = true
                        
                        if isScout && dist <= ScoutRadius {
                            canScout[x][y] = true
                        }
                    }
                }
            }
        }
    //calculate all enemies in visible region
    for _, otherBot := range engine.AllBots {
        if otherBot.OwnerID != playerID {
            if canSee[otherBot.x][otherBot.y] {
                enemy := EnemyBot{
                    ID:        otherBot.ID,
                    Location:  Point{X: otherBot.x, Y: otherBot.y},
                    Scraps:    otherBot.Scraps,
                    Abilities: otherBot.Abilities,
                }
                visibleEnemies = append(visibleEnemies, enemy)
            }
        }
    }
    //map of all algae in the region
    for x:= 0; x < 20; x++ {
        for y:= 0; y < 20; y++ {
                tile := engine.Grid[x][y]
                if tile.HasAlgae && canSee[x][y] {
                    poisonStatus := "UNKNOWN"
                    
                    if canScout[x][y] {
                        if tile.IsPoison {
                            poisonStatus = "TRUE"
                        }
                        else {
                            poisonStatus = "UNKNOWN"
                        }
                    }
                    algae := VisibleAlgae {
                        X: x,
                        Y: y,
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

func max(a, b int) int { if a > b { return a }; return b }
func min(a, b int) int { if a < b { return a }; return b }

