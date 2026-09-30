//go:build ignore

// Regenerate the public, synthetic decoder differential fixture.
package main
import("fmt";"encoding/json";"retroreverse.com/tools/cpu/x86")
func main(){fmt.Println("// Generated from tools/cpu/x86. Synthetic opcode vectors; no game bytes.\nstruct DecodeVector {bool wide;std::array<uint8_t,15> bytes;int length;const char*text;};\nstatic const DecodeVector decodeVectors[]={")
 for _,wide:=range []bool{false,true}{for op:=0;op<256;op++{var b [15]byte;for i:=range b{b[i]=byte(0x24+i*17)};b[0]=byte(op);d:=x86.Decode(b[:],0x1234);if wide{d=x86.Decode32(b[:],0x1234)};text,_:=json.Marshal(d.Text);fmt.Printf("{%v,{",wide);for i,v:=range b{if i>0{fmt.Print(",")};fmt.Print(v)};fmt.Printf("},%d,%s},\n",d.Len,text)}};fmt.Println("};")}
