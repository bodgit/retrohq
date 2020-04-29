package marquee

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"io"
	"strings"
)

const (
	// Extension is the file extension used
	Extension = ".mrq"
	// BoxWidth is the width of the box artwork image
	BoxWidth = 88
	// BoxHeight is the height of the box artwork image
	BoxHeight = 124
	// ScreenshotWidth is the width of the screenshot image
	ScreenshotWidth = 88
	// ScreenshotHeight is the height of the screenshot image
	ScreenshotHeight = 56
)

const (
	titleLength     = 48
	developerLength = 24
	publisherLength = 24
	yearLength      = 4
)

var (
	errInvalid           = errors.New("marquee: invalid data")
	errTooMuch           = errors.New("marquee: too much data")
	errInvalidBox        = errors.New("marquee: invalid box image")
	errInvalidScreenshot = errors.New("marquee: invalid screenshot image")
)

var signature = [4]byte{'M', 'Q', 0, 0}

type marqueeHeader struct {
	Signature [4]byte
}

func (h marqueeHeader) isValid() bool {
	return bytes.Equal(h.Signature[:], signature[:])
}

func newMarqueeHeader() marqueeHeader {
	return marqueeHeader{
		Signature: signature,
	}
}

type marqueeFields struct {
	marqueeHeader
	Title     [titleLength]byte
	Developer [developerLength]byte
	Publisher [publisherLength]byte
	Year      [yearLength]byte
	_         [10]byte // FIXME
}

// Marquee represents a .mrq file
type Marquee struct {
	marqueeFields
	Title      string
	Developer  string
	Publisher  string
	Year       string
	Box        *image.RGBA
	Screenshot *image.RGBA
}

// NewMarquee returns an empty Marquee with the two images initialised to
// the correct dimensions
func NewMarquee() (*Marquee, error) {
	return &Marquee{
		Box:        image.NewRGBA(image.Rect(0, 0, BoxWidth, BoxHeight)),
		Screenshot: image.NewRGBA(image.Rect(0, 0, ScreenshotWidth, ScreenshotHeight)),
	}, nil
}

func readImage(r io.Reader, m *image.RGBA) error {
	b := m.Bounds()
	for i := 0; i < b.Dx()*b.Dy()*4; i += 4 {
		var p uint16
		if err := binary.Read(r, binary.BigEndian, &p); err != nil {
			return err
		}
		// RRRRRBBBBBGGGGGG
		// fmt.Printf("%#016b\n", p)
		m.Pix[i+0] = uint8(p & 0xf800 >> 8)
		m.Pix[i+1] = uint8(p & 0x003f << 2)
		m.Pix[i+2] = uint8(p & 0x07c0 >> 3)
		m.Pix[i+3] = 255
	}

	return nil
}

// UnmarshalBinary decodes the marquee from binary form
func (m *Marquee) UnmarshalBinary(b []byte) error {
	r := bytes.NewReader(b)
	if err := binary.Read(r, binary.BigEndian, &m.marqueeFields); err != nil {
		return err
	}

	if !m.isValid() {
		return errInvalid
	}

	m.Title = strings.TrimRight(string(m.marqueeFields.Title[:]), "\x00")
	m.Developer = strings.TrimRight(string(m.marqueeFields.Developer[:]), "\x00")
	m.Publisher = strings.TrimRight(string(m.marqueeFields.Publisher[:]), "\x00")
	m.Year = strings.TrimRight(string(m.marqueeFields.Year[:]), "\x00")

	m.Box = image.NewRGBA(image.Rect(0, 0, BoxWidth, BoxHeight))
	if err := readImage(r, m.Box); err != nil {
		return err
	}

	m.Screenshot = image.NewRGBA(image.Rect(0, 0, ScreenshotWidth, ScreenshotHeight))
	if err := readImage(r, m.Screenshot); err != nil {
		return err
	}

	// There should be no more data to read
	var p [1]byte
	if n, _ := r.Read(p[:]); n != 0 {
		return errTooMuch
	}

	return nil
}

func writeImage(w io.Writer, m *image.RGBA) error {
	b := m.Bounds()
	for i := 0; i < b.Dx()*b.Dy()*4; i += 4 {
		// RRRRRBBBBBGGGGGG
		p := uint16(m.Pix[i+0]&0xf8)<<8 | uint16(m.Pix[i+1]&0xfc)>>2 | uint16(m.Pix[i+2]&0xf8)<<3
		// fmt.Printf("%#016b\n", p)
		_ = binary.Write(w, binary.BigEndian, p)
	}

	return nil
}

// MarshalBinary encodes the marquee into binary form and returns the
// result
func (m *Marquee) MarshalBinary() ([]byte, error) {
	b := m.Box.Bounds()
	if b.Dx() != BoxWidth || b.Dy() != BoxHeight {
		return nil, errInvalidBox
	}

	b = m.Screenshot.Bounds()
	if b.Dx() != ScreenshotWidth || b.Dy() != ScreenshotHeight {
		return nil, errInvalidScreenshot
	}

	m.marqueeFields = marqueeFields{}
	m.marqueeFields.marqueeHeader = newMarqueeHeader()

	copy(m.marqueeFields.Title[:], m.Title)
	copy(m.marqueeFields.Developer[:], m.Developer)
	copy(m.marqueeFields.Publisher[:], m.Publisher)
	copy(m.marqueeFields.Year[:], m.Year)

	w := new(bytes.Buffer)
	// Writes to bytes.Buffer never error
	_ = binary.Write(w, binary.BigEndian, &m.marqueeFields)
	_ = writeImage(w, m.Box)
	_ = writeImage(w, m.Screenshot)

	return w.Bytes(), nil
}
