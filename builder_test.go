package d2protocolbuilder

import "testing"

func TestBuild(t *testing.T) {
	_, err := Build("./fixtures/DofusInvoker.swf")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}
