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
		inGolden  string
		adition   string
		outGolden string
	}{
		{
			"imports-1-in.golden",
			`import "github.com/wawandco/fako"`,
			"imports-1-out.golden",
		},

		{
			"imports-2-in.golden",
			`import "github.com/wawandco/fako"`,
			"imports-2-out.golden",
		},
	}

	for _, tcase := range tcases {

		base, err := ioutil.ReadFile(filepath.Join("testdata", tcase.inGolden))
		r.NoError(err)

		tmp := os.TempDir()
		path := filepath.Join(tmp, "imports.go")
		ioutil.WriteFile(path, []byte(base), 0644)

		tr := NewTransformer(path)
		r.NoError(tr.AddImports(tcase.adition))

		src, err := ioutil.ReadFile(path)
		r.NoError(err)

		expected, err := ioutil.ReadFile(filepath.Join("testdata", tcase.outGolden))
		r.NoError(err)

		r.Equal(src, expected)
	}

}
