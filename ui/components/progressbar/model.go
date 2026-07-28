// Package progressbar wraps the bubbles progress component as a small,
// reusable grove component: a fixed-width bar in 10% increments followed by an
// exact, stable-width percentage. It renders statically from a done/total pair
// and owns no animation state.
package progressbar

import (
	"fmt"
	"math"

	"github.com/charmbracelet/bubbles/progress"
)

// cells is the bar width in characters. Ten cells means each filled cell is one
// 10% increment of progress.
const cells = 10

// fullColor is the accent used for filled cells, matching the loading spinner.
const fullColor = "117"

type Model struct {
	bar progress.Model
}

func New() Model {
	return Model{bar: progress.New(
		progress.WithWidth(cells),
		progress.WithFillCharacters('█', '░'),
		progress.WithoutPercentage(),
		progress.WithSolidFill(fullColor),
	)}
}

// View renders the 10-cell bar for done/total, a single space, then the exact
// percentage in a stable 4-character, left-aligned field ("0%  ", "47% ",
// "100%"), so the rendered width never changes as progress advances. The bar
// fills in 10% increments (rounded to the nearest cell) while the percentage
// stays exact. A zero total renders as 0%.
func (m Model) View(done, total int) string {
	fraction := 0.0
	if total > 0 {
		fraction = float64(done) / float64(total)
	}
	percent := int(math.Round(fraction * 100))
	label := fmt.Sprintf("%-4s", fmt.Sprintf("%d%%", percent))
	return m.bar.ViewAs(fraction) + " " + label
}
