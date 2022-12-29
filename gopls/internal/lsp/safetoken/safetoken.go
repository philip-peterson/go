// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package safetoken provides wrappers around methods in go/token,
// that return errors rather than panicking. It also provides a
// central place for workarounds in the underlying packages.
package safetoken

import (
	"fmt"
	"go/token"
)

// Offset returns f.Offset(pos), but first checks that the file
// contains the pos.
//
// The definition of "contains" here differs from that of token.File
// in order to work around a bug in the parser (issue #57490): during
// error recovery, the parser may create syntax nodes whose computed
// End position is 1 byte beyond EOF, which would cause
// token.File.Offset to panic. The workaround is that this function
// accepts a Pos that is exactly 1 byte beyond EOF and maps it to the
// EOF offset.
//
// The use of this function instead of (*token.File).Offset is
// mandatory in the gopls codebase; this is enforced by static check.
func Offset(f *token.File, pos token.Pos) (int, error) {
	if !inRange(f, pos) {
		// Accept a Pos that is 1 byte beyond EOF,
		// and map it to the EOF offset.
		// (Workaround for #57490.)
		if int(pos) == f.Base()+f.Size()+1 {
			return f.Size(), nil
		}

		return -1, fmt.Errorf("pos %d is not in range [%d:%d] of file %s",
			pos, f.Base(), f.Base()+f.Size(), f.Name())
	}
	return int(pos) - f.Base(), nil
}

// Pos returns f.Pos(offset), but first checks that the offset is
// non-negative and not larger than the size of the file.
func Pos(f *token.File, offset int) (token.Pos, error) {
	if !(0 <= offset && offset <= f.Size()) {
		return token.NoPos, fmt.Errorf("offset %d is not in range for file %s of size %d", offset, f.Name(), f.Size())
	}
	return token.Pos(f.Base() + offset), nil
}

// inRange reports whether file f contains position pos,
// according to the invariants of token.File.
//
// This function is not public because of the ambiguity it would
// create w.r.t. the definition of "contains". Use Offset instead.
func inRange(f *token.File, pos token.Pos) bool {
	return token.Pos(f.Base()) <= pos && pos <= token.Pos(f.Base()+f.Size())
}