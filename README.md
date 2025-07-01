<div align="center">
<h1>🚀BBQDeploy: Online Installer for GitHub Projects🚀</h1>
</div>

[РУССКИЙ](https://github.com/bazelik-null/BBQDeploy/blob/main/README_RU.md)

>[!CAUTION]
>
>🛠️ Work In Progress

## 📚 Table of Contents
- [Description](#description)
- [IMPORTANT](#important)
- [License](#license)

## ✨ Description
**BBQDeploy** is an easy-to-use cross-platform installer built on the [Fyne](https://fyne.io/) framework, designed specifically for GitHub projects.

It comes as a single executable file and installs all the necessary resources from the latest release on GitHub.

BBQDeploy offers a high degree of customization through its [module](#modules) system and full configuration in `config.toml`, allowing you to tailor it to your needs perfectly! 🎨

Example usage: [Localizer for ENA](https://github.com/bazelik-null/ENAbbq_rus)

## ⚠️ IMPORTANT
For a manual on installation, configuration, creating your own modules, and releasing, please visit the [documentation](https://github.com/bazelik-null/BBQDeploy/wiki) 📖

## 🛠️ Installation
1. Clone the installer into your project using the submodule system.
```bash
git submodule add https://github.com/bazelik-null/BBQDeploy
```
2. Initialize the submodule.
```bash
git submodule init
git submodule update
```
3. Copy the `resources` directory from the submodule and paste it into the root of your repository.
```bash
cp -r ./BBQDeploy/resources ./resources
```
4. Move the necessary files into it.
5. Open `resources/config/meta.json` and insert the path to the source file in the `src` field, and the path to the destination in the `dst` field. If there are multiple files, copy the structure with `src` and `dst` and paste it as many times as needed.
6. Open `resources/config/config.toml` and configure the installer.

## 📝 License
This project is licensed under the open-source [MIT](https://mit-license.org/) license. You are free to use, modify, and distribute this installer in accordance with the terms of the license. 🌟
