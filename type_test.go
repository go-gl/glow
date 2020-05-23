package main

import "testing"

func TestGoType(t *testing.T) {
	tt := []struct {
		in       Type
		expected string
	}{
		{
			in: Type{
				Name:         "uintptr_t",
				PointerLevel: 1,
				CDefinition:  "uintptr_t*",
				Cast:         "void *",
			},
			expected: "*uintptr",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.in.String(), func(t *testing.T) {
			goType := tc.in.GoType()
			if goType != tc.expected {
				t.Errorf("expected <%s>, got <%s>", tc.expected, goType)
			}
		})
	}
}
