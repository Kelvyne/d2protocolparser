package d2protocolparser

import "testing"

func BenchmarkBuild(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := Build("./fixtures/DofusInvoker.swf")
		if err != nil {
			b.Errorf("expected nil, got %v", err)
		}
	}
}

func TestBuild(t *testing.T) {
	_, err := Build("./fixtures/DofusInvoker.swf")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}
