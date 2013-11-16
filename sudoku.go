package cover

import (
	"fmt"
	"log"
	"math"
	"sort"
)

// Builds a constraint matrix for a sudoku of the given dimension.
// The constraint matrix can then be used by the DLX algorithm.
func SudokuConstraintMatrix(dim int) (matrix [][]int, headers []string) {
	// small dim, 3 for classic sudoku
	sdim := int(math.Sqrt(float64(dim)))
	// big dim, 81 for classic sudoku
	bdim := dim * dim
	rowCount := bdim * dim
	colCount := bdim * 4
	log.Printf("Building sparse matrix of %dx%d\n", rowCount, colCount)
	// constraint matrix headers
	// constraint order is existence, row, col, block
	headers = make([]string, colCount)
	for i, j := 0, 0; i < colCount; i++ {
		j = i % bdim
		if i < bdim {
			// constraint 1: existence
			headers[i] = fmt.Sprintf("%v,%v", i/dim, i%dim)
		} else if i < 2*bdim {
			// constraint 2: row
			headers[i] = fmt.Sprintf("%vr%v", j%dim+1, j/dim)
		} else if i < 3*bdim {
			// constraint 3: column
			headers[i] = fmt.Sprintf("%vc%v", j%dim+1, j/dim)
		} else {
			// constraint 4: block
			headers[i] = fmt.Sprintf("%vb%v", j%dim+1, j/dim)
		}
	}
	// constraint matrix
	matrix = make([][]int, rowCount)
	for i := 0; i < rowCount; i++ {
		matrix[i] = make([]int, colCount)
		digit := i%dim + 1
		dcell := i / dim
		drow := i / bdim
		dcol := (i / dim) % dim
		dblock := drow/sdim*sdim + dcol/sdim
		matrix[i][dcell] = digit
		matrix[i][bdim+drow*dim+i%dim] = digit
		matrix[i][bdim+bdim+i%bdim] = digit
		matrix[i][bdim+bdim+bdim+dblock*dim+i%dim] = digit
	}
	return
}

type SudokuSolver struct {
	*Solver
}

// Since the constraint matrix for a sudoku only depends on its size, this constructor
// encapsulate the matrix creation so that only the sudoku size is needed.
func NewSudokuSolver(dim int) *SudokuSolver {
	m, h := SudokuConstraintMatrix(dim)
	log.Println(len(h), h)
	log.Println(len(m))
	s := SudokuSolver{&Solver{matrix: NewSparseMatrix(m, h)}}
	return &s
}

// Translates the initial grid to a map of digit => cells.
// This enables the solver to safely initialize the matrix before
// actually running the DLX algorithm.
// The trick is to search only in the columns of the 1st constraint (existence)
// and start from the biggest digit. This way, the number of steps to find the correct
// row does not change and the digit can be found in a single column.
func (s *SudokuSolver) gridToCover(sudoku [][]int) map[int][]string {
	dim := len(sudoku)
	init := map[int][]string{}
	for i := 0; i < dim; i++ {
		for j := 0; j < dim; j++ {
			digit := sudoku[i][j]
			if digit > 0 {
				_, ok := init[digit]
				if !ok {
					init[digit] = make([]string, 0)
				}
				init[digit] = append(init[digit], fmt.Sprintf("%v,%v", i, j))
			}
		}
	}
	return init
}
func (s *SudokuSolver) Solve(sudoku [][]int) *Solution {
	partial := s.gridToCover(sudoku)
	// Iterate through the digits from biggest to smallest.
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
			// Find the column for existence constraint, so that all the digits are available inside.
			n := m.Col(c)
			n.Cover()
			// Move down by <digit> nodes to find the row to include in the solution.
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
