// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build freebsd
// +build arm

package runtime

import "internal/cpu"

const (
	_HWCAP_VFP   = 1 << 6
	_HWCAP_VFPv3 = 1 << 13
)

// AT_HWCAP is not available on FreeBSD-11.1-RELEASE or earlier.
// Default to mandatory VFP hardware support for arm being available.
// If AT_HWCAP is available goarmHWCap will be updated in archauxv.
// TODO(moehrmann) remove once all go supported FreeBSD versions support _AT_HWCAP.
var goarmHWCap uint = (_HWCAP_VFP | _HWCAP_VFPv3)

func checkgoarm() {
	// Update cpu.HWCap to match goarmHWCap in case they were not updated in archauxv.
	cpu.HWCap = goarmHWCap

	if goarm > 5 && cpu.HWCap&_HWCAP_VFP == 0 {
		print("runtime: this CPU has no floating point hardware, so it cannot run\n")
		print("this GOARM=", goarm, " binary. Recompile using GOARM=5.\n")
		exit(1)
	}
	if goarm > 6 && cpu.HWCap&_HWCAP_VFPv3 == 0 {
		print("runtime: this CPU has no VFPv3 floating point hardware, so it cannot run\n")
		print("this GOARM=", goarm, " binary. Recompile using GOARM=5 or GOARM=6.\n")
		exit(1)
	}

	// osinit not called yet, so ncpu not set: must use getncpu directly.
	if getncpu() > 1 && goarm < 7 {
		print("runtime: this system has multiple CPUs and must use\n")
		print("atomic synchronization instructions. Recompile using GOARM=7.\n")
		exit(1)
	}
}

func archauxv(tag, val uintptr) {
	switch tag {
	case _AT_TIMEKEEP:
		timekeepSharedPage = (*vdsoTimekeep)(unsafe.Pointer(val))
	case _AT_HWCAP:
		cpu.HWCap = uint(val)
		goarmHWCap = cpu.HWCap
	case _AT_HWCAP2:
		cpu.HWCap2 = uint(val)
	}
}

//go:nosplit
func cputicks() int64 {
	// Currently cputicks() is used in blocking profiler and to seed runtime·fastrand().
	// runtime·nanotime() is a poor approximation of CPU ticks that is enough for the profiler.
	// TODO: need more entropy to better seed fastrand.
	return nanotime()
}
