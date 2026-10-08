package main

import "strings"

// Keep the original fetch and shader loop available for Reference/observation.
func accelerateVertices(name, body string) string {
	if name != "n3ds_GPU_draw" {
		return body
	}
	replace := func(before, after string) {
		if strings.Count(body, before) != 1 {
			panic("missing or ambiguous 3DS vertex hook: " + before)
		}
		body = strings.Replace(body, before, after, 1)
	}
	replace("auto shadeVertex =", "rrvertex::Draw vertexFast(g,indexed,physBase,fmtWord,fixedMask,bufs,comps,idxAddr,idx16,count,trace);\nauto shadeVertex =")
	replace("if (indexed) {\nif (idx16)", "if (indexed && vertexFast.eligible) { vi=vertexFast.index(i); }\nelse if (indexed) {\nif (idx16)")
	start := strings.Index(body, "{int64_t a = cast<int64_t>(0ULL);for (;(a < cast<int64_t>(16ULL));a++){")
	end := strings.Index(body, "n3ds_mapAttrsToInputs(vin,attrs,inPerm,maxIn);")
	if start < 0 || end <= start {
		panic("missing 3DS attribute fetch block")
	}
	original := body[start:end]
	body = body[:start] + "if(vertexFast.reuse(vi,i,outs)){vertexProfile.reuse(vi);return true;}\nif(!vertexFast.fetch(vi,attrs)){\n" + original + "}\n" + body[end:]
	replace("n3ds_GPU_mapOutputs(g,vout,(&outs[i]));", "n3ds_GPU_mapOutputs(g,vout,(&outs[i]));\nvertexFast.remember(vi,i);")
	return body
}
