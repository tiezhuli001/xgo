// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goinfo

func CompareSemVer(v, w string) int {
	pv, ok1 := parseSemver(v)
	pw, ok2 := parseSemver(w)
	if !ok1 && !ok2 {
		return 0
	}
	if !ok1 {
		return -1
	}
	if !ok2 {
		return +1
	}
	if c := compareInt(pv.major, pw.major); c != 0 {
		return c
	}
	if c := compareInt(pv.minor, pw.minor); c != 0 {
		return c
	}
	if c := compareInt(pv.patch, pw.patch); c != 0 {
		return c
	}
	return comparePrerelease(pv.prerelease, pw.prerelease)
}

func IsValidSemVer(v string) bool {
	_, ok := parseSemver(v)
	return ok
}

type parsedSemver struct {
	major      string
	minor      string
	patch      string
	short      string
	prerelease string
	build      string
}

func parseSemver(v string) (p parsedSemver, ok bool) {
	if v == "" || v[0] != 'v' {
		return
	}
	p.major, v, ok = parseInt(v[1:])
	if !ok {
		return
	}
	if v == "" {
		p.minor = "0"
		p.patch = "0"
		p.short = ".0.0"
		return
	}
	if v[0] != '.' {
		ok = false
		return
	}
	p.minor, v, ok = parseInt(v[1:])
	if !ok {
		return
	}
	if v == "" {
		p.patch = "0"
		p.short = ".0"
		return
	}
	if v[0] != '.' {
		ok = false
		return
	}
	p.patch, v, ok = parseInt(v[1:])
	if !ok {
		return
	}
	if len(v) > 0 && v[0] == '-' {
		p.prerelease, v, ok = parsePrerelease(v)
		if !ok {
			return
		}
	}
	if len(v) > 0 && v[0] == '+' {
		p.build, v, ok = parseBuild(v)
		if !ok {
			return
		}
	}
	if v != "" {
		ok = false
		return
	}
	ok = true
	return
}

func parseBuild(v string) (t, rest string, ok bool) {
	if v == "" || v[0] != '+' {
		return
	}
	i := 1
	start := 1
	for i < len(v) {
		if !isIdentChar(v[i]) && v[i] != '.' {
			return
		}
		if v[i] == '.' {
			if start == i {
				return
			}
			start = i + 1
		}
		i++
	}
	if start == i {
		return
	}
	return v[:i], v[i:], true
}

func parseInt(v string) (t, rest string, ok bool) {
	if v == "" {
		return
	}
	if v[0] < '0' || '9' < v[0] {
		return
	}
	i := 1
	for i < len(v) && '0' <= v[i] && v[i] <= '9' {
		i++
	}
	if v[0] == '0' && i != 1 {
		return
	}
	return v[:i], v[i:], true
}

func parsePrerelease(v string) (t, rest string, ok bool) {
	// "A pre-release version MAY be denoted by appending a hyphen and
	// a series of dot separated identifiers immediately following the patch version.
	// Identifiers MUST comprise only ASCII alphanumerics and hyphen [0-9A-Za-z-].
	// Identifiers MUST NOT be empty. Numeric identifiers MUST NOT include leading zeroes."
	if v == "" || v[0] != '-' {
		return
	}
	i := 1
	start := 1
	for i < len(v) && v[i] != '+' {
		if !isIdentChar(v[i]) && v[i] != '.' {
			return
		}
		if v[i] == '.' {
			if start == i || isBadNum(v[start:i]) {
				return
			}
			start = i + 1
		}
		i++
	}
	if start == i || isBadNum(v[start:i]) {
		return
	}
	return v[:i], v[i:], true
}

func isIdentChar(c byte) bool {
	return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '-'
}

func isBadNum(v string) bool {
	i := 0
	for i < len(v) && '0' <= v[i] && v[i] <= '9' {
		i++
	}
	return i == len(v) && i > 1 && v[0] == '0'
}

func isNum(v string) bool {
	i := 0
	for i < len(v) && '0' <= v[i] && v[i] <= '9' {
		i++
	}
	return i == len(v)
}

func compareInt(x, y string) int {
	if x == y {
		return 0
	}
	if len(x) < len(y) {
		return -1
	}
	if len(x) > len(y) {
		return +1
	}
	if x < y {
		return -1
	} else {
		return +1
	}
}

func comparePrerelease(x, y string) int {
	// "When major, minor, and patch are equal, a pre-release version has
	// lower precedence than a normal version.
	// Example: 1.0.0-alpha < 1.0.0.
	// Precedence for two pre-release versions with the same major, minor,
	// and patch version MUST be determined by comparing each dot separated
	// identifier from left to right until a difference is found as follows:
	// identifiers consisting of only digits are compared numerically and
	// identifiers with letters or hyphens are compared lexically in ASCII
	// sort order. Numeric identifiers always have lower precedence than
	// non-numeric identifiers. A larger set of pre-release fields has a
	// higher precedence than a smaller set, if all of the preceding
	// identifiers are equal.
	// Example: 1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta <
	// 1.0.0-beta < 1.0.0-beta.2 < 1.0.0-beta.11 < 1.0.0-rc.1 < 1.0.0."
	if x == y {
		return 0
	}
	if x == "" {
		return +1
	}
	if y == "" {
		return -1
	}
	for x != "" && y != "" {
		x = x[1:] // skip - or .
		y = y[1:] // skip - or .
		var dx, dy string
		dx, x = nextIdent(x)
		dy, y = nextIdent(y)
		if dx != dy {
			numX := isNum(dx)
			numY := isNum(dy)
			if numX != numY {
				if numX {
					return -1
				} else {
					return +1
				}
			}
			if numX {
				if len(dx) < len(dy) {
					return -1
				}
				if len(dx) > len(dy) {
					return +1
				}
			}
			if dx < dy {
				return -1
			} else {
				return +1
			}
		}
	}
	if x == "" {
		return -1
	} else {
		return +1
	}
}

func nextIdent(x string) (dx, rest string) {
	i := 0
	for i < len(x) && x[i] != '.' {
		i++
	}
	return x[:i], x[i:]
}
