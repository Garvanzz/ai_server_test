// invoke_gen 根据 main_server/logic 下各 module 的 Register 调用，生成 main_server/invoke 的强类型 client 代码。
// 用法: 在项目根目录执行 go run ./main_server/tools/invoke_gen
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "invoke_gen: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// 加载 logic 下所有包（含子包如 activity、common、login）
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir:  ".",
	}
	pkgs, err := packages.Load(cfg, "./main_server/logic/...")
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, e := range pkg.Errors {
				fmt.Fprintf(os.Stderr, "pkg %s: %v\n", pkg.PkgPath, e)
			}
		}
	}

	outDir := filepath.Join("main_server", "invoke")
	generated := 0

	for _, pkg := range pkgs {
		if pkg.Name == "" || len(pkg.GoFiles) == 0 {
			continue
		}
		// 只处理直接子包：activity, common, login（即包含 Register 且含 GetType 的包）
		pkgPath := pkg.PkgPath
		if !strings.HasPrefix(pkgPath, "xfx/main_server/logic/") {
			continue
		}
		rel := strings.TrimPrefix(pkgPath, "xfx/main_server/logic/")
		if strings.Contains(rel, "/") {
			continue // 跳过 logic/activity/impl 等子包
		}

		moduleType, methods, importPaths, err := extractRegisterMethods(pkg)
		if err != nil {
			return fmt.Errorf("extract %s: %w", pkgPath, err)
		}
		if moduleType == "" || len(methods) == 0 {
			continue
		}

		clientName := toClientName(rel)
		structName := clientName + "ModClient"
		constructorName := clientName + "Client"
		outPath := filepath.Join(outDir, rel+".go")

		code, err := generateClient(moduleType, structName, constructorName, methods, importPaths)
		if err != nil {
			return fmt.Errorf("generate %s: %w", rel, err)
		}

		if err := os.WriteFile(outPath, code, 0644); err != nil {
			return err
		}
		fmt.Printf("wrote %s\n", outPath)
		generated++
	}

	if generated == 0 {
		return fmt.Errorf("no logic packages with Register() found")
	}
	return nil
}

func toClientName(pkgName string) string {
	if pkgName == "" {
		return ""
	}
	return strings.ToUpper(pkgName[:1]) + pkgName[1:]
}

// methodSpec 表示一个要生成的 client 方法
type methodSpec struct {
	RegisterName string   // 传给 Invoke 的 fn 名，如 "GetActivityStatus"
	ClientName   string   // 生成的 Go 方法名（PascalCase），如 "GetActivityStatus"
	Params       []param  // 参数（不含 receiver）
	Returns      []returnSpec
	ReturnsError bool
}

type param struct {
	Name string
	Type string
}

type returnSpec struct {
	Type string
}

func extractRegisterMethods(pkg *packages.Package) (moduleType string, methods []methodSpec, importPaths map[string]string, err error) {
	importPaths = make(map[string]string)
	// 1) 找 GetType() 方法的返回值，得到 define.ModuleXxx
	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Name.Name != "GetType" || len(fn.Body.List) == 0 {
				return true
			}
			for _, stmt := range fn.Body.List {
				if ret, ok := stmt.(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
					if sel, ok := ret.Results[0].(*ast.SelectorExpr); ok {
						if x, ok := sel.X.(*ast.Ident); ok && x.Name == "define" && sel.Sel != nil {
							moduleType = "define." + sel.Sel.Name
							return false
						}
					}
				}
			}
			return true
		})
		if moduleType != "" {
			break
		}
	}

	// 2) 找所有 Register("fn", methodValue) 调用
	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) != 2 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel == nil || sel.Sel.Name != "Register" {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			fnName := strings.Trim(lit.Value, `"`)
			methodExpr := call.Args[1]
			tv, ok := pkg.TypesInfo.Types[methodExpr]
			if !ok || tv.Type == nil {
				return true
			}
			sig, ok := tv.Type.Underlying().(*types.Signature)
			if !ok {
				return true
			}

			spec := methodSpec{
				RegisterName: fnName,
				ClientName:   toPascalCase(fnName),
			}
			params := sig.Params()
			for i := 0; i < params.Len(); i++ {
				v := params.At(i)
				spec.Params = append(spec.Params, param{
					Name: v.Name(),
					Type: typeString(v.Type(), pkg),
				})
				collectPackages(v.Type(), importPaths)
			}
			results := sig.Results()
			for i := 0; i < results.Len(); i++ {
				r := results.At(i)
				spec.Returns = append(spec.Returns, returnSpec{Type: typeString(r.Type(), pkg)})
				collectPackages(r.Type(), importPaths)
				if i == results.Len()-1 && isErrorType(r.Type()) {
					spec.ReturnsError = true
				}
			}
			methods = append(methods, spec)
	return true
	})
	}

	return moduleType, methods, importPaths, nil
}

