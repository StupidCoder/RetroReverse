package gekko

// spr.go reads and writes the special-purpose registers.
//
// Most of them are plain storage and could have been an array. They are a switch instead
// because several are not storage at all — writing them *does* something — and the ones
// that do are exactly the ones this machine depends on:
//
//	HID2   switches the paired-single unit, the locked cache and the write-gather pipe on
//	DMAL   triggers a locked-cache DMA transfer when its low bit is written
//	DEC    rearms the decrementer, which is how every timeout in the game is scheduled
//	the BATs   change the memory map underfoot
//
// An SPR this core does not know halts rather than silently reading zero. A zero is not a
// safe default: it is an answer, and a wrong one, and software that branches on it will go
// somewhere unaccountable rather than stopping where the mistake was made.

func (c *CPU) readSPR(n, pc uint32) uint32 {
	switch {
	case n == SPRXER:
		return c.XER
	case n == SPRLR:
		return c.LR
	case n == SPRCTR:
		return c.CTR
	case n == SPRDSISR:
		return c.DSISR
	case n == SPRDAR:
		return c.DAR
	case n == SPRDEC:
		return c.DEC
	case n == SPRSDR1:
		return c.SDR1
	case n == SPRSRR0:
		return c.SRR0
	case n == SPRSRR1:
		return c.SRR1
	case n >= SPRSPRG0 && n <= SPRSPRG3:
		return c.SPRG[n-SPRSPRG0]
	case n == SPRPVR:
		return c.PVR
	case n == SPRTBL:
		return uint32(c.TB)
	case n == SPRTBU:
		return uint32(c.TB >> 32)
	case n >= SPRIBAT0U && n <= SPRIBAT3L:
		i := (n - SPRIBAT0U) / 2
		return c.IBAT[i][(n-SPRIBAT0U)%2]
	case n >= SPRDBAT0U && n <= SPRDBAT3L:
		i := (n - SPRDBAT0U) / 2
		return c.DBAT[i][(n-SPRDBAT0U)%2]
	case n >= SPRGQR0 && n <= SPRGQR7:
		return c.GQR[n-SPRGQR0]
	case n == SPRHID0:
		return c.HID0
	case n == SPRHID1:
		return c.HID1
	case n == SPRHID2:
		return c.HID2
	case n == SPRHID4:
		return c.HID4
	case n == SPRWPAR:
		return c.WPAR
	case n == SPRDMAU:
		return c.DMAU
	case n == SPRDMAL:
		return c.DMAL
	case n == SPRL2CR:
		return c.L2CR
	case n == SPREAR:
		return 0
	case n == SPRDABR || n == SPRIABR || n == SPRICTC:
		return 0
	case n >= SPRTHRM1 && n <= SPRTHRM3:
		// The thermal sensors. Real ones; software reads them and does nothing useful
		// with the answer, so zero is honest here rather than a fiction.
		return 0
	}
	c.Halt("gekko: mfspr from the unknown SPR %d at 0x%08X", n, pc)
	return 0
}

func (c *CPU) writeSPR(n, v, pc uint32) {
	switch {
	case n == SPRXER:
		c.XER = v
	case n == SPRLR:
		c.LR = v
	case n == SPRCTR:
		c.CTR = v
	case n == SPRDSISR:
		c.DSISR = v
	case n == SPRDAR:
		c.DAR = v
	case n == SPRDEC:
		c.setDEC(v)
	case n == SPRSDR1:
		c.SDR1 = v
	case n == SPRSRR0:
		c.SRR0 = v
	case n == SPRSRR1:
		c.SRR1 = v
	case n >= SPRSPRG0 && n <= SPRSPRG3:
		c.SPRG[n-SPRSPRG0] = v
	case n == SPRTBL:
		c.TB = c.TB&0xFFFFFFFF00000000 | uint64(v)
	case n == SPRTBU:
		c.TB = c.TB&0x00000000FFFFFFFF | uint64(v)<<32
	case n >= SPRIBAT0U && n <= SPRIBAT3L:
		i := (n - SPRIBAT0U) / 2
		c.IBAT[i][(n-SPRIBAT0U)%2] = v
	case n >= SPRDBAT0U && n <= SPRDBAT3L:
		i := (n - SPRDBAT0U) / 2
		c.DBAT[i][(n-SPRDBAT0U)%2] = v
	case n >= SPRGQR0 && n <= SPRGQR7:
		c.GQR[n-SPRGQR0] = v
	case n == SPRHID0:
		c.HID0 = v
	case n == SPRHID1:
		c.HID1 = v
	case n == SPRHID2:
		c.setHID2(v) // this one turns hardware on
	case n == SPRHID4:
		c.HID4 = v
	case n == SPRWPAR:
		c.WPAR = v
	case n == SPRDMAU:
		c.DMAU = v
	case n == SPRDMAL:
		c.DMAL = v
		c.runDMA() // writing DMAL is what starts a locked-cache transfer
	case n == SPRL2CR:
		c.L2CR = v
	case n == SPREAR, n == SPRDABR, n == SPRIABR, n == SPRICTC, n == SPRPVR:
	case n >= SPRTHRM1 && n <= SPRTHRM3:
	default:
		c.Halt("gekko: mtspr to the unknown SPR %d (value 0x%08X) at 0x%08X", n, v, pc)
	}
}
