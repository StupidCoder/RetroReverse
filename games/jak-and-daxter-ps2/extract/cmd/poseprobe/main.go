package main

// poseprobe: compare a 1-frame idle anim against the skeleton's bind-pose
// locals — resolves the quaternion convention and validates the whole
// decode + hierarchy math in one shot.
import (
	"fmt"
	"math"
	"os"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
)

func main() {
	obj, _ := os.ReadFile(os.Args[1])
	geos := merc.FindGeos(obj, 0x640564)
	joints := merc.GeoJoints(obj, geos[0], 0x619CE4)
	anims := merc.FindAnims(obj, 0x640474)
	a, err := merc.DecodeJointAnim(obj, anims[0])
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d joints, anim %q %d poses\n", len(joints), a.Name, len(a.Frames[0]))
	same, conj, transErr := 0, 0, 0.0
	for j := 2; j < len(joints) && j < a.NumJoints; j++ {
		world := merc.Mat4(joints[j].Bind).Inverse()
		local := world
		if p := joints[j].Parent; p >= 0 {
			local = world.Mul(merc.Mat4(joints[p].Bind))
		}
		_, bq, _ := local.TRS()
		po := a.Frames[0][j]
		aq := po.Quat
		dSame := float64(bq[0]*aq[0] + bq[1]*aq[1] + bq[2]*aq[2] + bq[3]*aq[3])
		dConj := float64(-bq[0]*aq[0] - bq[1]*aq[1] - bq[2]*aq[2] + bq[3]*aq[3])
		if math.Abs(dSame) > math.Abs(dConj) {
			same++
		} else {
			conj++
		}
		bt := [3]float32{local[12], local[13], local[14]}
		dt := math.Hypot(math.Hypot(float64(bt[0]-po.Trans[0]), float64(bt[1]-po.Trans[1])), float64(bt[2]-po.Trans[2]))
		transErr += dt
		if j < 8 {
			fmt.Printf("  j%-2d %-10s bindq(%6.3f %6.3f %6.3f %6.3f) animq(%6.3f %6.3f %6.3f %6.3f) |dsame|=%.3f |dconj|=%.3f bindT(%7.1f %7.1f %7.1f) animT(%7.1f %7.1f %7.1f)\n",
				j, joints[j].Name, bq[0], bq[1], bq[2], bq[3], aq[0], aq[1], aq[2], aq[3],
				math.Abs(dSame), math.Abs(dConj), bt[0], bt[1], bt[2], po.Trans[0], po.Trans[1], po.Trans[2])
		}
	}
	fmt.Printf("same-convention wins: %d, conj wins: %d, mean transErr %.1f\n", same, conj, transErr/float64(len(joints)-2))
}
