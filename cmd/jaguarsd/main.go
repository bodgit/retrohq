package main

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/bodgit/retrohq/marquee"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print the version",
	}
}

func readMarquee(file string) (*marquee.Marquee, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	m, err := marquee.NewMarquee()
	if err != nil {
		return nil, err
	}

	if err := m.UnmarshalBinary(b); err != nil {
		return nil, err
	}

	return m, err
}

func readImage(file string) (image.Image, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func parseAddress(s string) (uint32, error) {
	var i uint32
	if n, err := fmt.Sscanf(strings.ToLower(s), "0x%08x", &i); n != 1 || err != nil {
		if n != 1 {
			return 0, errors.New("unable to parse address")
		}
		return 0, err
	}

	return i, nil
}

func writeMarquee(m *marquee.Marquee, file string) error {
	b, err := m.MarshalBinary()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(file, b, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func updateMarquee(m *marquee.Marquee, c *cli.Context) (err error) {
	if c.IsSet("title") {
		m.Title = c.String("title")
	}

	if c.IsSet("developer") {
		m.Developer = c.String("developer")
	}

	if c.IsSet("publisher") {
		m.Publisher = c.String("publisher")
	}

	if c.IsSet("year") {
		m.Year = c.String("year")
	}

	if c.IsSet("eeprom") {
		m.EEPROM = c.Uint("eeprom")
	}

	if c.IsSet("load-address") {
		if m.LoadAddr, err = parseAddress(c.String("load-address")); err != nil {
			return
		}
	}

	if c.IsSet("exec-address") {
		if m.ExecAddr, err = parseAddress(c.String("exec-address")); err != nil {
			return
		}
	}

	if c.IsSet("box") {
		if m.Box, err = readImage(c.String("box")); err != nil {
			return
		}
	}

	if c.IsSet("screenshot") {
		if m.Screenshot, err = readImage(c.String("screenshot")); err != nil {
			return
		}
	}

	return writeMarquee(m, c.Args().First())
}

func main() {
	app := cli.NewApp()

	app.Name = "jaguarsd"
	app.Usage = "RetroHQ Jaguar SD/GD utility"
	app.Version = fmt.Sprintf("%s, commit %s, built at %s", version, commit, date)

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  "title",
			Usage: "set title to `TITLE`",
		},
		&cli.StringFlag{
			Name:  "developer",
			Usage: "set developer to `DEVELOPER`",
		},
		&cli.StringFlag{
			Name:  "publisher",
			Usage: "set publisher to `PUBLISHER`",
		},
		&cli.StringFlag{
			Name:  "year",
			Usage: "set year to `YEAR`",
		},
		&cli.UintFlag{
			Name:  "eeprom",
			Usage: "set `EEPROM`",
		},
		&cli.StringFlag{
			Name:  "load-address",
			Usage: "set load address to `ADDRESS`",
		},
		&cli.StringFlag{
			Name:  "exec-address",
			Usage: "set exec address to `ADDRESS`",
		},
		&cli.StringFlag{
			Name:  "box",
			Usage: "set box art image to `FILE`",
		},
		&cli.StringFlag{
			Name:  "screenshot",
			Usage: "set screenshot image to `FILE`",
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:        "marquee",
			Usage:       "Manage " + marquee.Extension + " Marquee files",
			Description: "Manage " + marquee.Extension + " Marquee files.",
			Subcommands: []*cli.Command{
				{
					Name:        "create",
					Usage:       "Create a new " + marquee.Extension + " file",
					Description: "Create a new " + marquee.Extension + " file.",
					ArgsUsage:   "FILE",
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelpAndExit(c, c.Command.Name, 1)
						}

						m, err := marquee.NewMarquee()
						if err != nil {
							return cli.NewExitError(err, 1)
						}

						if err := updateMarquee(m, c); err != nil {
							return cli.NewExitError(err, 1)
						}

						return nil
					},
					Flags: flags,
				},
				{
					Name:        "edit",
					Usage:       "Edit an existing " + marquee.Extension + " file",
					Description: "Edit an existing " + marquee.Extension + " file.",
					ArgsUsage:   "FILE",
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelpAndExit(c, c.Command.Name, 1)
						}

						m, err := readMarquee(c.Args().First())
						if err != nil {
							return cli.NewExitError(err, 1)
						}

						if err := updateMarquee(m, c); err != nil {
							return cli.NewExitError(err, 1)
						}

						return nil
					},
					Flags: flags,
				},
				{
					Name:        "info",
					Usage:       "Show an existing " + marquee.Extension + " file",
					Description: "Show an existing " + marquee.Extension + " file.",
					ArgsUsage:   "FILE",
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelpAndExit(c, c.Command.Name, 1)
						}

						m, err := readMarquee(c.Args().First())
						if err != nil {
							return cli.NewExitError(err, 1)
						}

						table := tablewriter.NewWriter(os.Stdout)
						table.SetBorder(false)
						table.SetAutoWrapText(false)
						table.SetAlignment(tablewriter.ALIGN_LEFT)
						table.SetCenterSeparator("")
						table.SetColumnSeparator("")
						table.SetRowSeparator("")
						table.SetTablePadding(" ")
						table.SetNoWhiteSpace(true)

						table.Append([]string{"Title:", m.Title})
						table.Append([]string{"Developer:", m.Developer})
						table.Append([]string{"Publisher:", m.Publisher})
						table.Append([]string{"Year:", m.Year})

						eepromToString := map[uint]string{
							marquee.EEPROM128:        "128 bytes",
							marquee.EEPROM256or512:   "256/512 bytes",
							marquee.EEPROM1024or2048: "1024/2048 bytes",
							marquee.MemoryTrack:      "Memory Track",
						}

						if e, ok := eepromToString[m.EEPROM]; ok {
							table.Append([]string{"EEPROM:", e})
						}

						if m.LoadAddr > 0 || m.ExecAddr > 0 {
							table.Append([]string{"Load Address:", fmt.Sprintf("0x%08x", m.LoadAddr)})
							table.Append([]string{"Exec Address:", fmt.Sprintf("0x%08x", m.ExecAddr)})
						}

						table.Render()

						return nil
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
