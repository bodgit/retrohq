package marquee

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

func ExampleMarqueeUnmarshalBinary() {
	m, err := NewMarquee()
	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadFile(filepath.Join("testdata", "Tempest 2000"+Extension))
	if err != nil {
		panic(err)
	}

	if err := m.UnmarshalBinary(b); err != nil {
		panic(err)
	}

	fmt.Println(m.Title)
	// Output: Tempest 2000
}

func ExampleMarqueeMarshalBinary() {
	m, err := NewMarquee()
	if err != nil {
		panic(err)
	}

	m.Title = "Tempest 2000"
	m.Developer = "Llamasoft"
	m.Publisher = "Atari Corporation"
	m.Year = "1994"

	b, err := m.MarshalBinary()
	if err != nil {
		panic(err)
	}

	// Not including the box art and screenshot
	fmt.Print(hex.Dump(b[:0x74]))
	// Output: 00000000  4d 51 00 00 54 65 6d 70  65 73 74 20 32 30 30 30  |MQ..Tempest 2000|
	// 00000010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// 00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
	// 00000030  00 00 00 00 4c 6c 61 6d  61 73 6f 66 74 00 00 00  |....Llamasoft...|
	// 00000040  00 00 00 00 00 00 00 00  00 00 00 00 41 74 61 72  |............Atar|
	// 00000050  69 20 43 6f 72 70 6f 72  61 74 69 6f 6e 00 00 00  |i Corporation...|
	// 00000060  00 00 00 00 31 39 39 34  00 00 00 00 00 00 00 00  |....1994........|
	// 00000070  00 00 00 00                                       |....|
}
