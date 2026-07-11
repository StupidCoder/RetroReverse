// Command debugger is an interactive frame debugger for the RetroReverse oracles,
// built on the platform-agnostic tools/debug.DebugTarget surface. This is the M4
// shell: load a game, step it a video frame at a time, and inspect the captured
// frame — the rendered image, the RDP command stream, the CPU registers, and main
// memory. The command scrubber and pixel picker (M5) build on these panels.
//
// It lives in its own module (retroreverse.com/debugger) so its cgo/ImGui
// dependency stays out of the pure-stdlib tools module. Build it with the
// workspace disabled, since it pins a newer Go than the workspace:
//
//	GOWORK=off go run ./debugger -image "<rom>.z64"
package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	_ "github.com/AllenDang/cimgui-go/impl/glfw"    // registers the GLFW platform binding
	_ "github.com/AllenDang/cimgui-go/impl/opengl3" // registers the OpenGL3 renderer binding

	"retroreverse.com/tools/debug"
	"retroreverse.com/tools/debug/n64adapter"
)

// GLFW must own the main thread.
func init() { runtime.LockOSThread() }

var (
	be  backend.Backend[glfwbackend.GLFWWindowFlags]
	app *ui
)

// ui is the whole application state. There is one, driven from the render loop on
// the main goroutine; nothing here is touched concurrently.
type ui struct {
	adapter *n64adapter.Adapter

	fc       *debug.FrameCapture // the frame currently being inspected
	frameNo  int                 // how many fields we have stepped
	selected int                 // selected command index, or -1

	tex      *backend.Texture // GPU texture of the finished frame
	texDirty bool             // rebuild tex on the next loop (needs a live GL context)

	memAddr int32 // base address shown in the memory pane
	status  string
}

func main() {
	rom := flag.String("image", "", "N64 ROM (.z64) — required")
	state := flag.String("state", "", "optional savestate to load before stepping")
	flag.Parse()
	if *rom == "" {
		log.Fatal("debugger: -image is required")
	}

	a, err := n64adapter.New(*rom)
	if err != nil {
		log.Fatalf("debugger: %v", err)
	}
	if *state != "" {
		if err := a.LoadStateFile(*state); err != nil {
			log.Fatalf("debugger: loading state: %v", err)
		}
	}
	app = &ui{adapter: a, selected: -1, status: "ready"}
	app.stepToContent() // land on something worth showing before the window opens

	be, err = backend.CreateBackend(glfwbackend.NewGLFWBackend())
	if err != nil {
		log.Fatalf("debugger: creating backend: %v", err)
	}
	be.SetBgColor(imgui.NewVec4(0.10, 0.10, 0.12, 1.0))
	be.CreateWindow("RetroReverse Frame Debugger — "+a.Name(), 1280, 860)
	be.SetTargetFPS(60)
	be.Run(loop)
}

// --- stepping ---------------------------------------------------------------

func (u *ui) stepOne() {
	fc, err := u.adapter.StepFrame(true)
	if err != nil {
		u.status = "step error: " + err.Error()
		return
	}
	u.frameNo++
	u.setFrame(fc)
}

func (u *ui) stepToContent() {
	for i := 0; i < 800; i++ {
		fc, err := u.adapter.StepFrame(true)
		if err != nil {
			u.status = "step error: " + err.Error()
			return
		}
		u.frameNo++
		if len(fc.Commands) > 100 && fc.Prov != nil {
			u.setFrame(fc)
			return
		}
	}
	u.status = "no drawn frame found within the field budget"
}

func (u *ui) setFrame(fc *debug.FrameCapture) {
	u.fc = fc
	u.selected = -1
	u.texDirty = true
	u.status = fmt.Sprintf("field %d — %d RDP commands, %dx%d",
		u.frameNo, len(fc.Commands), fc.Width, fc.Height)
}

// --- render loop ------------------------------------------------------------

func loop() {
	u := app

	// A texture can only be created with a live GL context, which exists here in
	// the loop but not when a frame is first captured. Rebuild lazily.
	if u.texDirty && u.fc != nil && len(u.fc.Commands) > 0 {
		if img, err := u.adapter.RenderAfter(u.fc, len(u.fc.Commands)-1); err == nil {
			if u.tex != nil {
				u.tex.Release()
			}
			u.tex = backend.NewTextureFromRgba(img)
		}
		u.texDirty = false
	}

	drawControls(u)
	drawDisplay(u)
	drawCommands(u)
	drawCPU(u)
	drawMemory(u)
}

