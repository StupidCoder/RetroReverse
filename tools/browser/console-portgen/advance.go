package main

import (
	"regexp"
	"strings"
)

// Capture the PPU's real intermediate layer samples and final composition.
// These hooks observe execution; replay uses stored values and never reruns it.
func advanceBody(name, body string) string {
	switch name {
	case "gbamachine_biosSWI":
		for _, memory := range []string{"pal", "vram", "oam"} {
			body = strings.ReplaceAll(body, "clear(m->"+memory+");", "clear(m->"+memory+");if(rrcapture::trace.active)for(uint32_t i=0;i<m->"+memory+".n;i++)rrGBAMemWrite(m,m->"+memory+",i);")
		}
	case "gbamachine_ppu_renderLine":
		priority := "get(m->io,cast<uint32_t>((cast<uint32_t>(8ULL) + cast<uint32_t>((cast<uint32_t>(2ULL) * cast<uint32_t>(n))))))"
		body = strings.ReplaceAll(body, priority, "rrPriority[n]")
		body = "{std::array<uint16_t,4>rrPriority{};for(unsigned i=0;i<4;i++)rrPriority[i]=get(m->io,8u+2u*i);\n" + body + "}\n"

		body = "{if(rrcapture::trace.active)rrgba::beginLine(m,y);\n" + body + "}\n"
		body = strings.ReplaceAll(body, "out[x] = gbamachine_rgb15(c);", "out[x] = gbamachine_rgb15(c);\nif(rrcapture::trace.active)rrgba::compose(m,x,y,top,second,topC,secondC,c,ctl);")
		body = strings.ReplaceAll(body, "out[i] = cast<uint32_t>(4294967295ULL);", "out[i] = cast<uint32_t>(4294967295ULL);\nif(rrcapture::trace.active){rrgba::clean();rrcapture::trace.record(rrgba::frameBase+(y*240+i)*4,0xffffffff,4,m->Steps,m->cpu->R[15],rrcapture::trace.current);}")
	case "gbamachine_ppu_textLine":
		body = "{if(rrcapture::trace.active)rrgba::event(m,\"PPU text background\",n,y);\n" + body + "}\n"
		body = regexp.MustCompile(`bank = std::get<1>\(tmp\d+\);`).ReplaceAllString(body, "$0\nif(rrcapture::trace.active)rrgba::source(m,n,x,a,idx,bank,tx,ty);")
		body = strings.ReplaceAll(body, "bank = cast<int64_t>(shr<uint16_t>(entry,cast<int64_t>(12ULL)));", "bank = cast<int64_t>(shr<uint16_t>(entry,cast<int64_t>(12ULL)));\nif(rrcapture::trace.active)rrgba::source(m,n,x,a,idx,bank,tx,ty);")
		body = strings.ReplaceAll(body, "out->on[x] = true;", "out->on[x] = true;\nif(rrcapture::trace.active)rrgba::dot(m,n,x,out->c[x]);")
	case "gbamachine_ppu_affineLine":
		body = "{if(rrcapture::trace.active)rrgba::event(m,\"PPU affine background\",n,rrgba::currentY);\n" + body + "}\n"
		body = strings.ReplaceAll(body, "out->on[x] = true;", "out->on[x] = true;\nif(rrcapture::trace.active){rrgba::source(m,n,x,a,idx,0,tx,ty);rrgba::dot(m,n,x,out->c[x]);}")
	case "gbamachine_ppu_bitmapLine":
		body = "{if(rrcapture::trace.active)rrgba::event(m,\"PPU bitmap background\",2,rrgba::currentY);\n" + body + "}\n"
		body = strings.ReplaceAll(body, "out->on[x] = true;", "out->on[x] = true;\nif(rrcapture::trace.active){auto a=page+(ty*w+tx)*(mode==4?1:2);rrgba::source(m,2,x,a,mode==4?m->vram[a]:0,mode==4?0:-1,tx,ty);rrgba::dot(m,2,x,out->c[x]);}")
	case "gbamachine_ppu_objLine":
		body = strings.ReplaceAll(body, "gbamachine_ppu_objPixel(", "if(rrcapture::trace.active){rrgba::currentObject=i;rrgba::objectSource=a;rrgba::objectU=sx;rrgba::objectV=sy;}\ngbamachine_ppu_objPixel(")
	case "gbamachine_ppu_objPixel":
		body = strings.ReplaceAll(body, "c[x] = gbamachine_Machine_pal16(m,palBank,idx);", "c[x] = gbamachine_Machine_pal16(m,palBank,idx);\nif(rrcapture::trace.active){rrgba::event(m,\"PPU object texel\",4,rrgba::currentY);rrgba::source(m,4,x,rrgba::objectSource,idx,palBank,rrgba::objectU,rrgba::objectV);rrgba::dot(m,4,x,c[x]);}")
	}
	return body
}
