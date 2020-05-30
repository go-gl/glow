package main

import "testing"

func TestParseSignature(t *testing.T) {
	tt := []struct {
		input        string
		expectedName string
		expectedType Type
	}{
		{
			input:        "const void *<name>pointer</name>",
			expectedName: "pointer",
			expectedType: Type{
				Name:         "void",
				PointerLevel: 1,
				CDefinition:  "const void *",
			},
		},
		{
			input:        "<ptype>GLsizei</ptype> <name>stride</name>",
			expectedName: "stride",
			expectedType: Type{
				Name:         "GLsizei",
				PointerLevel: 0,
				CDefinition:  "GLsizei ",
			},
		},
		{
			input:        "const <ptype>GLuint</ptype> *<name>value</name>",
			expectedName: "value",
			expectedType: Type{
				Name:         "GLuint",
				PointerLevel: 1,
				CDefinition:  "const GLuint *",
			},
		},
		{
			input:        "<ptype>GLuint</ptype> <name>baseAndCount</name>[2]",
			expectedName: "baseAndCount",
			expectedType: Type{
				Name:         "GLuint",
				PointerLevel: 1,
				CDefinition:  "GLuint *",
			},
		},
		{
			input:        "uintptr_t **",
			expectedName: "",
			expectedType: Type{
				Name:         "uintptr_t",
				PointerLevel: 2,
				CDefinition:  "uintptr_t **",
			},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			name, ctype, err := parseSignature(xmlSignature(tc.input))
			failed := false
			if err != nil {
				t.Logf("parseSignature returned error: %v", err)
				failed = true
			}
			if name != tc.expectedName {
				t.Logf("name [%s] does not match expected [%s]", name, tc.expectedName)
				failed = true
			}
			if ctype != tc.expectedType {
				t.Logf("type [%v] does not match expected [%v]", ctype, tc.expectedType)
				failed = true
			}
			if failed {
				t.Fail()
			}
		})
	}
}
