# Far West

A western themed roguelike game.

## Survival run

The game is played in levels. Every level drops you on a randomly shaped
island, the ocean around it is the border of the world. You have to wipe out
every bandit to advance to the next level.

Bandits live in one to three bandit camps per island. Each camp has tents,
chests and a group of enemies that spawn and wander around it instead of
roaming the whole island. The minimap shows the location of the camps, not the
individual bandits.

The first level starts with 6 normal enemies. Every following level the enemies
get tougher — more health, harder hitting bullets and a faster fire rate — and
there are more of them, up to a cap. When a level is cleared an overview shows
your score and time, press `N` or the `A` button to continue to the next level,
where you respawn on a fresh island with your full health back.

Killing an enemy gives you points. Running out of health ends the run on the
level you are on and pressing `Space` starts a fresh run from level 1.

## Controls

| Action | Keyboard | Gamepad |
| --- | --- | --- |
| Move | `Z`/`W`, `Q`/`A`, `S`, `D` | Left stick |
| Aim | Mouse (free angle) | Right stick (free angle) |
| Shoot | `Space` or left mouse | Right bumper (`FTR`) |
| Reload | `R` | Right right face button (`RR`) |
| Dodge | `Left Shift` | Left bumper (`FTL`) |
| Dead Eye | `X` | Center left (`RB`) |
| Switch weapon | `0`-`4`, `Tab` | D-pad right (`LR`) |
| Pick up gun | `E` | Right top face button (`RT`) |
| Compare gun | `Q` (hold) | Right left face button (`RL`) |
| Pause | `P` | Right trigger (`FBR`) |
| Continue after level | `N` | Right bottom face button (`RB`) |
| Start / restart | `Space` | Right trigger (`FBR`) |

## TODO

- [X] animate player
- [X] spawn some cacti
- [X] player revolver
- [X] hitboxes
- [X] collision detection
- [X] shoot mechanics
- [X] bullet hit animation
- [ ] environment destructable?
- [X] enemies
- [ ] interesting enemy behaviour
- [X] minimap
- [X] healthbar
- [X] dodge mechanics / animation
- [X] dodge bar
- [X] ammunition
- [X] reload
- [X] random island world with an ocean border
- [X] level based progression with tougher enemies
- [ ] dead eye refill
    - [X] player cigar mechanics
    - [X] cigar as loot
    - [X] cigar in store
- [ ] island boss (with dynamite sticks as weapon)
- [ ] town buildings
    - [X] store
    - [X] jail (with bounty poster: bounty needs to be knocked out by fists)
- [X] bandit camps (tents, weapon chest, gold chest, ammo chest)
- [ ] bounty post
- [X] boat dock (with ticket to next island when camps are cleared)
- [ ] progress bar after each level when going to the next island
- [ ] world map
    - [ ] player can choose which island to go next on a map (with difficulty level color coded)
    - [ ] 4 islands on the map (increased difficulty: green -> yellow -> red and one boss island)
- [ ] more sounds (https://sfxr.me/)
- [ ] islands with different biomes (snow, grass)
    - [ ] snow (dead trees, pine trees ,rocks, iglos)
    - [ ] grass (normal trees, shrubs, tipis)
