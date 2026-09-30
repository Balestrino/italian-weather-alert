package territory

import "testing"

func TestProcessingCompatibility(t *testing.T) {
	for _, v := range []struct {
		region, product, profile string
		want                     bool
	}{{"09", "criticality", "toscana-cfr", true}, {"03", "criticality", "toscana-cfr", false}, {"03", "municipal", "municipal-html", true}, {"09", "municipal", "toscana-cfr", false}, {"03", "criticality", "unsupported", false}} {
		if Compatible(v.region, v.product, v.profile) != v.want {
			t.Fatal(v)
		}
	}
}
