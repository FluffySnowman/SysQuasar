package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/jroimartin/gocui"
)

type Command struct {
	Name string
	Cmd  string
}

// the old ones 

// var commands = []Command{
// 	{"List files", "ls"},
// 	{"pwd", "pwd"},
// 	{"Show Date and Time", "date"},
// 	{"Neofetch", "neofetch"},
// 	{"CPUs", "nproc"},
// }

var commandGroups = map[string][]Command{
    "general": {
        {"List files", "ls"},
        {"pwd", "pwd"},
        {"Show Date and Time", "date"},
        {"Neofetch", "neofetch"},
        {"CPUs", "nproc"},
    },
    "git": {
        {"Git Status", "git status"},
        {"Git Add", "git add ."},
        {"Git Push", "git push"},
        {"Git Pull", "git pull"},
    },
    "logs": {
        {"Apache Access (last 10)", "sudo tail -n 10 /var/log/apache2/access.log"},
        {"Apache Error (last 10)", "sudo tail -n 10 /var/log/apache2/error.log"},
    },
    "apache": {
        {"Start Apache", "sudo systemctl start apache2"},
        {"Stop Apache", "sudo systemctl stop apache2"},
    },
    "docker": {
        {"Start Docker", "sudo systemctl start docker.service docker.socket"},
        {"Stop Docker", "sudo systemctl stop docker.service docker.socket"},
    },
}

var commandGroupNames = []string{"general", "git", "logs", "apache", "docker"} 

var selectedGroup = "general" // default which is glitchy for some reasont

func main() {
		g, err := gocui.NewGui(gocui.OutputNormal)
    if err != nil {
        log.Panicln(err)
    }
    defer g.Close()

    g.Highlight = true
    g.SelFgColor = gocui.ColorGreen
    g.SelBgColor = gocui.ColorBlack

    g.SetManagerFunc(layout)


	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("right", gocui.KeyEnter, gocui.ModNone, executeCommand); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("right", 'c', gocui.ModNone, clearMiddlePane); err != nil {
			log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'h', gocui.ModNone, switchToView("left")); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'l', gocui.ModNone, switchToView("right")); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'j', gocui.ModNone, moveCursorDown); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'k', gocui.ModNone, moveCursorUp); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("left", gocui.KeyEnter, gocui.ModNone, switchToView("right")); err != nil {
			log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {

		maxX, maxY := g.Size()
    if v, err := g.SetView("left", 0, 0, maxX/4, maxY-1); err != nil {
        if err != gocui.ErrUnknownView {
            return err
        }
        v.Title = "Left"
        v.Wrap = true
        v.Editable = false
        v.Highlight = true
        v.SelBgColor = gocui.ColorGreen
        v.SelFgColor = gocui.ColorBlack


		// new new one 

    for _, group := range commandGroupNames { 
        fmt.Fprintln(v, group)
    }

		// old one

    // for _, command := range commands {
    //     fmt.Fprintln(v, command.Name)
    // }

    if _, err := g.SetCurrentView("left"); err != nil {
        return err
    }
}

if v, err := g.SetView("middle", maxX/4+1, 0, 3*maxX/4, maxY-1); err != nil {
    if err != gocui.ErrUnknownView {
        return err
    }
    v.Title = "Middle"
    v.Wrap = true
    v.Autoscroll = true
    v.Highlight = true
    v.SelBgColor = gocui.ColorGreen
    v.SelFgColor = gocui.ColorBlack
}

    if v, err := g.SetView("right", 3*maxX/4+1, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
				return err
		}
		v.Title = "Right"
		v.Wrap = true
		v.Editable = false
		v.Highlight = true
		v.SelBgColor = gocui.ColorYellow
		v.SelFgColor = gocui.ColorBlack
		for _, command := range commandGroups[selectedGroup] {
			fmt.Fprintln(v, command.Name)
		}
    if _, err := g.SetCurrentView("right"); err != nil {
        return err
    }
}

	return nil
}

func executeCommand(g *gocui.Gui, v *gocui.View) error {
    middleView, err := g.View("middle")
    if err != nil {
        return err
    }
		_, cy := v.Cursor()
    command := commandGroups[selectedGroup][cy].Cmd
    fmt.Fprintln(middleView, "\033[32m$ "+command+"\033[0m") // Add color escape codes
    cmd := exec.Command("/bin/sh", "-c", command)
    output, err := cmd.Output()
    if err != nil {
        return err
    }
    fmt.Fprintln(middleView, string(output))
    fmt.Fprintln(middleView, "\033[33m--------------------------------------------------\033[0m") // Add horizontal line
    return nil
}

func clearMiddlePane(g *gocui.Gui, v *gocui.View) error {
    middleView, err := g.View("middle")
    if err != nil {
        return err
    }
    middleView.Clear()
    middleView.SetCursor(0, 0)
    middleView.SetOrigin(0, 0)
    return nil
}

func refreshRightPane(g *gocui.Gui) error {
    v, err := g.View("right")
    if err != nil {
        return err
    }
    v.Clear()
    for _, command := range commandGroups[selectedGroup] {
        fmt.Fprintln(v, command.Name)
    }
    return nil
}

func switchToView(viewName string) func(g *gocui.Gui, v *gocui.View) error {
    return func(g *gocui.Gui, v *gocui.View) error {
        _, cy := v.Cursor()
        selectedGroup = v.BufferLines()[cy]
        if err := refreshRightPane(g); err != nil {
            return err
        }
        if viewName == "left" { // Added this block
            rightView, err := g.View("right")
            if err != nil {
                return err
            }
            rightView.Clear()
        }
        _, err := g.SetCurrentView(viewName)
        return err
    }
}



func moveCursorDown(g *gocui.Gui, v *gocui.View) error {
    cx, cy := v.Cursor()
    ox, oy := v.Origin()
    if cy+1 < len(v.BufferLines()) { // Added this check
        if err := v.SetCursor(cx, cy+1); err != nil && oy < len(v.BufferLines())-1 {
            if err := v.SetOrigin(ox, oy+1); err != nil {
                return err
            }
        }
        if v.Name() == "left" {
            selectedGroup = v.BufferLines()[cy+1]
            if err := refreshRightPane(g); err != nil {
                return err
            }
        }
    }
    return nil
}

func moveCursorUp(g *gocui.Gui, v *gocui.View) error {
    cx, cy := v.Cursor()
    ox, oy := v.Origin()
    if cy-1 >= 0 { // Added this check
        if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
            if err := v.SetOrigin(ox, oy-1); err != nil {
                return err
            }
        }
        if v.Name() == "left" {
            selectedGroup = v.BufferLines()[cy-1]
            if err := refreshRightPane(g); err != nil {
                return err
            }
        }
    }
    return nil
}
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}