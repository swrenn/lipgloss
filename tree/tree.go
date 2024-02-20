package tree

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// Node is a node in a tree.
type Node interface {
	Name() string
	String() string
	Children() []Node
}

// Atter returns a child node in a specified index.
type Atter interface {
	At(i int) Node
}

type atterImpl []Node

func (a atterImpl) At(i int) Node {
	if i >= 0 && i < len(a) {
		return a[i]
	}
	return nil
}

// StringNode is a node without children.
type StringNode string

// Children conforms with Node.
// StringNodes have no children.
func (StringNode) Children() []Node { return nil }

// Name conforms with Node.
// Returns the value of the string itself.
func (s StringNode) Name() string { return string(s) }

func (s StringNode) String() string { return s.Name() }

// TreeNode implements the Node interface with String data.
type TreeNode struct { //nolint:revive
	name         string
	renderer     *defaultRenderer
	rendererOnce sync.Once
	children     []Node
}

// Name returns the root name of this node.
func (n *TreeNode) Name() string { return n.name }

func (n *TreeNode) String() string {
	return n.ensureRenderer().Render(n, true, "")
}

// Item appends an item to a list.
//
// If the tree being added is a new TreeNode without a name, we add its
// children to the previous string node.
//
// This is mostly syntactic sugar for adding items to lists.
//
// Both of these should result in the same thing:
//
//	New("foo", "bar", New("", "zaz"))
//	New("foo", New("bar", "zaz"))
//
// The resulting tree would be:
// - foo
// - bar
//   - zaz
func (n *TreeNode) Item(item any) *TreeNode {
	switch item := item.(type) {
	case *TreeNode:
		newItem, rm := ensureParent(n.children, item)
		if rm >= 0 {
			n.children = remove(n.children, rm)
		}
		n.children = append(n.children, newItem)
	case Node:
		n.children = append(n.children, item)
	case string:
		s := StringNode(item)
		n.children = append(n.children, &s)
	}
	return n
}

// walks backwards in the existing nodes until it finds a string node, then
// remove it from the list and set it as the parent of the current node.
func ensureParent(nodes []Node, item *TreeNode) (*TreeNode, int) {
	if item.Name() != "" {
		return item, -1
	}
	for j := len(nodes) - 1; j >= 0; j-- {
		parent := nodes[j]
		switch parent := parent.(type) {
		case StringNode:
			item.name = parent.Name()
			return item, j
		case *StringNode:
			item.name = parent.Name()
			return item, j
		}
	}
	return item, -1
}

func remove(data []Node, i int) []Node {
	return append(data[:i], data[i+1:]...)
}

func (n *TreeNode) ensureRenderer() *defaultRenderer {
	n.rendererOnce.Do(func() {
		n.renderer = newDefaultRenderer()
	})
	return n.renderer
}

// EnumeratorStyle implements Renderer.
func (n *TreeNode) EnumeratorStyle(style lipgloss.Style) *TreeNode {
	n.ensureRenderer().style.enumeratorFunc = func(Atter, int) lipgloss.Style { return style }
	return n
}

// EnumeratorStyleFunc implements Renderer.
func (n *TreeNode) EnumeratorStyleFunc(fn StyleFunc) *TreeNode {
	if fn == nil {
		fn = func(Atter, int) lipgloss.Style { return lipgloss.NewStyle() }
	}
	n.ensureRenderer().style.enumeratorFunc = fn
	return n
}

// ItemStyle implements Renderer.
func (n *TreeNode) ItemStyle(style lipgloss.Style) *TreeNode {
	n.ensureRenderer().style.itemFunc = func(Atter, int) lipgloss.Style { return style }
	return n
}

// ItemStyleFunc implements Renderer.
func (n *TreeNode) ItemStyleFunc(fn StyleFunc) *TreeNode {
	if fn == nil {
		fn = func(Atter, int) lipgloss.Style { return lipgloss.NewStyle() }
	}
	n.ensureRenderer().style.enumeratorFunc = fn
	return n
}

// Enumerator implements Renderer.
func (n *TreeNode) Enumerator(enum Enumerator) *TreeNode {
	n.ensureRenderer().enumerator = enum
	return n
}

// Children returns the children of a string node.
func (n *TreeNode) Children() []Node {
	return n.children
}

// New returns a new tree.
func New(root string, data ...any) *TreeNode {
	t := &TreeNode{
		name: root,
	}
	for _, d := range data {
		t = t.Item(d)
	}
	return t
}