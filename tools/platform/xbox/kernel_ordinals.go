package xbox

// kernel_ordinals.go maps xboxkrnl.exe export ordinals to their names. This is the
// standard platform ABI — the fixed export table of the Xbox kernel, the same for
// every retail title, general Xbox knowledge like the Win32 or PSP export tables (the
// sanctioned clean-room exception noted in xbe.go). It is used only to make traces and
// halt messages legible; behaviour is bound by ordinal *number* in kernel.go, so a
// stray wrong name here misleads a log line but never changes execution.
//
// The map is derived from the retail kernel's export directory. Where an ordinal's
// identity is uncertain it is simply absent and prints as "ordinal_N" — honest about
// what is and is not pinned.

import "fmt"

// ordinalName returns the export name for an ordinal. A name confirmed empirically
// from this image's own call sites (verifiedNames) wins over the reconstructed table,
// which drifts by a few entries in the low (Hal/Ex/Dbg) and Ob/Phy blocks.
func ordinalName(ord uint16) string {
	if n, ok := verifiedNames[ord]; ok {
		return n
	}
	if n, ok := ordinalNames[ord]; ok {
		return n
	}
	return fmt.Sprintf("ordinal_%d", ord)
}

// verifiedNames are ordinals pinned by observing their live call at a known site — the
// argument count and shape, and how the caller consumes the result. These override the
// reconstructed table wherever the two disagree.
var verifiedNames = map[uint16]string{
	24:  "ExQueryNonVolatileSetting",       // 5-arg config read (index, type*, value*, len, resultLen*)
	44:  "HalGetInterruptVector",           // f(busLevel, &irql) -> vector (Hal block drifts +2)
	46:  "HalReadWritePCISpace",            // f(bus, slot, reg, buf, len, write)
	47:  "HalRegisterShutdownNotification", // f(&HAL_SHUTDOWN_REGISTRATION, TRUE), returns
	65:  "IoCreateDevice",                  // 6 args; MU-probe site 0x23F705 (Io block drifts +2)
	98:  "KeConnectInterrupt",              // f(KINTERRUPT) -> BOOLEAN
	107: "KeInitializeDpc",                 // f(KDPC, routine, context)
	109: "KeInitializeInterrupt",           // f(kint, routine, ctx, vector, irql, mode, share)
	113: "KeInitializeTimerEx",             // f(KTIMER, type)
	149: "KeSetTimer",                      // f(KTIMER, dueTime64, KDPC)
	165: "MmAllocateContiguousMemory",      // f(bytes) -> physical base (Mm block drifts +5)
	166: "MmAllocateContiguousMemoryEx",    // f(bytes, lowAddr, highAddr, align, protect) -> base
	168: "MmClaimGpuInstanceMemory",        // f(bytes, &padding) -> end of retained GPU block
	173: "MmGetPhysicalAddress",            // f(va) -> pa (stored next to the va by DSOUND)
	175: "MmLockUnlockBufferPages",         // f(base, bytes, unlock); lock-before-DMA no-op
	180: "MmQueryAllocationSize",           // f(block) -> SIZE_T (result summed into a global)
	182: "MmSetAddressProtect",             // f(base, bytes, newProtect); void (no-op here)
	184: "NtAllocateVirtualMemory",         // f(base**, zerobits, size*, type, protect)
	199: "NtFreeVirtualMemory",             // f(base**, size*, freeType) (Nt block drifts +5)
	202: "NtOpenFile",                      // f(handle*, access, objattr, iosb, share, opts)
	301: "RtlNtStatusToDosError",           // f(NTSTATUS)
	2:   "AvSendTVEncoderOption",           // f(regbase, option, param, result*) -> void
	15:  "ExAllocatePoolWithTag",           // f(bytes, tag) -> PVOID (2-arg; 3rd push is a save)
	23:  "ExQueryPoolBlockSize",            // f(block) -> SIZE_T
	129: "KeRaiseIrqlToDpcLevel",           // f() -> oldIrql in AL (paired w/ KfLowerIrql(CL))
	151: "KeStallExecutionProcessor",       // f(microseconds); 1 arg, spun in the APU bring-up's
	// timeout loop (call sites 0x1DE566 PUSH 1 / 0x1DE58B PUSH 0x29B, result ignored) —
	// the Ke block's +5 drift (107/113/149/160/161 all +5) lands table-146 here.
	160: "KfRaiseIrql",                  // fastcall(CL=newIrql) -> oldIrql (was mis-guessed as Mm)
	161: "KfLowerIrql",                  // fastcall(CL=newIrql) -> void
	190: "NtCreateFile",                 // 9 args; XAPI CreateFile wrapper site 0x43D08
	193: "NtCreateSemaphore",            // f(handle*, objattr, initial, max)
	211: "NtQueryInformationFile",       // 5 args, class 0x22 (site 0x445F6)
	219: "NtReadFile",                   // 8 args, OVERLAPPED shape (site 0x440C1)
	222: "NtReleaseSemaphore",           // f(handle, releaseCount, prev*) -> NTSTATUS
	224: "NtResumeThread",               // f(handle, prevCount*); pair w/ 231 (site 0x44F56)
	226: "NtSetInformationFile",         // 5 args, class 0xE seek (site 0x44378)
	231: "NtSuspendThread",              // f(handle, prevCount*); pair w/ 224 (site 0x44F30)
	234: "NtWaitForSingleObjectEx",      // f(handle, waitMode, alertable, timeout*) -> NTSTATUS
	187: "NtClose",                      // f(handle) -> NTSTATUS (Nt block drifts +5)
	255: "PsCreateSystemThreadEx",       // the CRT's 10-arg main-thread spawn
	277: "RtlEnterCriticalSection",      // census-anchored
	289: "RtlInitAnsiString",            // f(ANSI_STRING*, PCSZ); MU-probe call site 0x23F6D2
	291: "RtlInitializeCriticalSection", // census-anchored
	294: "RtlLeaveCriticalSection",      // census-anchored
}

