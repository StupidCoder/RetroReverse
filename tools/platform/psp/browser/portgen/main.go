// Port generator for the bounded PSP execution sources. Unsupported Go constructs
// fail explicitly; this is not a general Go compiler. Output is ordinary C++20.
package main

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var fs = token.NewFileSet()
var pkgs = map[string]*unit{}
var std = importer.Default()

type unit struct {
	p        *types.Package
	info     *types.Info
	files    []*ast.File
	selected []*ast.File
}
type imp struct{}

func (imp) Import(path string) (*types.Package, error) {
	if !strings.HasPrefix(path, "retroreverse.com/") {
		return std.Import(path)
	}
	if p := pkgs[path]; p != nil {
		return p.p, nil
	}
	dir := strings.TrimPrefix(path, "retroreverse.com/")
	u := &unit{info: &types.Info{Types: map[ast.Expr]types.TypeAndValue{}, Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}, Selections: map[*ast.SelectorExpr]*types.Selection{}}}
	ents, _ := os.ReadDir(dir)
	for _, e := range ents {
		n := e.Name()
		if !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") || strings.HasPrefix(n, "observe_") && n != "observe_default.go" {
			continue
		}
		f, err := parser.ParseFile(fs, filepath.Join(dir, n), nil, 0)
		must(err)
		u.files = append(u.files, f)
		if selected(dir, n) {
			u.selected = append(u.selected, f)
		}
	}
	c := types.Config{Importer: imp{}}
	p, err := c.Check(path, fs, u.files, u.info)
	must(err)
	u.p = p
	pkgs[path] = u
	return p, nil
}
func selected(dir, n string) bool {
	switch filepath.Base(dir) {
	case "allegrex":
		return n != "disasm.go"
	case "psp":
		return n != "state.go" && n != "cso.go" && n != "lru.go" && n != "sfo.go"
	}
	return false
}

func strLit(s string) string {
	q := "\""
	for _, b := range []byte(s) {
		if b == '"' || b == '\\' {
			q += "\\" + string(b)
		} else if b >= 32 && b < 127 {
			q += string(b)
		} else {
			q += fmt.Sprintf("\\%03o", b)
		}
	}
	return "std::string(" + q + "\"," + fmt.Sprint(len(s)) + ")"
}
func must(e error) {
	if e != nil {
		panic(e)
	}
}
func id(s string) string {
	switch s {
	case "sub", "new", "delete", "template", "typename", "operator", "register", "signed", "unsigned", "not", "and", "or", "xor", "default", "class", "union", "short", "long", "int", "float", "double", "auto", "switch", "case", "this", "char", "bool", "private", "public", "virtual":
		return s + "_"
	}
	return s
}
func q(o types.Object) string {
	if o == nil {
		panic("nil object")
	}
	if o.Pkg() != nil && o.Parent() == o.Pkg().Scope() {
		return o.Pkg().Name() + "_" + id(o.Name())
	}
	return id(o.Name())
}

var anon = map[string]string{}
var anonTypes = map[string]*types.Struct{}

func typ(t types.Type) string {
	switch t := t.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.Invalid:
			return "void"
		case types.Bool, types.UntypedBool:
			return "bool"
		case types.Int, types.UntypedInt:
			return "int64_t"
		case types.Uint:
			return "uint64_t"
		case types.Int8:
			return "int8_t"
		case types.Uint8:
			return "uint8_t"
		case types.Int16:
			return "int16_t"
		case types.Uint16:
			return "uint16_t"
		case types.Int32, types.UntypedRune:
			return "int32_t"
		case types.Uint32:
			return "uint32_t"
		case types.Int64:
			return "int64_t"
		case types.Uint64:
			return "uint64_t"
		case types.Float32:
			return "float"
		case types.Float64, types.UntypedFloat:
			return "double"
		case types.String, types.UntypedString:
			return "std::string"
		case types.UntypedNil:
			return "std::nullptr_t"
		}
	case *types.Alias:
		return typ(types.Unalias(t))
	case *types.Pointer:
		return typ(t.Elem()) + "*"
	case *types.Array:
		return fmt.Sprintf("std::array<%s,%d>", typ(t.Elem()), t.Len())
	case *types.Slice:
		return "Slice<" + typ(t.Elem()) + ">"
	case *types.Map:
		return "Map<" + typ(t.Key()) + "," + typ(t.Elem()) + ">"
	case *types.Named:
		if _, ok := t.Underlying().(*types.Interface); ok {
			if t.Obj().Name() == "error" {
				return "Error"
			}
			if t.Obj().Pkg() != nil && t.Obj().Pkg().Path() == "hash" {
				return "SHA1*"
			}
			if t.Obj().Name() == "blockSource" {
				return "BlockSource*"
			}
			return "psp_Machine*"
		}
		return q(t.Obj())
	case *types.Interface:
		return "std::any"
	case *types.Signature:
		return "std::function<" + ret(t.Results()) + "(" + argtypes(t.Params()) + ")>"
	case *types.Tuple:
		return ret(t)
	case *types.Struct:
		k := t.String()
		if anon[k] == "" {
			name := fmt.Sprintf("Anon%d", len(anon))
			anon[k] = name
			anonTypes[name] = t
		}
		return anon[k]
	}
	panic(fmt.Sprintf("type %T %v", t, t))
}
func argtypes(t *types.Tuple) string {
	a := []string{}
	for i := 0; i < t.Len(); i++ {
		a = append(a, typ(t.At(i).Type()))
	}
	return strings.Join(a, ",")
}
func ret(t *types.Tuple) string {
	if t.Len() == 0 {
		return "void"
	}
	if t.Len() == 1 {
		return typ(t.At(0).Type())
	}
	return "std::tuple<" + argtypes(t) + ">"
}

