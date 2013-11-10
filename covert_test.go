package cover

import "testing"

func TestNewNode(t *testing.T) {
	n := NewNode()
	if n.Up != n {
		t.Errorf("Node.Up points to %p (wants %p)", n.Up, n)
	}
	if n.Down != n {
		t.Errorf("Node.Down points to %p (wants %p)", n.Down, n)
	}
	if n.Left != n {
		t.Errorf("Node.Left points to %p (wants %p)", n.Left, n)
	}
	if n.Right != n {
		t.Errorf("Node.Right points to %p (wants %p)", n.Right, n)
	}
	if n.Col != nil {
		t.Errorf("Node.Col points to %p (wants %p)", n.Col, nil)
	}
}

func TestRowAppend(t *testing.T) {
	n := NewNode()
	p := NewNode()
	q := NewNode()
	n.RowAppend(p)
	n.RowAppend(q)
	if n.Right != p {
		t.Errorf("Node.Right points to %p (wants %p)", n.Right, p)
	}
	if p.Right != q {
		t.Errorf("Node.Right points to %p (wants %p)", p.Right, q)
	}
	if q.Right != n {
		t.Errorf("Node.Right points to %p (wants %p)", q.Right, n)
	}
	if n.Left != q {
		t.Errorf("Node.Left points to %p (wants %p)", n.Left, q)
	}
	if p.Left != n {
		t.Errorf("Node.Left points to %p (wants %p)", p.Left, n)
	}
	if q.Left != p {
		t.Errorf("Node.Left points to %p (wants %p)", q.Left, p)
	}
}

func TestColAppend(t *testing.T) {
	n := NewColNode("The col")
	if n.Meta.Size != 0 {
		t.Errorf("ColNode has size %v (wants %v)", n.Meta.Size, 0)
	}
	p := NewNode()
	q := NewNode()
	n.ColAppend(p)
	n.ColAppend(q)
	if n.Meta.Size != 2 {
		t.Errorf("ColNode has size %v (wants %v)", n.Meta.Size, 2)
	}
	if n.Col != nil {
		t.Errorf("Node.Col points to %p (wants %p)", n.Col, nil)
	}
	if p.Col != n {
		t.Errorf("Node.Col points to %p (wants %p)", p.Col, n)
	}
	if q.Col != n {
		t.Errorf("Node.Col points to %p (wants %p)", q.Col, n)
	}
	if n.Down != p {
		t.Errorf("Node.Down points to %p (wants %p)", n.Down, p)
	}
	if p.Down != q {
		t.Errorf("Node.Down points to %p (wants %p)", p.Down, q)
	}
	if q.Down != n {
		t.Errorf("Node.Down points to %p (wants %p)", q.Down, n)
	}
	if n.Up != q {
		t.Errorf("Node.Up points to %p (wants %p)", n.Up, q)
	}
	if p.Up != n {
		t.Errorf("Node.Up points to %p (wants %p)", p.Up, n)
	}
	if q.Up != p {
		t.Errorf("Node.Up points to %p (wants %p)", q.Up, p)
	}
}
