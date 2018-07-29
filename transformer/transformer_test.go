package transformer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Transformer_AddImports(t *testing.T) {
	r := require.New(t)

	tcases := []struct {
		goldensPrefix string
		adition       string
	}{
		{"imports-1", `import "github.com/wawandco/fako"`},
		{"imports-2", "import \"github.com/wawandco/fako\"\nimport \"other/package\""},
	}

	for _, tcase := range tcases {

		matches, err := matchesAfter(tcase.goldensPrefix, func(tr *Transformer) {
			r.NoError(tr.AddImports(tcase.adition))
		})

		r.NoError(err)
		r.True(matches)
	}

}

func Test_Transformer_Append(t *testing.T) {
	r := require.New(t)

	tcases := []struct {
		goldensPrefix string
		source        string
	}{
		{"append-1", `//Adding comment at the bottom`},
		{"append-2", "func other() {\n\t//does something else\n}"},
	}

	for _, tcase := range tcases {
		matches, err := matchesAfter(tcase.goldensPrefix, func(tr *Transformer) {
			r.NoError(tr.Append(tcase.source))
		})

		r.NoError(err)
		r.True(matches)
	}

}

func matchesAfter(prefix string, fn func(tr *Transformer)) (bool, error) {
	base, err := ioutil.ReadFile(filepath.Join("testdata", prefix+"-in.golden"))
	if err != nil {
		return false, err
	}

	tmp := os.TempDir()
	path := filepath.Join(tmp, "file.go")
	ioutil.WriteFile(path, []byte(base), 0644)

	tr := NewTransformer(path)
	fn(tr)

	src, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}

	expected, err := ioutil.ReadFile(filepath.Join("testdata", prefix+"-out.golden"))
	if err != nil {
		return false, err
	}

	return string(src) == string(expected), nil
}
