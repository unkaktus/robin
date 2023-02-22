package slurm

import (
	"testing"

	"github.com/matryer/is"
)

func TestParseNodeList(t *testing.T) {
	is := is.New(t)

	nodelistString := "i01r07c03s[10-12,14]opa,i01r07c04s[10-12],i01r07c05s[12-13]"
	nodelist, err := parseNodeList(nodelistString)
	is.NoErr(err)
	is.Equal(nodelist, []string{"i01r07c03s10opa", "i01r07c03s11opa", "i01r07c03s12opa", "i01r07c03s14opa", "i01r07c04s10", "i01r07c04s11", "i01r07c04s12", "i01r07c05s12", "i01r07c05s13"})

}
