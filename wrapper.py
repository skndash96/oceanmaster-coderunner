import inspect
import json
import sys
from dataclasses import asdict

import submission


def load_user_algo():
    for _, obj in inspect.getmembers(submission):
        if (
            inspect.isclass(obj)
            and issubclass(obj, submission.OurPythonLib)
            and obj is not submission.OurPythonLib
        ):
            return obj()
    raise Exception("No subclass of Game found in user.py")


def run_game():
    algo = load_user_algo()

    for line in sys.stdin:
        if line is None:
            break
        algo._clear_actions()
        line = line.rstrip("\n")
        data = json.loads(line)
        algo._set_state(submission.GameState(**data))
        algo.on_tick()

        out = [asdict(a) for a in algo._get_actions()]
        print(json.dumps(out), flush=True)


if __name__ == "__main__":
    run_game()
