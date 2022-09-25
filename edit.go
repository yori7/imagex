package imagex

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"strings"
)

func Save(img image.Image, fname string) error {
	var (
		fsplit  []string = strings.Split(fname, ".")
		outtail string   = fsplit[len(fsplit)-1]
		err     error
	)
	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()
	if outtail == "png" {
		err = png.Encode(file, img)
	} else if outtail == "jpg" {
		err = jpeg.Encode(file, img, &jpeg.Options{Quality: 100})
	}
	if err != nil {
		return err
	} else {
		return nil
	}
}

func Fill(img *image.RGBA, c color.Color) {
	var (
		width  int = img.Bounds().Dx()
		height int = img.Bounds().Dy()
	)
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			img.Set(i, j, c)
		}
	}
}

func resize_base(img image.Image, rate float64, bg color.Color) (*image.RGBA, int, int) {
	var (
		width  int        = int(float64(img.Bounds().Dx()) * rate)
		height int        = int(float64(img.Bounds().Dy()) * rate)
		outimg image.RGBA = *image.NewRGBA(image.Rect(0, 0, width, height))
	)
	Fill(&outimg, bg)

	return &outimg, width, height
}

func NEAREST_NEIBOR(img image.Image, rate float64, bg color.Color) *image.RGBA {
	outimg, width, height := resize_base(img, rate, bg)

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			var c color.Color = img.At(int(float64(i)/rate), int(float64(j)/rate))
			outimg.Set(i, j, c)
		}
	}
	return outimg
}

func BI_LINEAR(img image.Image, rate float64, bg color.Color) *image.RGBA {
	outimg, _, _ := resize_base(img, rate, bg)

	return outimg
}

func PIXEL_MIXING(img image.Image, rate float64, bg color.Color) *image.RGBA {
	outimg, width, height := resize_base(img, rate, bg)

	for w := 0; w < width; w++ {
		for h := 0; h < height; h++ {
			var (
				x0         float64 = float64(w) / rate
				x1         float64 = float64(w+1) / rate
				y0         float64 = float64(h) / rate
				y1         float64 = float64(h+1) / rate
				x0i        int     = int(x0)
				x1i        int     = int(math.Ceil(x1)) - 1
				y0i        int     = int(y0)
				y1i        int     = int(math.Ceil(y1)) - 1
				r, g, b, a float64
			)
			for x := x0i; x <= x1i; x++ {
				for y := y0i; y <= y1i; y++ {
					var (
						xw float64 = 1
						yw float64 = 1
					)
					if x == x0i {
						xw -= (x0 - float64(x0i))
					}
					if x == x1i {
						xw -= (1 + float64(x1i) - x1)
					}

					if y == y0i {
						yw -= (y0 - float64(y0i))
					}
					if y == y1i {
						yw -= (1 + float64(y1i) - y1)
					}

					var weight float64 = xw * yw
					var ri, gi, bi, ai = img.At(x, y).RGBA()
					r += float64(ri >> 8) * weight
					g += float64(gi >> 8) * weight
					b += float64(bi >> 8) * weight
					a += float64(ai >> 8) * weight
				}
			}
			var c color.RGBA = color.RGBA{
				uint8(r * (rate * rate)),
				uint8(g * (rate * rate)),
				uint8(b * (rate * rate)),
				uint8(a * (rate * rate)),
			}
			outimg.SetRGBA(w, h, c)
		}
	}

	return outimg
}

func Resize(img image.Image, rate float64, bg color.Color, fn func(image.Image, float64, color.Color) *image.RGBA) image.RGBA {
	var outimg *image.RGBA = fn(img, rate, bg)

	return *outimg
}

func Collage(fnames []string, bg string, margin float64) (*image.RGBA, error) {
	// 画像の読み込み
	var (
		figs    []image.Image
		aspect  float64
		mean_px int
		num     float64
	)
	for _, s := range fnames {
		file, err := os.Open(s)
		defer file.Close()
		if err != nil {
			return nil, err
		}
		img, _, err := image.Decode(file)
		if err != nil {
			return nil, err
		}
		var (
			w int = img.Bounds().Dx()
			h int = img.Bounds().Dy()
		)
		aspect += float64(w) / float64(h)
		mean_px += w * h
		figs = append(figs, img)
		num = float64(len(figs))
	}
	aspect /= num
	mean_px = int(mean_px / int(num))

	// 最終出力サイズの決定
	var (
		row    int         = int(math.Ceil(math.Sqrt(num)))
		col    int         = int(math.Ceil(num / float64(row)))
		height int         = int(math.Sqrt(float64(mean_px)/aspect) * margin)
		width  int         = int(math.Sqrt(float64(mean_px)*aspect) * margin)
		outimg *image.RGBA = image.NewRGBA(image.Rect(0, 0, width*row, height*col))
		c      color.Color = CMap(bg)
	)
	Fill(outimg, c)

	// 画像のリサイズと書き込み
	for n, img_n := range figs {
		var (
			i    int = n % row
			j    int = n / row
			rate float64
		)
		if aspect <= float64(img_n.Bounds().Dx())/float64(img_n.Bounds().Dy()) {
			rate = (float64(width) / margin) / float64(img_n.Bounds().Dx())
		} else {
			rate = (float64(height) / margin) / float64(img_n.Bounds().Dy())
		}
		var (
			img_resized image.RGBA      = Resize(img_n, rate, c, PIXEL_MIXING)
			rect_n      image.Rectangle = img_resized.Bounds()
			hspace      int             = (width - rect_n.Dx()) / 2
			vspace      int             = (height - rect_n.Dy()) / 2
			rect_np     image.Rectangle = image.Rect(i*width+hspace, j*height+vspace, (i+1)*width-1-hspace, (j+1)*height-1-vspace)
		)
		draw.Draw(outimg, rect_np, &img_resized, rect_n.Min, draw.Over)
	}
	return outimg, nil
}
