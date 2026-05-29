//go:build tools

// This file pins github.com/oklog/ulid/v2 as a direct dependency for the
// state-redesign milestones. M1 (internal/triggerctx) introduces the first
// production import of the library; until then the build tag keeps this file
// out of regular compilation while `go mod tidy` still treats the package as
// a direct require thanks to the blank import below.
//
// Remove this file (and the corresponding require line if it becomes
// indirect) once a real M1+ caller imports the package.

package statefs

import _ "github.com/oklog/ulid/v2"
