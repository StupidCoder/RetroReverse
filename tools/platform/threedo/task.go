package threedo

import (
	"fmt"

	"retroreverse.com/tools/cpu/arm60"
)

// task.go is a minimal cooperative scheduler for the Portfolio task model. The
// game is multitasking: it spawns a task (CreateItem of type 0x105 = TASKNODE)
// and waits for it to signal readiness. With one CPU and no preemption we run one
// task until it *yields* — it blocks in WaitSignal, exits, or busy-waits with no
// forward progress — then switch to another runnable task. Switches happen only
// between instructions (in the run loop), never inside a SWI handler mid
// PC-advance, so a suspended task always resumes cleanly.
//
// CreateItem(TASK) tag-args (CREATETASK_TAG_*, base TAG_ITEM_LAST=9): PC=0xA,
// STACKSIZE=0xC, ARGC=0xD, ARGP=0xE, SP=0xF. Signals (SIGF_*): IODONE=8,
// DEADTASK=0x10.

// taskExitTramp is the LR a spawned task returns to; the run loop treats a PC
// here as the task exiting.
const taskExitTramp = 0x0FF00000

type taskState int

const (
	stReady taskState = iota
	stRunning
	stWaiting
	stDone
)

type task struct {
	num       int32
	name      string
	ctx       arm60.Context // valid while not running
	state     taskState
	sig       uint32 // pending signal bits
	wait      uint32 // signals being awaited (while stWaiting)
	allocSigs uint32 // signal bits handed out by AllocSignal
}

// initTasks sets up the boot task as the initial running context.
func (m *Machine) initTasks() {
	m.tasks = []*task{{num: bootTaskNum, name: "boot", state: stRunning}}
	m.cur = 0
}

// spawnTask creates a runnable task from a CreateItem(TASK) tag-arg list and
// returns its item number. The task starts at its entry PC with its stack and
// argc/argp, and an exit trampoline in LR.
func (m *Machine) spawnTask(tagList uint32) int32 {
	entry, sp, argc, argp := m.parseTaskTags(tagList)
	m.nextItem++
	var ctx arm60.Context
	ctx.Mode = arm60.ModeSYS
	ctx.R[15] = entry
	ctx.R[13] = sp
	ctx.R[14] = taskExitTramp
	ctx.R[0] = argc
	ctx.R[1] = argp
	t := &task{num: m.nextItem, name: "task", ctx: ctx, state: stReady}
	m.tasks = append(m.tasks, t)
	m.items[t.num] = &item{num: t.num, typ: 0x105}
	m.note(fmt.Sprintf("spawnTask #%d entry=0x%08X sp=0x%08X argc=%d argp=0x%08X", t.num, entry, sp, argc, argp))
	return t.num
}

// TaskSummary reports each task's state and resume PC (diagnostics).
func (m *Machine) TaskSummary() []string {
	names := map[taskState]string{stReady: "ready", stRunning: "RUNNING", stWaiting: "waiting", stDone: "done"}
	var out []string
	for _, t := range m.tasks {
		pc := t.ctx.R[15]
		if t.state == stRunning {
			pc = m.CPU.Reg(15)
		}
		out = append(out, fmt.Sprintf("task #%d %-7s pc=0x%08X sig=0x%X wait=0x%X", t.num, names[t.state], pc, t.sig, t.wait))
	}
	return out
}

// parseTaskTags walks the TagArg list for the fields we need to launch a task.
func (m *Machine) parseTaskTags(p uint32) (entry, sp, argc, argp uint32) {
	for i := 0; i < 64 && p != 0; i++ {
		tag := m.readWord(p)
		arg := m.readWord(p + 4)
		switch tag {
		case 0: // TAG_END
			return
		case 0xA: // CREATETASK_TAG_PC
			entry = arg
		case 0xF: // CREATETASK_TAG_SP
			sp = arg
		case 0xD: // CREATETASK_TAG_ARGC
			argc = arg
		case 0xE: // CREATETASK_TAG_ARGP
			argp = arg
		}
		p += 8
	}
	return
}

// readWord reads a big-endian word from DRAM/VRAM (setup/inspection helper).
func (m *Machine) readWord(a uint32) uint32 {
	return uint32(m.Read(a))<<24 | uint32(m.Read(a+1))<<16 | uint32(m.Read(a+2))<<8 | uint32(m.Read(a+3))
}

// curTask returns the running task.
func (m *Machine) curTask() *task { return m.tasks[m.cur] }

// taskByNum returns the task with the given item number, or nil.
func (m *Machine) taskByNum(num int32) *task {
	for _, t := range m.tasks {
		if t.num == num {
			return t
		}
	}
	return nil
}

// switchTask saves the current context and resumes the next runnable task,
// round-robin. The caller sets the current task's state (Ready to keep it
// runnable, Waiting/Done otherwise) before calling. Returns false if no *other*
// task is runnable (the current context is left untouched).
func (m *Machine) switchTask() bool {
	saved := m.CPU.SaveContext()
	n := len(m.tasks)
	for i := 1; i <= n; i++ {
		j := (m.cur + i) % n
		if m.tasks[j].state == stReady {
			m.tasks[m.cur].ctx = saved
			m.CPU.RestoreContext(m.tasks[j].ctx)
			m.tasks[j].state = stRunning
			m.cur = j
			m.switches++
			return true
		}
	}
	return false
}

// sendSignal ORs bits into a task's pending signals and wakes it if it was
// waiting on any of them.
func (m *Machine) sendSignal(num int32, sigs uint32) {
	for _, t := range m.tasks {
		if t.num == num {
			t.sig |= sigs
			if t.state == stWaiting && t.sig&t.wait != 0 {
				t.state = stReady
			}
			return
		}
	}
}
