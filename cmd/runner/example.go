package main

const egCode = `
# this Game class will be provided by our library
class Game:
    def __init__(self):
        self.state = []
        self.actions = []

    def on_tick(self):
        raise NotImplementedError("User algorithm must implement on_tick()")


# user gives Extension of Game
class Fibonacci(Game):
    def __init__(self):
        super().__init__()

    def on_tick(self):
        state = self.state
        if len(state) == 0:
            self.actions = [0, 1]
        elif len(state) == 1:
            self.actions = [1, 1]
        else:
            self.actions.append(state[-1] + state[-2])

`