func drawControls(u *ui) {
	imgui.SetNextWindowPosV(imgui.NewVec2(8, 8), imgui.CondFirstUseEver, imgui.NewVec2(0, 0))
	imgui.SetNextWindowSizeV(imgui.NewVec2(480, 156), imgui.CondFirstUseEver)
	imgui.Begin("Controls")

	// Keyboard control, which is immune to any mouse hit-test offset:
	//   Space / N = step one field, Enter = step to the next drawn frame.
	if imgui.IsKeyPressedBool(imgui.KeySpace) || imgui.IsKeyPressedBool(imgui.KeyN) {
		u.stepOne()
	}
	if imgui.IsKeyPressedBool(imgui.KeyEnter) {
		u.stepToContent()
	}

	if imgui.ButtonV("Step Frame", imgui.NewVec2(150, 34)) {
		u.stepOne()
	}
	stepHovered := imgui.IsItemHovered()
	imgui.SameLine()
	if imgui.ButtonV("Step to drawn frame", imgui.NewVec2(190, 34)) {
		u.stepToContent()
	}
	imgui.Separator()
	text(u.status)
	text("keys: Space/N = step one field · Enter = step to a drawn frame")

	// Diagnostic for the mouse offset you saw. Move the cursor over the "Step
	// Frame" button: if stepHovered flips to YES only when the cursor is somewhere
	// else, imguiMouse vs your real cursor and contentScale tell us the factor.
	sx, sy := be.ContentScale()
	mp := imgui.MousePos()
	hov := "no"
	if stepHovered {
		hov = "YES"
	}
	text(fmt.Sprintf("diag: contentScale=%.2fx%.2f  imguiMouse=(%.0f,%.0f)  stepHovered=%s",
		sx, sy, mp.X, mp.Y, hov))
	imgui.End()
}

func drawDisplay(u *ui) {
	imgui.SetNextWindowPosV(imgui.NewVec2(8, 108), imgui.CondFirstUseEver, imgui.NewVec2(0, 0))
	imgui.SetNextWindowSizeV(imgui.NewVec2(704, 540), imgui.CondFirstUseEver)
	imgui.Begin("Display (draw target)")
	if u.tex != nil {
		imgui.Image(u.tex.ID, imgui.NewVec2(float32(u.tex.Width)*2, float32(u.tex.Height)*2))
	} else {
		text("no frame captured yet")
	}
	imgui.End()
}

func drawCommands(u *ui) {
	imgui.SetNextWindowPosV(imgui.NewVec2(720, 8), imgui.CondFirstUseEver, imgui.NewVec2(0, 0))
	imgui.SetNextWindowSizeV(imgui.NewVec2(552, 640), imgui.CondFirstUseEver)
	imgui.Begin("RDP Commands")
	if u.fc == nil {
		text("no frame")
		imgui.End()
		return
	}
	text(fmt.Sprintf("%d commands", len(u.fc.Commands)))
	imgui.Separator()
	imgui.BeginChildStrV("cmdlist", imgui.NewVec2(0, 0), imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
	for i := range u.fc.Commands {
		c := &u.fc.Commands[i]
		label := fmt.Sprintf("%5d  %-22s op=%#02x", c.Index, c.Name, c.Op)
		if imgui.SelectableBoolV(label, i == u.selected, 0, imgui.NewVec2(0, 0)) {
			u.selected = i
		}
	}
	imgui.EndChild()
	imgui.End()
}

func drawCPU(u *ui) {
	imgui.SetNextWindowPosV(imgui.NewVec2(8, 656), imgui.CondFirstUseEver, imgui.NewVec2(0, 0))
	imgui.SetNextWindowSizeV(imgui.NewVec2(704, 196), imgui.CondFirstUseEver)
	imgui.Begin("CPU")
	reg := u.adapter.CPU()
	text(fmt.Sprintf("PC = %016X", reg.PC))
	imgui.Separator()
	// 32 registers in a 4-column grid of name=value cells.
	if imgui.BeginTableV("regs", 8, imgui.TableFlagsBorders|imgui.TableFlagsRowBg, imgui.NewVec2(0, 0), 0) {
		for i := 0; i < len(reg.Names); i++ {
			imgui.TableNextColumn()
			text(fmt.Sprintf("%-4s %08X", reg.Names[i], reg.Vals[i]))
		}
		imgui.EndTable()
	}
	extras := []string{"hi", "lo", "Status", "Cause", "EPC"}
	line := ""
	for _, k := range extras {
		if v, ok := reg.Extra[k]; ok {
			line += fmt.Sprintf("%s=%08X  ", k, v)
		}
	}
	if line != "" {
		text(line)
	}
	imgui.End()
}

func drawMemory(u *ui) {
	imgui.SetNextWindowPosV(imgui.NewVec2(720, 656), imgui.CondFirstUseEver, imgui.NewVec2(0, 0))
	imgui.SetNextWindowSizeV(imgui.NewVec2(552, 196), imgui.CondFirstUseEver)
	imgui.Begin("Memory (RDRAM)")
	imgui.InputInt("addr", &u.memAddr)
	if u.memAddr < 0 {
		u.memAddr = 0
	}
	imgui.Separator()
	imgui.BeginChildStrV("hex", imgui.NewVec2(0, 0), imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
	base := uint32(u.memAddr) &^ 0xF
	data := u.adapter.ReadMem(base, 16*16)
	for row := 0; row < 16; row++ {
		off := row * 16
		hex, asc := "", ""
		for col := 0; col < 16; col++ {
			b := data[off+col]
			hex += fmt.Sprintf("%02X ", b)
			if b >= 0x20 && b < 0x7F {
				asc += string(rune(b))
			} else {
				asc += "."
			}
		}
		text(fmt.Sprintf("%08X  %s %s", base+uint32(off), hex, asc))
	}
	imgui.EndChild()
	imgui.End()
}

// text renders arbitrary text without treating it as a printf format string
// (imgui.Text does; a stray %02X in a hex dump would corrupt the display).
func text(s string) { imgui.TextUnformattedV(s) }
