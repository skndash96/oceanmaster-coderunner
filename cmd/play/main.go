package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/delta/code-runner/internal/engine"
)

func main() {
	gl := engine.NewGameLogger(os.Stdout)
	ge := engine.InitGameEngine(gl)
	reader := bufio.NewReader(os.Stdin)

	pendingMoves := engine.PlayerMoves{
		Spawns:  make(map[int]engine.SpawnCmd),
		Actions: make(map[int]engine.ActionCmd),
	}

	fmt.Println("Ocean Master CLI Simulator")
	printHelp()

	for {
		// Calculate whose turn it is for the *pending* moves
		// If Ticks=1, Next Update makes it 2 -> Player 0 moves.
		// So pending moves are for Player (Ticks % 2).
		playerID := (ge.Ticks + 1) % 2 + 1
		fmt.Printf("\n--- TICK %d | PENDING: PLAYER %d ---\n", ge.Ticks, playerID)

		printState(ge)

		// Print pending moves
		if len(pendingMoves.Spawns) > 0 || len(pendingMoves.Actions) > 0 {
			fmt.Println("Pending Moves:")
			for id, s := range pendingMoves.Spawns {
				fmt.Printf("  SPAWN Bot %d at %v with %v\n", id, s.Location, s.Abilities)
			}
			for id, a := range pendingMoves.Actions {
				fmt.Printf("  ACTION Bot %d %s %s\n", id, a.Direction, a.Action)
			}
		}

		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToUpper(parts[0])

		switch cmd {
		case "QUIT":
			return
		case "HELP":
			printHelp()
		case "NEXT":
			pendingMoves.Tick = ge.Ticks+1
			ge.UpdateState(pendingMoves)
			// Reset pending moves for next turn
			pendingMoves = engine.PlayerMoves{
				Spawns:  make(map[int]engine.SpawnCmd),
				Actions: make(map[int]engine.ActionCmd),
			}
		case "SPAWN":
			if len(parts) < 4 {
				fmt.Println("Usage: SPAWN <botID> <x> <y> <abilities...>")
				continue
			}
			botID, _ := strconv.Atoi(parts[1])
			x, _ := strconv.Atoi(parts[2])
			y, _ := strconv.Atoi(parts[3])
			abilities := parts[4:]

			// Validate abilities being upper case
			for i := range abilities {
				abilities[i] = strings.ToUpper(abilities[i])
			}

			pendingMoves.Spawns[botID] = engine.SpawnCmd{
				Location:  engine.Point{X: x, Y: y},
				Abilities: abilities,
			}
			fmt.Println("Spawn queued.")

		case "ACTION":
			if len(parts) < 4 {
				fmt.Println("Usage: ACTION <botID> <direction> <verb>")
				continue
			}
			botID, _ := strconv.Atoi(parts[1])
			direction := strings.ToUpper(parts[2])
			verb := strings.ToUpper(parts[3])

			pendingMoves.Actions[botID] = engine.ActionCmd{
				Direction: direction,
				Action:    verb,
			}
			fmt.Println("Action queued.")

		default:
			fmt.Println("Unknown command. Type HELP.")
		}
	}
}

func printState(ge *engine.GameEngine) {
	// Print Stats
	fmt.Printf("Scraps: A=%d, B=%d | Algae: A=%d, B=%d | Total Algae: %d\n",
		ge.Scraps[0], ge.Scraps[1],
		ge.PermanentAlgae[0], ge.PermanentAlgae[1],
		ge.AlgaeCount)

    for bankID, bank := range ge.Banks {
        fmt.Printf("Bank ID %d at %d %d -> DepositOccuring=%t | DepositTicksLeft=%d | DepositOwner=%d | Deposit Amount=%d\n",bankID, bank.Location.X, bank.Location.Y, bank.DepositOccuring, bank.DepositTicksLeft, bank.DepositOwner, bank.DepositAmount )
    }

    for padID, pad := range ge.EnergyPads {
        fmt.Printf("EnergyPad ID %d at %d %d -> Available=%t | TicksLeft=%d\n", padID, pad.Location.X, pad.Location.Y, pad.Available, pad.TicksLeft)
    }

	// Print Bots
	var bots []*engine.Bot
	for _, b := range ge.AllBots {
		bots = append(bots, b)
	}
	sort.Slice(bots, func(i, j int) bool {
		return bots[i].ID < bots[j].ID
	})

	fmt.Println("Bots:")
	if len(bots) == 0 {
		fmt.Println("  (No bots on map)")
	}
	for _, bot := range bots {
		fmt.Printf("  Bot %d [P%d] @ %v | Energy: %.1f | Scraps: %d | Held: %d | Abilities: %v\n",
			bot.ID, bot.OwnerID, bot.Location, bot.Energy, bot.Scraps, bot.AlgaeHeld, bot.Abilities)
	}

    for y := range 20 {
        for x := range 20 {
            if ge.LocationOccupied(engine.Point{x, y}){
                fmt.Printf("o ")
            } else if ge.Grid[x][y].HasAlgae {
                if ge.Grid[x][y].IsPoison {
                    fmt.Printf("x ")
                } else {
                fmt.Printf("* ")
                }
            } else {
                fmt.Printf(". ")
            }
        }
        fmt.Printf("\n")
    }
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  SPAWN <botID> <x> <y> <abilities...>  (e.g., SPAWN 1 5 5 SCOUT)")
	fmt.Println("  ACTION <botID> <dir> <verb>           (e.g., ACTION 1 NORTH HARVEST)")
	fmt.Println("    Dirs: NORTH, SOUTH, EAST, WEST, NULL")
	fmt.Println("    Verbs: HARVEST, DEPOSIT, LOCKPICK, POISON, SELFDESTRUCT, NIL")
	fmt.Println("  NEXT                                  (Commit moves and advance turn)")
	fmt.Println("  QUIT")
}
