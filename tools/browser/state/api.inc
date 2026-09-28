// Each core provides stateWrite/archive and stateRead/archive before inclusion.
static std::vector<uint8_t> stateOutput,stateInput;
extern "C" {
uint8_t* rr_state_input(uint32_t n){try{if(n>rrstate::limit)throw std::runtime_error("State too large");stateInput.resize(n);return stateInput.data();}catch(...){return nullptr;}}
uint32_t rr_state_save(){try{rrstate::Archive a;stateWrite(a);stateOutput=std::move(a.bytes);return stateOutput.size();}catch(...){return 0;}}
uint8_t* rr_state_data(){return stateOutput.data();}
int rr_state_load(uint32_t n){try{if(n!=stateInput.size())throw std::runtime_error("State length mismatch");rrstate::Archive a(stateInput.data(),n);stateRead(a);return 1;}catch(...){return 0;}}
}
