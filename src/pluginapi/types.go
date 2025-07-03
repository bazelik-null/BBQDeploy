package pluginapi

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
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

// Page0Package — payload for hook "Page0"
type Page0Package struct {
	MainLabel  *widget.Label
	TeamLabel  *canvas.Text
	ErrorLabel *canvas.Text
	Container  *fyne.Container
}

// PageInstallPackage — payload for hook "PageInstall"
type PageInstallPackage struct {
	Path       string
	LabelPath  *widget.Label
	Container  *fyne.Container
	BtnInstall *widget.Button
}

// PageEndPackage — payload for hook "PageEnd"
type PageEndPackage struct {
	Label     *widget.Label
	Container *fyne.Container
}

// AfterInstallPayload — for hook "AfterInstall"
type AfterInstallPayload string
