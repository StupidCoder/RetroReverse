// doorcheck opens the SHIPPED level documents and door GLBs and asserts what a
// player would see: every door holds itself open when clicked, the two leaves of
// a double door swing the SAME way, and every star gate is a PAIR of halves that
// can slide the width of one half apart.
//
// It re-derives the pairs from the shipped placements, reads the clip each
// placement's onClick names out of the exported GLB, samples it at the frame
// the placement holds at, and compares the direction the two free edges move.
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type level struct {
	Placements []struct {
		Object  string    `json:"object"`
		Pos     []float64 `json:"pos"`
		Rot     []float64 `json:"rot"`
		Props   map[string]any
		OnClick *struct {
			Action string  `json:"action"`
			Clip   string  `json:"clip"`
			HoldAt float64 `json:"holdAt"`
		} `json:"onClick"`
	} `json:"placements"`
}

type objDoc struct {
	Model string `json:"model"`
}

type gltf struct {
	Accessors []struct {
		BufferView    int    `json:"bufferView"`
		ByteOffset    int    `json:"byteOffset"`
		ComponentType int    `json:"componentType"`
		Count         int    `json:"count"`
		Type          string `json:"type"`
	} `json:"accessors"`
	BufferViews []struct {
		ByteOffset int `json:"byteOffset"`
		ByteLength int `json:"byteLength"`
	} `json:"bufferViews"`
	Animations []struct {
		Name     string `json:"name"`
		Channels []struct {
			Sampler int `json:"sampler"`
			Target  struct {
				Node int    `json:"node"`
				Path string `json:"path"`
			} `json:"target"`
		} `json:"channels"`
		Samplers []struct {
			Input  int `json:"input"`
			Output int `json:"output"`
		} `json:"samplers"`
	} `json:"animations"`
}

// readGLB splits a .glb into its JSON document and binary chunk.
func readGLB(path string) (*gltf, []byte, error) {
	d, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	if len(d) < 12 || string(d[:4]) != "glTF" {
		return nil, nil, fmt.Errorf("%s: not a glb", path)
	}
	var doc *gltf
	var bin []byte
	for p := 12; p+8 <= len(d); {
		n := int(binary.LittleEndian.Uint32(d[p:]))
		typ := binary.LittleEndian.Uint32(d[p+4:])
		body := d[p+8 : p+8+n]
		switch typ {
		case 0x4E4F534A:
			doc = &gltf{}
			if err := json.Unmarshal(body, doc); err != nil {
				return nil, nil, err
			}
		case 0x004E4942:
			bin = body
		}
		p += 8 + n
	}
	if doc == nil {
		return nil, nil, fmt.Errorf("%s: no json chunk", path)
	}
	return doc, bin, nil
}

// quatAt returns the rotation quaternion the named clip holds at fraction f.
func quatAt(doc *gltf, bin []byte, clip string, f float64) ([4]float64, bool) {
	for _, a := range doc.Animations {
		if a.Name != clip {
			continue
		}
		for _, ch := range a.Channels {
			if ch.Target.Path != "rotation" {
				continue
			}
			acc := doc.Accessors[a.Samplers[ch.Sampler].Output]
			if acc.Type != "VEC4" || acc.ComponentType != 5126 {
				continue
			}
			bv := doc.BufferViews[acc.BufferView]
			base := bv.ByteOffset + acc.ByteOffset
			k := int(math.Round(f * float64(acc.Count-1)))
			if k < 0 || k >= acc.Count {
				return [4]float64{}, false
			}
			var q [4]float64
			for c := 0; c < 4; c++ {
				q[c] = float64(math.Float32frombits(
					binary.LittleEndian.Uint32(bin[base+k*16+c*4:])))
			}
			return q, true
		}
	}
	return [4]float64{}, false
}

