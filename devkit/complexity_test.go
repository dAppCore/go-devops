// LEK-1 | lthn.ai | EUPL-1.2
package devkit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyseComplexitySource_SimpleFunc_Good(t *testing.T) {
	src := `package example

func simple() {
	x := 1
	_ = x
}
`
	results, err := AnalyseComplexitySource(src, "simple.go", 1)
	require.NoError(t, err)
	// Complexity = 1 (just the function body, no branches), threshold = 1
	assert.Len(t, results, 1)
	assert.Equal(t, "simple", results[0].FuncName)
	assert.Equal(t, "example", results[0].Package)
	assert.Equal(t, 1, results[0].Complexity)
}

func TestAnalyseComplexitySource_IfElse_Good(t *testing.T) {
	src := `package example

func branches(x int) string {
	if x > 0 {
		return "positive"
	} else if x < 0 {
		return "negative"
	}
	return "zero"
}
`
	results, err := AnalyseComplexitySource(src, "branches.go", 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	// 1 (base) + 1 (if) + 1 (else if) = 3
	assert.Equal(t, 3, results[0].Complexity)
	assert.Equal(t, "branches", results[0].FuncName)
}

func TestAnalyseComplexitySource_ForLoop_Good(t *testing.T) {
	src := `package example

func loopy(items []int) int {
	total := 0
	for _, v := range items {
		total += v
	}
	for i := 0; i < 10; i++ {
		total += i
	}
	return total
}
`
	results, err := AnalyseComplexitySource(src, "loops.go", 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	// 1 (base) + 1 (range) + 1 (for) = 3
	assert.Equal(t, 3, results[0].Complexity)
}

func TestAnalyseComplexitySource_SwitchCase_Good(t *testing.T) {
	src := `package example

func switcher(x int) string {
	switch x {
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	default:
		return "other"
	}
}
`
	results, err := AnalyseComplexitySource(src, "switch.go", 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	// 1 (base) + 3 (case 1, 2, 3; default has nil List) = 4
	assert.Equal(t, 4, results[0].Complexity)
}

func TestAnalyseComplexitySource_LogicalOperators_Good(t *testing.T) {
	src := `package example

func complex(a, b, c bool) bool {
	if a && b || c {
		return true
	}
	return false
}
`
	results, err := AnalyseComplexitySource(src, "logical.go", 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	// 1 (base) + 1 (if) + 1 (&&) + 1 (||) = 4
	assert.Equal(t, 4, results[0].Complexity)
}

func TestAnalyseComplexitySource_MethodReceiver_Good(t *testing.T) {
	src := `package example

type Server struct{}

func (s *Server) Handle(req int) string {
	if req > 0 {
		return "ok"
	}
	return "bad"
}
`
	results, err := AnalyseComplexitySource(src, "method.go", 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Server.Handle", results[0].FuncName)
	assert.Equal(t, 2, results[0].Complexity)
}

func TestAnalyseComplexitySource_HighComplexity_Good(t *testing.T) {
	// Build a function with high complexity to test threshold filtering.
	src := `package example

func monster(x, y, z int) int {
	result := 0
	if x > 0 {
		if y > 0 {
			if z > 0 {
				result = 1
			} else if z < -10 {
				result = 2
			}
		} else if y < -5 {
			result = 3
		}
	} else if x < -10 {
		result = 4
	}
	for i := 0; i < x; i++ {
		for j := 0; j < y; j++ {
			if i > j && j > 0 {
				result += i
			} else if i == j || i < 0 {
				result += j
			}
		}
	}
	switch result {
	case 1:
		result++
	case 2:
		result--
	case 3:
		result *= 2
	}
	return result
}
`
	// With threshold 15 — should be above it
	results, err := AnalyseComplexitySource(src, "monster.go", 15)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "monster", results[0].FuncName)
	assert.GreaterOrEqual(t, results[0].Complexity, 15)
}

func TestAnalyseComplexitySource_BelowThreshold_Good(t *testing.T) {
	src := `package example

func simple() int {
	return 42
}
`
	results, err := AnalyseComplexitySource(src, "simple.go", 5)
	require.NoError(t, err)
	assert.Empty(t, results) // Complexity 1, below threshold 5
}

func TestAnalyseComplexitySource_MultipleFuncs_Good(t *testing.T) {
	src := `package example

func low() { }

func medium(x int) {
	if x > 0 {
		if x > 10 {
			_ = x
		}
	}
}

func high(a, b, c, d int) int {
	if a > 0 {
		if b > 0 {
			if c > 0 {
				if d > 0 {
					return 1
				}
			}
		}
	}
	return 0
}
`
	results, err := AnalyseComplexitySource(src, "multi.go", 3)
	require.NoError(t, err)
	// low: 1, medium: 3, high: 5
	assert.Len(t, results, 2) // medium and high
	assert.Equal(t, "medium", results[0].FuncName)
	assert.Equal(t, 3, results[0].Complexity)
	assert.Equal(t, "high", results[1].FuncName)
	assert.Equal(t, 5, results[1].Complexity)
}

func TestAnalyseComplexitySource_SelectStmt_Good(t *testing.T) {
	src := `package example

func selecter(ch1, ch2 chan int) int {
	select {
	case v := <-ch1:
		return v
	case v := <-ch2:
		return v
	}
}
`
	results, err := AnalyseComplexitySource(src, "select.go", 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	// 1 (base) + 1 (select) + 2 (comm clauses) = 4
	assert.Equal(t, 4, results[0].Complexity)
}

func TestAnalyseComplexitySource_TypeSwitch_Good(t *testing.T) {
	src := `package example

func typeSwitch(v interface{}) string {
	switch v.(type) {
	case int:
		return "int"
	case string:
		return "string"
	default:
		return "unknown"
	}
}
`
	results, err := AnalyseComplexitySource(src, "typeswitch.go", 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	// 1 (base) + 1 (type switch) + 2 (case int, case string; default has nil List) = 4
	assert.Equal(t, 4, results[0].Complexity)
}

func TestAnalyseComplexitySource_EmptyBody_Good(t *testing.T) {
	// Interface methods or abstract funcs have nil body
	src := `package example

type Iface interface {
	DoSomething(x int) error
}
`
	results, err := AnalyseComplexitySource(src, "iface.go", 1)
	require.NoError(t, err)
	assert.Empty(t, results) // No FuncDecl in interface
}

func TestAnalyseComplexitySource_ParseError_Bad(t *testing.T) {
	src := `this is not valid go code at all!!!`
	_, err := AnalyseComplexitySource(src, "bad.go", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestAnalyseComplexity_DefaultThreshold_Good(t *testing.T) {
	cfg := DefaultComplexityConfig()
	assert.Equal(t, 15, cfg.Threshold)
	assert.Equal(t, ".", cfg.Path)
}

func TestAnalyseComplexity_ZeroThreshold_Good(t *testing.T) {
	// Zero threshold should default to 15
	src := `package example
func f() { }
`
	results, err := AnalyseComplexitySource(src, "zero.go", 0)
	require.NoError(t, err)
	assert.Empty(t, results) // complexity 1, default threshold 15
}

func TestAnalyseComplexity_Directory_Good(t *testing.T) {
	dir := t.TempDir()

	// Write a Go file with a complex function
	src := `package example

func complex(a, b, c, d, e int) int {
	if a > 0 {
		if b > 0 {
			if c > 0 {
				return 1
			}
		}
	}
	if d > 0 || e > 0 {
		return 2
	}
	return 0
}
`
	err := os.WriteFile(filepath.Join(dir, "example.go"), []byte(src), 0644)
	require.NoError(t, err)

	cfg := ComplexityConfig{Threshold: 3, Path: dir}
	results, err := AnalyseComplexity(cfg)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "complex", results[0].FuncName)
	// 1 (base) + 3 (if x>0, if y>0, if z>0) + 1 (if d>0||e>0) + 1 (||) = 6
	assert.Equal(t, 6, results[0].Complexity)
}

func TestAnalyseComplexity_SingleFile_Good(t *testing.T) {
	dir := t.TempDir()
	src := `package example

func branchy(x int) {
	if x > 0 { }
	if x > 1 { }
	if x > 2 { }
}
`
	path := filepath.Join(dir, "single.go")
	err := os.WriteFile(path, []byte(src), 0644)
	require.NoError(t, err)

	cfg := ComplexityConfig{Threshold: 1, Path: path}
	results, err := AnalyseComplexity(cfg)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 4, results[0].Complexity) // 1 + 3 ifs
}

func TestAnalyseComplexity_SkipsTestFiles_Good(t *testing.T) {
	dir := t.TempDir()

	// Production file — should be analysed
	prod := `package example
func prodFunc(x int) {
	if x > 0 { }
	if x > 1 { }
}
`
	err := os.WriteFile(filepath.Join(dir, "prod.go"), []byte(prod), 0644)
	require.NoError(t, err)

	// Test file — should be skipped
	test := `package example
func TestHelper(x int) {
	if x > 0 { }
	if x > 1 { }
	if x > 2 { }
	if x > 3 { }
}
`
	err = os.WriteFile(filepath.Join(dir, "prod_test.go"), []byte(test), 0644)
	require.NoError(t, err)

	cfg := ComplexityConfig{Threshold: 1, Path: dir}
	results, err := AnalyseComplexity(cfg)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "prodFunc", results[0].FuncName)
}

func TestAnalyseComplexity_SkipsVendor_Good(t *testing.T) {
	dir := t.TempDir()

	// Create vendor dir with a Go file
	vendorDir := filepath.Join(dir, "vendor")
	err := os.MkdirAll(vendorDir, 0755)
	require.NoError(t, err)

	vendorSrc := `package lib
func VendorFunc(x int) {
	if x > 0 { }
	if x > 1 { }
}
`
	err = os.WriteFile(filepath.Join(vendorDir, "lib.go"), []byte(vendorSrc), 0644)
	require.NoError(t, err)

	cfg := ComplexityConfig{Threshold: 1, Path: dir}
	results, err := AnalyseComplexity(cfg)
	require.NoError(t, err)
	assert.Empty(t, results) // vendor dir should be skipped
}

func TestAnalyseComplexity_NonexistentPath_Bad(t *testing.T) {
	cfg := ComplexityConfig{Threshold: 1, Path: "/nonexistent/path/xyz"}
	_, err := AnalyseComplexity(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stat")
}

func TestAnalyseComplexitySource_NestedLogicalOps_Good(t *testing.T) {
	src := `package example

func nested(a, b, c, d bool) bool {
	return (a && b) || (c && d)
}
`
	results, err := AnalyseComplexitySource(src, "nested.go", 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	// 1 (base) + 2 (&&) + 1 (||) = 4
	assert.Equal(t, 4, results[0].Complexity)
}

// LEK-1 | lthn.ai | EUPL-1.2
