package crdt

import (
	"bytes"
	"testing"
)

func TestCrdt(t *testing.T) {
	start, a, b, merged := TestStates()

	doc := NewDoc(start)

	t.Log("Merge a with start")
	doc.Merge(a)
	if !bytes.Equal(doc.State(), a) {
		t.Error("merging a with start didn't result in a")
	}

	t.Log(doc.GetTextValue("text"))

	t.Log("Merge b with a")
	doc.Merge(b)
	if !bytes.Equal(doc.State(), merged) {
		t.Error("merging b with a didn't result in merged")
	}

	t.Log(doc.GetTextValue("text"))
}
