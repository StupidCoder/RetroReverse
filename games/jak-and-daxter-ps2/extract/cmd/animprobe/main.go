package main

// animprobe: decode every art-joint-anim in a linked art group and report
// sanity stats — quaternion norms ~1 and scales ~1 are the desync detector.
import (
	"fmt"
	"math"
	"os"

	"retroreverse.com/games/jak-and-daxter-ps2/extract/merc"
)

func main() {
	obj, _ := os.ReadFile(os.Args[1])
	joints := merc.Joints(obj, 0x619CE4)
	fmt.Printf("%d joints", len(joints))
	if len(joints) > 0 {
		fmt.Printf("; first: %q num %d parent %d; last: %q", joints[0].Name, joints[0].Number, joints[0].Parent, joints[len(joints)-1].Name)
	}
	fmt.Println()
	for _, p := range merc.FindAnims(obj, 0x640474) {
		a, err := merc.DecodeJointAnim(obj, p)
		if err != nil {
			fmt.Println("ERR:", err)
			continue
		}
		qmin, qmax := 99.0, 0.0
		smin, smax := 99.0, -99.0
		tmax := 0.0
		for _, fr := range a.Frames {
			for _, po := range fr[2:] {
				n := po.QuatNorm()
				if n < qmin {
					qmin = n
				}
				if n > qmax {
					qmax = n
				}
				for _, s := range po.Scale {
					if float64(s) < smin {
						smin = float64(s)
					}
					if float64(s) > smax {
						smax = float64(s)
					}
				}
				t := math.Sqrt(float64(po.Trans[0]*po.Trans[0] + po.Trans[1]*po.Trans[1] + po.Trans[2]*po.Trans[2]))
				if t > tmax {
					tmax = t
				}
			}
		}
		fmt.Printf("  %-24s %3d frames %2d joints  quat|%.4f..%.4f|  scale[%.3f..%.3f]  maxT %.1f\n",
			a.Name, len(a.Frames), a.NumJoints, qmin, qmax, smin, smax, tmax)
	}
}
