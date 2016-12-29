
# Gomd

Gomd is a command-line server written in go to render and preview markdown files locally. It is using the [blackfriday](github.com/russross/blackfriday) package for markdown rendering and [websockets](https://github.com/gorilla/websocket) for continuous updating of the preview upon every change of the previewed file. It was inspired by python's [grip](https://github.com/joeyespo/grip), but currently still lacks advanced features like code highlighting.

## Getting started

### Install

    $ go get github.com/dron22/gomd

### Preview markdown file

    $ gomd /path/to/markdownfile.md

