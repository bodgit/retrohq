package marquee

import (
	"bytes"
	"image"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalBinary(t *testing.T) {
	tables := []struct {
		got func() (*Marquee, error)
		err error
	}{
		{
			NewMarquee,
			nil,
		},
		{
			func() (*Marquee, error) {
				return &Marquee{
					Box:        image.NewRGBA(image.Rect(0, 0, BoxWidth+1, BoxHeight-1)),
					Screenshot: image.NewRGBA(image.Rect(0, 0, ScreenshotWidth, ScreenshotHeight)),
				}, nil
			},
			errInvalidBox,
		},
		{
			func() (*Marquee, error) {
				return &Marquee{
					Box:        image.NewRGBA(image.Rect(0, 0, BoxWidth, BoxHeight)),
					Screenshot: image.NewRGBA(image.Rect(0, 0, ScreenshotWidth+1, ScreenshotHeight-1)),
				}, nil
			},
			errInvalidScreenshot,
		},
	}

	for _, table := range tables {
		m, err := table.got()
		assert.Nil(t, err)

		m.Title = "Tempest 2000"
		m.Developer = "Llamasoft"
		m.Publisher = "Atari Corporation"
		m.Year = "1994"

		_, err = m.MarshalBinary()
		assert.Equal(t, table.err, err)
	}
}

func TestUnmarshalBinary(t *testing.T) {
	tables := []struct {
		got []byte
		err error
	}{
		{
			bytes.Join([][]byte{signature[:], bytes.Repeat([]byte{0}, 31790)}, nil),
			nil,
		},
		// Not enough bytes to read the fields
		{
			signature[:],
			io.ErrUnexpectedEOF,
		},
		// Invalid signature at the beginning of the file
		{
			bytes.Repeat([]byte{0}, 31794),
			errInvalid,
		},
		// Not enough bytes to read the box art image
		{
			bytes.Join([][]byte{signature[:], bytes.Repeat([]byte{0}, 110)}, nil),
			io.EOF,
		},
		// Not enough bytes to read the screenshot image
		{
			bytes.Join([][]byte{signature[:], bytes.Repeat([]byte{0}, 110+BoxWidth*BoxHeight*2)}, nil),
			io.EOF,
		},
		// Too many bytes
		{
			bytes.Join([][]byte{signature[:], bytes.Repeat([]byte{0}, 31791)}, nil),
			errTooMuch,
		},
	}

	for _, table := range tables {
		m, err := NewMarquee()
		assert.Nil(t, err)

		err = m.UnmarshalBinary(table.got)
		assert.Equal(t, table.err, err)
	}
}
