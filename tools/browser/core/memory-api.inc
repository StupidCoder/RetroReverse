extern "C" {
const char*rr_inspect_regions(){rrMemoryRegions();return rrmem::describe();}
const uint8_t*rr_inspect_data(uint32_t region){return region<rrmem::regions.size()?rrmem::regions[region].data:nullptr;}
void rr_activity_begin(uint32_t mask){rrmem::events.clear();rrmem::dropped=0;rrmem::mask=mask&7;rrmem::active=rrmem::supported;}
void rr_activity_end(){rrmem::active=false;}
uint32_t rr_activity_count(){return rrmem::events.size();}
uint32_t rr_activity_dropped(){return rrmem::dropped;}
const uint8_t*rr_activity_data(){return reinterpret_cast<const uint8_t*>(rrmem::events.data());}
}
