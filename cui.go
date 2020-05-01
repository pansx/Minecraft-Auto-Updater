package main

import (
	"fmt"
	"time"

	"github.com/jroimartin/gocui"
)

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if menu, err := g.SetView("menu", 0, 0, 20, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		menu.Highlight = true
		menu.SelFgColor = gocui.ColorBlack
		menu.SelBgColor = gocui.ColorGreen
		menu.Title = "菜单"
		fmt.Fprintln(menu, "启动游戏(Enter)")
		fmt.Fprintln(menu, "修复游戏")
		fmt.Fprintln(menu, "更换游戏版本")
		fmt.Fprintln(menu, "游戏维护人员工具")
	}

	if notice, err := g.SetView("notice", 21, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		notice.Title = "公告"
		notice.Wrap = true
		fmt.Fprintln(notice, "欢迎使用minecraft自动更新器")
		fmt.Fprintln(notice, "请勿将该窗口大小调的太小以免程序闪退")
		fmt.Fprintln(notice, "用鼠标点击对应的菜单项或按下菜单项对应的按键可以使用对应功能")
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func menuOnClick(g *gocui.Gui, v *gocui.View) error {
	g.SetCurrentView(v.Name())
	_, y := v.Cursor()
	if y == 0 {
		autoUpdateCUI(g, v)
	}
	return nil
}

func autoUpdateCUI(g *gocui.Gui, v *gocui.View) error {
	notice, _ := g.View("notice")
	notice.Clear()
	notice.Title = "控制台输出"
	go AutoUpdate(false, notice)
	go updateCUI(g)
	return nil
}

func updateCUI(g *gocui.Gui) {
	for true {
		time.Sleep(time.Millisecond * 100)
		g.Update(func(gg *gocui.Gui) error { return nil })
	}
}

func cui() {
	g, _ := gocui.NewGui(gocui.OutputNormal)
	defer g.Close()
	g.ASCII = true
	g.Mouse = true
	g.SetManagerFunc(layout)
	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, autoUpdateCUI)
	g.SetKeybinding("menu", gocui.MouseLeft, gocui.ModNone, menuOnClick)
	g.MainLoop()
}
