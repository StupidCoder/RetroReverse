package main

import "strings"

func x86Body(name, body string) string {
	scopes := map[string]string{"xbox_Machine_Run": "0,\"x86 CPU and devices\"", "xbox_Machine_runPusher": "1,\"NV2A commands\"", "xbox_pgraph_runDraw": "2,\"Vertex processing\"", "xbox_pgraph_assemble": "3,\"Software rasterizer\"", "xbox_pgraph_clearSurface": "4,\"Surface clears\"", "xbox_pgraph_texDecode": "5,\"Texture decoding\"", "dos_Machine_handleInt": "1,\"DOS and BIOS services\"", "dos_PM_handleInt": "1,\"DOS and DPMI services\""}
	if name == "xbox_pgraph_texDecode" {
		body = strings.ReplaceAll(body, "[&]", "[=]")
		body = strings.ReplaceAll(body, "arenaNew(xbox_texImage{", "rrTextureNew(xbox_texImage{")
	}
	if name == "xbox_pgraph_cacheTex" {
		body = strings.ReplaceAll(body, "arenaNew(xbox_texEntry{", "rrTextureEntryNew(xbox_texEntry{")
	}
	if scope := scopes[name]; scope != "" {
		body = "{rrprof::Scope timing(" + scope + ");\n" + body + "}\n"
	}
	if name == "dos_PM_enforceBaseAddress" {
		body = "{if(p->rrQuakeBase){" + body + "}}"
	}
	if name == "xbox_pgraph_runDraw" || name == "xbox_pgraph_clearSurface" || name == "xbox_pgraph_blitExecute" {
		body = "{rrconsole::EventScope restore(rrcapture::trace.current); if(rrcapture::trace.active)rrXboxDraw(g,\"" + name + "\");\n" + body + "}\n"
	}
	// These source readers never retain the slice. Avoid shared_ptr churn per scalar load.
	if strings.HasPrefix(name, "xbox_") {
		for _, fn := range []string{"xbox_le32", "xbox_le16", "be_Uint32", "le_Uint32"} {
			body = strings.ReplaceAll(body, fn+"(sub(", fn+"(rrBorrow(")
		}
	}
	return body
}
