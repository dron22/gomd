package html

import "fmt"

var html = `<!DOCTYPE html>
<html lang="en" >
    <head>
        <script>

            document.addEventListener('DOMContentLoaded', function() {
                // Create a new WebSocket.
                var socket = new WebSocket('ws://%v:%v/ws');

                // Handle messages sent by the server.
                socket.onmessage = function(event) {
                  var message = event.data;
                  console.log("File changed!")
                  document.getElementById('main').innerHTML = message;
                };
                console.log('Document loaded')
            })

        </script>
    </head>
    <body>
        <div id="main"></div>
    </body>
</html>`

func GetHTML(host string, port int) string {
    return fmt.Sprintf(html, host, port)
}

