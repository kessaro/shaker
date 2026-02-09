package shaker

import "testing"

func TestErrors(t *testing.T) {
	tests := []struct {
		name                string
		err                 error
		expectedErrorString string
	}{
		{
			name:                "Not found without resource name",
			err:                 ErrNotFound(),
			expectedErrorString: "resource not found",
		},
		{
			name:                "Not found with resource name",
			err:                 ErrNotFoundf("user"),
			expectedErrorString: "user not found",
		},
	}

	for _, test := range tests {
		if got := test.err.Error(); got != test.expectedErrorString {
			t.Fatalf("the returned error string does not match with the expected one : %s != %s", got, test.expectedErrorString)
		}
	}
}
