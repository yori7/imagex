package imagex

import (
	"errors"
	"image"
	"strconv"
	"strings"
)

type Aspect struct {
	w, h int
}

func NewAspect(s string) (Aspect, error) {
	var (
		wh  []string = strings.Split(s, ":")
		err error    = nil
	)
	if len(wh) != 2 {
		err = errors.New("Aspect string must be described as width:height")
	}
	var (
		w, err2 = strconv.Atoi(wh[0])
		h, err3 = strconv.Atoi(wh[1])
	)
	if err2 != nil || err3 != nil {
		err = errors.New("Aspect string must be described as width:height")
	}
	var aspect Aspect = Aspect{w, h}

	return aspect, err
}

func (asp Aspect) ToRect(rate float64) image.Rectangle {
	return image.Rect(0, 0, int(float64(asp.w)*rate), int(float64(asp.h)*rate))
}
