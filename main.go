package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

func main() {
	var buildFlags = flag.String("build", "", "go build flags, space-separated")
	flag.Parse()

	mode := packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes |
		packages.NeedTypesInfo
	cfg := &packages.Config{Mode: mode, BuildFlags: strings.Split(*buildFlags, " ")}
	pkgs, err := packages.Load(cfg, flag.Args()...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(1)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	logToPos := make(map[string][]string)
	for _, pkg := range pkgs {
		getStringLiterals(pkg, logToPos)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		fullLine := line
		if !strings.HasPrefix(line, "202") {
			continue
		}
		bracketPos := strings.IndexByte(line, ']')
		if bracketPos == -1 {
			continue
		}
		line = line[bracketPos:]

		colonPos := strings.IndexByte(line, ':')
		if colonPos == -1 {
			continue
		}
		colonPos++
		for len(line) > colonPos && line[colonPos] == ' ' {
			colonPos++
		}
		line = line[colonPos:]
		if len(line) == 0 {
			continue
		}

		lastColonPos := strings.LastIndex(line, ": ")
		if lastColonPos != -1 {
			line = line[:lastColonPos]
		}

		pos := logToPos[`"`+line+`"`]
		switch len(pos) {
		case 0:
			fmt.Println(fullLine)
		case 1:
			fmt.Println(fullLine, pos[0])
		default:
			fmt.Println(fullLine, strings.Join(pos, " - "))
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func getStringLiterals(pkg *packages.Package, logToPos map[string][]string) {
	firstArgStart, firstArgEnd := token.NoPos, token.NoPos
	var funcName string
	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CallExpr:
				switch node := x.Fun.(type) {
				case *ast.Ident:
					// Simple function call, e.g. append().
					//name = node.Name
					return true
				case *ast.SelectorExpr:
					// Either a qualified identifier, e.g. fmt.Printf, or some
					// other expression with a method call applied to it.  We
					// don't care about the former.
					if _, ok := node.X.(*ast.Ident); ok {
						//name = pkg.TypesInfo.Uses[i].Name()
						return true
					}

					name := pkg.TypesInfo.Uses[node.Sel].Name()
					switch name {
					case "Trace", "Debug", "Info", "Warn", "Error":
					default:
						return true
					}
					parents, _ := astutil.PathEnclosingInterval(f, n.Pos(), n.End())
					for _, par := range parents {
						if fun, ok := par.(*ast.FuncDecl); ok {
							funcName = fun.Name.Name
							break
						}
					}

				default:
					return true
				}

				//log.Println(x, name)
				if len(x.Args) > 0 {
					firstArgStart, firstArgEnd = x.Args[0].Pos(), x.Args[0].End()
				}
			case *ast.BasicLit:
				if x.Pos() == firstArgStart && x.End() == firstArgEnd && x.Kind == token.STRING {
					logToPos[x.Value] = append(logToPos[x.Value], pkg.Fset.Position(n.Pos()).String()+" ("+funcName+")")
				}
			}
			return true
		})
	}
}
