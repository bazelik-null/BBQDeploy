package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/BurntSushi/toml"
	"github.com/bazelik-null/BBQDeploy/src/plugin"
	"github.com/bazelik-null/BBQDeploy/src/pluginapi"

	"fmt"
	"image/color"
	"os"
	"path/filepath"
)

var conf pluginapi.Config

func main() {
	a := app.New()
	w := a.NewWindow("BBQDeploy")
	w.Resize(fyne.NewSize(800, 600))

	loadingLabel := widget.NewLabel("Downloading resources, please wait...")
	loadingWidget := widget.NewActivity()
	loadingWidget.Start()

	initContainer := container.New(layout.NewCenterLayout(),
		container.New(layout.NewVBoxLayout(),
			loadingLabel,
			loadingWidget,
		),
	)
	w.SetContent(initContainer)
	w.Show()

	go func() {
		err := download()
		if err != 0 {
			w.SetContent(pageERR(w, err))
		} else {
			// Extract data from config
			appDir, _ := os.Getwd()
			tomlPath := filepath.Join(appDir, "resources", "config", "config.toml")
			tomlData, _ := os.ReadFile(tomlPath)
			_, err := toml.Decode(string(tomlData), &conf)
			if err != nil {
				fmt.Println(err)
			}

			// Import and load plugins
			pluginDir := filepath.Join(appDir, "resources", "plugins")
			plugin.Global.LoadPlugins(pluginDir)

			fyne.Do(func() {
				w.SetContent(page0(w))
			})
		}
	}()

	fyne.CurrentApp().Run()
}

func page0(w fyne.Window) *fyne.Container {
	pkg := &pluginapi.Page0Package{
		MainLabel:  widget.NewLabel(conf.MainLabel),
		TeamLabel:  canvas.NewText(conf.TeamLabel, color.RGBA{R: 169, G: 169, B: 169, A: 255}),
		ErrorLabel: canvas.NewText("", color.RGBA{R: 255, A: 255}),
	}

	btnContinue := widget.NewButton(conf.ButtonContinue, func() {
		w.SetContent(pageInstall(w))
	})

	pkg.Container = container.New(layout.NewCenterLayout(),
		container.New(layout.NewVBoxLayout(),
			pkg.MainLabel,
			pkg.TeamLabel,
			btnContinue,
			pkg.ErrorLabel,
		),
	)

	// Integrity check
	checkIntegrity(btnContinue, pkg.ErrorLabel)

	plugin.Global.Entry("Page0", pkg)

	return pkg.Container
}

func pageERR(_ fyne.Window, err int) *fyne.Container {
	errLabel := canvas.NewText("[FATL]: Download failed", color.RGBA{R: 255, A: 255})
	errCode := canvas.NewText("[FATL]: Error "+fmt.Sprint(err), color.RGBA{R: 255, A: 255})

	buttonClose := widget.NewButtonWithIcon(conf.ButtonClose, theme.WindowCloseIcon(), func() {
		fyne.CurrentApp().Quit()
	})

	pageERRContainer := container.New(layout.NewCenterLayout(),
		container.New(layout.NewVBoxLayout(),
			errLabel,
			errCode,
			buttonClose,
		),
	)
	return pageERRContainer
}

func pageInstall(w fyne.Window) *fyne.Container {
	pkg := &pluginapi.PageInstallPackage{}

	choosePathLabel := widget.NewLabel(conf.ChoosePathLabel)
	pkg.LabelPath = widget.NewLabel("")
	pkg.BtnInstall = widget.NewButtonWithIcon(conf.ButtonInstall, theme.DownloadIcon(), func() {
		w.SetContent(pageEnd(pkg.Path))
	})
	pkg.BtnInstall.Disable()

	btnBrowse := widget.NewButtonWithIcon(conf.ButtonBrowse, theme.SearchIcon(), func() {
		browseFile(w, func(sel string) {
			pkg.Path = sel
			pkg.LabelPath.SetText(conf.LabelPath + sel)
			pkg.BtnInstall.Enable()
		})
	})

	pkg.Container = container.New(layout.NewCenterLayout(),
		container.New(layout.NewVBoxLayout(),
			choosePathLabel,
			btnBrowse,
			pkg.LabelPath,
			pkg.BtnInstall,
		),
	)

	plugin.Global.Entry("PageInstall", pkg)
	return pkg.Container
}

func checkIntegrity(btnContinue *widget.Button, errorLabel *canvas.Text) {
	appDir, _ := os.Getwd()
	resourcesPath := filepath.Join(appDir, "resources", "config", "meta.json")
	if _, err := os.Stat(resourcesPath); os.IsNotExist(err) {
		btnContinue.Disable()
		errorLabel.Text = conf.IntegrityError
		errorLabel.Refresh()
	} else {
		btnContinue.Enable()
		errorLabel.Text = ""
		errorLabel.Refresh()
	}
}

func browseFile(w fyne.Window, onPathSelected func(string)) {
	dialog.ShowFolderOpen(func(folder fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if folder != nil {
			onPathSelected(folder.Path())
		}
	}, w)
}

func pageEnd(path string) *fyne.Container {
	appDir, _ := os.Getwd()

	err := install(path)

	plugin.Global.Entry("AfterInstall", pluginapi.AfterInstallPayload(path))

	_ = os.RemoveAll(filepath.Join(appDir, "resources"))

	if err != nil {
		errLabel := canvas.NewText(conf.InjectionError, color.RGBA{R: 255, A: 255})
		errCode := canvas.NewText("[ERROR]: Error "+fmt.Sprint(err), color.RGBA{R: 255, A: 255})

		buttonClose := widget.NewButtonWithIcon(conf.ButtonClose, theme.WindowCloseIcon(), func() {
			fyne.CurrentApp().Quit()
		})

		return container.New(layout.NewCenterLayout(),
			container.New(layout.NewVBoxLayout(),
				errLabel,
				errCode,
				buttonClose,
			),
		)
	}

	pkg := &pluginapi.PageEndPackage{
		Label: widget.NewLabel(conf.ThanksLabel),
	}

	buttonClose := widget.NewButtonWithIcon(conf.ButtonClose, theme.WindowCloseIcon(), func() {
		fyne.CurrentApp().Quit()
	})

	pkg.Container = container.New(layout.NewCenterLayout(),
		container.New(layout.NewVBoxLayout(),
			pkg.Label,
			buttonClose,
		),
	)

	plugin.Global.Entry("PageEnd", pkg)

	return pkg.Container
}
