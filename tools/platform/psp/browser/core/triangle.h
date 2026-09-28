#pragma once
// Keep operation order identical to the scalar renderer. Reusing barycentric
// weights avoids constructing/copying three vertices and evaluating three
// edges again just to compute the current fragment's texture coordinates.
inline std::pair<float,float>rrUV(const psp_vert&a,const psp_vert&b,const psp_vert&c,float l0,float l1,float l2){
 if(!a.clip)return {l0*a.u+l1*b.u+l2*c.u,l0*a.v+l1*b.v+l2*c.v};
 float iw=l0*a.invW+l1*b.invW+l2*c.invW;if(!iw)return {0,0};
 return {(l0*a.u*a.invW+l1*b.u*b.invW+l2*c.u*c.invW)/iw,(l0*a.v*a.invW+l1*b.v*b.invW+l2*c.v*c.invW)/iw};
}
inline void psp_Machine_rasterTri(psp_Machine*m,psp_geState*s,psp_vert a,psp_vert b,psp_vert c){
 int minX=psp_clampI(cast<int64_t>(psp_fmin3(a.x,b.x,c.x)),0,480),maxX=psp_clampI(cast<int64_t>(psp_fmax3(a.x,b.x,c.x))+1,0,480);
 int minY=psp_clampI(cast<int64_t>(psp_fmin3(a.y,b.y,c.y)),0,272),maxY=psp_clampI(cast<int64_t>(psp_fmax3(a.y,b.y,c.y))+1,0,272);
 float area=psp_edge(a,b,c);if(!area)return;
 if(s->cullOn&&!psp_geNoCull&&!s->clearOn&&!psp_vert_through(a))if(((s->cullFace==0)==(area<0))!=psp_geCullFlip)return;
 // These differences do not depend on the fragment. Preserve each subtraction
 // and multiplication rather than accumulating edge values (which drifts).
 rrFloat4 ca={float(a.r),float(a.g),float(a.b),float(a.a)},cb={float(b.r),float(b.g),float(b.b),float(b.a)},cc={float(c.r),float(c.g),float(c.b),float(c.a)};
 float e0x=c.x-b.x,e0y=c.y-b.y,e1x=a.x-c.x,e1y=a.y-c.y,e2x=b.x-a.x,e2y=b.y-a.y;
 bool textured=s->texEnable&&!s->clearOn,fog=s->fogOn&&a.clip&&!s->clearOn;
 for(int y=minY;y<maxY;y++){
  float py=float(y)+0.5f;
  float e0row=(py-b.y)*e0x,e1row=(py-c.y)*e1x,e2row=(py-a.y)*e2x;
  for(int x=minX;x<maxX;x++){
   float px=float(x)+0.5f,w0=(px-b.x)*e0y-e0row,w1=(px-c.x)*e1y-e1row,w2=(px-a.x)*e2y-e2row;
   if((w0<0||w1<0||w2<0)&&(w0>0||w1>0||w2>0))continue;
   float l0=w0/area,l1=w1/area,l2=w2/area;
   float z=l0*a.z+l1*b.z+l2*c.z;
   auto color=ca*l0+cb*l1+cc*l2;
   uint8_t r=cast<uint8_t>(color[0]),g=cast<uint8_t>(color[1]),bl=cast<uint8_t>(color[2]),al=cast<uint8_t>(color[3]);
   if(textured){float u,v,rho=1;
    if(s->texMaxLvl){
     rrFloat4 xx={px,float(x)+1.5f,px,px},yy={py,py,float(y)+1.5f,py};
     rrFloat4 q0=((xx-b.x)*e0y-(yy-b.y)*e0x)/area,q1=((xx-c.x)*e1y-(yy-c.y)*e1x)/area,q2=((xx-a.x)*e2y-(yy-a.y)*e2x)/area;
     rrFloat4 uu,vv;
     if(a.clip){auto iw=q0*a.invW+q1*b.invW+q2*c.invW;uu=(q0*a.u*a.invW+q1*b.u*b.invW+q2*c.u*c.invW)/iw;vv=(q0*a.v*a.invW+q1*b.v*b.invW+q2*c.v*c.invW)/iw;for(int i=0;i<3;i++)if(iw[i]==0)uu[i]=vv[i]=0;}
     else{uu=q0*a.u+q1*b.u+q2*c.u;vv=q0*a.v+q1*b.v+q2*c.v;}
     u=uu[0];v=vv[0];float dx=psp_hypot32((uu[1]-u)*float(s->texW),(vv[1]-v)*float(s->texH)),dy=psp_hypot32((uu[2]-u)*float(s->texW),(vv[2]-v)*float(s->texH));rho=dy>dx?dy:dx;
    }else std::tie(u,v)=rrUV(a,b,c,l0,l1,l2);
    std::tie(r,g,bl,al)=psp_modTex(m,s,u,v,rho,r,g,bl,al);
   }
   if(fog)std::tie(r,g,bl)=psp_applyFog(s,l0*a.fog+l1*b.fog+l2*c.fog,r,g,bl);
   psp_Machine_putPixel(m,s,x,y,z,r,g,bl,al);
  }
 }
}
