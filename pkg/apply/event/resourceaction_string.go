// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Code generated by "stringer -type=ResourceAction -linecomment"; DO NOT EDIT.

package event

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ApplyAction-0]
	_ = x[PruneAction-1]
	_ = x[DeleteAction-2]
	_ = x[WaitAction-3]
	_ = x[InventoryAction-4]
}

const _ResourceAction_name = "ApplyPruneDeleteWaitInventory"

var _ResourceAction_index = [...]uint8{0, 5, 10, 16, 20, 29}

func (i ResourceAction) String() string {
	if i < 0 || i >= ResourceAction(len(_ResourceAction_index)-1) {
		return "ResourceAction(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ResourceAction_name[_ResourceAction_index[i]:_ResourceAction_index[i+1]]
}
