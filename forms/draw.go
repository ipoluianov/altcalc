package forms

import (
	"image/color"
	"math"

	"github.com/ipoluianov/nui/ui"
)

// fillRoundRect fills the rectangle with rounded corners of the radius.
// Canvas.FillRoundedRect makes a mask of the whole window for each call,
// too slow for the dozens of keys repainted as the mouse moves over them:
// here the straight parts are rectangles and only the pixels of the
// corners are computed, with antialiasing.
func fillRoundRect(cnv *ui.Canvas, x, y, w, h, radius int, col color.RGBA) {
	radius = min(radius, w/2, h/2)
	if radius <= 0 {
		cnv.FillRect(x, y, w, h, col)
		return
	}
	cnv.FillRect(x+radius, y, w-2*radius, h, col)
	cnv.FillRect(x, y+radius, radius, h-2*radius, col)
	cnv.FillRect(x+w-radius, y+radius, radius, h-2*radius, col)

	// The corners in the pixels of the image, which has more of them than
	// logical ones on a HiDPI screen
	s := cnv.Scale()
	ox, oy := float64(cnv.TranslatedX()+x)*s, float64(cnv.TranslatedY()+y)*s
	pw, ph, pr := float64(w)*s, float64(h)*s, float64(radius)*s
	n := int(math.Ceil(pr))
	corners := [4][2]float64{
		{ox + pr, oy + pr}, {ox + pw - pr, oy + pr},
		{ox + pr, oy + ph - pr}, {ox + pw - pr, oy + ph - pr},
	}
	for i, c := range corners {
		// The square of the corner: left or right of the center, above or below
		x0, y0 := int(math.Floor(c[0]))-n, int(math.Floor(c[1]))-n
		if i%2 == 1 {
			x0 = int(math.Floor(c[0]))
		}
		if i >= 2 {
			y0 = int(math.Floor(c[1]))
		}
		for py := y0; py < y0+n; py++ {
			for px := x0; px < x0+n; px++ {
				d := math.Hypot(float64(px)+0.5-c[0], float64(py)+0.5-c[1])
				cover := math.Max(0, math.Min(1, pr-d+0.5))
				if cover <= 0 {
					continue
				}
				pc := col
				pc.A = uint8(float64(col.A)*cover + 0.5)
				cnv.MixPixel(px, py, pc)
			}
		}
	}
}

// strokeRoundRect draws the border of the rounded rectangle, one logical
// pixel wide, over the fill of inside
func strokeRoundRect(cnv *ui.Canvas, x, y, w, h, radius int, border, inside color.RGBA) {
	fillRoundRect(cnv, x, y, w, h, radius, border)
	fillRoundRect(cnv, x+1, y+1, w-2, h-2, radius-1, inside)
}
