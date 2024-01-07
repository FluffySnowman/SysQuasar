package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/jroimartin/gocui"
)

func main() {
    g, err := gocui.NewGui(gocui.OutputNormal)
    if err != nil {
        log.Panicln(err)
    }
    defer g.Close()

    g.SetManagerFunc(layout)

    if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
        log.Panicln(err)
    }

    if err := g.SetKeybinding("right", gocui.KeyEnter, gocui.ModNone, executeCommand); err != nil {
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

    if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
        log.Panicln(err)
    }
}

func layout(g *gocui.Gui) error {
    maxX, maxY := g.Size()
    if v, err := g.SetView("left", 0, 0, maxX/3, maxY-1); err != nil {
        if err != gocui.ErrUnknownView {
            return err
        }
        v.Title = "Left"
    }

    if v, err := g.SetView("middle", maxX/3+1, 0, 2*maxX/3, maxY-1); err != nil {
        if err != gocui.ErrUnknownView {
            return err
        }
        v.Title = "Middle"
        v.Wrap = true
    }

    if v, err := g.SetView("right", 2*maxX/3+1, 0, maxX-1, maxY-1); err != nil {
        if err != gocui.ErrUnknownView {
            return err
        }
        v.Title = "Right"
        v.Highlight = true
        v.SelBgColor = gocui.ColorGreen
        v.SelFgColor = gocui.ColorBlack
        v.Wrap = true
        v.Editable = false
        fmt.Fprintln(v, "ls")
        fmt.Fprintln(v, "pwd")
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
    command := v.Buffer()
    cmd := exec.Command("/bin/sh", "-c", command)
    output, err := cmd.Output()
    if err != nil {
        return err
    }
    middleView.Clear()
    fmt.Fprintln(middleView, string(output))
    return nil
}

func switchToView(viewName string) func(g *gocui.Gui, v *gocui.View) error {
    return func(g *gocui.Gui, v *gocui.View) error {
        _, err := g.SetCurrentView(viewName)
        return err
    }
}

func moveCursorDown(g *gocui.Gui, v *gocui.View) error {
    cx, cy := v.Cursor()
    ox, oy := v.Origin()
    if err := v.SetCursor(cx, cy+1); err != nil && oy < len(v.BufferLines())-1 {
        if err := v.SetOrigin(ox, oy+1); err != nil {
            return err
        }
    }
    return nil
}

func moveCursorUp(g *gocui.Gui, v *gocui.View) error {
    cx, cy := v.Cursor()
    ox, oy := v.Origin()
    if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
        if err := v.SetOrigin(ox, oy-1); err != nil {
            return err
        }
    }
    return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
    return gocui.ErrQuit
}
