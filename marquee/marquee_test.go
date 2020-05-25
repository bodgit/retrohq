package marquee

import (
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalBinary(t *testing.T) {
	tables := map[string]struct {
		got  func() (*Marquee, error)
		err  error
		want string
	}{
		"good": {
			func() (*Marquee, error) {
				m, err := NewMarquee()
				if err != nil {
					return nil, err
				}

				m.Title = "Tempest 2000"
				m.Developer = "Llamasoft"
				m.Publisher = "Atari Corporation"
				m.Year = "1994"

				box, err := os.Open(filepath.Join("testdata", "box.png"))
				if err != nil {
					return nil, err
				}
				defer box.Close()

				m.Box, err = png.Decode(box)
				if err != nil {
					return nil, err
				}

				screenshot, err := os.Open(filepath.Join("testdata", "screenshot.png"))
				if err != nil {
					return nil, err
				}
				defer screenshot.Close()

				m.Screenshot, err = png.Decode(screenshot)
				if err != nil {
					return nil, err
				}

				return m, nil
			},
			nil,
			"Tempest 2000.mrq",
		},
		"bad box image": {
			func() (*Marquee, error) {
				return &Marquee{
					Box:        image.NewRGBA(image.Rect(0, 0, BoxWidth+1, BoxHeight-1)),
					Screenshot: image.NewRGBA(image.Rect(0, 0, ScreenshotWidth, ScreenshotHeight)),
				}, nil
			},
			errInvalidBox,
			"",
		},
		"box format conversion": {
			func() (*Marquee, error) {
				return &Marquee{
					Box:        image.NewYCbCr(image.Rect(0, 0, BoxWidth, BoxHeight), image.YCbCrSubsampleRatio444),
					Screenshot: image.NewRGBA(image.Rect(0, 0, ScreenshotWidth, ScreenshotHeight)),
				}, nil
			},
			nil,
			"",
		},
		"bad screenshot image": {
			func() (*Marquee, error) {
				return &Marquee{
					Box:        image.NewRGBA(image.Rect(0, 0, BoxWidth, BoxHeight)),
					Screenshot: image.NewRGBA(image.Rect(0, 0, ScreenshotWidth+1, ScreenshotHeight-1)),
				}, nil
			},
			errInvalidScreenshot,
			"",
		},
		"screenshot format conversion": {
			func() (*Marquee, error) {
				return &Marquee{
					Box:        image.NewRGBA(image.Rect(0, 0, BoxWidth, BoxHeight)),
					Screenshot: image.NewYCbCr(image.Rect(0, 0, ScreenshotWidth, ScreenshotHeight), image.YCbCrSubsampleRatio444),
				}, nil
			},
			nil,
			"",
		},
	}

	for name, table := range tables {
		t.Run(name, func(t *testing.T) {
			m, err := table.got()
			if err != nil {
				t.Fatal(err)
			}

			b, err := m.MarshalBinary()
			assert.Equal(t, table.err, err)

			if err == nil && table.want != "" {
				want, err := ioutil.ReadFile(filepath.Join("testdata", table.want))
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, want, b)
			}
		})
	}
}

func TestUnmarshalBinary(t *testing.T) {
	tables := map[string]struct {
		file       string
		err        error
		title      string
		developer  string
		publisher  string
		year       string
		loadAddr   uint32
		execAddr   uint32
		box        string
		screenshot string
	}{
		"good": {
			file:       "Tempest 2000.mrq",
			err:        nil,
			title:      "Tempest 2000",
			developer:  "Llamasoft",
			publisher:  "Atari Corporation",
			year:       "1994",
			box:        "box.png",
			screenshot: "screenshot.png",
		},
		"short fields": {
			file: "err0.mrq",
			err:  io.ErrUnexpectedEOF,
		},
		"invalid signature": {
			file: "err1.mrq",
			err:  errInvalid,
		},
		"box art failure": {
			file: "err2.mrq",
			err:  io.EOF,
		},
		"screenshot failure": {
			file: "err3.mrq",
			err:  io.EOF,
		},
		"too big": {
			file: "err4.mrq",
			err:  errTooMuch,
		},
	}

	for name, table := range tables {
		t.Run(name, func(t *testing.T) {
			b, err := ioutil.ReadFile(filepath.Join("testdata", table.file))
			if err != nil {
				t.Fatal(err)
			}

			m, err := NewMarquee()
			assert.Nil(t, err)

			err = m.UnmarshalBinary(b)
			assert.Equal(t, table.err, err)

			if err == nil {
				assert.Equal(t, table.title, m.Title)
				assert.Equal(t, table.developer, m.Developer)
				assert.Equal(t, table.publisher, m.Publisher)
				assert.Equal(t, table.year, m.Year)
				assert.Equal(t, table.loadAddr, m.LoadAddr)
				assert.Equal(t, table.execAddr, m.ExecAddr)

				box, err := os.Open(filepath.Join("testdata", table.box))
				if err != nil {
					t.Fatal(err)
				}
				defer box.Close()

				rgba, err := png.Decode(box)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, rgba, m.Box)

				screenshot, err := os.Open(filepath.Join("testdata", table.screenshot))
				if err != nil {
					t.Fatal(err)
				}
				defer screenshot.Close()

				rgba, err = png.Decode(screenshot)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, rgba, m.Screenshot)
			}
		})
	}
}
