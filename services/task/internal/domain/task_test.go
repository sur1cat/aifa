package domain

import "testing"

func TestKind_Valid(t *testing.T) {
	valid := []Kind{KindTodo, KindBill, KindIncome}
	for _, k := range valid {
		if !k.Valid() {
			t.Errorf("expected %q to be valid", k)
		}
	}
	invalid := []Kind{"", "unknown", "TODO", "expense"}
	for _, k := range invalid {
		if k.Valid() {
			t.Errorf("expected %q to be invalid", k)
		}
	}
}
