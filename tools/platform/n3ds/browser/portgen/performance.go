package main

import "strings"

// Keep browser-only diagnostic boundaries reproducible after regeneration.
func instrumentPerformance(name, body string) string {
	if name != "n3ds_GPU_draw" {
		return body
	}
	replace := func(before, after string) {
		if strings.Count(body, before) != 1 {
			panic("missing or ambiguous 3DS performance hook: " + before)
		}
		body = strings.Replace(body, before, after, 1)
	}
	replace("uint32_t vi =", "rrperf::Vertex vertexProfile;\nuint32_t vi =")
	replace("if ((!n3ds_GPU_shaderRun(g,vin,vout,entry))) {", "vertexProfile.shade(vi);\nif ((!n3ds_GPU_shaderRun(g,vin,vout,entry))) {")
	replace("n3ds_GPU_mapOutputs(g,vout,(&outs[i]));", "vertexProfile.finish();\nn3ds_GPU_mapOutputs(g,vout,(&outs[i]));")
	replace("time_Time tr =", "rrperf::Clip clipProfile;\ntime_Time tr =")
	replace("n3ds_GPU_fill(g,(&fb),(&ls),(&tev),tris);", "clipProfile.finish();\nn3ds_GPU_fill(g,(&fb),(&ls),(&tev),tris);")
	return "rrperf::Draw drawProfile(g,indexed);\n" + body
}
