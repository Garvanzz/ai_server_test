// invoke_gen 根据指定输入目录下各 module 的 Register 调用，生成输出目录的强类型 client 代码。
// 用法: 在项目根目录执行 go run ./main_server/tools/invoke_gen -input main_server/logic -output main_server/invoke
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

var inputDir = "../logic"
var outputDir = "../invoke"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "invoke_gen: %v\n", err)
		os.Exit(1)
	}
}

// exeDir 返回当前可执行文件所在目录（便于 exe 放到 run 等目录时，相对路径仍以 exe 所在目录为基准）。
func exeDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exe), nil
}

// findGoModRoot 从 dir 开始向上查找包含 go.mod 的目录（项目根）。
func findGoModRoot(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(abs, "go.mod")); err == nil {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", fmt.Errorf("no go.mod found from %s upward", dir)
		}
		abs = parent
	}
}

func run() error {
	exe, err := exeDir()
	if err != nil {
		return fmt.Errorf("exe dir: %w", err)
	}
	projectRoot, err := findGoModRoot(exe)
	if err != nil {
		return err
	}
	// 相对路径以 exe 所在目录为基准，这样 exe 放到 main_server/run 时 ../logic、../invoke 仍正确
	absInput := filepath.Join(exe, inputDir)
	absOutput := filepath.Join(exe, outputDir)
	absInput = filepath.Clean(absInput)
	absOutput = filepath.Clean(absOutput)

	// 相对于项目根的输入路径，用于 packages 和 import prefix
	relInput, err := filepath.Rel(projectRoot, absInput)
	if err != nil {
		return fmt.Errorf("input dir not under project root: %w", err)
	}
	relInputSlash := filepath.ToSlash(relInput)
	if relInputSlash == "." {
		relInputSlash = ""
	}
	loadPattern := "./" + relInputSlash + "/..."
	if loadPattern == "./..." {
		loadPattern = "./..."
	}

	modulePath, err := getModulePath(projectRoot)
	if err != nil {
		return fmt.Errorf("get module path: %w", err)
	}
	inputPrefix := modulePath + "/"
	if relInputSlash != "" {
		inputPrefix += relInputSlash + "/"
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir:  projectRoot,
	}
	pkgs, err := packages.Load(cfg, loadPattern)
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

	outDir := absOutput
	pkgName := filepath.Base(outDir)
	if pkgName == "." || pkgName == "" {
		pkgName = "invoke"
	}
	generated := 0

	for _, pkg := range pkgs {
		if pkg.Name == "" || len(pkg.GoFiles) == 0 {
			continue
		}
		pkgPath := pkg.PkgPath
		if !strings.HasPrefix(pkgPath, inputPrefix) {
			continue
		}
		rel := strings.TrimPrefix(pkgPath, inputPrefix)
		if strings.Contains(rel, "/") {
			continue // 跳过子包，如 logic/activity/impl
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

		code, err := generateClient(pkgName, moduleType, structName, constructorName, methods, importPaths)
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

// getModulePath 从 projectRoot 的 go.mod 读取 module 路径。
func getModulePath(projectRoot string) (string, error) {
	data, err := os.ReadFile(filepath.Join(projectRoot, "go.mod"))
	if err != nil {
		return "", err
	}
	const prefix = "module "
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			name := strings.TrimSpace(strings.TrimPrefix(line, prefix))
			if idx := strings.IndexAny(name, " \t"); idx >= 0 {
				name = name[:idx]
			}
			return name, nil
		}
	}
	return "", fmt.Errorf("go.mod missing module directive")
}

func toClientName(pkgName string) string {
	if pkgName == "" {
		return ""
	}
	return strings.ToUpper(pkgName[:1]) + pkgName[1:]
}

// methodSpec 表示一个要生成的 client 方法
type methodSpec struct {
	RegisterName string  // 传给 Invoke 的 fn 名，如 "GetActivityStatus"
	ClientName   string  // 生成的 Go 方法名（PascalCase），如 "GetActivityStatus"
	Params       []param // 参数（不含 receiver）
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

func generateClient(outPkgName, moduleType, structName, constructorName string, methods []methodSpec, importPaths map[string]string) ([]byte, error) {
	imports := make(map[string]string)
	imports["xfx/core/define"] = "define"
	for path, name := range importPaths {
		imports[path] = name
	}
	// 若方法签名中出现 proto.XXX（如 proto.Message）而 collectPackages 未收集到（接口类型多为 *types.Interface），则补上 proto 包
	for _, m := range methods {
		for _, p := range m.Params {
			if strings.HasPrefix(p.Type, "proto.") {
				imports["github.com/golang/protobuf/proto"] = "proto"
				break
			}
		}
		for _, r := range m.Returns {
			if strings.HasPrefix(r.Type, "proto.") {
				imports["github.com/golang/protobuf/proto"] = "proto"
				break
			}
		}
	}
	// 生成代码中只使用 github.com/golang/protobuf/proto，不使用 gogo 版本
	delete(imports, "github.com/gogo/protobuf/proto")
	// 生成代码中会使用 log 打印 Invoke/As 错误
	imports["xfx/pkg/log"] = "log"

	var buf bytes.Buffer
	buf.WriteString("// Code generated by main_server/tools/invoke_gen. DO NOT EDIT.\n\n")
	buf.WriteString("package " + outPkgName + "\n\n")
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
	buf.WriteString(fmt.Sprintf("\t\tlog.Error(\"invoke failed: type=%%s method=%%s err=%%v\", c.Type, %q, err)\n", m.RegisterName))
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
		buf.WriteString(fmt.Sprintf("\tif err2 != nil {\n\t\tlog.Error(\"invoke result type assert failed: type=%%s method=%%s err=%%v\", c.Type, %q, err2)\n\t\treturn err2\n\t}\n", m.RegisterName))
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
		buf.WriteString(fmt.Sprintf("\tif err2 != nil {\n\t\tlog.Error(\"invoke result type assert failed: type=%%s method=%%s err=%%v\", c.Type, %q, err2)\n\t\treturn %s, err2\n\t}\n", m.RegisterName, zeroValue(mainRet)))
		buf.WriteString("\treturn v, nil\n}\n\n")
	} else {
		buf.WriteString(fmt.Sprintf("\tv, err2 := As[%s](result)\n", mainRet))
		buf.WriteString(fmt.Sprintf("\tif err2 != nil {\n\t\tlog.Error(\"invoke result type assert failed: type=%%s method=%%s err=%%v\", c.Type, %q, err2)\n\t\treturn %s\n\t}\n", m.RegisterName, zeroValue(mainRet)))
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
		// 具名类型：指针/接口别名用 nil，枚举用 0，struct 用 T{}
		if idx := strings.LastIndex(typ, "."); idx >= 0 {
			typeName := typ[idx+1:]
			// 指针类型别名（如 agent.PID -> *actor.PID）零值为 nil
			if typeName == "PID" || typeName == "Context" {
				return "nil"
			}
			// 枚举/整型风格：含 CODE 或后缀 Id/ID（且非 PID）用 0
			if strings.Contains(typeName, "CODE") ||
				strings.HasSuffix(typeName, "Id") || (strings.HasSuffix(typeName, "ID") && typeName != "PID") {
				return "0"
			}
			// 全大写且非 PID：多为枚举，用 0
			if typeName != "PID" && typeName == strings.ToUpper(typeName) {
				return "0"
			}
			return typ + "{}"
		}
		return "nil"
	}
}
