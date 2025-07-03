package main

import (
	"fmt"
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
	"image/color"
	"os"
	"path/filepath"
)

type Config struct {
	// Global
	ButtonClose string
	// First page
	MainLabel      string
	TeamLabel      string
	ButtonContinue string
	// Second page
	ChoosePathLabel string
	ButtonInstall   string
	LabelPath       string
	ButtonBrowse    string
	// Integrity error
	IntegrityError string
	// Injection error
	InjectionError string
	// Third page
	ThanksLabel string
}

var conf Config

func main() {
	// Init app
	a := app.New()
	w := a.NewWindow("BBQDeploy")
	w.Resize(fyne.NewSize(800, 600))

	loadingLabel := widget.NewLabel("Downloading resources, please wait...")
	loadingWidget := widget.NewActivity()
	loadingWidget.Start()

	init := container.New(layout.NewCenterLayout(),
		container.New(layout.NewVBoxLayout(),
			loadingLabel,
			loadingWidget,
		),
	)

	w.SetContent(init)
	w.Show()

	// Download files in goroutine
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
	type Package struct {
		mainLabel  *widget.Label
		teamLabel  *canvas.Text
		errorLabel *canvas.Text
		page0      *fyne.Container
	}

	var pkg Package

	pkg.mainLabel = widget.NewLabel(conf.MainLabel)

	pkg.teamLabel = canvas.NewText(conf.TeamLabel, color.RGBA{R: 169, G: 169, B: 169, A: 255})
	pkg.teamLabel.TextSize = 12

	pkg.errorLabel = canvas.NewText("", color.RGBA{R: 255, A: 255})

	btnContinue := widget.NewButton(conf.ButtonContinue, func() {
		w.SetContent(pageInstall(w))
	})

	pkg.page0 = container.New(layout.NewCenterLayout(),
		container.New(layout.NewVBoxLayout(),
			pkg.mainLabel,
			pkg.teamLabel,
			btnContinue,
			pkg.errorLabel,
		),
	)

	// Check integrity of downloaded files
	checkIntegrity(btnContinue, pkg.errorLabel)

	plugin.Global.Entry("Page0", pkg)

	return pkg.page0
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
	type Package struct {
		path        string
		labelPath   *widget.Label
		pageInstall *fyne.Container
	}

	var pkg Package

	choosePathLabel := widget.NewLabel(conf.ChoosePathLabel)
	pkg.labelPath = widget.NewLabel("")

	btnInstall := widget.NewButtonWithIcon(conf.ButtonInstall, theme.DownloadIcon(), func() {
		w.SetContent(pageEnd(pkg.path))
	})
	btnInstall.Disable()

	btnBrowse := widget.NewButtonWithIcon(conf.ButtonBrowse, theme.SearchIcon(), func() {
		browseFile(w, func(selectedPath string) {
			pkg.path = selectedPath
			// Display chosen path
			pkg.labelPath.SetText(conf.LabelPath + pkg.path)
			btnInstall.Enable()
		})
	})

	pkg.pageInstall = container.New(layout.NewCenterLayout(),
		container.New(layout.NewVBoxLayout(),
			choosePathLabel,
			btnBrowse,
			pkg.labelPath,
			btnInstall,
		),
	)

	plugin.Global.Entry("PageInstall", pkg)

	return pkg.pageInstall
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
	plugin.Global.Entry("AfterInstall", path)

	_ = os.RemoveAll(filepath.Join(appDir, "resources"))

	if err != nil {
		errLabel := canvas.NewText(conf.InjectionError, color.RGBA{R: 255, A: 255})
		errCode := canvas.NewText("[ERROR]: Error "+fmt.Sprint(err), color.RGBA{R: 255, A: 255})

		buttonClose := widget.NewButtonWithIcon(conf.ButtonClose, theme.WindowCloseIcon(), func() {
			fyne.CurrentApp().Quit()
		})

		pageEndContainer := container.New(layout.NewCenterLayout(),
			container.New(layout.NewVBoxLayout(),
				errLabel,
				errCode,
				buttonClose,
			),
		)
		return pageEndContainer
	} else {

		type Package struct {
			label            *widget.Label
			pageEndContainer *fyne.Container
		}

		var pkg Package

		pkg.label = widget.NewLabel(conf.ThanksLabel)
		buttonClose := widget.NewButtonWithIcon(conf.ButtonClose, theme.WindowCloseIcon(), func() {
			fyne.CurrentApp().Quit()
		})

		pkg.pageEndContainer = container.New(layout.NewCenterLayout(),
			container.New(layout.NewVBoxLayout(),
				pkg.label,
				buttonClose,
			),
		)

		plugin.Global.Entry("PageEnd", pkg)

		return pkg.pageEndContainer
	}
}
