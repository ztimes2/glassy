package ui

import (
	. "github.com/maragudk/gomponents"
)

// mapIndex is like [github.com/maragudk/gomponents.Map] but also passes the index to the callback function.
func mapIndex[T any](ts []T, fn func(int, T) Node) []Node {
	var nodes []Node
	for i, t := range ts {
		nodes = append(nodes, fn(i, t))
	}
	return nodes
}
