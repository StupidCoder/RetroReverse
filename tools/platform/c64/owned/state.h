#pragma once
#include "system.h"
#include <string>

namespace rr::c64 {
// SHA-256 identities supplied by the host after hashing the actual inputs.
// Zero means absent. Disk identifies the original image, not its mutable tracks.
// configuration distinguishes host choices affecting execution (PAL, drive,
// write protection, etc.). A build/core digest must change when semantics change.
struct StateIdentity {
 using Digest = std::array<uint8_t,32>;
 Digest core{},basic{},kernal{},characters{},driveRom{},tape{},disk{};
 uint32_t configuration=0;
 bool operator==(const StateIdentity&)const=default;
};
// Field-wise, versioned little-endian snapshots, never C++ object/heap dumps.
// Media/firmware stay with the caller; disk mutations are included. Host-side
// debugger UI, breakpoint lists and observation history are not hardware state.
// Loading is atomic: failure leaves every live machine field untouched.
std::vector<uint8_t> saveState(const Board&,const StateIdentity&);
std::vector<uint8_t> saveState(const System&,const StateIdentity&);
bool loadState(Board&,std::span<const uint8_t>,const StateIdentity&,std::string& error);
bool loadState(System&,std::span<const uint8_t>,const StateIdentity&,std::string& error);
}
