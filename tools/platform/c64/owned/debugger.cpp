#include "debugger.h"
#include <algorithm>
#include <cstdio>
namespace rr::c64 {
const CpuState& Debugger::cpu(Processor p)const{return p==Processor::Drive?machine.drive.state.cpu:machine.board.state.cpu;}
uint8_t Debugger::peek(Processor p,uint16_t a)const{return p==Processor::Drive?machine.drive.peek(a):machine.board.peek(a);}
void Debugger::breakpoint(Processor p,uint16_t a,bool enabled){auto it=std::find_if(points.begin(),points.end(),[&](const auto& b){return b.processor==p&&b.address==a;});if(enabled){if(it==points.end())points.push_back({p,a});}else if(it!=points.end())points.erase(it);}
void Debugger::ignoreCurrent(Processor p){const auto& c=cpu(p);bypass=Bypass{p,c.bus.address,c.retired};}
void Debugger::resume(){if(paused&&last.reason==StopReason::Breakpoint)ignoreCurrent(last.processor);paused=false;}
void Debugger::restore(const SystemSnapshot& snapshot){machine.restore(snapshot);paused=true;bypass.reset();last={};clearEvents();}
bool Debugger::hit(Processor p){
 const auto& c=cpu(p);
 if(bypass&&bypass->processor==p){if(c.stage==Stage::Fetch&&c.bus.address==bypass->address&&c.retired==bypass->retired)return false;bypass.reset();}
 if(c.stage!=Stage::Fetch)return false;
 return std::any_of(points.begin(),points.end(),[&](const auto& b){return b.processor==p&&b.address==c.bus.address;});
}
DebugStop Debugger::run(uint64_t n){if(paused)return {StopReason::Paused,last.processor,last.address,0};return execute(n,{});}
DebugStop Debugger::step(Processor p,uint64_t n){ignoreCurrent(p);paused=false;const auto result=execute(n,p);paused=true;return result;}
DebugStop Debugger::execute(uint64_t n,std::optional<Processor> stepping){
 const auto retired=stepping?cpu(*stepping).retired:0;
 for(uint64_t i=0;i<n;i++){
  const auto p=machine.nextProcessor();const auto address=cpu(p).bus.address;
  if(hit(p)){paused=true;return last={StopReason::Breakpoint,p,address,i};}
  const auto transitions=machine.iec.transitions;
  if(!machine.tickEdge()){paused=true;return last={StopReason::Fault,p,address,i+1};}
  if(machine.iec.transitions!=transitions&&capacity){if(history.size()==capacity){history.pop_front();++dropped;}history.push_back(machine.iec.last);}
  if(stepping&&cpu(*stepping).retired!=retired&&cpu(*stepping).stage==Stage::Fetch)return last={StopReason::Step,*stepping,cpu(*stepping).bus.address,i+1};
 }
 const auto p=stepping.value_or(machine.nextProcessor());return last={StopReason::Budget,p,cpu(p).bus.address,n};
}
Disassembly Debugger::disassemble(Processor p,uint16_t a)const{
 static constexpr const char* names[]={"???","JAM","SLO","RLA","SRE","RRA","SAX","LAX","DCP","ISC","ANC","ALR","ARR","AXS","LAS","ADC","AND","ASL","BCC","BCS","BEQ","BIT","BMI","BNE","BPL","BRK","BVC","BVS","CLC","CLD","CLI","CLV","CMP","CPX","CPY","DEC","DEX","DEY","EOR","INC","INX","INY","JMP","JSR","LDA","LDX","LDY","LSR","NOP","ORA","PHA","PHP","PLA","PLP","ROL","ROR","RTI","RTS","SBC","SEC","SED","SEI","STA","STX","STY","TAX","TAY","TSX","TXA","TXS","TYA"};
 static_assert(std::size(names)==unsigned(Op::TYA)+1);
 Disassembly out;out.address=a;out.bytes[0]=peek(p,a);const auto instruction=instructions[out.bytes[0]];out.op=instruction.op;out.mode=instruction.mode;
 switch(out.mode){case Mode::imp:case Mode::acc:out.length=1;break;case Mode::abs:case Mode::abx:case Mode::aby:case Mode::ind:out.length=3;break;default:out.length=2;}
 for(unsigned i=1;i<out.length;i++)out.bytes[i]=peek(p,uint16_t(a+i));
 const unsigned byte=out.bytes[1],word=byte|(unsigned(out.bytes[2])<<8);char operand[32]={};
 switch(out.mode){
 case Mode::imp:break;case Mode::acc:std::snprintf(operand,sizeof operand," A");break;
 case Mode::imm:std::snprintf(operand,sizeof operand," #$%02X",byte);break;
 case Mode::zp:std::snprintf(operand,sizeof operand," $%02X",byte);break;
 case Mode::zpx:std::snprintf(operand,sizeof operand," $%02X,X",byte);break;
 case Mode::zpy:std::snprintf(operand,sizeof operand," $%02X,Y",byte);break;
 case Mode::izx:std::snprintf(operand,sizeof operand," ($%02X,X)",byte);break;
 case Mode::izy:std::snprintf(operand,sizeof operand," ($%02X),Y",byte);break;
 case Mode::rel:std::snprintf(operand,sizeof operand," $%04X",unsigned(uint16_t(a+2+int8_t(byte))));break;
 case Mode::abs:std::snprintf(operand,sizeof operand," $%04X",word);break;
 case Mode::abx:std::snprintf(operand,sizeof operand," $%04X,X",word);break;
 case Mode::aby:std::snprintf(operand,sizeof operand," $%04X,Y",word);break;
 case Mode::ind:std::snprintf(operand,sizeof operand," ($%04X)",word);break;
 }
 out.text=std::string(names[unsigned(out.op)])+operand;return out;
}
}
