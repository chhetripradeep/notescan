package notescan

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
)

type Pixel struct {
	R uint8
	G uint8
	B uint8

	H float64
	S float64
	V float64
}

// pack pixel data into an integer
func Pack(p *Pixel) int {
	return int(p.R)<<16 | int(p.G)<<8 | int(p.B)
}

// restore to original rgb format
func UnPack(v int) (uint8, uint8, uint8) {
	r := uint8((v >> 16) & 0xFF)
	g := uint8((v >> 8) & 0xFF)
	b := uint8(v & 0xFF)
	return r, g, b
}

// generate pixel from color
func NewPixel(c color.Color) *Pixel {
	cc, err := convertColor(c)
	if err != nil {
		return nil
	}
	return NewPixelRGB(cc.R, cc.G, cc.B)
}

// generate pixel from RGB
func NewPixelRGB(r, g, b uint8) *Pixel {
	p := &Pixel{}
	p.R = r
	p.G = g
	p.B = b
	p.H, p.S, p.V = RGB2HSV(p.R, p.G, p.B)
	return p
}

// generate pixel from HSV
func NewPixelHSV(h, s, v float64) *Pixel {
	c := HSV2RGBA(h, s, v)
	return NewPixelRGB(c.R, c.G, c.B)
}

// get HSV space distance
func (p Pixel) DistanceHSV(src *Pixel) (float64, float64, float64) {
	h := math.Abs(src.H - p.H)
	s := math.Abs(src.S - p.S)
	v := math.Abs(src.V - p.V)
	return h, s, v
}

// get RGB space distance
func (p Pixel) DistanceRGB(src *Pixel) float64 {
	all := 0.0
	r := float64(src.R) - float64(p.R)
	g := float64(src.G) - float64(p.G)
	b := float64(src.B) - float64(p.B)
	all += r * r
	all += g * g
	all += b * b
	return all
}

// shift operation
func (p Pixel) Shift(shift uint) *Pixel {
	r := uint8((p.R >> shift) << shift)
	g := uint8((p.G >> shift) << shift)
	b := uint8((p.B >> shift) << shift)
	return NewPixelRGB(r, g, b)
}

// generate color
func (p Pixel) Color() *color.RGBA {
	return UIntRGBA(p.R, p.G, p.B)
}

// generate string
func (p Pixel) String() string {
	rtn := fmt.Sprintf("R[%d]G[%d]B[%d] = H[%f]S[%f]V[%f]", p.R, p.G, p.B, p.H, p.S, p.V)
	return rtn
}

type Pixels []*Pixel

// get the most populous color
func (p Pixels) Most() *Pixel {
	counter := make(map[int]int)
	for _, pix := range p {
		val := Pack(pix)
		counter[val]++
	}

	max := 0
	val := 0
	for key, count := range counter {
		if count > max {
			max = count
			val = key
		}
	}
	return NewPixelRGB(UnPack(val))
}

// get the color that rounds all data
func (p Pixels) Quantize(s int) (Pixels, error) {
	if s >= 8 {
		return nil, fmt.Errorf("Shift can't be over 8")
	}

	shift := uint(s)
	quantize := make([]*Pixel, len(p))
	for idx, pix := range p {
		quantize[idx] = pix.Shift(shift)
	}
	return quantize, nil
}

// get the average color
func (p Pixels) Average() (*Pixel, error) {
	if p == nil {
		return nil, fmt.Errorf("Pixels is nil")
	}

	length := len(p)
	if length == 0 {
		return nil, fmt.Errorf("Pixels length is zero")
	}

	r, g, b := 0.0, 0.0, 0.0
	for _, d := range p {
		r += float64(d.R)
		g += float64(d.G)
		b += float64(d.B)
	}

	avg := 1.0 / float64(length)
	r = r * avg
	g = g * avg
	b = b * avg

	c := FloatRGBA(r, g, b)
	return NewPixel(c), nil
}

// create image
func (p Pixels) ToImage(cols, rows int) (image.Image,error) {
	idx := 0
	img := image.NewRGBA(image.Rect(0, 0, cols, rows))
	for col := 0; col < cols; col++ {
		for row := 0; row < rows; row++ {
			img.Set(col, row, p[idx].Color())
			idx++
		}
	}
	return img,nil
}

// sorting
func (p Pixels) Sort() error {
	sort.Slice(p, func(i, j int) bool {
		pi := Pack(p[i])
		pj := Pack(p[j])
		iRGB := int((pi>>16)&0xFF) +
				int((pi>>8)&0xFF) +
				int(pi&0xFF)
		jRGB := int((pj>>16)&0xFF) +
				int((pj>>8)&0xFF) +
				int(pj&0xFF)
		return iRGB < jRGB
	})
	return nil
}


// create palette for debugging
func (p Pixels) debug(f string) error {
	length := len(p)
	if length > 20 {
		return fmt.Errorf("Length of Pixels greater than 20 isn't supported")
	}

	img,err := p.ToImage(length, 1)
	if err != nil {
		return err
	}
	return OutputPNG(f, img)
}

// output
func (p Pixels) output(f string, cols, rows int) error {
	img,err := p.ToImage(cols, rows)
	if err != nil {
		return err
	}
	return OutputPNG(f, img)
}