# CS3031_Lab1
CS3031 Lab 1: Web Proxy

Author: Kyrill Zorin

Web proxy for HTTP and HTTPS traffic written in Go.
It uses the standard web proxy protocol.
HTTPS uses the CONNECT method but currently does not function properly.
It creates a tcp tunnel between the client and destination however in it's current state web browsers seem to just reject the connection.
I will hopefully find a solution to this issue soon.
The server handles multiple concurrent requests using goroutines.
The server prints all requests to stdout.
The server uses a blocklist to block certain web pages if the URL contains blocked phrases.
The blocklist is stored in boltdb, an embedded Go database.
The server provides an online blocklist management console accessible at http://management.console while using the proxy.
To add an entry to the blocklist simply fill in and submit the form in the management console.
I attempted to use web sockets for the management console howevere for some reason the browser kept rejecting the connection.
As a result I decided to use standard HTTP for it instead.
Caching is currently not implemented as I didn't have enough time due to trying to fix HTTPS and Websockets.
Caching could use an LRU cache built on top of boltdb to store content.
The HTTP HEAD request could be used to fetch content headers.
The headers could then be compared to those of the cache to determine whether the cache is stale and if appropriate fetch a new copy of the web page.
