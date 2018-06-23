package auth

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
)

//SourceOperator has helper classes to modify go source code
type SourceOperator struct {
	filePath string
}

//NewSourceOperator creates a SourceManipulator from a provided file
func NewSourceOperator(filePath string) *SourceOperator {
	return &SourceOperator{filePath}
}

//AddImports allows to add imports to a .go file
func (sm *SourceOperator) AddImports(im ...string) error {
	src, err := ioutil.ReadFile(sm.filePath)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, sm.filePath, string(src), 0)
	if err != nil {
		return err
	}

	var end = -1

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ImportSpec:
			end = fset.Position(x.End()).Line
			return true
		}
		return true
	})

	lines := strings.Split(string(src), "\n")
	c := append(lines[:end], append(im, lines[end:]...)...)
	fileContent := strings.Join(c, "\n")

	err = ioutil.WriteFile(sm.filePath, []byte(fileContent), 0755)
	return err
}

//Append appends to a source file
func (sm *SourceOperator) Append(content []string) error {
	src, err := ioutil.ReadFile(sm.filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(src), "\n")
	c := append(lines, content...)
	fileContent := strings.Join(c, "\n")

	err = ioutil.WriteFile(sm.filePath, []byte(fileContent), 0755)
	return err
}

//RemoveLine removes a line starting with some passed code
func (sm *SourceOperator) RemoveLine(starting string) error {
	src, err := ioutil.ReadFile(sm.filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(src), "\n")
	lineNum := -1
	for index, line := range lines {
		if strings.HasPrefix(line, starting) {
			lineNum = index
			break
		}
	}

	if lineNum > 0 {
		return nil
	}

	c := append(lines[:lineNum], lines[lineNum:]...)
	fileContent := strings.Join(c, "\n")

	err = ioutil.WriteFile(sm.filePath, []byte(fileContent), 0755)
	return err
}

//RemoveBlock removes a block starting with passed expression
func (sm *SourceOperator) RemoveBlock(starting string) error {
	start, end, err := sm.FindBlockFor(starting)
	if err != nil {
		return err
	}

	if end < 0 {
		return errors.New("could not find desired block on the app.go file")
	}

	src, err := ioutil.ReadFile(sm.filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(src), "\n")
	c := append(lines[:start-1], lines[end:]...)
	fileContent := strings.Join(c, "\n")

	err = ioutil.WriteFile(sm.filePath, []byte(fileContent), 0755)
	return err
}

//InsertInBlock replaces block body starting with string
func (sm *SourceOperator) InsertInBlock(starting string, content []string) error {
	start, end, err := sm.FindBlockFor(starting)
	if err != nil {
		return err
	}

	if end < 0 {
		return errors.New("could not find desired block on the app.go file")
	}

	src, err := ioutil.ReadFile(sm.filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(src), "\n")
	c := append(lines[:start], append(content, lines[end-1:]...)...)
	fileContent := strings.Join(c, "\n")

	err = ioutil.WriteFile(sm.filePath, []byte(fileContent), 0755)
	return err
}

//InsertBeforeBlockEnd adds source before block ends
func (sm *SourceOperator) InsertBeforeBlockEnd(startingExpr string, content []string) error {
	_, end, err := sm.FindBlockFor(startingExpr)
	if err != nil {
		return err
	}

	src, err := ioutil.ReadFile(sm.filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(src), "\n")
	c := append(lines[:end-1], append(content, lines[end-1:]...)...)
	fileContent := strings.Join(c, "\n")

	err = ioutil.WriteFile(sm.filePath, []byte(fileContent), 0755)
	return err
}

//FindBlockFor finds a block line start and end
func (sm *SourceOperator) FindBlockFor(startingExpr string) (int, int, error) {
	end, start := -1, -1

	src, err := ioutil.ReadFile(sm.filePath)
	if err != nil {
		return start, end, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, sm.filePath, string(src), 0)
	if err != nil {
		return start, end, err
	}

	lines := strings.Split(string(src), "\n")
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {

		case *ast.StructType:
			line := fset.Position(x.Pos()).Line
			structDeclaration := fmt.Sprintf("%s\n", lines[line-1])

			if strings.Contains(structDeclaration, startingExpr) {
				start = line
				end = fset.Position(x.End()).Line
				return false
			}

		case *ast.BlockStmt:
			line := fset.Position(x.Lbrace).Line
			blockDeclaration := fmt.Sprintf("%s\n", lines[line-1])

			if strings.Contains(blockDeclaration, startingExpr) {
				start = line
				end = fset.Position(x.Rbrace).Line
			}

		}
		return true
	})

	return start, end, nil
}
