package ast

// Node defines an ast node.
type Node interface {
	Parent() Node
	Children() []Node
	SetParent(Node)
	SetChildren([]Node)
	SetToken(t Token)
	Token() Token
}

// Append appends child to the children of parent. It panics if either node is nil.
func Append(parent Node, child Node) {
	child.SetParent(parent)
	if parent.Children() == nil {
		parent.SetChildren([]Node{child})
	} else {
		cs := append(parent.Children(), child)
		parent.SetChildren(cs)
	}
}

// Remove removes the child from parent at index i.
func RemoveChild(parent Node, index int) {
	cs := parent.Children()
	if index >= len(cs) {
		return
	}
	cs = append(cs[:index], cs[index+1:]...)
	parent.SetChildren(cs)
}

// Container is a type of node that can contain children.
type Container struct {
	parent   Node
	children []Node
	token    Token
}

func (c *Container) Parent() Node          { return c.parent }
func (c *Container) Children() []Node      { return c.children }
func (c *Container) SetParent(p Node)      { c.parent = p }
func (c *Container) SetChildren(cs []Node) { c.children = cs }
func (c *Container) SetToken(t Token)      { c.token = t }
func (c *Container) Token() Token          { return c.token }

// Leaf is a type of node that cannot have children.
type Leaf struct {
	parent Node
	token  Token
}

func (l *Leaf) Parent() Node          { return l.parent }
func (l *Leaf) Children() []Node      { return nil }
func (l *Leaf) SetParent(p Node)      { l.parent = p }
func (l *Leaf) SetChildren(cs []Node) { panic("ast: leaf can't have children") }
func (l *Leaf) SetToken(t Token)      { l.token = t }
func (l *Leaf) Token() Token          { return l.token }

// New returns a new Node, with an optional token.
func New(n Node, t ...Token) Node {
	if len(t) > 0 {
		n.SetToken(t[0])
	}
	return n
}
