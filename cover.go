/*
Package cover implements DLX algorithm from Donald Knuth.
The paper can be found at: http://www-cs-faculty.stanford.edu/~knuth/musings.html,
search for "Dancing links" in the page.

It also includes tools to solve sudoku using Knuth's algorithm.
*/
package cover

import (
	"fmt"
	"log"
)

// Used for column nodes to remember their name and size.
type Meta struct {
	Size uint
	Name string
}

// Element of the four-way linked list.
type Node struct {
	Right, Up, Left, Down *Node
	Col                   *Node
	*Meta
}

// Initializes a node with neighbours pointing to itself.
func NewNode() *Node {
	n := &Node{}
	n.Left = n
	n.Right = n
	n.Up = n
	n.Down = n
	return n
}

// Initializes a column node as a normal node + meta.
func NewColNode(s string) *Node {
	n := NewNode()
	n.Meta = &Meta{Name: s}
	return n
}

// Appends a node to a row by putting it before the current node.
func (r *Node) RowAppend(n *Node) {
	n.Right = r
	n.Left = r.Left
	// inserts the node at the en of the row
	r.Left.Right = n
	r.Left = n
}

// Appends a node to a column by putting it before the current node.
// Note that the current node has to own meta in order to update the size.
func (c *Node) ColAppend(n *Node) {
	n.Col = c
	n.Down = c
	n.Up = c.Up
	// inserts the node at bottom of the col
	c.Up.Down = n
	c.Up = n
	c.Size++
}

func (n *Node) String() string {
	return fmt.Sprintf("&{Right:%p Up:%p Left:%p Down:%p Col:%p Meta:%+v}", n.Right, n.Up, n.Left, n.Down, n.Col, n.Meta)
}

// Reduces the matrix in a non-destructive way by hiding the column
// from the matrix headers as well as the intersecting rows.
func (c *Node) Cover() {
	log.Println("Cover col", c.Name)
	c.Right.Left = c.Left
	c.Left.Right = c.Right
	for i := c.Down; i != c; i = i.Down {
		for j := i.Right; j != i; j = j.Right {
			j.Down.Up = j.Up
			j.Up.Down = j.Down
			j.Col.Size--
		}
	}
}

// Expands the matrix bz restoring the columns and its intersecting rows.
// Beware that the order is important to properly undo a Cover() step.
func (c *Node) Uncover() {
	log.Println("Uncover col", c.Name)
	for i := c.Up; i != c; i = i.Up {
		for j := i.Left; j != i; j = j.Left {
			j.Col.Size++
			j.Down.Up = j
			j.Up.Down = j
		}
	}
	c.Right.Left = c
	c.Left.Right = c
}

// Embeds the root node to provide a clean interface.
type SparseMatrix struct {
	*Node
}

/*
Given a binary matrix like:

  A  B  C  D  E  F  G
[[0, 0, 1, 0, 1, 1, 0] (3: CEF)
 [1, 0, 0, 1, 0, 0, 1]
 [0, 1, 1, 0, 0, 1, 0]
 [1, 0, 0, 1, 0, 0, 0] (1: AD)
 [0, 1, 0, 0, 0, 0, 1] (2: BG)
 [0, 0, 0, 1, 1, 0, 1]]

it return a sparse matrix made of horizontally and vertically
double linked nodes for 1 values.
*/
func NewSparseMatrix(matrix [][]int, headers []string) *SparseMatrix {
	rowCount := len(matrix)
	colCount := len(headers)
	root := &Node{Meta: &Meta{Name: "root"}}
	root.Left = root
	root.Right = root
	// create the columns
	for _, h := range headers {
		head := NewColNode(h)
		root.RowAppend(head)
	}
	for i := 0; i < rowCount; i++ {
		var prev, head *Node
		head = root.Right
		for j := 0; j < colCount; j++ {
			if matrix[i][j] > 0 {
				node := NewNode()
				head.ColAppend(node)
				if prev != nil {
					prev.RowAppend(node)
				} else {
					prev = node
				}
			}
			head = head.Right
		}
	}
	return &SparseMatrix{root}
}

// Returns the column having the smallest number of intersecting rows.
// It used to reduce the branching in the Search() method.
func (m *SparseMatrix) SmallestCol() *Node {
	var r *Node
	min := ^uint(0)
	// we want the underlying node rather than the matrix for comparison
	root := m.Root()
	for col := root.Right; col != root; col = col.Right {
		if col.Size < min {
			r = col
			min = col.Size
		}
	}
	return r
}

// Get the root element of the matrix.
func (m *SparseMatrix) Root() *Node {
	return m.Left.Right
}

// Returns the column of the specified name. Panics it not found.
func (m *SparseMatrix) Col(name string) *Node {
	root := m.Root()
	for col := root.Right; col != root; col = col.Right {
		if col.Name == name {
			return col
		}
	}
	panic(fmt.Sprintf("Column \"%v\" not found", name))
}

// Heart of the DLX algorithm.
func (m *SparseMatrix) Search(O *Solution, k int, g Guesser) {
	log.Println(k)
	root := m.Root()
	if root.Right == root {
		fmt.Println(O)
		return
	}
	c, bt := g.ChooseCol(k)
	c.Cover()
	for r := c.Down; r != c; r = r.Down {
		O.Set(k, r)
		for j := r.Right; j != r; j = j.Right {
			j.Col.Cover()
		}
		m.Search(O, k+1, g)
		if !bt {
			return
		}
		r = O.Get(k)
		c = r.Col
		for j := r.Left; j != r; j = j.Left {
			j.Col.Uncover()
		}
	}
	c.Uncover()
}

// Embeds a sparse matrix to provide clean interface.
type Solver struct {
	matrix *SparseMatrix
}

func NewSolver(m [][]int, h []string) *Solver {
	s := Solver{matrix: NewSparseMatrix(m, h)}
	return &s
}
func (s *Solver) Solve() *Solution {
	O := new(Solution)
	s.matrix.Search(O, 0, s)
	return O
}

// A guesser is an object able to choose a specific column for the DLX algorithm.
type Guesser interface {
	// Given a specific level, returns a node for the current step and a boolean
	// telling wether this step should backtracked or not.
	ChooseCol(int) (*Node, bool)
}

// Chooses the column havng the smallest number of interesecting rows and always
// asks for backtracking.
func (s *Solver) ChooseCol(k int) (*Node, bool) {
	m := s.matrix
	log.Println("guess is", m.SmallestCol().Name, "(", m.SmallestCol().Size, "), bt", true)
	return m.SmallestCol(), true
}

// Aliases a Node pointer array to provide a nice interface.
type Solution []*Node

func (s *Solution) Set(i int, n *Node) {
	if i < len(*s) {
		(*s)[i] = n
	} else {
		(*s) = append((*s), n)
	}
}
func (s *Solution) Get(i int) *Node {
	return (*s)[i]
}
func (s *Solution) Len() int {
	return len(*s)
}

func (s *Solution) String() string {
	o := ""
	for _, n := range *s {
		if n != nil {
			o += n.Col.Name
			for m := n.Right; n != m; m = m.Right {
				o += " " + m.Col.Name
			}
			o += "\n"
		}
	}
	return o
}
