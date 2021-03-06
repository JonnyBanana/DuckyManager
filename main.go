package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/nsf/termbox-go"
)

func init() {

	// Load lang
	if err := parseLang(); err != nil {
		fmt.Println(err.Error())
		os.Exit(errExitCode)
	}

	// Load config
	cf, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println(translate.ErrOpeningConfig + ": " + err.Error())
		os.Exit(errExitCode)
	}

	err = json.Unmarshal(cf, &config)
	if err != nil {
		fmt.Println(errStr + translate.ErrParsingConfig + ": " + err.Error())
		os.Exit(errExitCode)
	}

	// Init log
	f, err := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(errStr + translate.ErrParsingConfig + ": " + err.Error())
		os.Exit(errExitCode)
	}

	l = log.New(f, "", log.Ltime)

	l.SetOutput(f)

}

func main() {

	position, positionUpper := 0, 0

	// Load scripts
	l.Println("+------------------------------+")
	l.Println("Loading local scripts")

	scripts, valid, deleted, modified, newOnes, err := CheckLocal(config.LocalDBFile, config.ScriptsPath)

	if err != nil {
		fmt.Println(errStr + translate.ErrCheckingLocal + " : " + err.Error())
		l.Println(errStr + translate.ErrCheckingLocal + " : " + err.Error())
		os.Exit(errExitCode)
	}
	defer Save(config.LocalDBFile, scripts)

	l.Println("[" + strconv.Itoa(int(valid)) + "] " + translate.Valid + " , " +
		"[" + strconv.Itoa(int(deleted)) + "] " + translate.Deleted + " , " +
		"[" + strconv.Itoa(int(modified)) + "] " + translate.Modified + " , " +
		"[" + strconv.Itoa(int(newOnes)) + "] " + translate.NewScripts)

	// GUI
	err = termbox.Init()
	if err != nil {
		fmt.Println(errStr + translate.ErrTermboxInit + ": " + err.Error())
		l.Println(errStr + translate.ErrTermboxInit + ": " + err.Error())
		os.Exit(errExitCode)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc)
	termbox.SetOutputMode(termbox.Output256)
	l.Println(okStr + translate.TermInputMode + ": Input ESC || " + translate.TermOutputMode + ": Output256")

	mainLoop(positionUpper, position, scripts)

}

func mainLoop(positionUpper, position int, scripts []Script) {
	saveOn := false
	var tmpPosition, tmpPositionUpper int
	var tmpSave []Script
	for {
		l.Println(translate.MainLoop)

		if err := redrawMain(positionUpper, position, scripts); err != nil {
			l.Println(errStr + translate.ErrDrawing + ": " + err.Error())
			os.Exit(errExitCode)
		}

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				if saveOn {
					saveOn = false
					scripts = tmpSave
					position = tmpPosition
					positionUpper = tmpPositionUpper
				} else {
					return
				}

			case termbox.KeyArrowDown:

				if position+1 < len(scripts) {
					position++

					_, h := termbox.Size()

					if position-positionUpper > h-1 {
						positionUpper++
					}
				}

			case termbox.KeyArrowUp:
				if position-1 >= 0 {
					position--

					if position < positionUpper {
						positionUpper--
					}
				}

			case termbox.KeyHome:
				position = 0
				positionUpper = 0

			case termbox.KeyEnd:
				position = len(scripts) - 1
				_, h := termbox.Size()
				positionUpper = len(scripts) - h
				if positionUpper < 0 {
					positionUpper = 0
				}

			default:
				if ev.Ch != 0 {
					switch ev.Ch {
					case 's', 'S':

						res := search(scripts)
						if len(res) != 0 {
							saveOn = true

							tmpSave = scripts
							tmpPosition = position
							tmpPositionUpper = positionUpper

							scripts = res
							position = 0
							positionUpper = 0
						} else {
							showErrorMsg(translate.NoMatch)
						}

					case 'e', 'E':
						edit(position, scripts)
					}
				}
			}
		case termbox.EventError:
			l.Println(errStr + translate.ErrEvent + ": " + ev.Err.Error())
			os.Exit(errExitCode)
		}

	}
}
