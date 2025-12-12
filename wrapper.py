import inspect
import json
import sys

import submission


def load_user_algo():
    for name, obj in inspect.getmembers(submission):
        if (
            inspect.isclass(obj)
            and issubclass(obj, submission.Game)
            and obj is not submission.Game
        ):
            return obj()
    raise Exception("No subclass of Game found in user.py")


def run_game():
    algo = load_user_algo()

    for line in sys.stdin:
        if line is None:
            break
        algo.actions = []
        line = line.rstrip("\n")
        algo.state = json.loads(line)
        algo.on_tick()
        print(json.dumps(algo.actions), flush=True)


if __name__ == "__main__":
    run_game()