type gen struct {
	function string
	u        *unit
	sig      *types.Signature
	serial   int
	breaks   []string
	writing  bool
}

func (g *gen) tmp() string { g.serial++; return fmt.Sprintf("tmp%d", g.serial) }
func (g *gen) obj(i *ast.Ident) types.Object {
	if o := g.u.info.Uses[i]; o != nil {
		return o
	}
	return g.u.info.Defs[i]
}
func (g *gen) t(e ast.Expr) types.Type { return g.u.info.TypeOf(e) }
func (g *gen) e(e ast.Expr) string {
	if e == nil {
		return ""
	}
	// Go evaluates constant expressions at arbitrary precision before conversion.
	// In particular uint16(0x3fe00/8) must not truncate the numerator first.
	if tv := g.u.info.Types[e]; tv.Value != nil && tv.Value.Kind() == constant.Int {
		return "cast<" + typ(tv.Type) + ">(" + tv.Value.ExactString() + "ULL)"
	}
	switch x := e.(type) {
	case *ast.ArrayType, *ast.MapType, *ast.StructType:
		return typ(g.t(e))
	case *ast.Ident:
		if x.Name == "nil" {
			return "{}"
		}
		if x.Name == "true" || x.Name == "false" {
			return x.Name
		}
		return q(g.obj(x))
	case *ast.BasicLit:
		if x.Kind == token.STRING {
			return strLit(constant.StringVal(g.u.info.Types[x].Value))
		}
		if x.Kind == token.INT {
			v := g.u.info.Types[x].Value
			if v != nil {
				return "cast<" + typ(g.t(x)) + ">(" + v.ExactString() + "ULL)"
			}
		}
		return x.Value
	case *ast.ParenExpr:
		return "(" + g.e(x.X) + ")"
	case *ast.SelectorExpr:
		if s := g.u.info.Selections[x]; s != nil {
			if s.Kind() == types.MethodVal {
				return g.method(x)
			}
			op := "."
			if _, ok := g.t(x.X).(*types.Pointer); ok {
				op = "->"
			}
			return g.e(x.X) + op + id(x.Sel.Name)
		}
		return g.external(x)
	case *ast.IndexExpr:
		w := g.writing
		g.writing = false
		a, b := g.e(x.X), g.e(x.Index)
		g.writing = w
		if _, ok := g.t(x.X).Underlying().(*types.Map); ok && !w {
			return "get(" + a + "," + b + ")"
		}
		if p, ok := g.t(x.X).(*types.Pointer); ok {
			if _, ok := p.Elem().Underlying().(*types.Array); ok {
				a = "(*" + a + ")"
			}
		}
		return a + "[" + b + "]"
	case *ast.SliceExpr:
		hi := g.e(x.High)
		if hi == "" {
			hi = "len(" + g.e(x.X) + ")"
		}
		lo := g.e(x.Low)
		if lo == "" {
			lo = "0"
		}
		if x.Max != nil {
			return "fullSub(" + g.e(x.X) + "," + lo + "," + hi + "," + g.e(x.Max) + ")"
		}
		return "sub(" + g.e(x.X) + "," + lo + "," + hi + ")"
	case *ast.UnaryExpr:
		op := x.Op.String()
		if x.Op == token.XOR {
			op = "~"
		}
		if x.Op == token.AND {
			if c, ok := x.X.(*ast.CompositeLit); ok {
				return "arenaNew(" + g.e(c) + ")"
			}
		}
		if x.Op == token.XOR || x.Op == token.SUB || x.Op == token.ADD {
			return "cast<" + typ(g.t(x)) + ">(" + op + g.e(x.X) + ")"
		}
		return "(" + op + g.e(x.X) + ")"
	case *ast.StarExpr:
		return "(*" + g.e(x.X) + ")"
	case *ast.BinaryExpr:
		a, b := g.e(x.X), g.e(x.Y)
		if x.Op == token.SHL || x.Op == token.SHR {
			f := "shl"
			if x.Op == token.SHR {
				f = "shr"
			}
			return f + "<" + typ(g.t(x)) + ">(" + a + "," + b + ")"
		}
		if x.Op == token.QUO || x.Op == token.REM {
			if z, ok := g.t(x).Underlying().(*types.Basic); ok && z.Info()&types.IsInteger != 0 {
				f := "divi"
				if x.Op == token.REM {
					f = "modi"
				}
				return f + "<" + typ(g.t(x)) + ">(" + a + "," + b + ")"
			}
		}
		if x.Op == token.AND_NOT {
			return "(" + a + " & ~(" + b + "))"
		}
		if i, ok := x.Y.(*ast.Ident); ok && i.Name == "nil" {
			if x.Op == token.EQL {
				return "(!" + a + ")"
			}
			return "bool(" + a + ")"
		}
		v := "(" + a + " " + x.Op.String() + " " + b + ")"
		if z, ok := g.t(x).Underlying().(*types.Basic); ok && z.Info()&types.IsInteger != 0 {
			return "cast<" + typ(g.t(x)) + ">(" + v + ")"
		}
		return v
	case *ast.CallExpr:
		return g.call(x)
	case *ast.CompositeLit:
		return g.lit(x)
	case *ast.FuncLit:
		s := g.t(x).(*types.Signature)
		old := g.sig
		g.sig = s
		body := g.block(x.Body)
		g.sig = old
		capture := "[&]"
		if g.function == "" {
			capture = "[]"
		}
		if g.function == "psp_handlerFor" {
			capture = "[=]"
		}
		return capture + "(" + params(s.Params()) + ")->" + ret(s.Results()) + body
	}
	panic(fmt.Sprintf("expr %T at %s", e, fs.Position(e.Pos())))
}
func (g *gen) external(x *ast.SelectorExpr) string {
	o := g.obj(x.Sel)
	if o.Pkg() != nil && !strings.HasPrefix(o.Pkg().Path(), "retroreverse.com/") {
		return "go_" + o.Pkg().Name() + "_" + o.Name()
	}
	return q(o)
}
func methodname(o *types.Func) string {
	s := o.Type().(*types.Signature)
	t := s.Recv().Type()
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	return typ(t) + "_" + id(o.Name())
}
func (g *gen) method(x *ast.SelectorExpr) string {
	s := g.u.info.Selections[x]
	o := s.Obj().(*types.Func)
	r := g.e(x.X)
	recv := o.Type().(*types.Signature).Recv().Type()
	if _, ok := recv.Underlying().(*types.Interface); ok {
		return "[=](auto...args){return psp_Machine_" + x.Sel.Name + "(" + r + ",args...);}"
	}
	_, rp := recv.(*types.Pointer)
	_, xp := g.t(x.X).(*types.Pointer)
	if rp && !xp {
		r = "&(" + r + ")"
	}
	if !rp && xp {
		r = "*(" + r + ")"
	}
	return "[=](auto...args){return " + methodname(o) + "(" + r + ",args...);}"
}
func (g *gen) call(x *ast.CallExpr) string {
	a := []string{}
	for _, e := range x.Args {
		a = append(a, g.e(e))
	}
	join := strings.Join(a, ",")
	if g.u.info.Types[x.Fun].IsType() {
		if len(x.Args) == 1 {
			if i, ok := x.Args[0].(*ast.Ident); ok && i.Name == "nil" {
				return typ(g.t(x)) + "{}"
			}
		}
		t := typ(g.t(x))
		return "cast<" + t + ">(" + join + ")"
	}
	if f, ok := x.Fun.(*ast.Ident); ok {
		switch f.Name {
		case "delete":
			return "removeKey(" + join + ")"
		case "len", "cap":
			return "len(" + join + ")"
		case "copy":
			return "gcopy(" + join + ")"
		case "append":
			return "append(" + join + ")"
		case "make":
			t := g.t(x)
			if s, ok := t.(*types.Slice); ok {
				return "Slice<" + typ(s.Elem()) + ">::make(" + strings.Join(a[1:], ",") + ")"
			}
			return typ(t) + "{}"
		case "panic":
			return "throw " + join
		case "min", "max":
			return "g" + f.Name + "(" + join + ")"
		}
	}
	if s, ok := x.Fun.(*ast.SelectorExpr); ok {
		if z, ok := s.X.(*ast.SelectorExpr); ok {
			if i, ok := z.X.(*ast.Ident); ok && i.Name == "binary" {
				prefix := "le_"
				if z.Sel.Name == "BigEndian" {
					prefix = "be_"
				}
				return prefix + s.Sel.Name + "(" + join + ")"
			}
		}
		if sel := g.u.info.Selections[s]; sel != nil && sel.Kind() == types.MethodVal {
			o := sel.Obj().(*types.Func)
			r := g.e(s.X)
			recv := o.Type().(*types.Signature).Recv().Type()
			name := methodname(o)
			if o.Pkg() != nil && o.Pkg().Path() == "encoding/binary" {
				prefix := "le_"
				if strings.Contains(types.TypeString(recv, nil), "bigEndian") {
					prefix = "be_"
				}
				return prefix + s.Sel.Name + "(" + join + ")"
			}
			if _, ok := recv.Underlying().(*types.Interface); ok {
				if typ(g.t(s.X)) == "SHA1*" {
					name = "hash_" + s.Sel.Name
				} else if s.Sel.Name == "ReadBlock" {
					name = "blockSource_ReadBlock"
				} else {
					name = "psp_Machine_" + s.Sel.Name
				}
			} else {
				_, rp := recv.(*types.Pointer)
				_, xp := g.t(s.X).(*types.Pointer)
				if rp && !xp {
					r = "&(" + r + ")"
				}
				if !rp && xp {
					r = "*(" + r + ")"
				}
			}
			if len(a) > 0 {
				r += "," + join
			}
			return name + "(" + r + ")"
		}
		if z, ok := s.X.(*ast.SelectorExpr); ok {
			if i, ok := z.X.(*ast.Ident); ok && i.Name == "binary" {
				prefix := "le_"
				if z.Sel.Name == "BigEndian" {
					prefix = "be_"
				}
				return prefix + s.Sel.Name + "(" + join + ")"
			}
		}
	}
	return g.e(x.Fun) + "(" + join + ")"
}
func (g *gen) lit(x *ast.CompositeLit) string {
	t := g.t(x)
	if p, ok := t.(*types.Pointer); ok {
		old := g.u.info.Types[x]
		v := old
		v.Type = p.Elem()
		g.u.info.Types[x] = v
		out := "arenaNew(" + g.lit(x) + ")"
		g.u.info.Types[x] = old
		return out
	}
	name := typ(t)
	u := t.Underlying()
	a := []string{}
	if s, ok := u.(*types.Struct); ok {
		vals := map[string]string{}
		keyed := len(x.Elts) > 0
		if keyed {
			_, keyed = x.Elts[0].(*ast.KeyValueExpr)
		}
		if keyed {
			for _, e := range x.Elts {
				kv := e.(*ast.KeyValueExpr)
				vals[kv.Key.(*ast.Ident).Name] = g.e(kv.Value)
			}
			for i := 0; i < s.NumFields(); i++ {
				v := vals[s.Field(i).Name()]
				if v == "" {
					v = "{}"
				}
				if b, ok := s.Field(i).Type().Underlying().(*types.Basic); ok && b.Kind() == types.Float32 && v != "{}" {
					v = "cast<float>(" + v + ")"
				}
				a = append(a, v)
			}
		} else {
			for _, e := range x.Elts {
				a = append(a, g.e(e))
			}
		}
		return name + "{" + strings.Join(a, ",") + "}"
	}
	if _, ok := u.(*types.Map); ok {
		for _, e := range x.Elts {
			kv := e.(*ast.KeyValueExpr)
			a = append(a, "{"+g.e(kv.Key)+","+g.e(kv.Value)+"}")
		}
		return name + "{" + strings.Join(a, ",") + "}"
	}
	keyed := false
	for _, e := range x.Elts {
		if _, ok := e.(*ast.KeyValueExpr); ok {
			keyed = true
		}
	}
	if keyed {
		out := "[](){" + name + " v{};"
		i := int64(0)
		for _, e := range x.Elts {
			if kv, ok := e.(*ast.KeyValueExpr); ok {
				i, _ = constant.Int64Val(g.u.info.Types[kv.Key].Value)
				e = kv.Value
			}
			out += fmt.Sprintf("v[%d]=%s;", i, g.e(e))
			i++
		}
		return out + "return v;}()"
	}
	for _, e := range x.Elts {
		a = append(a, g.e(e))
	}
	return name + "{" + strings.Join(a, ",") + "}"
}
func params(t *types.Tuple) string {
	a := []string{}
	for i := 0; i < t.Len(); i++ {
		v := t.At(i)
		n := id(v.Name())
		if n == "" || n == "_" {
			n = fmt.Sprintf("unused%d", i)
		}
		a = append(a, typ(v.Type())+" "+n)
	}
	return strings.Join(a, ",")
}
func (g *gen) block(b *ast.BlockStmt) string {
	s := "{\n"
	for _, v := range b.List {
		s += g.st(v)
	}
	return s + "}\n"
}
func (g *gen) lhs(e ast.Expr) string { g.writing = true; s := g.e(e); g.writing = false; return s }
func (g *gen) assign(x *ast.AssignStmt) string {
	if len(x.Lhs) == 1 {
		if i, ok := x.Lhs[0].(*ast.Ident); ok && i.Name == "_" {
			return "(void)(" + g.e(x.Rhs[0]) + ");\n"
		}
		l := g.lhs(x.Lhs[0])
		if x.Tok == token.DEFINE {
			if _, ok := x.Rhs[0].(*ast.FuncLit); ok {
				return "auto " + l + " = " + g.e(x.Rhs[0]) + ";\n"
			}
			return typ(g.t(x.Lhs[0])) + " " + l + " = " + g.e(x.Rhs[0]) + ";\n"
		}
		if x.Tok == token.AND_NOT_ASSIGN {
			return l + " &= ~(" + g.e(x.Rhs[0]) + ");\n"
		}
		if x.Tok == token.SHL_ASSIGN || x.Tok == token.SHR_ASSIGN {
			f := "shl"
			if x.Tok == token.SHR_ASSIGN {
				f = "shr"
			}
			return l + " = " + f + "<" + typ(g.t(x.Lhs[0])) + ">(" + l + "," + g.e(x.Rhs[0]) + ");\n"
		}
		return l + " " + x.Tok.String() + " " + g.e(x.Rhs[0]) + ";\n"
	}
	tmp := g.tmp()
	r := []string{}
	for _, e := range x.Rhs {
		r = append(r, g.e(e))
	}
	rhs := strings.Join(r, ",")
	if len(r) > 1 {
		rhs = "std::make_tuple(" + rhs + ")"
	} else if idx, ok := x.Rhs[0].(*ast.IndexExpr); ok {
		if _, ok := g.t(idx.X).Underlying().(*types.Map); ok {
			rhs = "lookup(" + g.e(idx.X) + "," + g.e(idx.Index) + ")"
		}
	}
	s := "auto " + tmp + " = " + rhs + ";\n"
	for i, e := range x.Lhs {
		if idn, ok := e.(*ast.Ident); ok && idn.Name == "_" {
			continue
		}
		prefix := ""
		if idn, ok := e.(*ast.Ident); ok && g.u.info.Defs[idn] != nil {
			prefix = typ(g.t(e)) + " "
		}
		s += fmt.Sprintf("%s%s = std::get<%d>(%s);\n", prefix, g.lhs(e), i, tmp)
	}
	return s
}
func (g *gen) st(s ast.Stmt) string {
	if s == nil {
		return ""
	}
	switch x := s.(type) {
	case *ast.BlockStmt:
		return g.block(x)
	case *ast.ExprStmt:
		return g.e(x.X) + ";\n"
	case *ast.AssignStmt:
		return g.assign(x)
	case *ast.ReturnStmt:
		a := []string{}
		for _, e := range x.Results {
			a = append(a, g.e(e))
		}
		if len(a) == 0 && g.sig.Results().Len() > 0 {
			for i := 0; i < g.sig.Results().Len(); i++ {
				a = append(a, id(g.sig.Results().At(i).Name()))
			}
		}
		if len(a) > 1 {
			return "return {" + strings.Join(a, ",") + "};\n"
		}
		return "return " + strings.Join(a, ",") + ";\n"
	case *ast.IncDecStmt:
		return g.lhs(x.X) + x.Tok.String() + ";\n"
	case *ast.IfStmt:
		r := ""
		if x.Init != nil {
			r = "{\n" + g.st(x.Init)
		}
		r += "if (" + g.e(x.Cond) + ") " + g.block(x.Body)
		if x.Else != nil {
			r += "else " + g.st(x.Else)
		}
		if x.Init != nil {
			r += "}\n"
		}
		return r
	case *ast.ForStmt:
		g.breaks = append(g.breaks, "")
		defer func() { g.breaks = g.breaks[:len(g.breaks)-1] }()
		init := strings.TrimSpace(g.st(x.Init))
		init = strings.TrimSuffix(init, ";")
		post := strings.TrimSpace(g.st(x.Post))
		post = strings.TrimSuffix(post, ";")
		if strings.Contains(post, "\n") {
			post = "[&](){" + post + ";}()"
		}
		return "{" + init + ";for (;" + g.e(x.Cond) + ";" + post + ")" + g.block(x.Body) + "}"
	case *ast.RangeStmt:
		g.breaks = append(g.breaks, "")
		defer func() { g.breaks = g.breaks[:len(g.breaks)-1] }()
		r := g.e(x.X)
		v := g.tmp()
		out := "{auto&& " + v + " = " + r + ";\n"
		if _, ok := g.t(x.X).Underlying().(*types.Map); ok {
			k, val := g.tmp(), g.tmp()
			out += "for(auto [" + k + "," + val + "]:" + v + "){\n"
			if x.Key != nil && g.e(x.Key) != "_" {
				out += "auto " + g.e(x.Key) + "=" + k + ";"
			}
			if x.Value != nil && g.e(x.Value) != "_" {
				out += "auto " + g.e(x.Value) + "=" + val + ";"
			}
		} else {
			k := g.tmp()
			out += "for(int64_t " + k + "=0;" + k + "<len(" + v + ");++" + k + "){\n"
			if x.Key != nil && g.e(x.Key) != "_" {
				out += "auto " + g.e(x.Key) + "=" + k + ";"
			}
			if x.Value != nil && g.e(x.Value) != "_" {
				out += "auto " + g.e(x.Value) + "=" + v + "[" + k + "];"
			}
		}
		for _, s := range x.Body.List {
			out += g.st(s)
		}
		return out + "}}\n"
	case *ast.SwitchStmt:
		if x.Tag != nil {
			if b, ok := g.t(x.Tag).Underlying().(*types.Basic); !ok || b.Info()&(types.IsString|types.IsFloat) == 0 {
				return g.tagSwitch(x)
			}
		}
		label := g.tmp()
		g.breaks = append(g.breaks, label)
		defer func() { g.breaks = g.breaks[:len(g.breaks)-1] }()
		out := "{\n" + g.st(x.Init)
		tag := ""
		if x.Tag != nil {
			tag = g.tmp()
			out += "auto " + tag + "=" + g.e(x.Tag) + ";\n"
		}
		first := true
		for _, c := range x.Body.List {
			c := c.(*ast.CaseClause)
			if len(c.List) == 0 {
				if !first {
					out += "else "
				}
			} else {
				if !first {
					out += "else "
				}
				a := []string{}
				for _, e := range c.List {
					v := g.e(e)
					if tag != "" {
						v = tag + "==(" + v + ")"
					}
					a = append(a, v)
				}
				out += "if (" + strings.Join(a, " || ") + ")"
			}
			out += "{\n"
			for _, s := range c.Body {
				if b, ok := s.(*ast.BranchStmt); ok && b.Tok == token.BREAK {
					continue
				}
				out += g.st(s)
			}
			out += "}\n"
			first = false
		}
		return out + "}\n" + label + ":;\n"
	case *ast.BranchStmt:
		if x.Tok == token.BREAK && len(g.breaks) > 0 && g.breaks[len(g.breaks)-1] != "" {
			return "goto " + g.breaks[len(g.breaks)-1] + ";\n"
		}
		return x.Tok.String() + ";\n"
	case *ast.DeclStmt:
		return g.decl(x.Decl.(*ast.GenDecl), false)
	case *ast.EmptyStmt:
		return ""
	case *ast.DeferStmt:
		call := g.e(x.Call)
		return "auto " + g.tmp() + "=defer([&](){" + call + ";});\n"
	}
	panic(fmt.Sprintf("stmt %T at %s", s, fs.Position(s.Pos())))
}
func (g *gen) tagSwitch(x *ast.SwitchStmt) string {
	g.breaks = append(g.breaks, "")
	defer func() { g.breaks = g.breaks[:len(g.breaks)-1] }()
	out := "{\n" + g.st(x.Init) + "switch(" + g.e(x.Tag) + "){\n"
	for _, c := range x.Body.List {
		c := c.(*ast.CaseClause)
		if len(c.List) == 0 {
			out += "default:"
		} else {
			for _, e := range c.List {
				out += "case " + g.e(e) + ":"
			}
		}
		out += "{\n"
		for _, s := range c.Body {
			out += g.st(s)
		}
		out += "break;}\n"
	}
	return out + "}}\n"
}
func (g *gen) decl(d *ast.GenDecl, global bool) string {
	out := ""
	if d.Tok == token.IMPORT || d.Tok == token.TYPE {
		return ""
	}
	for _, s := range d.Specs {
		v := s.(*ast.ValueSpec)
		if len(v.Values) == 1 && len(v.Names) > 1 {
			if _, ok := g.t(v.Values[0]).(*types.Tuple); ok {
				names := []string{}
				for _, n := range v.Names {
					if n.Name == "_" {
						names = append(names, g.tmp())
					} else {
						names = append(names, q(g.u.info.Defs[n]))
					}
				}
				out += "auto [" + strings.Join(names, ",") + "]=" + g.e(v.Values[0]) + ";\n"
				continue
			}
		}
		for i, n := range v.Names {
			if n.Name == "_" {
				continue
			}
			o := g.u.info.Defs[n]
			if c, ok := o.(*types.Const); ok {
				value := c.Val().ExactString()
				if c.Val().Kind() == constant.String {
					value = strLit(constant.StringVal(c.Val()))
				}
				if c.Val().Kind() == constant.Float {
					f, _ := constant.Float64Val(c.Val())
					value = fmt.Sprintf("%.17e", f)
				}
				if c.Val().Kind() == constant.Int {
					value += "ULL"
				}
				qual := "constexpr "
				if c.Val().Kind() == constant.String {
					qual = "const "
				}
				out += qual + typ(o.Type()) + " " + q(o) + "=" + value + ";\n"
			} else {
				val := "{}"
				if i < len(v.Values) {
					val = g.e(v.Values[i])
				}
				out += typ(o.Type()) + " " + q(o) + "=" + val + ";\n"
			}
		}
	}
	return out
}

