# srv

A simple replacement for "python2 -m SimpleHTTPServer" but with concurrency and WebSocket-based auto-refresh.

    go get -u github.com/holizz/srv
    srv

It has a vast array of options:

    srv -d /etc -p 9999

-d for directory, -p for port.
