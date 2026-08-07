// trapprobe asks the game about the castle's double trapdoor — actor 36, whose
// placed record is a SPAWNER and whose two leaves run a state machine out of a
// table in overlay 10's BSS.
//
//	trapprobe [-rom img] [-extracted dir] [-states N] [-place] [-swing]
//
// -states dumps the state table (it only exists after the overlay's static
// initialisers have run, which LoadConfig does for real). -place lists every
// actor-36 placement with the two leaf transforms its init derives. -swing
// prints the opening angle per tick.
//
// What this CANNOT do is step a leaf and watch it swing. Every state resolves
// the OTHER leaf first ($0211139C: obj+$3AC is a handle, $02010F3C turns it
// into an object) and a leaf that cannot find its partner destroys itself and
// returns — and an actor the oracle spawns on its own is never entered in the
// game's object registry at $0209B468, so its handle is 0. A single-leaf run
// therefore sits at rotZ 0 for as many ticks as you give it. The swing below is
// a reimplementation of state 1's four lines of integration; what corroborates
// it is state 4, which hard-codes the held-open angle as -$3C00 independently,
// and which the integration passes through exactly one tick before it clamps.
package main

import (
	"flag"
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"retroreverse.com/games/super-mario-64-ds/extract/sm64ds"
)

const (
	trapActor = 36
	trapCfg   = 10
	stateTab  = 0x02112D28 // the step at $0211160C indexes this by obj+$3A0
	numStates = 5

	// state 1 ($021112B4), the opening integration
	openSpeed = 0x400  // what state 0 puts in obj+$3A8
	gravity   = 0x100  // taken off it every tick
	openStop  = 0x3D00 // |obj+$90| at which it clamps and enters state 2
	heldOpen  = 0x3C00 // state 4's independent constant for the same pose

	trapHalf = 349.0 // $15D — placement to each leaf's hinge, in world units
)

func main() {
	rom := flag.String("rom", "Super Mario 64 DS (Europe) (En,Fr,De,Es,It).nds", "cartridge image")
	ext := flag.String("extracted", "extracted", "extracted binaries dir")
	states := flag.Bool("states", false, "dump the state table out of overlay 10's BSS")
	place := flag.Bool("place", false, "every placement, with the two leaf transforms its init derives")
	swing := flag.Bool("swing", false, "the opening angle per tick")
	flag.Parse()

	if *swing {
		dumpSwing()
		return
	}

	ls, err := sm64ds.OpenLevels(*rom, *ext)
	if err != nil {
		sm64ds.Die(err)
	}
	if *place {
		listPlacements(ls)
		return
	}
	if !*states {
		flag.Usage()
		return
	}
	o, err := sm64ds.NewOracle(ls)
	if err != nil {
		sm64ds.Die(err)
	}
	if err := o.InitEngine(); err != nil {
		sm64ds.Die(err)
	}
	if err := o.LoadConfig(trapCfg); err != nil {
		sm64ds.Die(err)
	}
	fmt.Printf("state table @%08X\n", stateTab)
	for i := 0; i < numStates; i++ {
		a := uint32(stateTab + i*8)
		fmt.Printf("  state %d: fn=%08X adj=%08X\n", i, o.R32(a), o.R32(a+4))
	}
}

// listPlacements prints each actor-36 record and the two leaves its own init
// spawns: the parent's position offset by trapHalf along the yaw's X axis in
// both senses, with leaf 1 turned 180 degrees ($02111814).
func listPlacements(ls *sm64ds.LevelSet) {
	fmt.Printf("%-20s %-6s %-8s %s\n", "level", "par1", "rotY", "placement / leaf 0 / leaf 1 (world units)")
	for id := 0; id < sm64ds.NumLevels; id++ {
		lv, err := ls.Level(id)
		if err != nil {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(lv.BMDPath), ".bmd")
		for _, ob := range lv.Objects {
			if ob.Actor != trapActor {
				continue
			}
			yaw := ob.RotY * math.Pi / 180
			dx, dz := trapHalf*math.Cos(yaw), -trapHalf*math.Sin(yaw)
			fmt.Printf("%-20s %04X   %-8.3f %8.1f %8.1f %8.1f\n", stem, ob.Params[0], ob.RotY, ob.X, ob.Y, ob.Z)
			fmt.Printf("%-20s %-6s %-8.3f %8.1f %8.1f %8.1f   leaf 0\n", "", "", ob.RotY, ob.X-dx, ob.Y, ob.Z-dz)
			fmt.Printf("%-20s %-6s %-8.3f %8.1f %8.1f %8.1f   leaf 1\n", "", "", ob.RotY+180, ob.X+dx, ob.Y, ob.Z+dz)
		}
	}
}

func dumpSwing() {
	fmt.Printf("%-5s %8s %10s\n", "tick", "obj+$90", "degrees")
	v, a := openSpeed, 0
	for t := 0; ; t++ {
		fmt.Printf("%-5d %8d %10.3f", t, a, float64(a)*360/0x10000)
		switch a {
		case -heldOpen:
			fmt.Print("   <- state 4's held-open angle")
		case -openStop:
			fmt.Print("   <- state 1 clamps here and enters state 2")
		}
		fmt.Println()
		if a == -openStop {
			return
		}
		v -= gravity
		a += v
		if a < -openStop {
			a = -openStop
		}
	}
}
