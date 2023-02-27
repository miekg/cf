package ast

// Node defines an ast node.
type Node interface {
	Parent() Node
	Children() []Node
	SetParent(Node)
	SetChildren([]Node)
	SetToken(t Token)
	Token() Token
	Type() string // return type, container, or leaf
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
func Remove(parent Node, index int) {
	cs := parent.Children()
	if index >= len(cs) {
		return
	}
	cs = append(cs[:index], cs[index+1:]...)
	parent.SetChildren(cs)
}

// Reverse reverse the order of the child nodes in parent. Only use is for ArgList.
func Reverse(arglist Node) {
	cs := arglist.Children()
	rs := make([]Node, len(cs))
	j := 0
	for i := len(cs) - 1; i >= 0; i-- {
		rs[j] = cs[i]
		j++
	}
	arglist.SetChildren(rs)

}

// Container is a type of node that can contain children.
type Container struct {
	parent   Node
	children []Node
	token    Token
}

const container = "container"

func (c *Container) Parent() Node          { return c.parent }
func (c *Container) Children() []Node      { return c.children }
func (c *Container) SetParent(p Node)      { c.parent = p }
func (c *Container) SetChildren(cs []Node) { c.children = cs }
func (c *Container) Token() Token          { return c.token }
func (c *Container) Type() string          { return container }
func (c *Container) SetToken(t Token) {
	c.token.Lit = t.Lit
	c.token.Tok = t.Tok
	// prepend or postpend?
	c.token.Comment = append(t.Comment, c.token.Comment...)
}

// Leaf is a type of node that cannot have children.
type Leaf struct {
	parent Node
	token  Token
}

const leaf = "leaf"

func (l *Leaf) Parent() Node          { return l.parent }
func (l *Leaf) Children() []Node      { return nil }
func (l *Leaf) SetParent(p Node)      { l.parent = p }
func (l *Leaf) SetChildren(cs []Node) { panic("ast: leaf can't have children") }
func (l *Leaf) Token() Token          { return l.token }
func (c *Leaf) Type() string          { return leaf }
func (c *Leaf) SetToken(t Token) {
	c.token.Lit = t.Lit
	c.token.Tok = t.Tok
	c.token.Comment = append(t.Comment, c.token.Comment...)
}

// New returns a new Node, with an optional token.
func New(n Node, t ...Token) Node {
	if len(t) > 0 {
		n.SetToken(t[0])
	}
	return n
}