var skips = map[string]bool{"psp_u16": true, "psp_Image_Close": true, "psp_OpenImage": true, "psp_Module_Describe": true, "psp_Image_LoadExecutable": true, "psp_Machine_Screenshot": true, "psp_Machine_ScreenshotAt": true, "psp_init": true, "psp_Machine_read32": true, "psp_Machine_write16": true, "psp_Machine_write32": true, "psp_Machine_note": true, "psp_fmtPrintf": true, "psp_Machine_DisasmAt": true, "psp_Machine_Run": true, "psp_Machine_WriteShot": true, "psp_Machine_RenderDepth": true, "psp_Machine_RenderTexture": true, "psp_cbcDecryptZero": true, "psp_OpenISO": true, "psp_Machine_callGuest": true, "allegrex_CPU_Halt": true, "allegrex_NewCPU": true, "allegrex_CPU_read16": true, "allegrex_CPU_read32": true, "allegrex_CPU_write16": true, "allegrex_CPU_write32": true}

func funcname(o *types.Func) string {
	if o.Type().(*types.Signature).Recv() != nil {
		return methodname(o)
	}
	return q(o)
}
func main() {
	_, err := (imp{}).Import("retroreverse.com/tools/platform/psp")
	must(err)
	out := "#include \"runtime.h\"\n"
	units := []*unit{pkgs["retroreverse.com/tools/cpu/allegrex"], pkgs["retroreverse.com/tools/platform/psp"]}
	named := []*types.Named{}
	for _, u := range units {
		for _, f := range u.selected {
			ast.Inspect(f, func(n ast.Node) bool {
				if ts, ok := n.(*ast.TypeSpec); ok {
					o := u.info.Defs[ts.Name]
					if o.Parent() != u.p.Scope() {
						if t, ok := o.Type().(*types.Named); ok {
							named = append(named, t)
						}
					}
				}
				if e, ok := n.(ast.Expr); ok {
					if t := u.info.TypeOf(e); t != nil {
						typ(t)
					}
				}
				return true
			})
			for _, d := range f.Decls {
				if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.TYPE {
					for _, sp := range gd.Specs {
						ts := sp.(*ast.TypeSpec)
						t := u.info.Defs[ts.Name].Type()
						if n, ok := t.(*types.Named); ok {
							named = append(named, n)
						}
					}
				}
			}
		}
	}
	for _, n := range named {
		if _, ok := n.Underlying().(*types.Struct); ok {
			out += "struct " + typ(n) + ";\n"
		}
	}
	anonNames := []string{}
	for name := range anonTypes {
		anonNames = append(anonNames, name)
	}
	sort.Strings(anonNames)
	for _, name := range anonNames {
		out += "struct " + name + ";\n"
	}

	done := map[string]bool{}
	var emit func(types.Type)
	emit = func(t types.Type) {
		switch t := t.(type) {
		case *types.Named:
			if t.Obj().Pkg() != nil && !strings.HasPrefix(t.Obj().Pkg().Path(), "retroreverse.com/") {
				return
			}
			n := typ(t)
			if done[n] {
				return
			}
			done[n] = true
			switch u := t.Underlying().(type) {
			case *types.Struct:
				for i := 0; i < u.NumFields(); i++ {
					emit(u.Field(i).Type())
				}
				out += "struct " + n + "{\n"
				for i := 0; i < u.NumFields(); i++ {
					v := u.Field(i)
					out += typ(v.Type()) + " " + id(v.Name()) + "{};\n"
				}
				out += "};\n"
			case *types.Interface:
			default:
				emit(u)
				out += "using " + n + "=" + typ(u) + ";\n"
			}
		case *types.Array:
			emit(t.Elem())
		case *types.Slice:
			emit(t.Elem())
		case *types.Struct:
			n := typ(t)
			if done[n] {
				return
			}
			done[n] = true
			for i := 0; i < t.NumFields(); i++ {
				emit(t.Field(i).Type())
			}
			out += "struct " + n + "{"
			for i := 0; i < t.NumFields(); i++ {
				v := t.Field(i)
				out += typ(v.Type()) + " " + id(v.Name()) + "{};"
			}
			out += "};\n"
		}
	}
	for _, n := range named {
		emit(n)
	}
	for _, name := range anonNames {
		emit(anonTypes[name])
	}
	for _, u := range units {
		for _, f := range u.selected {
			for _, d := range f.Decls {
				if gd, ok := d.(*ast.GenDecl); ok && (gd.Tok == token.VAR || gd.Tok == token.CONST) {
					for _, sp := range gd.Specs {
						for _, n := range sp.(*ast.ValueSpec).Names {
							emit(u.info.Defs[n].Type())
						}
					}
				}
			}
		}
	}

	protos, bodies, globals := "", "", ""
	for _, u := range units {
		g := &gen{u: u}
		for _, f := range u.selected {
			for _, d := range f.Decls {
				switch d := d.(type) {
				case *ast.GenDecl:
					g.function = ""
					globals += g.decl(d, true)
				case *ast.FuncDecl:
					o := u.info.Defs[d.Name].(*types.Func)
					name := funcname(o)
					if skips[name] {
						continue
					}
					sig := o.Type().(*types.Signature)
					g.sig = sig
					g.function = name
					args := params(sig.Params())
					if sig.Recv() != nil {
						r := sig.Recv()
						pre := typ(r.Type()) + " " + id(r.Name())
						if args != "" {
							pre += "," + args
						}
						args = pre
					}
					signature := ret(sig.Results()) + " " + name + "(" + args + ")"
					protos += signature + ";\n"
					body := g.block(d.Body)
					if name == "psp_Machine_sampleTexLvl" || name == "psp_Machine_putPixel" || name == "psp_Machine_rasterTri" || name == "psp_modTex" {
						signature = strings.Replace(signature, name, name+"_Reference", 1)
						protos += signature + ";\n"
					}
					if name == "psp_Machine_drawPrim" || name == "psp_Machine_drawPatch" {
						body = "{rrTextureScope textures(m,s); rrFragmentScope fragments(m,s);\n" + body + "}\n"
					}

					if name == "psp_Machine_Read" {
						body = "{if(!m->OnRead)if(auto*p=rrMemory(m,addr,1))return *p;\n" + body + "}\n"
					}
					if name == "psp_Machine_storePixel" {
						body = strings.ReplaceAll(body, "psp_Machine_Write(m,addr,cast<uint8_t>(p));\npsp_Machine_Write(m,cast<uint32_t>((addr + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint16_t>(p,cast<int64_t>(8ULL))));", "psp_Machine_write16(m,addr,p);")
					}

					if name == "psp_Machine_rasterList" {
						body = "{rrGETraceScope traceScope;\n" + body + "}\n"
					}
					scopes := map[string]string{"psp_Machine_rasterList": "1,\"GE command processor\"", "psp_Machine_drawPrim": "2,\"GE vertices and software rasterizer\"", "psp_Machine_drawPatch": "2,\"GE vertices and software rasterizer\"", "psp_Machine_handleSyscall": "3,\"Kernel HLE\""}
					if scope := scopes[name]; scope != "" {
						body = "{rrprof::Scope timing(" + scope + ");\n" + body + "}\n"
					}
					if name == "psp_handlerFor" {
						for i, axis := range []string{"rrAnalogX", "rrAnalogY"} {
							body = strings.ReplaceAll(body, fmt.Sprintf("(p + cast<uint32_t>(%dULL))),cast<uint8_t>(128ULL)", 8+i), fmt.Sprintf("(p + cast<uint32_t>(%dULL))),%s", 8+i, axis))
						}
					}

					namedret := ""
					for i := 0; i < sig.Results().Len(); i++ {
						v := sig.Results().At(i)
						if v.Name() != "" {
							namedret += typ(v.Type()) + " " + id(v.Name()) + "{};\n"
						}
					}
					bodies += "// " + fs.Position(d.Pos()).String() + "\n" + signature + "{\n" + namedret + body + "}\n"
				}
			}
		}
	}
	_ = sort.Strings
	out += "\n#include \"adapters-decl.h\"\n" + protos + globals + "\n#include \"adapters.h\"\n" + bodies + "\n#include \"fast.h\"\n"
	must(os.WriteFile("tools/platform/psp/browser/core/generated.cpp", []byte(out), 0644))
}
