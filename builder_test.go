package d2protocolparser

import (
	"reflect"
	"testing"
)

func BenchmarkBuild(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := Build("./fixtures/DofusInvoker.swf")
		if err != nil {
			b.Errorf("expected nil, got %v", err)
		}
	}
}

func TestBuild(t *testing.T) {
	p, err := Build("./fixtures/DofusInvoker.swf")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	expectedVersion := Version{2, 39, 0, 117122, 0}
	if !reflect.DeepEqual(p.Version, expectedVersion) {
		t.Errorf("expected %v, got %v", expectedVersion, p.Version)
	}
}
