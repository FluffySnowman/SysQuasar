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
		{"Last (logins)[10]", "sudo last -n 10"},
		{"Apache Access (-n 10)", "sudo tail -n 10 /var/log/apache2/access.log"},
		{"Apache Error (-n 10)", "sudo tail -n 10 /var/log/apache2/error.log"},
		{"Nginx Access (-n 10)", "sudo tail -n 10 /var/log/nginx/access.log"},
		{"Nginx Error (-n 10)", "sudo tail -n 10 /var/log/nginx/error.log"},
		// {"Apache Acess (stream)", "sudo tail -f /var/log/apache2/access.log"},
	},
	"apache": {
		{"Start Apache", "sudo systemctl start apache2"},
		{"Stop Apache", "sudo systemctl stop apache2"},
		{"Status Apache", "sudo systemctl status apache2"},
		{"Apache Access (-n 10)", "sudo tail -n 10 /var/log/apache2/access.log"},
		{"Apache Error (-n 10)", "sudo tail -n 10 /var/log/apache2/error.log"},
		// {"Edit Config", "sudo vi /etc/apache2/apache2.conf"},
		{"Edit Config", "echo 'feature WIP'"},
	},
	"nginx": {
		{"Start Nginx", "sudo systemctl start nginx"},
		{"Stop Nginx", "sudo systemctl stop nginx"},
		{"Status Nginx", "sudo systemctl status nginx"},
		{"Edit Config", "echo 'feature WIP'"},
		// {"Edit Config", "sudo vi /etc/apache2/apache2.conf"},
	},
	"docker": {
		{"Start Docker", "sudo systemctl start docker.service docker.socket"},
		{"Stop Docker", "sudo systemctl stop docker.service docker.socket"},
	},
}

var commandGroupNames = []string{"general", "git", "logs", "apache", "nginx", "docker"}

var selectedGroup = "general" // default which is glitchy for some reasont

var sudoPassword string // passed as args to the commands that require sudo
var isPasswordPopupActive bool
var passwordPopup *gocui.View

var previousView string


// KEYBINDS

type Keybinding struct {
    ViewName string
    Key      interface{}
    Mod      gocui.Modifier
    Handler  func(*gocui.Gui, *gocui.View) error
}

var keybindings = []Keybinding{
    {"", gocui.KeyCtrlC, gocui.ModNone, quit},
    {"", 'q', gocui.ModNone, quit},
    {"right", gocui.KeyEnter, gocui.ModNone, executeCommand},
    {"right", 'c', gocui.ModNone, clearMiddlePane},
    {"", 'h', gocui.ModNone, switchToView("left")},
    {"", 'l', gocui.ModNone, switchToView("right")},
    {"", 'j', gocui.ModNone, moveCursorDown},
    {"", 'k', gocui.ModNone, moveCursorUp},
    {"", 'p', gocui.ModNone, getPassword},
    {"", 'i', gocui.ModNone, selectMiddlePane},
    {"", 'b', gocui.ModNone, switchToPreviousView},
    {"middle", 'K', gocui.ModNone, scrollUp},
    {"middle", 'J', gocui.ModNone, scrollDown},
    {"passwordPopup", gocui.KeyEnter, gocui.ModNone, handlePassword},
    {"left", gocui.KeyEnter, gocui.ModNone, switchToView("right")},
}

func enableKeybindings(g *gocui.Gui, viewName string) error {
    for _, kb := range keybindings {
        if kb.ViewName == viewName {
            if err := g.SetKeybinding(kb.ViewName, kb.Key, kb.Mod, kb.Handler); err != nil {
                return err
            }
        }
    }
    return nil
}


func disableKeybindings(g *gocui.Gui, viewName string) error {
    for _, kb := range keybindings {
        if kb.ViewName == viewName {
            if err := g.DeleteKeybinding(kb.ViewName, kb.Key, kb.Mod); err != nil {
                return err
            }
        }
    }
    return nil
}