// collectPackages 从类型中收集被引用的包路径，用于生成 import
func collectPackages(t types.Type, out map[string]string) {
	switch x := t.(type) {
	case *types.Named:
		if p := x.Obj().Pkg(); p != nil && p.Path() != "" {
			out[p.Path()] = p.Name()
		}
	case *types.Pointer:
		collectPackages(x.Elem(), out)
	case *types.Slice:
		collectPackages(x.Elem(), out)
	case *types.Map:
		collectPackages(x.Key(), out)
		collectPackages(x.Elem(), out)
	case *types.Signature:
		for i := 0; i < x.Params().Len(); i++ {
			collectPackages(x.Params().At(i).Type(), out)
		}
		for i := 0; i < x.Results().Len(); i++ {
			collectPackages(x.Results().At(i).Type(), out)
		}
	case *types.Interface, *types.Basic, *types.Struct:
		// 无额外包
	default:
		// 其他类型可再扩展
	}
}

func isErrorType(t types.Type) bool {
	if n, ok := t.(*types.Named); ok {
		return n.Obj().Name() == "error"
	}
	return false
}

func typeString(t types.Type, pkg *packages.Package) string {
	return types.TypeString(t, func(p *types.Package) string {
		if p == nil || p.Path() == "" {
			return ""
		}
		// 使用包名而非路径，便于生成可读代码
		return p.Name()
	})
}

func toPascalCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func generateClient(moduleType, structName, constructorName string, methods []methodSpec, importPaths map[string]string) ([]byte, error) {
	imports := make(map[string]string)
	imports["xfx/core/define"] = "define"
	for path, name := range importPaths {
		imports[path] = name
	}

	var buf bytes.Buffer
	buf.WriteString("// Code generated by main_server/tools/invoke_gen. DO NOT EDIT.\n\n")
	buf.WriteString("package invoke\n\n")
	buf.WriteString("import (\n")
	sortedPaths := make([]string, 0, len(imports))
	for path := range imports {
		sortedPaths = append(sortedPaths, path)
	}
	sort.Strings(sortedPaths)
	for _, path := range sortedPaths {
		name := imports[path]
		if name != "" && name != filepath.Base(path) {
			buf.WriteString(fmt.Sprintf("\t%s %q\n", name, path))
		} else {
			buf.WriteString(fmt.Sprintf("\t%q\n", path))
		}
	}
	buf.WriteString(")\n\n")

	buf.WriteString(fmt.Sprintf("type %s struct {\n\tinvoke Invoker\n\tType   string\n}\n\n", structName))
	buf.WriteString(fmt.Sprintf("func %s(invoker Invoker) %s {\n\treturn %s{\n\t\tinvoke: invoker,\n\t\tType:   %s,\n\t}\n}\n\n", constructorName, structName, structName, moduleType))

	for _, m := range methods {
		if err := writeMethod(&buf, structName, m); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func writeMethod(buf *bytes.Buffer, structName string, m methodSpec) error {
	// 方法签名
	paramParts := make([]string, len(m.Params))
	for i, p := range m.Params {
		if p.Name == "" || p.Name == "_" {
			paramParts[i] = fmt.Sprintf("arg%d %s", i, m.Params[i].Type)
		} else {
			paramParts[i] = fmt.Sprintf("%s %s", p.Name, p.Type)
		}
	}
	var sig string
	if len(m.Returns) == 0 {
		sig = fmt.Sprintf("func (c %s) %s(%s)",
			structName, m.ClientName, strings.Join(paramParts, ", "))
	} else {
		retParts := make([]string, len(m.Returns))
		for i, r := range m.Returns {
			retParts[i] = r.Type
		}
		sig = fmt.Sprintf("func (c %s) %s(%s) (%s)",
			structName, m.ClientName, strings.Join(paramParts, ", "), strings.Join(retParts, ", "))
	}
	buf.WriteString(sig + " {\n")

	// 参数列表用于 Invoke
	argNames := make([]string, len(m.Params))
	for i, p := range m.Params {
		if p.Name != "" && p.Name != "_" {
			argNames[i] = p.Name
		} else {
			argNames[i] = fmt.Sprintf("arg%d", i)
		}
	}
	invokeArgs := strings.Join(argNames, ", ")
	if invokeArgs != "" {
		invokeArgs = ", " + invokeArgs
	}
	buf.WriteString(fmt.Sprintf("\tresult, err := c.invoke.Invoke(c.Type, %q%s)\n", m.RegisterName, invokeArgs))
	buf.WriteString("\tif err != nil {\n")
	// err != nil 时：返回零值，其中 error 位直接返回 err
	if len(m.Returns) == 0 {
		buf.WriteString("\t\treturn\n")
	} else if len(m.Returns) == 1 && m.Returns[0].Type == "error" {
		buf.WriteString("\t\treturn err\n")
	} else if m.ReturnsError && len(m.Returns) >= 2 {
		zeroRet := make([]string, len(m.Returns))
		for i := range m.Returns {
			if m.Returns[i].Type == "error" {
				zeroRet[i] = "err"
			} else {
				zeroRet[i] = zeroValue(m.Returns[i].Type)
			}
		}
		buf.WriteString(fmt.Sprintf("\t\treturn %s\n", strings.Join(zeroRet, ", ")))
	} else if len(m.Returns) == 1 {
		buf.WriteString(fmt.Sprintf("\t\treturn %s\n", zeroValue(m.Returns[0].Type)))
	}
	buf.WriteString("\t}\n\n")

	if len(m.Returns) == 0 {
		buf.WriteString("\t_ = result\n\treturn\n}\n\n")
		return nil
	}

	// 仅返回 error 的情况
	if len(m.Returns) == 1 && m.Returns[0].Type == "error" {
		buf.WriteString("\tif result == nil {\n\t\treturn nil\n\t}\n")
		buf.WriteString("\tv, err2 := As[error](result)\n")
		buf.WriteString("\tif err2 != nil {\n\t\treturn err2\n\t}\n")
		buf.WriteString("\treturn v\n}\n\n")
		return nil
	}

	buf.WriteString("\tif result == nil {\n")
	if m.ReturnsError && len(m.Returns) >= 2 {
		zeroRet := make([]string, len(m.Returns))
		for i := range m.Returns {
			zeroRet[i] = zeroValue(m.Returns[i].Type)
		}
		buf.WriteString(fmt.Sprintf("\t\treturn %s\n", strings.Join(zeroRet, ", ")))
	} else {
		buf.WriteString(fmt.Sprintf("\t\treturn %s\n", zeroValue(m.Returns[0].Type)))
	}
	buf.WriteString("\t}\n\n")

	mainRet := m.Returns[0].Type
	if m.ReturnsError {
		buf.WriteString(fmt.Sprintf("\tv, err2 := As[%s](result)\n", mainRet))
		buf.WriteString("\tif err2 != nil {\n")
		buf.WriteString(fmt.Sprintf("\t\treturn %s, err2\n", zeroValue(mainRet)))
		buf.WriteString("\t}\n")
		buf.WriteString("\treturn v, nil\n}\n\n")
	} else {
		buf.WriteString(fmt.Sprintf("\tv, _ := As[%s](result)\n", mainRet))
		buf.WriteString("\treturn v\n}\n\n")
	}
	return nil
}

func zeroValue(typ string) string {
	switch {
	case strings.HasPrefix(typ, "*") || strings.HasPrefix(typ, "[]") || strings.Contains(typ, "map["):
		return "nil"
	case typ == "bool":
		return "false"
	case typ == "error":
		return "nil"
	case strings.Contains(typ, "int"):
		return "0"
	default:
		return "nil"
	}
}
