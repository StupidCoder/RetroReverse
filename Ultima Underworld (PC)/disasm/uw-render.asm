; ============================================================================
; uw-render.asm — Ultima Underworld 3D renderer: view/camera-basis setup
; ============================================================================
;
; Captured from the ORACLE's live (relocated, overlay-paged) memory with the
; game in the dungeon. Static tracing of the raw UW.EXE cannot reach this code:
; it is paged in as an overlay, so its file bytes differ from the run-time
; image. Recapture from "Ultima Underworld (PC)/extract" with the character-
; creation -> dungeon key script and:
;
;   -bp 07F7:4DCA -dis 07F7:4B90:02B0   ; the view-basis routine (below)
;   -bp 07F7:4DCA -dis 214A:0A30:0090   ; the shared math helpers (sqrt, lerp)
;   -bp 07F7:4DCA -dis 07F7:5BC0:0030   ; the divide-error (#DE) handler
;
; Runtime segments: 07F7 = renderer overlay code; 499D = its data (the view
; matrix at [499D:1600..1618], 1.15 fixed point); 214A = shared math/utility
; code. The #DE trampoline is IVT[0]=589E:04D0 -> JMPF [589E:04D5]; this routine
; arms [589E:04D5] to 07F7:5BD1 before its saturating divides. Full commentary
; in uw-render.annotations.txt.
; ============================================================================

; ==== 07F7 — orthonormalise the camera basis =================================
; Scales the view-matrix rows by reciprocal lengths, then normalises each basis
; vector: length = isqrt(x^2+y^2) via CALLF 214A:0A30, component = (c<<15)/length
; with the result saturated to ~1.0 (the #DE handler catches the IDIV overflow).
  00004B90  A1 00 7C            MOV AX, [$7C00]
  00004B93  52                  PUSH DX
  00004B94  87 D3               XCHG BX, DX
  00004B96  91                  XCHG AX, CX
  00004B97  D1 EA               SHR DX, 1
  00004B99  D1 D8               RCR AX, 1
  00004B9B  F7 FB               IDIV BX
  00004B9D  8B C8               MOV CX, AX
  00004B9F  89 0E 28 27         MOV [$2728], CX
  00004BA3  A1 02 16            MOV AX, [$1602]
  00004BA6  F7 E9               IMUL CX
  00004BA8  D1 E0               SHL AX, 1
  00004BAA  D1 D2               RCL DX, 1
  00004BAC  89 16 02 16         MOV [$1602], DX
  00004BB0  A1 08 16            MOV AX, [$1608]
  00004BB3  F7 E9               IMUL CX
  00004BB5  D1 E0               SHL AX, 1
  00004BB7  D1 D2               RCL DX, 1
  00004BB9  89 16 08 16         MOV [$1608], DX
  00004BBD  A1 0E 16            MOV AX, [$160E]
  00004BC0  F7 E9               IMUL CX
  00004BC2  D1 E0               SHL AX, 1
  00004BC4  D1 D2               RCL DX, 1
  00004BC6  89 16 0E 16         MOV [$160E], DX
  00004BCA  A1 16 16            MOV AX, [$1616]
  00004BCD  F7 E9               IMUL CX
  00004BCF  D1 E0               SHL AX, 1
  00004BD1  D1 D2               RCL DX, 1
  00004BD3  89 16 16 16         MOV [$1616], DX
  00004BD7  A1 18 16            MOV AX, [$1618]
  00004BDA  F7 E9               IMUL CX
  00004BDC  D1 E0               SHL AX, 1
  00004BDE  D1 D2               RCL DX, 1
  00004BE0  89 16 18 16         MOV [$1618], DX
  00004BE4  EB 4D               JMP $00004C33
  00004BE6  D1 EA               SHR DX, 1
  00004BE8  D1 D8               RCR AX, 1
  00004BEA  F7 FB               IDIV BX
  00004BEC  8B C8               MOV CX, AX
  00004BEE  89 0E 2A 27         MOV [$272A], CX
  00004BF2  A1 04 16            MOV AX, [$1604]
  00004BF5  F7 E9               IMUL CX
  00004BF7  D1 E0               SHL AX, 1
  00004BF9  D1 D2               RCL DX, 1
  00004BFB  89 16 04 16         MOV [$1604], DX
  00004BFF  A1 0A 16            MOV AX, [$160A]
  00004C02  F7 E9               IMUL CX
  00004C04  D1 E0               SHL AX, 1
  00004C06  D1 D2               RCL DX, 1
  00004C08  89 16 0A 16         MOV [$160A], DX
  00004C0C  A1 10 16            MOV AX, [$1610]
  00004C0F  F7 E9               IMUL CX
  00004C11  D1 E0               SHL AX, 1
  00004C13  D1 D2               RCL DX, 1
  00004C15  89 16 10 16         MOV [$1610], DX
  00004C19  A1 14 16            MOV AX, [$1614]
  00004C1C  F7 E9               IMUL CX
  00004C1E  D1 E0               SHL AX, 1
  00004C20  D1 D2               RCL DX, 1
  00004C22  89 16 14 16         MOV [$1614], DX
  00004C26  A1 18 16            MOV AX, [$1618]
  00004C29  F7 E9               IMUL CX
  00004C2B  D1 E0               SHL AX, 1
  00004C2D  D1 D2               RCL DX, 1
  00004C2F  89 16 18 16         MOV [$1618], DX
  00004C33  8B 0E 26 27         MOV CX, [$2726]
  00004C37  0B C9               OR CX, CX
  00004C39  79 03               JNS $00004C3E
  00004C3B  E9 88 00            JMP $00004CC6
  00004C3E  89 0E 30 27         MOV [$2730], CX
  00004C42  A1 06 16            MOV AX, [$1606]
  00004C45  F7 E9               IMUL CX
  00004C47  D1 E0               SHL AX, 1
  00004C49  D1 D2               RCL DX, 1
  00004C4B  89 16 06 16         MOV [$1606], DX
  00004C4F  A1 0C 16            MOV AX, [$160C]
  00004C52  F7 E9               IMUL CX
  00004C54  D1 E0               SHL AX, 1
  00004C56  D1 D2               RCL DX, 1
  00004C58  89 16 0C 16         MOV [$160C], DX
  00004C5C  A1 12 16            MOV AX, [$1612]
  00004C5F  F7 E9               IMUL CX
  00004C61  D1 E0               SHL AX, 1
  00004C63  D1 D2               RCL DX, 1
  00004C65  89 16 12 16         MOV [$1612], DX
  00004C69  A1 16 16            MOV AX, [$1616]
  00004C6C  F7 E9               IMUL CX
  00004C6E  D1 E0               SHL AX, 1
  00004C70  D1 D2               RCL DX, 1
  00004C72  89 16 16 16         MOV [$1616], DX
  00004C76  A1 14 16            MOV AX, [$1614]
  00004C79  F7 E9               IMUL CX
  00004C7B  D1 E0               SHL AX, 1
  00004C7D  D1 D2               RCL DX, 1
  00004C7F  89 16 14 16         MOV [$1614], DX
  00004C83  8B C1               MOV AX, CX
  00004C85  F7 E8               IMUL AX
  00004C87  8B E8               MOV BP, AX
  00004C89  8B F2               MOV SI, DX
  00004C8B  A1 28 27            MOV AX, [$2728]
  00004C8E  A3 2C 27            MOV [$272C], AX
  00004C91  F7 E8               IMUL AX
  00004C93  03 C5               ADD AX, BP
  00004C95  13 D6               ADC DX, SI
  00004C97  8B DA               MOV BX, DX
  00004C99  33 C9               XOR CX, CX
  00004C9B  9A 30 0A 4A 21      CALLF $214A:$0A30
  00004CA0  8B C7               MOV AX, DI
  00004CA2  86 E0               XCHG AL, AH
  00004CA4  A3 28 27            MOV [$2728], AX
  00004CA7  A1 2A 27            MOV AX, [$272A]
  00004CAA  A3 2E 27            MOV [$272E], AX
  00004CAD  F7 E8               IMUL AX
  00004CAF  03 C5               ADD AX, BP
  00004CB1  13 D6               ADC DX, SI
  00004CB3  8B DA               MOV BX, DX
  00004CB5  33 C9               XOR CX, CX
  00004CB7  9A 30 0A 4A 21      CALLF $214A:$0A30
  00004CBC  8B C7               MOV AX, DI
  00004CBE  86 C4               XCHG AH, AL
  00004CC0  A3 2A 27            MOV [$272A], AX
  00004CC3  E9 BF 00            JMP $00004D85
  00004CC6  C7 06 30 27 FF 7F   MOV WORD [$2730], $7FFF
  00004CCC  F7 D9               NEG CX
  00004CCE  A1 02 16            MOV AX, [$1602]
  00004CD1  F7 E9               IMUL CX
  00004CD3  D1 E0               SHL AX, 1
  00004CD5  D1 D2               RCL DX, 1
  00004CD7  89 16 02 16         MOV [$1602], DX
  00004CDB  A1 08 16            MOV AX, [$1608]
  00004CDE  F7 E9               IMUL CX
  00004CE0  D1 E0               SHL AX, 1
  00004CE2  D1 D2               RCL DX, 1
  00004CE4  89 16 08 16         MOV [$1608], DX
  00004CE8  A1 0E 16            MOV AX, [$160E]
  00004CEB  F7 E9               IMUL CX
  00004CED  D1 E0               SHL AX, 1
  00004CEF  D1 D2               RCL DX, 1
  00004CF1  89 16 0E 16         MOV [$160E], DX
  00004CF5  A1 04 16            MOV AX, [$1604]
  00004CF8  F7 E9               IMUL CX
  00004CFA  D1 E0               SHL AX, 1
  00004CFC  D1 D2               RCL DX, 1
  00004CFE  89 16 04 16         MOV [$1604], DX
  00004D02  A1 0A 16            MOV AX, [$160A]
  00004D05  F7 E9               IMUL CX
  00004D07  D1 E0               SHL AX, 1
  00004D09  D1 D2               RCL DX, 1
  00004D0B  89 16 0A 16         MOV [$160A], DX
  00004D0F  A1 10 16            MOV AX, [$1610]
  00004D12  F7 E9               IMUL CX
  00004D14  D1 E0               SHL AX, 1
  00004D16  D1 D2               RCL DX, 1
  00004D18  89 16 10 16         MOV [$1610], DX
  00004D1C  A1 18 16            MOV AX, [$1618]
  00004D1F  F7 E9               IMUL CX
  00004D21  D1 E0               SHL AX, 1
  00004D23  D1 D2               RCL DX, 1
  00004D25  89 16 18 16         MOV [$1618], DX
  00004D29  8B F1               MOV SI, CX
  00004D2B  A1 28 27            MOV AX, [$2728]
  00004D2E  F7 E9               IMUL CX
  00004D30  D1 E0               SHL AX, 1
  00004D32  D1 D2               RCL DX, 1
  00004D34  8B EA               MOV BP, DX
  00004D36  D1 E0               SHL AX, 1
  00004D38  D1 D2               RCL DX, 1
  00004D3A  89 16 2C 27         MOV [$272C], DX
  00004D3E  8B D5               MOV DX, BP
  00004D40  8B C2               MOV AX, DX
  00004D42  F7 E8               IMUL AX
  00004D44  8B DA               MOV BX, DX
  00004D46  81 C3 FF 3F         ADD BX, $3FFF
  00004D4A  33 C9               XOR CX, CX
  00004D4C  9A 30 0A 4A 21      CALLF $214A:$0A30
  00004D51  8B C7               MOV AX, DI
  00004D53  86 E0               XCHG AL, AH
  00004D55  A3 28 27            MOV [$2728], AX
  00004D58  A1 2A 27            MOV AX, [$272A]
  00004D5B  F7 EE               IMUL SI
  00004D5D  D1 E0               SHL AX, 1
  00004D5F  D1 D2               RCL DX, 1
  00004D61  8B EA               MOV BP, DX
  00004D63  D1 E0               SHL AX, 1
  00004D65  D1 D2               RCL DX, 1
  00004D67  89 16 2E 27         MOV [$272E], DX
  00004D6B  8B D5               MOV DX, BP
  00004D6D  8B C2               MOV AX, DX
  00004D6F  F7 E8               IMUL AX
  00004D71  8B DA               MOV BX, DX
  00004D73  81 C3 FF 3F         ADD BX, $3FFF
  00004D77  33 C9               XOR CX, CX
  00004D79  9A 30 0A 4A 21      CALLF $214A:$0A30
  00004D7E  8B C7               MOV AX, DI
  00004D80  86 C4               XCHG AH, AL
  00004D82  A3 2A 27            MOV [$272A], AX
  00004D85  36 C7 06 D5 04 D1 …  MOV WORD [SS:$04D5], $5BD1
  00004D8C  A1 06 16            MOV AX, [$1606]
  00004D8F  F7 2E 06 16         IMUL WORD [$1606]
  00004D93  8B CA               MOV CX, DX
  00004D95  8B D8               MOV BX, AX
  00004D97  A1 12 16            MOV AX, [$1612]
  00004D9A  F7 2E 12 16         IMUL WORD [$1612]
  00004D9E  03 D8               ADD BX, AX
  00004DA0  13 CA               ADC CX, DX
  00004DA2  0A ED               OR CH, CH
  00004DA4  74 31               JZ $00004DD7
  00004DA6  9A 30 0A 4A 21      CALLF $214A:$0A30
  00004DAB  8B 16 06 16         MOV DX, [$1606]
  00004DAF  33 C0               XOR AX, AX
  00004DB1  D1 FA               SAR DX, 1
  00004DB3  D1 D8               RCR AX, 1
  00004DB5  F7 FF               IDIV DI
  00004DB7  3D 00 80            CMP AX, $8000
  00004DBA  75 01               JNZ $00004DBD
  00004DBC  40                  INC AX
  00004DBD  A3 CC 26            MOV [$26CC], AX
  00004DC0  8B 16 12 16         MOV DX, [$1612]
  00004DC4  33 C0               XOR AX, AX
  00004DC6  D1 FA               SAR DX, 1
  00004DC8  D1 D8               RCR AX, 1
  00004DCA  F7 FF               IDIV DI
  00004DCC  3D 00 80            CMP AX, $8000
  00004DCF  75 01               JNZ $00004DD2
  00004DD1  40                  INC AX
  00004DD2  A3 CE 26            MOV [$26CE], AX
  00004DD5  EB 54               JMP $00004E2B
  00004DD7  A1 04 16            MOV AX, [$1604]
  00004DDA  F7 2E 04 16         IMUL WORD [$1604]
  00004DDE  8B CA               MOV CX, DX
  00004DE0  8B D8               MOV BX, AX
  00004DE2  A1 10 16            MOV AX, [$1610]
  00004DE5  F7 2E 10 16         IMUL WORD [$1610]
  00004DE9  03 D8               ADD BX, AX
  00004DEB  13 CA               ADC CX, DX
  00004DED  9A 30 0A 4A 21      CALLF $214A:$0A30
  00004DF2  8B 16 04 16         MOV DX, [$1604]
  00004DF6  33 C0               XOR AX, AX
  00004DF8  D1 FA               SAR DX, 1
  00004DFA  D1 D8               RCR AX, 1
  00004DFC  F7 FF               IDIV DI
  00004DFE  3D 00 80            CMP AX, $8000
  00004E01  75 01               JNZ $00004E04
  00004E03  40                  INC AX
  00004E04  8B D8               MOV BX, AX
  00004E06  8B 16 10 16         MOV DX, [$1610]
  00004E0A  33 C0               XOR AX, AX
  00004E0C  D1 FA               SAR DX, 1
  00004E0E  D1 D8               RCR AX, 1
  00004E10  F7 FF               IDIV DI
  00004E12  3D 00 80            CMP AX, $8000
  00004E15  75 01               JNZ $00004E18
  00004E17  40                  INC AX
  00004E18  F7 06 0C 16 FF FF   TEST WORD [$160C], $FFFF
  00004E1E  78 04               JS $00004E24
  00004E20  F7 DB               NEG BX
  00004E22  F7 D8               NEG AX
  00004E24  89 1E CC 26         MOV [$26CC], BX
  00004E28  A3 CE 26            MOV [$26CE], AX
  00004E2B  A1 CE 26            MOV AX, [$26CE]
  00004E2E  F7 2E 02 16         IMUL WORD [$1602]
  00004E32  8B D8               MOV BX, AX
  00004E34  8B CA               MOV CX, DX
  00004E36  A1 CC 26            MOV AX, [$26CC]
  00004E39  F7 D8               NEG AX
  00004E3B  F7 2E 0E 16         IMUL WORD [$160E]

; ==== 214A — shared math helpers ============================================
; 0A30 (far): CALL 0A78 = isqrt32(CX:BX) -> DI (Newton's method, 5 iterations).
; 0A34 (far): CALL 0A38 = linear interpolation between two DS-relative tables at
;             [BX+04E0] and [BX+0560] by fraction CX (purpose per-caller).
  00000A30  E8 45 00            CALL $00000A78
  00000A33  CB                  RETF
  00000A34  E8 01 00            CALL $00000A38
  00000A37  CB                  RETF
  00000A38  8B CB               MOV CX, BX
  00000A3A  32 ED               XOR CH, CH
  00000A3C  8A DF               MOV BL, BH
  00000A3E  32 FF               XOR BH, BH
  00000A40  D1 E3               SHL BX, 1
  00000A42  8B AF E0 04         MOV BP, [BX+$4E0]
  00000A46  8B 87 E2 04         MOV AX, [BX+$4E2]
  00000A4A  2B C5               SUB AX, BP
  00000A4C  F7 E9               IMUL CX
  00000A4E  8A C4               MOV AL, AH
  00000A50  8A E2               MOV AH, DL
  00000A52  03 C5               ADD AX, BP
  00000A54  50                  PUSH AX
  00000A55  8B AF 60 05         MOV BP, [BX+$560]
  00000A59  8B 87 62 05         MOV AX, [BX+$562]
  00000A5D  2B C5               SUB AX, BP
  00000A5F  F7 E9               IMUL CX
  00000A61  8A DC               MOV BL, AH
  00000A63  8A FA               MOV BH, DL
  00000A65  03 DD               ADD BX, BP
  00000A67  58                  POP AX
  00000A68  C3                  RET
  00000A69  8A DF               MOV BL, BH
  00000A6B  32 FF               XOR BH, BH
  00000A6D  D1 E3               SHL BX, 1
  00000A6F  8B 87 E0 04         MOV AX, [BX+$4E0]
  00000A73  8B 9F 60 05         MOV BX, [BX+$560]
  00000A77  C3                  RET
  00000A78  0A ED               OR CH, CH
  00000A7A  74 3D               JZ $00000AB9
  00000A7C  BF 00 40            MOV DI, $4000
  00000A7F  3B CF               CMP CX, DI
  00000A81  72 03               JB $00000A86
  00000A83  BF FF FF            MOV DI, $FFFF
  00000A86  8B D1               MOV DX, CX
  00000A88  8B C3               MOV AX, BX
  00000A8A  F7 F7               DIV DI
  00000A8C  03 F8               ADD DI, AX
  00000A8E  D1 DF               RCR DI, 1
  00000A90  8B D1               MOV DX, CX
  00000A92  8B C3               MOV AX, BX
  00000A94  F7 F7               DIV DI
  00000A96  03 F8               ADD DI, AX
  00000A98  D1 DF               RCR DI, 1
  00000A9A  8B D1               MOV DX, CX
  00000A9C  8B C3               MOV AX, BX
  00000A9E  F7 F7               DIV DI
  00000AA0  03 F8               ADD DI, AX
  00000AA2  D1 DF               RCR DI, 1
  00000AA4  8B D1               MOV DX, CX
  00000AA6  8B C3               MOV AX, BX
  00000AA8  F7 F7               DIV DI
  00000AAA  03 F8               ADD DI, AX
  00000AAC  D1 DF               RCR DI, 1
  00000AAE  8B D1               MOV DX, CX
  00000AB0  8B C3               MOV AX, BX
  00000AB2  F7 F7               DIV DI
  00000AB4  03 F8               ADD DI, AX
  00000AB6  D1 DF               RCR DI, 1
  00000AB8  C3                  RET
  00000AB9  0A C9               OR CL, CL
  00000ABB  74 36               JZ $00000AF3
  00000ABD  BF 00 04            MOV DI, $0400

; ==== 07F7:5BD1 — the renderer's divide-error (#DE) handler ==================
; Reached from IVT[0]. INC WORD [BP] twice advances the pushed return IP past
; the 2-byte faulting IDIV (this needs 286+ #DE semantics: the pushed IP is the
; faulting instruction's, which tools/x86 now models); AX := 0x7FFF saturates the
; quotient; [CS:5BCF] counts the saturations. (5BC0-5BCC is a small neighbour.)
  00005BC0  2B CA               SUB CX, DX
  00005BC2  81 FF 00 80         CMP DI, $8000
  00005BC6  75 03               JNZ $00005BCB
  00005BC8  BF 01 80            MOV DI, $8001
  00005BCB  5B                  POP BX
  00005BCC  C3                  RET
  00005BCD  00 00               ADD [BX+SI], AL
  00005BCF  00 00               ADD [BX+SI], AL
  00005BD1  2E 89 2E CD 5B      MOV [CS:$5BCD], BP
  00005BD6  8B EC               MOV BP, SP
  00005BD8  FF 46 00            INC WORD [BP]
  00005BDB  FF 46 00            INC WORD [BP]
  00005BDE  B8 FF 7F            MOV AX, $7FFF
  00005BE1  33 D2               XOR DX, DX
  00005BE3  2E 8B 2E CD 5B      MOV BP, [CS:$5BCD]
  00005BE8  2E FF 06 CF 5B      INC WORD [CS:$5BCF]
  00005BED  CF                  IRET
  00005BEE  2E                  .byte $2E


; ============================================================================
; PART 2 — the rasterisation pipeline (segment 01A0, the resident graphics
; engine) and the off-screen buffer. Found by profiling framebuffer + buffer
; writes in the dungeon (bootoracle -vgaprof / -profrange): the 3D overlay 07F7
; and the primitives below all draw into a CHUNKY off-screen buffer at 41C5
; (one byte per pixel), which the Mode X blit then deinterleaves into A000.
; ============================================================================

; ==== 01A0:02CE — perspective/affine texture-mapped span ====================
; Walks CX pixels of one span, stepping a fixed-point texture coordinate and
; copying a texel per pixel into the chunky buffer. The step is self-modifying
; (the $0000 immediates at 02EE/02F2 are patched per span with the gradient).
; Texture is 32 texels wide: SI = (V & 0x3E0) + U, with V in the coord high word
; and U in DH. [07AF]=texture segment, [07BB/07BD]=coord, ES:DI=buffer.
  000002CE  A1 BD 07            MOV AX, [$07BD]
  000002D1  8B 16 BB 07         MOV DX, [$07BB]
  000002D5  33 DB               XOR BX, BX
  000002D7  8B 2E C1 07         MOV BP, [$07C1]
  000002DB  8E 1E AF 07         MOV DS, [$07AF]
  000002DF  8B F0               MOV SI, AX
  000002E1  81 E6 E0 03         AND SI, $03E0
  000002E5  8A DE               MOV BL, DH
  000002E7  03 F3               ADD SI, BX
  000002E9  A4                  MOVSB
  000002EA  81 C2 3B 01         ADD DX, $013B
  000002EE  81 C5 00 00         ADD BP, $0000
  000002F2  15 00 00            ADC AX, $0000
  000002F5  E2 E8               LOOP $000002DF
  000002F7  36 8E 1E D4 07      MOV DS, [SS:$07D4]
  000002FC  FF 0E CF 07         DEC WORD [$07CF]
  00000300  81 06 BF 07 00 00   ADD WORD [$07BF], $0000
  00000306  81 16 BB 07 00 00   ADC WORD [$07BB], $0000
  0000030D  06                  PUSH ES

; ==== 01A0:0B96 — chunky -> planar (Mode X) blit to A000 =====================
; For each scanline group: for each of the 4 planes, set the sequencer map-mask
; (OUT DX,AL with 01/02/04/08) and CALL 0BEC — an unrolled `MOVSB; ADD SI,3`
; that reads every 4th source byte (plane P reads buffer[base+P], +4, +8, ...).
; [BP+95E] is a per-scanline source-offset table; dest advances 0x50 (80) bytes
; per Mode X scanline. (0A58 REP STOSW is the flat span-fill primitive; 0A5C is
; a LODSW-driven display-list command interpreter that dispatches these.)
  00000B96  D1 E5               SHL BP, 1
  00000B98  8B 8E 5E 09         MOV CX, [BP+$95E]
  00000B9C  B0 08               MOV AL, $08
  00000B9E  EE                  OUT DX, AL
  00000B9F  8B FB               MOV DI, BX
  00000BA1  BE 03 00            MOV SI, $0003
  00000BA4  03 F1               ADD SI, CX
  00000BA6  E8 43 00            CALL $00000BEC
  00000BA9  B0 02               MOV AL, $02
  00000BAB  EE                  OUT DX, AL
  00000BAC  8B FB               MOV DI, BX
  00000BAE  BE 01 00            MOV SI, $0001
  00000BB1  03 F1               ADD SI, CX
  00000BB3  E8 36 00            CALL $00000BEC
  00000BB6  B0 04               MOV AL, $04
  00000BB8  EE                  OUT DX, AL
  00000BB9  8B FB               MOV DI, BX
  00000BBB  BE 02 00            MOV SI, $0002
  00000BBE  03 F1               ADD SI, CX
  00000BC0  E8 29 00            CALL $00000BEC
  00000BC3  B0 01               MOV AL, $01
  00000BC5  EE                  OUT DX, AL
  00000BC6  8B FB               MOV DI, BX
  00000BC8  8B F1               MOV SI, CX
  00000BCA  E8 1F 00            CALL $00000BEC
  00000BCD  83 C3 50            ADD BX, $0050
  00000BD0  D1 ED               SHR BP, 1
  00000BD2  4D                  DEC BP
  00000BD3  36 3B 2E F8 3D      CMP BP, [SS:$3DF8]
  00000BD8  7F BC               JG $00000B96
  00000BDA  FA                  CLI
  00000BDB  2E 8E 16 08 0B      MOV SS, [CS:$0B08]
  00000BE0  2E 8B 26 0A 0B      MOV SP, [CS:$0B0A]
  00000BE5  FB                  STI
  00000BE6  5F                  POP DI
  00000BE7  5E                  POP SI
  00000BE8  1F                  POP DS
  00000BE9  07                  POP ES
  00000BEA  5D                  POP BP
  00000BEB  CB                  RETF
  00000BEC  A4                  MOVSB
  00000BED  83 C6 03            ADD SI, $0003
  00000BF0  A4                  MOVSB


; ============================================================================
; PART 3 — the affine textured-polygon rasteriser (segment 01A0), the two DDA
; levels that feed the texture span (PART 2). All interpolation is affine
; (linear), with the per-step deltas held as SELF-MODIFYING immediates patched
; by the polygon setup from the projected vertices (the projection itself is in
; the 07F7 overlay, not yet mapped). Inputs/state live in DS=3BDD (the graphics
; driver data segment): [07B7/07B9] left-edge X int/frac, [07C3/07C5] right-edge
; X, [07BB/07BD] start tex U/V, [07C7..07CD] end tex U/V, [07CF] scanline.
; ============================================================================

; ==== 01A0:0296 — per-span horizontal texture DDA setup =====================
; CX = span length (endX-startX); the U gradient (endU-startU)/CX and the V
; gradient (endV-startV)/CX (split integer/fractional) are computed and PATCHED
; into the span loop's step immediates at [CS:02EC]/[CS:02F0]/[CS:02F3].
  00000296  A1 C7 07            MOV AX, [$07C7]
  00000299  2B 06 BB 07         SUB AX, [$07BB]
  0000029D  99                  CWD
  0000029E  F7 F9               IDIV CX
  000002A0  2E A3 EC 02         MOV [CS:$02EC], AX
  000002A4  A1 C9 07            MOV AX, [$07C9]
  000002A7  2B 06 BD 07         SUB AX, [$07BD]
  000002AB  99                  CWD
  000002AC  F7 F9               IDIV CX
  000002AE  2E A3 F3 02         MOV [CS:$02F3], AX
  000002B2  33 C0               XOR AX, AX
  000002B4  D1 FA               SAR DX, 1
  000002B6  D1 D8               RCR AX, 1
  000002B8  F7 F9               IDIV CX
  000002BA  99                  CWD
  000002BB  D1 E0               SHL AX, 1
  000002BD  D1 D2               RCL DX, 1
  000002BF  2E A3 F0 02         MOV [CS:$02F0], AX
  000002C3  2E 01 16 F3 02      ADD [CS:$02F3], DX
  000002C8  41                  INC CX
  000002C9  7F 03               JG $000002CE
  000002CB  E9 D7 00            JMP $000003A5

; ==== 01A0:0312 — per-scanline vertical edge/texture DDA ====================
; Advances every polygon interpolant by a constant per-scanline delta: the left
; and right edge X, and the span-endpoint texture coordinates, then loops until
; the scanline counter [07CF] passes the polygon's last row. The FF99/8000/FFFE/
; B334... immediates are the constant slopes, patched per polygon by its setup.
  00000312  81 16 BD 07 99 FF   ADC WORD [$07BD], $FF99
  00000318  81 06 B9 07 00 80   ADD WORD [$07B9], $8000
  0000031E  81 16 B7 07 FE FF   ADC WORD [$07B7], $FFFE
  00000324  81 06 CB 07 00 00   ADD WORD [$07CB], $0000
  0000032A  81 16 C7 07 00 00   ADC WORD [$07C7], $0000
  00000330  81 06 CD 07 34 B3   ADD WORD [$07CD], $B334
  00000336  81 16 C9 07 99 FF   ADC WORD [$07C9], $FF99
  0000033C  81 06 C5 07 00 80   ADD WORD [$07C5], $8000
  00000342  81 16 C3 07 FF FF   ADC WORD [$07C3], $FFFF
  00000349  0E                  PUSH CS


; ============================================================================
; PART 4 — THE PERSPECTIVE PROJECTION (segment 07F7, the geometry heart).
; Turns the view-space points (already rotated by the camera basis, PART 1)
; into the screen-space vertices the 01A0 rasteriser consumes (PART 3). One
; perspective divide per axis per vertex; DS=499D (renderer data), ES=3BDD:0659
; (the vertex list). Projection constants live at 499D:26B0 — captured live:
;   [26B0]=56 scaleY  [26B2]=86 scaleX  [26B4]=86 centreX  [26B6]=56 centreY
;   => screenX = X*86/Z + 86 ; screenY = Y*56/Z + 56  (a 172x112 viewport,
;      centred at 86,56 — the size of the dungeon 3D view).
; Source point (at DS:[0307], stepped to [0305]): X +0, Y +2, Z +4, U +8, V +A.
; Dest vertex (3BDD:0659, 14 bytes): screenX +0, screenY +2, U +4, V +6,
;   raw axis +8 (scratch), +A, Z +C. At 6129 it arms its OWN #DE handler
;   ([SS:04D5]=692C) so a near-zero Z saturates instead of trapping.
; ============================================================================

  00006129  C7 06 D5 04 2C 69   MOV WORD [SS:$04D5], $692C   ; arm this routine's #DE handler
  00006148  56                  PUSH SI
  00006149  06                  PUSH ES
  0000614A  B8 DD 3B            MOV AX, $3BDD
  0000614D  8E C0               MOV ES, AX
  0000614F  BF 59 06            MOV DI, $0659
  00006152  8B 36 07 03         MOV SI, [$0307]
  00006156  33 C9               XOR CX, CX
  00006158  8B 6C 04            MOV BP, [SI+$4]
  0000615B  26 89 6D 0C         MOV [ES:DI+$C], BP
  0000615F  AD                  LODSW
  00006160  26 89 45 08         MOV [ES:DI+$8], AX
  00006164  F7 2E B2 26         IMUL WORD [$26B2]
  00006168  F7 FD               IDIV BP
  0000616A  03 06 B4 26         ADD AX, [$26B4]
  0000616E  AB                  STOSW
  0000616F  AD                  LODSW
  00006170  26 89 45 08         MOV [ES:DI+$8], AX
  00006174  F7 2E B0 26         IMUL WORD [$26B0]
  00006178  F7 FD               IDIV BP
  0000617A  03 06 B6 26         ADD AX, [$26B6]
  0000617E  AB                  STOSW
  0000617F  83 C6 04            ADD SI, $0004
  00006182  A5                  MOVSW
  00006183  A5                  MOVSW
  00006184  83 C7 06            ADD DI, $0006
  00006187  41                  INC CX
  00006188  3B 36 05 03         CMP SI, [$0305]


; ============================================================================
; PART 5 — the geometry-prep pipeline: how tiles become drawn polygons.
; ============================================================================
; The visible geometry is a DISPLAY LIST (built once per view change from the
; tile map, then re-rendered each frame; it is cached — 0 writers while the
; player stands still). It is interpreted at 07F7:2BF0 (LODSW a command word;
; JMP [BX+2738] through a ~24-entry jump table at 499D:2738), which for each
; vertex calls 07F7:5096 to turn a tile-space vertex into a view-space one and
; stores it in a pool at 499D:1620 (8 bytes/vertex: X,Y,Z,W). A per-polygon
; gather (07F7:5E2A) then copies a polygon's vertices by index out of that pool,
; attaches texture coords, and hands them to the projection (PART 4). So:
;   tile map -> display list -> [per vertex] 5096 tile->view -> pool 1620
;            -> [per polygon] 5E2A gather -> projection 6148 -> 01A0 raster.

; ==== 07F7:5096 — tile-space vertex -> world (then camera transform) =========
; Reads a vertex as bytes (tileX, tileY, height) from the display list, scales
; each by 32 (SHL 5 = the world units per tile), subtracts the camera position
; (the 0078/0080/03AE offsets), and doubles; 50BE is the same for word coords.
; The result falls through to the camera-matrix multiply (at 50D8).
  00005096  B1 01               MOV CL, $01
  00005098  AC                  LODSB
  00005099  98                  CBW
  0000509A  C1 E0 05            SHL AX, $05
  0000509D  2D 78 00            SUB AX, $0078
  000050A0  D3 E0               SHL AX, CL
  000050A2  8B D8               MOV BX, AX
  000050A4  AC                  LODSB
  000050A5  98                  CBW
  000050A6  C1 E0 05            SHL AX, $05
  000050A9  2D 80 00            SUB AX, $0080
  000050AC  D3 E0               SHL AX, CL
  000050AE  8B E8               MOV BP, AX
  000050B0  AC                  LODSB
  000050B1  98                  CBW
  000050B2  C1 E0 05            SHL AX, $05
  000050B5  2D AE 03            SUB AX, $03AE
  000050B8  D3 E0               SHL AX, CL
  000050BA  8B C8               MOV CX, AX
  000050BC  EB 1A               JMP $000050D8
  000050BE  B1 01               MOV CL, $01

; ==== 01A0:0210 — per-polygon draw setup: THE TEXTURE REFERENCE =============
; The polygon descriptor (at ES:[B07C]) carries, inline, its texture: this reads
; the vertex count, then the texture SEGMENT into [07AF] (which the texture span
; 01A0:02CE samples) and a texture parameter patched into the span at [CS:02E3].
; The texture segment traces back to the tile texture index (word0 10-15 floor /
; word1 0-5 wall) -> the per-level texture list -> the loaded .TR bitmap. It then
; scans the screen vertices at 3BDD:0659 for the polygon's Y extent.
  00000210  41                  INC CX
  00000211  89 0E D1 07         MOV [$07D1], CX
  00000215  26 8B 36 7C B0      MOV SI, [ES:$B07C]
  0000021A  26 AD               LODSW
  0000021C  48                  DEC AX
  0000021D  A3 AD 07            MOV [$07AD], AX
  00000220  83 C6 02            ADD SI, $0002
  00000223  26 AD               LODSW
  00000225  A3 AF 07            MOV [$07AF], AX
  00000228  26 AD               LODSW
  0000022A  2E A3 E3 02         MOV [CS:$02E3], AX
  0000022E  BB F0 D8            MOV BX, $D8F0
  00000231  BD 10 27            MOV BP, $2710
  00000234  BE 59 06            MOV SI, $0659
  00000237  8B FE               MOV DI, SI
  00000239  49                  DEC CX
  0000023A  AD                  LODSW
  0000023B  3B D8               CMP BX, AX
  0000023D  7D 02               JGE $00000241
  0000023F  8B D8               MOV BX, AX
  00000241  3B E8               CMP BP, AX
  00000243  7E 02               JLE $00000247
  00000245  8B E8               MOV BP, AX
  00000247  AD                  LODSW
  00000248  3B 45 02            CMP AX, [DI+$2]
  0000024B  7E 03               JLE $00000250
  0000024D  8D 7C FC            LEA DI, [SI-$4]
  00000250  83 C6 0A            ADD SI, $000A
  00000253  E2 E5               LOOP $0000023A


; ============================================================================
; PART 6 — the display-list BUILDER (segment 1FF9): tiles -> polygons, PER FRAME.
; ============================================================================
; CORRECTION to PART 5: the display list is NOT cached and NOT built "once per
; view change" — that was an artifact of reading a stale SI (065D) from the main
; loop after a breakpoint (07F7:2BF0) that never actually fires. The real display
; list is at 499D:744A and is REBUILT EVERY FRAME by segment 1FF9 (which also
; reads the tile map). The 07F7 side only interprets + transforms it; its dispatch
; runs through handlers at 07F7:2B16 / 5E04 / 79CE (JMP [BX+2738]), not 2BF0.
;
; So the geometry-prep needs no player movement to observe: 1FF9 walks the
; visible tiles each frame, reads each tile's fields, and emits draw commands.

; ==== 1FF9:0E56 — read a tile's fields to decide its geometry ================
; ES:BX from [2F54] points at the current tile. It reads floor HEIGHT (byte0 >>4
; & 0F = word0 bits 4-7), tile TYPE (byte0 & 0F = word0 bits 0-3) and a texture
; field (byte1 >>2 & 0F) — i.e. the exact fields extract/lev decodes — then
; branches on them (CMP ...,$08) to choose which polygons the tile contributes.
  00000E56  C7 06 4A 2F C8 00   MOV WORD [$2F4A], $00C8
  00000E5C  C4 1E 54 2F         LES BX, [$2F54]
  00000E60  26 8A 07            MOV AL, [ES:BX]
  00000E63  C1 E8 04            SHR AX, $04
  00000E66  25 0F 00            AND AX, $000F
  00000E69  88 46 FF            MOV [BP-$1], AL
  00000E6C  C4 1E 3C 2C         LES BX, [$2C3C]
  00000E70  26 8A 47 01         MOV AL, [ES:BX+$1]
  00000E74  24 0F               AND AL, $0F
  00000E76  A2 52 2F            MOV [$2F52], AL
  00000E79  80 3E 52 2F 08      CMP BYTE [$2F52], $08
  00000E7E  73 2A               JNB $00000EAA
  00000E80  C4 1E 54 2F         LES BX, [$2F54]
  00000E84  26 8A 07            MOV AL, [ES:BX]
  00000E87  25 0F 00            AND AX, $000F
  00000E8A  88 46 E9            MOV [BP-$17], AL
  00000E8D  C4 1E 54 2F         LES BX, [$2F54]
  00000E91  26 8A 5F 01         MOV BL, [ES:BX+$1]
  00000E95  C1 EB 02            SHR BX, $02
  00000E98  81 E3 0F 00         AND BX, $000F
  00000E9C  D1 E3               SHL BX, 1
  00000E9E  8A 46 E9            MOV AL, [BP-$17]

; ==== 1FF9:0AD7 — emit a polygon into the display list ======================
; [7250] is the display-list write pointer (LES BX,[7250]; MOV [ES:BX],word; ADD
; [7250],2). It emits command words (e.g. $3E, $02 — the interpreter dispatches
; word/2 through the 2738 jump table) interleaved with parameters (a texture id
; from [2F5A], a height/param from [BP+A], a vertex ref from [2C3A]). One such
; block per polygon; the 07F7 interpreter then transforms and draws it (PART 5).
  00000AD7  8A 46 0A            MOV AL, [BP+$A]
  00000ADA  04 30               ADD AL, $30
  00000ADC  88 46 0A            MOV [BP+$A], AL
  00000ADF  C4 1E 50 72         LES BX, [$7250]
  00000AE3  26 C7 07 3E 00      MOV WORD [ES:BX], $003E
  00000AE8  83 06 50 72 02      ADD WORD [$7250], $0002
  00000AED  C4 1E 50 72         LES BX, [$7250]
  00000AF1  A1 5A 2F            MOV AX, [$2F5A]
  00000AF4  26 89 07            MOV [ES:BX], AX
  00000AF7  83 06 50 72 02      ADD WORD [$7250], $0002
  00000AFC  8A 46 0A            MOV AL, [BP+$A]
  00000AFF  B4 00               MOV AH, $00
  00000B01  C4 1E 50 72         LES BX, [$7250]
  00000B05  26 89 07            MOV [ES:BX], AX
  00000B08  83 06 50 72 02      ADD WORD [$7250], $0002
  00000B0D  C4 1E 50 72         LES BX, [$7250]
  00000B11  A1 3A 2C            MOV AX, [$2C3A]
  00000B14  26 89 07            MOV [ES:BX], AX
  00000B17  83 06 50 72 02      ADD WORD [$7250], $0002
  00000B1C  C4 1E 50 72         LES BX, [$7250]
  00000B20  26 C7 07 02 00      MOV WORD [ES:BX], $0002


; ============================================================================
; PART 7 — OBJECTS and the 3D-MODEL BYTECODE (07F7 display-list interpreter).
; ============================================================================
; The whole 3D view is a bytecode program: 07F7 interprets a display list of
; opcodes via a jump table at 499D:2738 (110+ entries; opcode word = byte offset;
; entries >=0xDC alias the 0x78+ block).
; Every handler ends by fetching+dispatching the next opcode. 3D object models
; (e.g. doors) are sub-programs of the SAME opcodes; a face is CULLED by
; advancing the program pointer past it.

; ==== 07F7:16F8 — the opcode fetch/decode (ends every handler) ===============
  000016F8  AD                  LODSW                 ; next opcode word from the program (DS:SI)
  000016F9  93                  XCHG AX, BX
  000016FA  FF A7 38 27         JMP [BX+$2738]        ; dispatch via the 499D:2738 table

; ==== 07F7:15AC — opcode 15: register an object into the sorted draw list ====
; Transforms the object's world point to camera space and appends a record to
; the distance-ordered list at [2860]; nothing is drawn here.
  000015AC  8D 44 0E            LEA AX, [SI+$E]       ; keep the object's program pointer
  000015B0  AD                  LODSW                 ; (per-vertex link/shift setup)
  000015B3  03 FE               ADD DI, SI
  000015B5  8E 45 F8            MOV ES, [DI-$8]        ; ES = per-object power-of-two shift (distance LOD)
  000015B8  8B 16 BA 26         MOV DX, [$26BA]        ; camera X (lo)   [26BA/26BC]=camX 32-bit
  000015BC  AD                  LODSW
  000015BD  2B D0               SUB DX, AX             ; camX - vertex.x
  000015C3  8B 1E BC 26         MOV BX, [$26BC]
  000015C7  AD                  LODSW
  000015C8  1B D8               SBB BX, AX             ; 32-bit subtract
  000015CF  8C C1               MOV CX, ES             ; shift count -> scale the delta by 2^ES
  000015D1  E3 14               JCXZ $000015E7
  000015D3  79 0A               JNS $000015E1          ; CX>=0 -> shift right (SAR/RCR); <0 -> left
;   ... same for camera Y ([26BE/26C0]) and Z ([26C2/26C4]) -> view X/Y/Z in [01C8]/[01CA]/[01CC]
  0000167C  8B 75 F6            MOV SI, [DI-$A]
  000016AA  E8 C3 FC            CALL $00001370         ; project + range test (CF=1 -> reject)
  000016AD  72 60               JB $0000170F           ; off-screen -> skip this object
  000016B5  C4 3E 60 28         LES DI, [$2860]        ; append record: viewX,viewY,type=0,progptr,...
  000016BF  B8 00 00            MOV AX, $0000          ; type 0 (sprite/model marker)
  000016E5  89 3E 60 28         MOV [$2860], DI        ; advance list head

; ==== 07F7:18FC — the object DRAW pass (walks the [2860] sorted list) ========
; Per object: push the 3x3 camera matrix [1602..1612], rotate the model's frame
; by its heading (5426/558E/56F6), load its base transform (5834), run its
; opcode program, then restore the camera matrix.
  00001941  E8 FC 36            CALL $00005040         ; setup
  00001944  FF 36 12 16         PUSH WORD [$1612]      ; save the camera matrix (9 words 1602..1612)
  00001968  89 2E C6 01         MOV [$01C6], BP        ; BP = the object's model record
  00001984  FF 76 00            PUSH WORD [BP]
  0000198A  E8 A7 3E            CALL $00005834         ; load the model transform / run program
  0000199B  E8 88 3A            CALL $00005426         ; heading rotation about axis 0 (if flagged)
  000019A7  E8 E4 3B            CALL $0000558E         ; ... axis 1
  000019B2  E8 41 3D            CALL $000056F6         ; ... axis 2
  00001A58  8F 06 02 16         POP WORD [$1602]       ; restore the camera matrix afterwards

; ==== 07F7:5426 — compose object heading (one axis) with the camera matrix ===
  00005426  9A 34 0A 4A 21      CALLF $214A:$0A34      ; angle -> sin/cos (table interp)
  00005433  A3 0C 27            MOV [$270C], AX        ; cos
  00005436  89 1E 0A 27         MOV [$270A], BX        ; sin
  00005445  F7 2E 02 16         IMUL WORD [$1602]      ; rotate camera-matrix rows, saturate to +-7FFF
  00005468  89 0E D4 26         MOV [$26D4], CX        ; -> composed matrix [26D4..]

; ==== 07F7:1BE4 — opcode 28: define/transform a model VERTEX =================
  00001BE4  C7 06 82 28 00 00   MOV WORD [$2882], $0000
  00001BEA  AD                  LODSW                 ; relative link to next record (DI = SI + word)
  00001BED  03 FE               ADD DI, SI
  00001BEF  AD                  LODSW                 ; shift (CX) for this vertex
  00001BF8  AD                  LODSW
  00001BF9  2B 06 BA 26         SUB AX, [$26BA]        ; vertex - camera (32-bit), then scale by 2^CX
  00001C0B  E3 14               JCXZ $00001C21         ; (same camera-relative + power-of-two path)

; ==== 07F7:1A8D (op 0x6A) — FACE-begin: plane check, then CALL the face's ====
; draw sub-program at the record tail; cull = skip. (Excerpt from 1BAC on; the
; full record layout is in annotations §10b/§10c. NOTE: 1BAA is NOT an
; instruction boundary — 1BA8 is MOV BP,[288A]; an earlier mid-instruction
; disassembly here produced a phantom instruction, since corrected.)
  00001BAF  E8 BE F7            CALL $00001370         ; project + range test
  00001BB2  72 28               JB $00001BDC           ; vertex off-screen -> drop
  00001BC0  AD                  LODSW
  00001BC2  FF 97 38 27         CALL [BX+$2738]        ; draw the textured polygon (opcode)
  00001BC6  AD                  LODSW
  00001BC9  FF A7 38 27         JMP [BX+$2738]         ; next opcode
; --- the CULL path: a failed plane/visibility check advances SI past the face:
  00001BCD  C7 06 82 28 FF FF   MOV WORD [$2882], $FFFF ; culled marker
  00001BD3  83 C6 0C            ADD SI, $000C          ; SKIP the culled polygon's 12 bytes
  00001BD6  AD                  LODSW
  00001BD8  FF A7 38 27         JMP [BX+$2738]         ; continue the program
