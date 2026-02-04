# main.py
import json
import sys
from typing import Callable

from oceanmaster.api import GameAPI
from oceanmaster.models.player_view import PlayerView
from oceanmaster.context.bot_context import BotContext
from oceanmaster.botbase import BotController
from oceanmaster.constants import Ability
from submission import spawn_policy as _spawn_policy # in sandbox submission dir will present and main.py inside represents the user code

class _WrapperState:
    def __init__(self):
        self.bot_strategies: dict[int, BotController] = {}
        self.spawn_policy: Callable[[GameAPI], list[dict]] = _spawn_policy
        self.curr_bot_id: int = -1

_STATE = _WrapperState()

def play(api: GameAPI):
    tick = api.get_tick()

    # linearize tick for the user algo
    # because user algo might do `if tick % 10 then spawn bot`
    if tick % 2 == 1:
        api.view.tick = tick//2 + 1
    else:
        api.view.tick = tick//2

    spawns: dict[str, dict] = {}
    actions: dict[str, dict] = {}

    if _STATE.curr_bot_id == -1:
        _STATE.curr_bot_id = api.view.bot_id_seed

    # ---- SPAWN PHASE (EVERY TICK) ----
    for spec in _STATE.spawn_policy(api):
        strategy_cls = spec["strategy"]

        if not issubclass(strategy_cls, BotController):
            raise TypeError(
                f"Invalid strategy class in spawn_policy: {strategy_cls}"
            )

        abilities = strategy_cls.ABILITIES

        # it's up to the engine to limit
        # if api.view.bot_count >= api.view.max_bots:
        #     continue

        bot_id = _STATE.curr_bot_id
        _STATE.curr_bot_id += 1

        spawns[str(bot_id)] = {
            "abilities": abilities,
            "location": {"x": 0, "y": spec["location"]},
        }
        _STATE.bot_strategies[int(bot_id)] = strategy_cls(None)

    # ---- ACTION PHASE ----
    alive_ids: set[int] = set()

    for bot in api.get_my_bots():
        alive_ids.add(bot.id)

        strategy = _STATE.bot_strategies.get(bot.id)
        if strategy is None:
            raise RuntimeError(
                f"Bot {bot.id} exists without a registered strategy."
            )

        ctx = BotContext(api, bot)
        strategy.ctx = ctx

        try:
            action = strategy.act()
        except Exception as exc:
            import traceback
            print(
                f"[USER_CODE] Error in bot {bot.id}: {exc}\n{traceback.format_exc()}",
                file=sys.stderr,
            )
            action = None

        if action is not None:
            actions[str(bot.id)] = action.to_dict()

    return {
        "tick": tick,
        "spawns": spawns,
        "actions": actions,
    }

def main():
    print("\"__READY_V1__\"", flush=True)
    while True:
        line = sys.stdin.readline()
        if not line:
            break

        data = json.loads(line)

        view = PlayerView.from_dict(data)

        api = GameAPI(view)
        out = play(api)

        print(json.dumps(out))
        sys.stdout.flush()


if __name__ == "__main__":
    main()
