package gomvc

import "testing"

func TestReplaceDashes(t *testing.T) {
	type testpair struct{ in, out string }

	tests := []testpair{
		{"hello-world", "helloWorld"},
		{"register", "register"},
		{"view-room-", "viewRoom"},
		{"best-web-page-in-the-world", "bestWebPageInTheWorld"},
	}

	for _, test := range tests {

		if res := replaceDashes(test.in); res != test.out {
			t.Errorf("replaceDashes(%q) = %q, want %q", test.in, res, test.out)

		}
	}
}
