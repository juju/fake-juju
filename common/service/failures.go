// Track which entities should fail

package service

import (
	"fmt"
	"strings"
)

// The given entity will fail as soon as possible.
func SetFailure(entity string) {
	failures[entity] = true
}

// Whether the given entity should fail
func ShouldFail(kind, id string) bool {
	id = strings.Replace(id, "/", "-", -1)
	_, ok := failures[fmt.Sprintf("%s-%s", kind, id)]
	return ok
}

// Clear all scheduled failures
func ClearFailures() {
	for key := range failures {
		delete(failures, key)
	}
}

var failures = make(map[string]bool)
