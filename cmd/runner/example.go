package main

const egCode = `
from dataclasses import dataclass, field


@dataclass(frozen=True)
class GameState:
    tick: int = 0
    series: list[int] = field(default_factory=list)

    @staticmethod
    def from_json(json_data: dict):
        return GameState(
            tick=int(json_data.get("tick", 0)), series=json_data.get("series", [])
        )


@dataclass(frozen=True)
class Action:
    element: int = 0

    def to_json(self):
        return {"element": self.element}


class OurPythonLib:
    def __init__(self):
        self._state = GameState()
        self._actions: list[Action] = []

    @property
    def state(self) -> GameState:
        return self._state

    def _set_state(self, state: GameState):
        self._state = state

    def _get_actions(self) -> list[Action]:
        return self._actions

    def _clear_actions(self):
        self._actions.clear()

    def add_element(self, v: int):
        self._actions.append(Action(v))

    def on_tick(self):
        """
        Must return a list of Action objects.
        """
        raise NotImplementedError


###### Above is python lib
###### Below is user implementation


class UserImplementation(OurPythonLib):
    def __init__(self):
        super().__init__()

    def on_tick(self):
        s = self.state.series

        if len(s) == 0:
            self.add_element(0)
        elif len(s) == 1:
            self.add_element(1)
        else:
            self.add_element(s[-1] + s[-2])
`
