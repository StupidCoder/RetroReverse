#pragma once
// Portfolio movie playback is explicitly HLE: decode the stream requested by
// the guest, while suspending its unsupported DSP/DataStreamer player.
// It uses the same VRAM, capture journal and display boundaries as the game.
inline double presentationSeconds = 0, movieSeconds = 0;
inline bool movieActive(threedo_Machine* m) {
  return m->MovieHLE && m->moviePos < m->movieQueue.n;
}
inline double movieDuration(threedo_CvidMovie* movie,int64_t i) {
  if(i+1<movie->Times.n && movie->Times[i+1]>movie->Times[i])
    return (movie->Times[i+1]-movie->Times[i])/240.0;
  if(i<movie->Durations.n && movie->Durations[i]>0 && movie->Durations[i]<=2400)
    return movie->Durations[i]/240.0;
  return 1.0/(movie->FPS>0?movie->FPS:15);
}
inline bool presentMovie(threedo_Machine* m) {
  if(!movieActive(m))return false;
  rrprof::Scope clock(6,"Cinepak movie HLE");
  // Drop completed payloads explicitly: the translated object arena otherwise
  // retains every attract-loop movie until reset.
  while(movieActive(m)) {
    auto movie=m->movieQueue[m->moviePos]->mv;
    if(m->movieDec && m->movieFrameIdx>=movie->Frames.n) {
      movie->Frames={};movie->Times={};movie->Durations={};
      if(m->movieDec->img)m->movieDec->img->Pix={};
      m->movieDec=nullptr;m->moviePos++;m->movieFrameIdx=0;
      continue;
    }
    const auto index=m->movieDec?m->movieFrameIdx:0;
    if(rrcapture::trace.active) {
      std::ostringstream event;
      event<<"{\"kind\":\"Cinepak movie HLE\",\"frame\":"<<index
           <<",\"width\":"<<movie->Width<<",\"height\":"<<movie->Height<<"}";
      rrcapture::trace.event(m->CPU->Instrs,m->CPU->cur,event.str());
      rrcapture::trace.source=rrcapture::trace.palette=rrcapture::trace.texel=0;
    }
    const auto duration=movieDuration(movie,index);
    auto result=threedo_Machine_StepMovieFrame(m);
    if(!std::get<4>(result))return false;
    ++m->frame;presentationSeconds+=duration;movieSeconds+=duration;
    return true;
  }
  return false;
}
