# Themer

Themer is a CLI for creating and switching between desktop themes. It is still in its early stages and currently only supports theming of (Termite)[https://wiki.archlinux.org/index.php/Termite].

## Installation

Install and configure go. Then run:

```bash
go get -u github.com/mmuldo/themer
```

## Supported applications

### Terminals

* Termite

## Usage

To get the full usage:

```bash
themer --help
```

### Create a new theme

Currently, themer only supports creating new themes based on a supplied image. To create a new theme:

```bash
themer create -i {image_file} [-n {theme_name}]
```

This will output the theme to the terminal. If a name is specified, the theme will be saved to `~/.config/themer/themes/`.

### Switch desktop to a saved theme

To switch to a saved theme:

```bash
themer switch -n {theme_name}
```

## Configuration

Configuration can be specified in a file named `config` with a supported extension. This file must be located at `~/.config/themer/`.

### Supported file types

* YAML
* TOML
* JSON

### Example

```YAML
# ~/.config/themer/config.yaml
terminal: termite
```