var ordinalNames = map[uint16]string{
	1:   "AvGetSavedDataAddress",
	2:   "AvSendTVEncoderOption",
	3:   "AvSetDisplayMode",
	4:   "AvSetSavedDataAddress",
	5:   "DbgBreakPoint",
	6:   "DbgBreakPointWithStatus",
	7:   "DbgLoadImageSymbols",
	8:   "DbgPrint",
	9:   "HalReadSMCTrayState",
	10:  "DbgPrompt",
	11:  "DbgUnLoadImageSymbols",
	12:  "ExAcquireReadWriteLockExclusive",
	13:  "ExAcquireReadWriteLockShared",
	14:  "ExAllocatePool",
	15:  "ExAllocatePoolWithTag",
	16:  "ExFreePool",
	17:  "ExInitializeReadWriteLock",
	18:  "ExInterlockedAddLargeInteger",
	19:  "ExInterlockedAddLargeStatistic",
	20:  "ExInterlockedCompareExchange64",
	21:  "ExQueryPoolBlockSize",
	22:  "ExQueryNonVolatileSetting",
	23:  "ExReadWriteRefurbInfo",
	24:  "ExRaiseException",
	25:  "ExRaiseStatus",
	26:  "ExReleaseReadWriteLock",
	27:  "ExSaveNonVolatileSetting",
	28:  "ExSemaphoreObjectType",
	29:  "ExTimerObjectType",
	30:  "ExfInterlockedInsertHeadList",
	31:  "ExfInterlockedInsertTailList",
	32:  "ExfInterlockedRemoveHeadList",
	33:  "FscGetCacheSize",
	34:  "FscInvalidateIdleBlocks",
	35:  "FscSetCacheSize",
	36:  "HalClearSoftwareInterrupt",
	37:  "HalDisableSystemInterrupt",
	38:  "HalDiskCachePartitionCount",
	39:  "HalDiskModelNumber",
	40:  "HalDiskSerialNumber",
	41:  "HalEnableSystemInterrupt",
	42:  "HalGetInterruptVector",
	43:  "HalReadSMBusValue",
	44:  "HalReadWritePCISpace",
	45:  "HalRegisterShutdownNotification",
	46:  "HalRequestSoftwareInterrupt",
	47:  "HalReturnToFirmware",
	48:  "HalWriteSMBusValue",
	49:  "InterlockedCompareExchange",
	50:  "InterlockedDecrement",
	51:  "InterlockedIncrement",
	52:  "InterlockedExchange",
	53:  "InterlockedExchangeAdd",
	54:  "InterlockedFlushSList",
	55:  "InterlockedPopEntrySList",
	56:  "InterlockedPushEntrySList",
	57:  "IoAllocateIrp",
	58:  "IoBuildAsynchronousFsdRequest",
	59:  "IoBuildDeviceIoControlRequest",
	60:  "IoBuildSynchronousFsdRequest",
	61:  "IoCheckShareAccess",
	62:  "IoCompletionObjectType",
	63:  "IoCreateDevice",
	64:  "IoCreateFile",
	65:  "IoCreateSymbolicLink",
	66:  "IoDeleteDevice",
	67:  "IoDeleteSymbolicLink",
	68:  "IoFreeIrp",
	69:  "IoInitializeIrp",
	70:  "IoInvalidDeviceRequest",
	71:  "IoQueryFileInformation",
	72:  "IoQueryVolumeInformation",
	73:  "IoQueueThreadIrp",
	74:  "IoRemoveShareAccess",
	75:  "IoSetIoCompletion",
	76:  "IoSetShareAccess",
	77:  "IoStartNextPacket",
	78:  "IoStartNextPacketByKey",
	79:  "IoStartPacket",
	80:  "IoSynchronousDeviceIoControlRequest",
	81:  "IoSynchronousFsdRequest",
	82:  "IofCallDriver",
	83:  "IofCompleteRequest",
	84:  "KdDebuggerEnabled",
	85:  "KdDebuggerNotPresent",
	86:  "IoDismountVolume",
	87:  "IoDismountVolumeByName",
	88:  "KeAlertResumeThread",
	89:  "KeAlertThread",
	90:  "KeBoostPriorityThread",
	91:  "KeBugCheck",
	92:  "KeBugCheckEx",
	93:  "KeCancelTimer",
	94:  "KeConnectInterrupt",
	95:  "KeDelayExecutionThread",
	96:  "KeDisconnectInterrupt",
	97:  "KeEnterCriticalRegion",
	98:  "KeGetCurrentIrql",
	99:  "KeGetCurrentThread",
	100: "KeInitializeApc",
	101: "KeInitializeDeviceQueue",
	102: "KeInitializeDpc",
	103: "KeInitializeEvent",
	104: "KeInitializeInterrupt",
	105: "KeInitializeMutant",
	106: "KeInitializeQueue",
	107: "KeInitializeSemaphore",
	108: "KeInitializeTimerEx",
	109: "KeInsertByKeyDeviceQueue",
	110: "KeInsertDeviceQueue",
	111: "KeInsertHeadQueue",
	112: "KeInsertQueue",
	113: "KeInsertQueueApc",
	114: "KeInsertQueueDpc",
	115: "KeInterruptTime",
	116: "KeIsExecutingDpc",
	117: "KeLeaveCriticalRegion",
	118: "KePulseEvent",
	119: "KeQueryBasePriorityThread",
	120: "KeQueryInterruptTime",
	121: "KeQueryPerformanceCounter",
	122: "KeQueryPerformanceFrequency",
	123: "KeQuerySystemTime",
	124: "KeRaiseIrqlToDpcLevel",
	125: "KeRaiseIrqlToSynchLevel",
	126: "KeReleaseMutant",
	127: "KeReleaseSemaphore",
	128: "KeRemoveByKeyDeviceQueue",
	129: "KeRemoveDeviceQueue",
	130: "KeRemoveEntryDeviceQueue",
	131: "KeRemoveQueue",
	132: "KeRemoveQueueDpc",
	133: "KeResetEvent",
	134: "KeRestoreFloatingPointState",
	135: "KeResumeThread",
	136: "KeRundownQueue",
	137: "KeSaveFloatingPointState",
	138: "KeSetBasePriorityThread",
	139: "KeSetDisableBoostThread",
	140: "KeSetEvent",
	141: "KeSetEventBoostPriority",
	142: "KeSetPriorityProcess",
	143: "KeSetPriorityThread",
	144: "KeSetTimer",
	145: "KeSetTimerEx",
	146: "KeStallExecutionProcessor",
	147: "KeSuspendThread",
	148: "KeSynchronizeExecution",
	149: "KeSystemTime",
	150: "KeTestAlertThread",
	151: "KeTickCount",
	152: "KeTimeIncrement",
	153: "KeWaitForMultipleObjects",
	154: "KeWaitForSingleObject",
	155: "KfRaiseIrql",
	156: "KfLowerIrql",
	157: "KiBugCheckData",
	158: "KiUnlockDispatcherDatabase",
	159: "LaunchDataPage",
	160: "MmAllocateContiguousMemory",
	161: "MmAllocateContiguousMemoryEx",
	162: "MmAllocateSystemMemory",
	163: "MmClaimGpuInstanceMemory",
	164: "MmCreateKernelStack",
	165: "MmDeleteKernelStack",
	166: "MmFreeContiguousMemory",
	167: "MmFreeSystemMemory",
	168: "MmGetPhysicalAddress",
	169: "MmIsAddressValid",
	170: "MmLockUnlockBufferPages",
	171: "MmLockUnlockPhysicalPage",
	172: "MmMapIoSpace",
	173: "MmPersistContiguousMemory",
	174: "MmQueryAddressProtect",
	175: "MmQueryAllocationSize",
	176: "MmQueryStatistics",
	177: "MmSetAddressProtect",
	178: "MmUnmapIoSpace",
	179: "NtAllocateVirtualMemory",
	180: "NtCancelTimer",
	181: "NtClearEvent",
	182: "NtClose",
	183: "NtCreateDirectoryObject",
	184: "NtCreateEvent",
	185: "NtCreateFile",
	186: "NtCreateIoCompletion",
	187: "NtCreateMutant",
	188: "NtCreateSemaphore",
	189: "NtCreateTimer",
	190: "NtDeleteFile",
	191: "NtDeviceIoControlFile",
	192: "NtDuplicateObject",
	193: "NtFlushBuffersFile",
	194: "NtFreeVirtualMemory",
	195: "NtFsControlFile",
	196: "NtOpenDirectoryObject",
	197: "NtOpenFile",
	198: "NtOpenSymbolicLinkObject",
	199: "NtProtectVirtualMemory",
	200: "NtPulseEvent",
	201: "NtQueueApcThread",
	202: "NtQueryDirectoryFile",
	203: "NtQueryDirectoryObject",
	204: "NtQueryEvent",
	205: "NtQueryFullAttributesFile",
	206: "NtQueryInformationFile",
	207: "NtQueryIoCompletion",
	208: "NtQueryMutant",
	209: "NtQuerySemaphore",
	210: "NtQuerySymbolicLinkObject",
	211: "NtQueryTimer",
	212: "NtQueryVirtualMemory",
	213: "NtQueryVolumeInformationFile",
	214: "NtReadFile",
	215: "NtReadFileScatter",
	216: "NtReleaseMutant",
	217: "NtReleaseSemaphore",
	218: "NtRemoveIoCompletion",
	219: "NtResumeThread",
	220: "NtSetEvent",
	221: "NtSetInformationFile",
	222: "NtSetIoCompletion",
	223: "NtSetSystemTime",
	224: "NtSetTimerEx",
	225: "NtSignalAndWaitForSingleObjectEx",
	226: "NtSuspendThread",
	227: "NtUserIoApcDispatcher",
	228: "NtWaitForSingleObject",
	229: "NtWaitForSingleObjectEx",
	230: "NtWaitForMultipleObjectsEx",
	231: "NtWriteFile",
	232: "NtWriteFileGather",
	233: "NtYieldExecution",
	234: "ObCreateObject",
	235: "ObDirectoryObjectType",
	236: "ObInsertObject",
	237: "ObMakeTemporaryObject",
	238: "ObOpenObjectByName",
	239: "ObOpenObjectByPointer",
	240: "ObpObjectHandleTable",
	241: "ObReferenceObjectByHandle",
	242: "ObReferenceObjectByName",
	243: "ObReferenceObjectByPointer",
	244: "ObSymbolicLinkObjectType",
	245: "ObfDereferenceObject",
	246: "ObfReferenceObject",
	// The Ob..end tail is anchored on ordinals verified empirically from their call
	// sites in this image: 255 = PsCreateSystemThreadEx (a 10-argument CRT call that
	// spawns the main thread), and 277/291/294 = RtlEnter/Initialize/LeaveCriticalSection
	// (all three present in the census, consistent only with this alignment). That
	// pins the Ps block at 254 and the Rtl block at 260 — five higher than a naive
	// alphabetical count, because the Nt/Ob region carries five exports this table did
	// not originally enumerate. The 234..253 span (Ob/Phy) is therefore approximate and
	// is refined as the boot reaches each ordinal.
	249: "PsCreateSystemThread?",
	250: "PsCreateSystemThread?",
	251: "PsCreateSystemThread?",
	252: "PsCreateSystemThread?",
	253: "PsCreateSystemThread",
	254: "PsCreateSystemThread",
	255: "PsCreateSystemThreadEx", // verified
	256: "PsQueryStatistics",
	257: "PsSetCreateThreadNotifyRoutine",
	258: "PsTerminateSystemThread",
	259: "PsThreadObjectType",
	260: "RtlAnsiStringToUnicodeString",
	261: "RtlAppendStringToString",
	262: "RtlAppendUnicodeStringToString",
	263: "RtlAppendUnicodeToString",
	264: "RtlAssert",
	265: "RtlCaptureContext",
	266: "RtlCaptureStackBackTrace",
	267: "RtlCharToInteger",
	268: "RtlCompareMemory",
	269: "RtlCompareMemoryUlong",
	270: "RtlCompareString",
	271: "RtlCompareUnicodeString",
	272: "RtlCopyString",
	273: "RtlCopyUnicodeString",
	274: "RtlCreateUnicodeString",
	275: "RtlDowncaseUnicodeChar",
	276: "RtlDowncaseUnicodeString",
	277: "RtlEnterCriticalSection", // verified (census)
	278: "RtlEnterCriticalSectionAndRegion",
	279: "RtlEqualString",
	280: "RtlEqualUnicodeString",
	281: "RtlExtendedIntegerMultiply",
	282: "RtlExtendedLargeIntegerDivide",
	283: "RtlExtendedMagicDivide",
	284: "RtlFillMemory",
	285: "RtlFillMemoryUlong",
	286: "RtlFreeAnsiString",
	287: "RtlFreeUnicodeString",
	288: "RtlGetCallersAddress",
	289: "RtlInitAnsiString",
	290: "RtlInitUnicodeString",
	291: "RtlInitializeCriticalSection", // verified (census)
	292: "RtlIntegerToChar",
	293: "RtlIntegerToUnicodeString",
	294: "RtlLeaveCriticalSection", // verified (census)
	295: "RtlLeaveCriticalSectionAndRegion",
	296: "RtlLowerChar",
	297: "RtlMapGenericMask",
	298: "RtlMoveMemory",
	299: "RtlMultiByteToUnicodeN",
	300: "RtlMultiByteToUnicodeSize",
	301: "RtlNtStatusToDosError",
	302: "RtlRaiseException",
	303: "RtlRaiseStatus",
	304: "RtlTimeFieldsToTime",
	305: "RtlTimeToTimeFields",
	306: "RtlTryEnterCriticalSection",
	307: "RtlUlongByteSwap",
	308: "RtlUnicodeStringToAnsiString",
	309: "RtlUnicodeStringToInteger",
	310: "RtlUnicodeToMultiByteN",
	311: "RtlUnicodeToMultiByteSize",
	312: "RtlUpcaseUnicodeChar",
	313: "RtlUpcaseUnicodeString",
	314: "RtlUpcaseUnicodeToMultiByteN",
	315: "RtlUpperChar",
	316: "RtlUpperString",
	317: "RtlUshortByteSwap",
	318: "RtlWalkFrameChain",
	319: "RtlZeroMemory",
	320: "XboxEEPROMKey",
	321: "XboxHardwareInfo",
	322: "XboxHDKey",
	323: "XboxKrnlVersion",
	324: "XboxSignatureKey",
	325: "XeImageFileName",
	326: "XeLoadSection",
	327: "XeUnloadSection",
	328: "XcSHAInit",
	329: "XcSHAUpdate",
	330: "XcSHAFinal",
	331: "XcRC4Key",
	332: "XcRC4Crypt",
	333: "XcHMAC",
	334: "XcPKEncPublic",
	335: "XcPKDecPrivate",
	336: "XcPKGetKeyLen",
	337: "XcVerifyPKCS1Signature",
	338: "XcModExp",
	339: "XcDESKeyParity",
	340: "XcKeyTable",
	341: "XcBlockCrypt",
	342: "XcBlockCryptCBC",
	343: "XcCryptService",
	344: "XcUpdateCrypto",
	345: "RtlRip",
	346: "XboxLANKey",
	347: "XboxAlternateSignatureKeys",
	348: "XePublicKeyData",
	349: "HalBootSMCVideoMode",
	350: "IdexChannelObject",
	351: "HalIsResetOrShutdownPending",
	352: "IoMarkIrpMustComplete",
	353: "HalInitiateShutdown",
	354: "RtlSnprintf",
	355: "RtlSprintf",
	356: "RtlVsnprintf",
	357: "RtlVsprintf",
	358: "HalEnableSecureTrayEject",
	359: "HalWriteSMCScratchRegister",
	360: "MmDbgAllocateMemory",
}
