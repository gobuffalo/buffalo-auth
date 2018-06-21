package auth

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"strings"
)

//SourceManipulator has helper classes to modify go source code
type SourceManipulator struct {
	filePath string
}

//NewSourceManipulator creates a SourceManipulator from a provided file
func NewSourceManipulator(filePath string) *SourceManipulator {
	return &SourceManipulator{filePath}
}

func (sm *SourceManipulator) AddImports(im ...string) error {
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

func (sm *SourceManipulator) Append(content []string) error {
	src, err := ioutil.ReadFile(sm.filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(src), "\n")
	c := append(lines, content...)
	fileContent := strings.Join(c, "\n")

	log.Println(fileContent)

	err = ioutil.WriteFile(sm.filePath, []byte(fileContent), 0755)
	return err
}

func (sm *SourceManipulator) RemoveLine(starting string) error {
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

func (sm *SourceManipulator) RemoveBlock(starting string) error {
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

func (sm *SourceManipulator) InsertInBlock(starting string, content []string) error {
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

	log.Println(fileContent)

	err = ioutil.WriteFile(sm.filePath, []byte(fileContent), 0755)
	return err
}

func (sm *SourceManipulator) InsertBeforeBlockEnd(startingExpr string, content []string) error {
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

func (sm *SourceManipulator) FindBlockFor(startingExpr string) (int, int, error) {
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
