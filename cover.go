package cover

import (
	"fmt"
	"log"
	"sort"
)

type Meta struct {
	Size uint
	Name string
}

type Node struct {
	Right, Up, Left, Down *Node
	Col                   *Node
	Meta                  *Meta
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

// Appends a node to a col by putting it before the current node.
// Note that the current node has to own meta in order to update the size.
func (c *Node) ColAppend(n *Node) {
	n.Col = c
	n.Down = c
	n.Up = c.Up
	// inserts the node at bottom of the col
	c.Up.Down = n
	c.Up = n
	c.Meta.Size++
}

func (n *Node) String() string {
	return fmt.Sprintf("&{Right:%p Up:%p Left:%p Down:%p Col:%p Meta:%+v}", n.Right, n.Up, n.Left, n.Down, n.Col, n.Meta)
}

func (c *Node) Cover() {
	log.Println("Cover col", c.Meta.Name)
	c.Right.Left = c.Left
	c.Left.Right = c.Right
	for i := c.Down; i != c; i = i.Down {
		for j := i.Right; j != i; j = j.Right {
			j.Down.Up = j.Up
			j.Up.Down = j.Down
			j.Col.Meta.Size--
		}
	}
}

func (c *Node) Uncover() {
	log.Println("Uncover col", c.Meta.Name)
	for i := c.Up; i != c; i = i.Up {
		for j := i.Left; j != i; j = j.Left {
			j.Col.Meta.Size++
			j.Down.Up = j
			j.Up.Down = j
		}
	}
	c.Right.Left = c
	c.Left.Right = c
}

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
return a sparse matrix made of horizontally and vertically
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

func (m *SparseMatrix) SmallestCol() *Node {
	var r *Node
	min := ^uint(0)
	// we want the underlying node rather than the matrix for comparison
	root := m.Root()
	for col := root.Right; col != root; col = col.Right {
		if col.Meta.Size < min {
			r = col
			min = col.Meta.Size
		}
	}
	return r
}

// Get the root element of the matrix.
func (m *SparseMatrix) Root() *Node {
	return m.Left.Right
}
func (m *SparseMatrix) Col(name string) *Node {
	root := m.Root()
	for col := root.Right; col != root; col = col.Right {
		if col.Meta.Name == name {
			return col
		}
	}
	panic(fmt.Sprintf("Column \"%v\" not found", name))
}

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

type Solver struct {
	matrix *SparseMatrix
}

func NewSolver(m [][]int, h []string) *Solver {
	s := Solver{matrix: NewSparseMatrix(m, h)}
	return &s
}
func (s *Solver) Solve(partial map[int][]string) *Solution {
	keys := make([]int, 0, len(partial))
	for key, _ := range partial {
		keys = append(keys, key)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	log.Println("Initial config is", partial)
	O := new(Solution)
	k := 0
	m := s.matrix
	for _, digit := range keys {
		for _, c := range partial[digit] {
			n := m.Col(c)
			n.Cover()
			for j := 0; j < digit; j++ {
				n = n.Down
			}
			O.Set(k, n)
			for o := n.Right; o != n; o = o.Right {
				o.Col.Cover()
			}
			k++
		}
	}
	fmt.Printf("Initial solution is\n%v", O)
	s.matrix.Search(O, k, s)
	return O
}

type Guesser interface {
	ChooseCol(int) (*Node, bool)
}

// Given a specific level, returns a node for the next step and a boolean
// telling wether this step should backtracked or not.
func (s *Solver) ChooseCol(k int) (*Node, bool) {
	m := s.matrix
	log.Println("guess is", m.SmallestCol().Meta.Name, "(", m.SmallestCol().Meta.Size, "), bt", true)
	return m.SmallestCol(), true
}

type Solution struct {
	items []*Node
}

func NewSolution() *Solution {
	s := Solution{items: make([]*Node, 0)}
	return &s
}
func (s *Solution) Set(i int, n *Node) {
	if i < len(s.items) {
		s.items[i] = n
	} else {
		s.items = append(s.items, n)
	}
}
func (s *Solution) Get(i int) *Node {
	return s.items[i]
}

func (s *Solution) Len() int {
	return len(s.items)
}

func (s *Solution) String() string {
	o := ""
	for _, n := range s.items {
		if n != nil {
			o += n.Col.Meta.Name
			for m := n.Right; n != m; m = m.Right {
				o += " " + m.Col.Meta.Name
			}
			o += "\n"
		}
	}
	return o
}