func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.Mouse = true
	g.SelFgColor = gocui.ColorGreen
	g.SelBgColor = gocui.ColorBlack

	g.SetManagerFunc(layout)

	for _, kb := range keybindings {
			if err := g.SetKeybinding(kb.ViewName, kb.Key, kb.Mod, kb.Handler); err != nil {
					log.Panicln(err)
			}
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
		v.Title = "Categories"
		v.Wrap = true
		v.Editable = false
		v.Highlight = true
		v.SelBgColor = gocui.ColorBlue
		v.SelFgColor = gocui.ColorWhite

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
		v.Title = "Output"
		// v.Wrap = false
		v.Autoscroll = true
		v.Highlight = true
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack
	}

	if v, err := g.SetView("right", 3*maxX/4+1, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Options"
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

	if v, err := g.SetView("password", 30, 2, 70, 4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		if _, err := g.SetViewOnTop("middle"); err != nil {
				return err
		}

		v.Editable = true
		v.Wrap = true
		v.Title = "Enter sudo password"
		v.Frame = false
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

	// if strings.HasPrefix(command, "sudo ") {


	// 	getPassword(g, v)
	// 	return nil
	// }

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

func selectMiddlePane(g *gocui.Gui, v *gocui.View) error {
    currentView := g.CurrentView()
    if currentView.Name() == "middle" {
        _, err := g.SetCurrentView("left")
        return err
    } else {
        _, err := g.SetCurrentView("middle")
        return err
    }
}

func switchToView(viewName string) func(*gocui.Gui, *gocui.View) error {
    return func(g *gocui.Gui, v *gocui.View) error {
        previousView = v.Name() // saving the previousvwiew
        _, err := g.SetCurrentView(viewName)
        return err
    }
}

func switchToPreviousView(g *gocui.Gui, v *gocui.View) error {
    _, err := g.SetCurrentView(previousView)
    return err
}

// func switchToView(viewName string) func(g *gocui.Gui, v *gocui.View) error {
// 	return func(g *gocui.Gui, v *gocui.View) error {
// 		if v != nil {
// 			v.FgColor = gocui.ColorWhite
// 		}

// 		newView, err := g.SetCurrentView(viewName)
// 		if err != nil {
// 			return err
// 		}
// 		newView.FgColor = gocui.ColorBlue

// 		_, cy := v.Cursor()
// 		selectedGroup = v.BufferLines()[cy]
// 		if err := refreshRightPane(g); err != nil {
// 			return err
// 		}
// 		if viewName == "left" {
// 			rightView, err := g.View("right")
// 			if err != nil {
// 				return err
// 			}
// 			rightView.Clear()
// 		}

// 		_, err = g.SetCurrentView(viewName)
// 		return err

// 	}
// }

func moveCursorDown(g *gocui.Gui, v *gocui.View) error {
	if !isPasswordPopupActive {
		cx, cy := v.Cursor()
		ox, oy := v.Origin()
		if cy+1 < len(v.BufferLines()) && cy+1 < len(commandGroupNames) { // Added check for commandGroupNames
			if err := v.SetCursor(cx, cy+1); err != nil && oy < len(v.BufferLines())-1 {
				if err := v.SetOrigin(ox, oy+1); err != nil {
					return err
				}
			}
			if v.Name() == "left" {
				selectedGroup = commandGroupNames[cy+1]
				if err := refreshRightPane(g); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func moveCursorUp(g *gocui.Gui, v *gocui.View) error {
	if !isPasswordPopupActive {
		cx, cy := v.Cursor()
		ox, oy := v.Origin()
		if cy-1 >= 0 { // Added this check
			if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
				if err := v.SetOrigin(ox, oy-1); err != nil {
					return err
				}
			}
			if v.Name() == "left" && cy-1 >= 0 {
				selectedGroup = commandGroupNames[cy-1]
				if err := refreshRightPane(g); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// // scrolling for the middle pane

func scrollUp(g *gocui.Gui, v *gocui.View) error {
    if v != nil {
        cx, cy := v.Cursor()
        if err := v.SetCursor(cx, cy-1); err != nil && cy > 0 {
            if err := v.SetOrigin(cx, cy-1); err != nil {
                return err
            }
        }
    }
    return nil
}

func scrollDown(g *gocui.Gui, v *gocui.View) error {
    if v != nil {
        cx, cy := v.Cursor()
        if err := v.SetCursor(cx, cy+1); err != nil {
            ox, oy := v.Origin()
            if err := v.SetOrigin(ox, oy+1); err != nil {
                return err
            }
        }
    }
    return nil
}

func getPassword(g *gocui.Gui, v *gocui.View) error {
	max_x, max_y := g.Size()
	if v, err := g.SetView("passwordPopup", max_x/4, max_y/4, 3*max_x/4, 3*max_y/4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		passwordPopup = v 
		v.Title = "Enter sudo password"
		v.Editable = true
		v.Wrap = true
		v.Mask = '*'
		if _, err := g.SetCurrentView("passwordPopup"); err != nil {
			return err
		}
	}
	return nil
}

func hidePassword(g *gocui.Gui, v *gocui.View) error {
	sudoPassword = v.Buffer()
	if _, err := g.SetViewOnBottom("password"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("right"); err != nil {
		return err
	}
	return nil
}

func handlePassword(g *gocui.Gui, v *gocui.View) error {
	sudoPassword = v.Buffer()
	if err := g.DeleteView("passwordPopup"); err != nil {
		return err
	}

	passwordPopup = nil 
	if _, err := g.SetCurrentView("right"); err != nil {
		return err
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
