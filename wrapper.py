import inspect
import json
import sys

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

    print(json.dumps("__READY__"), flush=True)

    for line in sys.stdin:
        if line is None:
            break

        algo._clear_actions()

        raw_json = json.loads(line.rstrip("\n"))

        new_game_state = submission.GameState.from_json(raw_json)

        algo._set_state(new_game_state)

        algo.on_tick()

        raw_actions = [a.to_json() for a in algo._get_actions()]

        print(json.dumps(raw_actions), flush=True)


if __name__ == "__main__":
    run_game()
