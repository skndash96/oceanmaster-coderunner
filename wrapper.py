# main.py
import json
import sys
import types

from oceanmaster.api.game_api import GameAPI
from oceanmaster.models.player_view import PlayerView
from submission import spawn_policy


def player_view_from_dict(data):
    view = PlayerView()
    view.tick = data["tick"]
    view.scraps = data["scraps"]
    view.algae = data["algae"]
    view.bot_count = data["bot_count"]
    view.max_bots = data["max_bots"]
    view.width = data["width"]
    view.height = data["height"]

    # ---------------- BOTS ----------------
    from oceanmaster.models.bot import Bot
    from oceanmaster.models.point import Point

    view.bots = []
    for b in data["bots"]:
        bot = Bot()
        bot.id = b["id"]
        bot.owner_id = b["owner_id"]
        bot.location = Point(**b["location"])
        bot.energy = b["energy"]
        bot.scraps = b["scraps"]
        bot.abilities = b["abilities"]
        bot.algae_held = b["algae_held"]
        view.bots.append(bot)

    # ---------------- VISIBLE ENTITIES ----------------
    from oceanmaster.models.visible_entities import VisibleEntities
    from oceanmaster.models.visible_scrap import VisibleScrap

    view.visible_entities = VisibleEntities()
    view.visible_entities.enemies = []  # filled later by backend
    view.visible_entities.scraps = []

    for s in data["visible_entities"]["scraps"]:
        scrap = VisibleScrap()
        scrap.location = Point(**s["location"])
        scrap.amount = s["amount"]
        view.visible_entities.scraps.append(scrap)

    # ---------------- PERMANENT ENTITIES ----------------
    from oceanmaster.models.permanent_entities import PermanentEntities
    from oceanmaster.models.bank import Bank
    from oceanmaster.models.energy_pad import EnergyPad
    from oceanmaster.models.algae import Algae

    view.permanent_entities = PermanentEntities()

    view.permanent_entities.banks = []
    for b in data["permanent_entities"]["banks"]:
        bank = Bank()
        bank.id = b["id"]
        bank.location = Point(**b["location"])
        bank.deposit_occuring = b["deposit_occuring"]
        bank.deposit_amount = b["deposit_amount"]
        bank.deposit_owner = b["deposit_owner"]
        bank.depositticksleft = b["depositticksleft"]
        view.permanent_entities.banks.append(bank)

    view.permanent_entities.energypads = []
    for e in data["permanent_entities"]["energypads"]:
        pad = EnergyPad()
        pad.id = e["id"]
        pad.location = Point(**e["location"])
        pad.available = e["available"]
        pad.ticksleft = e["ticksleft"]
        view.permanent_entities.energypads.append(pad)

    view.permanent_entities.walls = [
        Point(**w) for w in data["permanent_entities"]["walls"]
    ]

    view.permanent_entities.algae = []
    for a in data["permanent_entities"]["algae"]:
        algae = Algae()
        algae.location = Point(**a["location"])
        algae.is_poison = a["is_poison"]
        view.permanent_entities.algae.append(algae)

    return view


def main():
    from oceanmaster.wrapper import play

    # Handshake
    print(json.dumps("__READY_V1__"), flush=True)

    while True:
        line = sys.stdin.readline()
        if not line:
            break

        data = json.loads(line)
        view = player_view_from_dict(data)
        api = GameAPI(view)

        out = play(api, spawn_policy)
        print(json.dumps(out))
        sys.stdout.flush()

if __name__ == "__main__":
    main()
