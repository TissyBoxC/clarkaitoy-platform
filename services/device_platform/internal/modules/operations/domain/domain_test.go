package domain

import "testing"

// Equal plain versions previously panicked because the comparator indexed a
// pre-release segment that SplitN never produced. The OTA and service-version
// modules both call this function, so the regression is release-blocking.
func TestCompareVersionsEqualPlainVersions(t *testing.T) {
	t.Parallel()

	if got := CompareVersions("0.12.6", "0.12.6"); got != 0 {
		t.Fatalf("CompareVersions(0.12.6, 0.12.6) = %d, want 0", got)
	}
	if got := CompareVersions("v0.12.6", "0.12.6"); got != 0 {
		t.Fatalf("CompareVersions(v0.12.6, 0.12.6) = %d, want 0", got)
	}
}

func TestCompareVersionsOrdering(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		left  string
		right string
		want  int
	}{
		{name: "older patch", left: "0.12.5", right: "0.12.6", want: -1},
		{name: "newer minor", left: "0.13.0", right: "0.12.9", want: 1},
		{name: "missing patch segment", left: "0.12", right: "0.12.0", want: 0},
		{name: "release sorts after prerelease", left: "1.0.0", right: "1.0.0-rc.1", want: 1},
		{name: "prerelease sorts before release", left: "1.0.0-beta", right: "1.0.0", want: -1},
		{name: "prerelease lexicographic", left: "1.0.0-alpha", right: "1.0.0-beta", want: -1},
		{name: "leading v prefix ignored", left: "v1.2.3", right: "1.2.3", want: 0},
	}

	for _, testCase := range cases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := CompareVersions(testCase.left, testCase.right); got != testCase.want {
				t.Fatalf(
					"CompareVersions(%q, %q) = %d, want %d",
					testCase.left,
					testCase.right,
					got,
					testCase.want,
				)
			}
		})
	}
}
