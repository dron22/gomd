
# Gomd

Render and preview markdown files locally.

Gomd is a command-line server written in go using the [blackfriday](github.com/russross/blackfriday) package to render markdown and [websockets](https://github.com/gorilla/websocket) to update the preview upon change of the previewed file. It was inspired by python's [grip](https://github.com/joeyespo/grip), but currently still lacks advanced features like (code) highlighting etc. of [kramdown](http://github.com/gettalong/kramdown).

## Getting started

### Install

    $ go get github.com/dron22/gomd

### Start server

    $ gomd /path/to/markdownfile.md

