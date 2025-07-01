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
	"image/color"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	// Global
	ButtonClose    string
	ExecutableName string
	Name           string
	// Download error
	DownloadError string
	// First page
	MainLabel      string
	TeamLabel      string
	ButtonContinue string
	// Second page
	isSteamEnabled           bool
	isCheckExecutableEnabled bool
	ChoosePathLabel          string
	ButtonInstall            string
	LabelPath                string
	ButtonBrowse             string
	// Executable error
	ExecutableError string
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
			fyne.Do(func() {
				w.SetContent(page0(w))
			})
		}
	}()

	fyne.CurrentApp().Run()
}

func page0(w fyne.Window) *fyne.Container {
	mainLabel := widget.NewLabel(conf.MainLabel)

	teamLabel := canvas.NewText(conf.TeamLabel, color.RGBA{R: 169, G: 169, B: 169, A: 255})
	teamLabel.TextSize = 12

	errorLabel := canvas.NewText("", color.RGBA{R: 255, A: 255})

	btnContinue := widget.NewButton(conf.ButtonContinue, func() {
		w.SetContent(pageInstall(w))
	})

	page0 := container.New(layout.NewCenterLayout(),
		container.New(layout.NewVBoxLayout(),
			mainLabel,
			teamLabel,
			btnContinue,
			errorLabel,
		),
	)

	// Check integrity of downloaded files
	checkIntegrity(btnContinue, errorLabel)

	return page0
}

func pageERR(_ fyne.Window, err int) *fyne.Container {
	errLabel := canvas.NewText(conf.DownloadError, color.RGBA{R: 255, A: 255})
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
	var path string

	choosePathLabel := widget.NewLabel(conf.ChoosePathLabel)
	labelPath := widget.NewLabel("")
	errorLabel := canvas.NewText("", color.RGBA{R: 255, A: 255})

	btnInstall := widget.NewButtonWithIcon(conf.ButtonInstall, theme.DownloadIcon(), func() {
		w.SetContent(pageEnd(path))
	})
	btnInstall.Disable()

	btnSteam := widget.NewButtonWithIcon("Steam", theme.ComputerIcon(), func() {
		// Choose default path to game depending on OS
		if runtime.GOOS == "windows" {
			path = filepath.Join("C:\\", "Program Files (x86)", "Steam", "steamapps", "common", conf.Name)
		} else {
			homeDir := os.Getenv("HOME")
			path = filepath.Join(homeDir, ".steam", "root", "steamapps", "common", conf.Name)
		}
		// Check if there is executable game file
		if conf.isCheckExecutableEnabled == false {
			checkExecutable(path, btnInstall, errorLabel)
		}
		// Display chosen path
		labelPath.SetText(conf.LabelPath + path)
	})

	btnBrowse := widget.NewButtonWithIcon(conf.ButtonBrowse, theme.SearchIcon(), func() {
		browseFile(w, func(selectedPath string) {
			path = selectedPath
			// Check if there is executable game file
			if conf.isCheckExecutableEnabled == false {
				checkExecutable(path, btnInstall, errorLabel)
			}
			// Display chosen path
			labelPath.SetText(conf.LabelPath + path)
		})
	})

	if conf.isCheckExecutableEnabled == true {
		btnInstall.Enable()
	}

	pageInstall := container.New(layout.NewCenterLayout(),
		container.New(layout.NewVBoxLayout(),
			choosePathLabel,
			btnSteam,
			btnBrowse,
			labelPath,
			errorLabel,
			btnInstall,
		),
	)

	if conf.isSteamEnabled == true {
		btnSteam.Hide()
	}

	return pageInstall
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

func checkExecutable(selectedPath string, btnInstall *widget.Button, errorLabel *canvas.Text) {
	executablePath := filepath.Join(selectedPath, conf.ExecutableName)
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		btnInstall.Disable()
		errorLabel.Text = fmt.Sprintf(conf.ExecutableError, conf.ExecutableName)
		errorLabel.Refresh()
	} else {
		btnInstall.Enable()
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

	_ = os.RemoveAll(filepath.Join(appDir, "resources"))

	if err != nil {
		errLabel := canvas.NewText(conf.InjectionError, color.RGBA{R: 255, A: 255})
		errCode := canvas.NewText("[FATL]: Error "+fmt.Sprint(err), color.RGBA{R: 255, A: 255})

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
		label := widget.NewLabel(conf.ThanksLabel)
		buttonClose := widget.NewButtonWithIcon(conf.ButtonClose, theme.WindowCloseIcon(), func() {
			fyne.CurrentApp().Quit()
		})

		pageEndContainer := container.New(layout.NewCenterLayout(),
			container.New(layout.NewVBoxLayout(),
				label,
				buttonClose,
			),
		)
		return pageEndContainer
	}
}