func main() {
	root := flag.String("root", "../../site/public/super-mario-64-ds", "shipped export root")
	flag.Parse()

	objModel := map[string]string{}
	objs, _ := filepath.Glob(filepath.Join(*root, "objects", "*.json"))
	for _, p := range objs {
		var o objDoc
		d, err := os.ReadFile(p)
		if err != nil || json.Unmarshal(d, &o) != nil {
			continue
		}
		objModel[strings.TrimSuffix(filepath.Base(p), ".json")] = o.Model
	}

	// vec3At reads a clip's translation output at its last key.
	endShift := func(obj, clip string) (float64, bool) {
		doc, bin, err := readGLB(filepath.Join(*root, "objects", objModel[obj]))
		if err != nil {
			return 0, false
		}
		for _, a := range doc.Animations {
			if a.Name != clip {
				continue
			}
			for _, ch := range a.Channels {
				if ch.Target.Path != "translation" {
					continue
				}
				acc := doc.Accessors[a.Samplers[ch.Sampler].Output]
				bv := doc.BufferViews[acc.BufferView]
				base := bv.ByteOffset + acc.ByteOffset + (acc.Count-1)*12
				return float64(math.Float32frombits(
					binary.LittleEndian.Uint32(bin[base:]))), true
			}
		}
		return 0, false
	}

	levels, _ := filepath.Glob(filepath.Join(*root, "levels", "*.json"))
	sort.Strings(levels)
	doors, pairs, fails, noHold := 0, 0, 0, 0
	gates := 0
	for _, lp := range levels {
		var lv level
		d, err := os.ReadFile(lp)
		if err != nil || json.Unmarshal(d, &lv) != nil {
			continue
		}
		type leaf struct {
			pos  [3]float64
			yaw  float64
			clip string
			hold float64
			obj  string
		}
		var ls []leaf
		for _, p := range lv.Placements {
			act, _ := p.Props["actor"].(float64)
			if int(act) != 353 || len(p.Pos) < 3 {
				continue
			}
			doors++
			yaw := 0.0
			if len(p.Rot) > 1 {
				yaw = p.Rot[1]
			}
			l := leaf{pos: [3]float64{p.Pos[0], p.Pos[1], p.Pos[2]}, yaw: yaw, obj: p.Object}
			if p.OnClick != nil {
				l.clip, l.hold = p.OnClick.Clip, p.OnClick.HoldAt
			}
			if l.clip != "" && l.hold <= 0 {
				fmt.Printf("  %s: %s has clip %s but holds at 0 — it slams shut again\n",
					filepath.Base(lp), l.obj, l.clip)
				noHold++
			}
			ls = append(ls, l)
		}
		// a star gate is one record drawn twice: same position, yaws 180 apart,
		// each half able to slide its own width (18.750 model units)
		type half struct {
			yaw  float64
			obj  string
			clip string
			hold float64
		}
		byPos := map[string][]half{}
		for _, p := range lv.Placements {
			act, _ := p.Props["actor"].(float64)
			if int(act) != 354 || len(p.Pos) < 3 {
				continue
			}
			yaw := 0.0
			if len(p.Rot) > 1 {
				yaw = p.Rot[1]
			}
			h := half{yaw: yaw, obj: p.Object}
			if p.OnClick != nil {
				h.clip, h.hold = p.OnClick.Clip, p.OnClick.HoldAt
			}
			k := fmt.Sprintf("%.3f/%.3f/%.3f", p.Pos[0], p.Pos[1], p.Pos[2])
			byPos[k] = append(byPos[k], h)
		}
		var keys []string
		for k := range byPos {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			hs := byPos[k]
			gates++
			if len(hs) != 2 {
				fmt.Printf("  %s: star gate at %s has %d half/halves, not 2\n",
					filepath.Base(lp), k, len(hs))
				fails++
				continue
			}
			d := math.Abs(math.Mod(math.Abs(hs[0].yaw-hs[1].yaw), 2*math.Pi) - math.Pi)
			if d > 0.01 {
				fmt.Printf("  %s: star gate at %s: halves are not 180 apart\n",
					filepath.Base(lp), k)
				fails++
			}
			for _, h := range hs {
				sh, ok := endShift(h.obj, h.clip)
				if !ok || math.Abs(math.Abs(sh)-18.75) > 1e-3 || h.hold <= 0 {
					fmt.Printf("  %s: star gate half %s does not slide a half width (%.3f, hold %.3f)\n",
						filepath.Base(lp), h.obj, sh, h.hold)
					fails++
				}
			}
		}

		// the same pairing rule the exporter uses, on the shipped numbers
		// The two records sit 0.300 apart (two 0.150 leaves), but each placement
		// carries the hinge compensation along its OWN local +X, and those
		// oppose — so the shipped positions are 0.150 apart.
		const pairDist = 0.150
		for i := range ls {
			for j := i + 1; j < len(ls); j++ {
				a, b := ls[i], ls[j]
				if math.Abs(a.pos[1]-b.pos[1]) > 1e-6 {
					continue
				}
				dx, dz := b.pos[0]-a.pos[0], b.pos[2]-a.pos[2]
				if math.Abs(math.Hypot(dx, dz)-pairDist) > 0.002 {
					continue
				}
				dy := math.Mod(math.Abs(a.yaw-b.yaw), 2*math.Pi)
				if math.Abs(dy-math.Pi) > 0.01 {
					continue
				}
				pairs++
				var d [2][3]float64
				ok := true
				for k, l := range [2]leaf{a, b} {
					mp := objModel[l.obj]
					doc, bin, err := readGLB(filepath.Join(*root, "objects", mp))
					if err != nil {
						ok = false
						break
					}
					q, found := quatAt(doc, bin, l.clip, l.hold)
					if !found {
						ok = false
						break
					}
					th := 2 * math.Atan2(q[1], q[3]) // Y rotation of the hinge bone
					// where the free edge goes: local (L,0,0) swung by th, then
					// the placement's own yaw
					const L = 1.0
					lx, lz := L*math.Cos(th)-L, -L*math.Sin(th)
					c, s := math.Cos(l.yaw), math.Sin(l.yaw)
					d[k] = [3]float64{lx*c + lz*s, 0, -lx*s + lz*c}
				}
				if !ok {
					fmt.Printf("  %s: could not read a leaf's clip\n", filepath.Base(lp))
					fails++
					continue
				}
				dot := d[0][0]*d[1][0] + d[0][2]*d[1][2]
				if dot <= 0 {
					fmt.Printf("  %s: %s + %s open in OPPOSITE directions (dot %.3f)\n",
						filepath.Base(lp), a.obj, b.obj, dot)
					fails++
				}
			}
		}
	}
	fmt.Printf("%d door placements, %d double doors, %d star gates, %d without a hold, %d failures\n",
		doors, pairs, gates, noHold, fails)
	if fails+noHold > 0 {
		log.Fatal("door check failed")
	}
}
