# Far West

A western themed roguelike game.
gameplay loop:
You leave your ranch and cattle behind to pick up your life as bounty hunter again.
You get thrown into a region with a `town` and a `big bounty target`:

<img src="assets/minimap.png" width="300px" style="image-rendering: pixelated;"/>

In the town there are `shops` were you can by stuff:
- **Grocery store**: consumables for health, stamina, bonus stats etc
- **Stable**: horses
- **Gunshop**: weapons / upgrades / ammo
- **Saloon**: for recruiting extra gang members
- **clothing**: for increasing charisma

## TODO

- [X] animate player
- [X] spawn some cacti
- [X] player revolver
- [X] hitboxes
- [X] collision detection
- [X] shoot mechanics
- [ ] bullet hit animation
- [ ] environment destructable?
- [X] enemies
- [ ] interesting enemy behaviour
- [X] minimap
- [X] healthbar
- [ ] dash mechanics / animation
- [ ] stamina bar
- [X] ammunition
- [X] reload
- [X] random island world with an ocean border
- [X] endless enemy spawning with a growing enemy cap
- [X] survival timer, score and game over / restart
- [ ] town buildings
- [ ] spawn random town

## Survival run

Every run generates a random shaped island, the ocean around it is the border of
the world. Enemies keep spawning as long as you are alive: a fixed amount is on
the field at a time and that maximum grows the longer you survive. Killing an
enemy gives you points, running out of health ends the run and you can start a
fresh one right away.

## Controls

| Action | Keyboard | Gamepad |
| --- | --- | --- |
| Move | `Z`/`W`, `Q`/`A`, `S`, `D` | Left stick |
| Aim | Mouse (free angle) | Right stick (free angle) |
| Shoot | `Space` or left mouse | Right top shoulder |
| Reload | `R` | Right cluster left button (`X`) |
| Dodge | `Left Shift` | Left top shoulder |
| Switch weapon | `0` (fists), `1` (revolver) | Right top face button |
| Pause | `P` | Start |
| Start / restart | `Space` | Start |
