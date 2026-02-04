#!/usr/bin/env node

import fs from "fs";
import path from "path";
import readline from "readline";

/* ----------------------------- Rendering ----------------------------- */

function printState(view) {
  console.clear();
  console.log("=".repeat(80));
  console.log(`TICK ${view.tick}`);
  console.log("=".repeat(80));

  // Stats
  console.log(
    `Scraps: A=${view.scraps[0]}, B=${view.scraps[1]} | ` +
    `Algae: A=${view.algae_count[0]}, B=${view.algae_count[1]} | ` +
    `Bots: ${view.bot_count}/${view.max_bots}`
  );
  console.log();

  // Banks
  for (const bank of Object.values(view.permanent_entities.banks)) {
    console.log(
      `Bank ID ${bank.id} at ${bank.location.x} ${bank.location.y} -> ` +
      `DepositOccuring=${bank.deposit_occuring} | ` +
      `DepositTicksLeft=${bank.deposit_ticks_left} | ` +
      `DepositOwner=${bank.deposit_owner} | ` +
      `Deposit Amount=${bank.deposit_amount}`
    );
  }

  // Energy Pads
  for (const pad of Object.values(view.permanent_entities.energy_pads)) {
    console.log(
      `EnergyPad ID ${pad.id} at ${pad.location.x} ${pad.location.y} -> ` +
      `Available=${pad.available} | TicksLeft=${pad.ticks_left}`
    );
  }

  console.log();

  // Bots
  const bots = Object.values(view.bots).sort((a, b) => a.id - b.id);
  console.log("Bots:");
  if (bots.length === 0) {
    console.log("  (No bots on map)");
  }
  for (const bot of bots) {
    console.log(
      `  Bot ${bot.id} [P${bot.owner_id}] @ (${bot.location.x},${bot.location.y}) | ` +
      `Energy: ${bot.energy.toFixed(1)} | Scraps: ${bot.scraps} | ` +
      `Held: ${bot.algae_held} | Abilities: ${bot.abilities}`
    );
  }

  console.log();

  // Grid
  const grid = Array.from({ length: view.height }, () =>
    Array.from({ length: view.width }, () => ". ")
  );

  for (const w of view.permanent_entities.walls) {
    grid[w.y][w.x] = "W ";
  }

  for (const a of view.algae) {
    grid[a.location.y][a.location.x] =
      a.is_poison === "TRUE" ? "x " : "* ";
  }

  for (const b of Object.values(view.bots)) {
    grid[b.location.y][b.location.x] = `${b.owner_id} `;
  }

  for (const bank of Object.values(view.permanent_entities.banks)) {
    grid[bank.location.y][bank.location.x] = "B ";
  }

  for (const pad of Object.values(view.permanent_entities.energy_pads)) {
    grid[pad.location.y][pad.location.x] = "E ";
  }

  console.log("Map:");
  for (let y = 0; y < grid.length; y++) {
    console.log(grid[y].join(""));
  }

  console.log("\nLegend: . empty | * algae | x poison | W wall | B bank | E energy | 0/1 bot");
}

function printLog(entry) {
  if (entry.typ === "WARN") {
    console.log("\x1b[33m[WARN]\x1b[0m", entry.msg.join(" "));
  } else if (entry.typ === "ERROR") {
    console.log("\x1b[31m[ERROR]\x1b[0m", entry.msg.join(" "));
  } else if (entry.typ === "DEBUG") {
    console.log("\x1b[36m[DEBUG]\x1b[0m", entry.msg.join(" "));
  } else if (entry.typ === "MOVE") {
    console.log("[MOVE]", JSON.stringify(entry.msg[0]));
  }
}

/* --------------------------- Replay Parsing --------------------------- */

function loadTicks(logPath) {
  const lines = fs.readFileSync(logPath, "utf8").trim().split("\n");

  const ticks = [];
  let current = null;

  let i = 0
  for (const line of lines) {
    i += 1;
    const entry = JSON.parse(line);

    if (i < 20) console.log(entry)
    if (entry.typ === "VIEW") {
      if (current) ticks.push(current);
      current = {
        view: entry.msg[0],
        logs: [],
      };
    } else if (current) {
      current.logs.push(entry);
    }
  }

  if (current) ticks.push(current);
  return ticks;
}

/* ------------------------------- CLI -------------------------------- */

const submissionId = process.argv[2];
if (!submissionId) {
  console.error("Usage: node replay-viewer.js <SUBMISSION_ID>");
  process.exit(1);
}

const logPath = path.join(".submissions", submissionId, "log.txt");
if (!fs.existsSync(logPath)) {
  console.error("Log file not found:", logPath);
  process.exit(1);
}

const ticks = loadTicks(logPath);
let idx = 0;

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  prompt: "> ",
});

function showTick() {
  const tick = ticks[idx];
  if (!tick) {
    console.log(`No tick ${idx} data available.`);
    return;
  }

  if (tick.view) printState(tick.view);
  for (const log of tick.logs) {
    printLog(log);
  }
  console.log(`\n[Tick ${idx + 1}/${ticks.length}]`);
}

console.log(`Loaded ${ticks.length} ticks.`);
console.log("Commands: NEXT [n], BACK [n], START, END, QUIT\n");

rl.prompt();

rl.on("line", (line) => {
  const [cmd, n] = line.trim().toUpperCase().split(/\s+/);
  const step = Math.max(1, parseInt(n) || 1);

  if (cmd === "CLEAR" || cmd === "C") {
    console.clear();
  } else if (cmd === "NEXT" || cmd === "N") {
    idx = Math.min(ticks.length - 1, idx + step);
    showTick();
  } else if (cmd === "BACK" || cmd === "B") {
    idx = Math.max(0, idx - step);
    showTick();
  } else if (cmd === "START" || cmd === "S") {
    idx = 0;
    showTick();
  } else if (cmd === "END" || cmd === "E") {
    idx = ticks.length - 1;
    showTick();
  } else if (cmd === "QUIT" || cmd === "Q") {
    process.exit(0);
  } else {
    console.log("Unknown command.");
  }

  rl.prompt();
});
