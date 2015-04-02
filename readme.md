# riemann-ping

`riemann-ping` is a tool for running pings against specified urls and sending that data through [Riemann][1]. It
performs http gets against the specified url and then sends the number of milliseconds that the request took to
complete to [riemann][1]. It makes heavy use of built in functionality in [Riemann][1] and therefore leaves it up to
Riemann what threshold should be considered warning or critical thresholds. It also makes use of expired events for
handling the case where a url is not available.

[1]:http://riemann.io

# Installing

You can either build from source or you can download from the releases section on github.

# Developing

- Use the [most recent go release][2]
- Use [`godep`][3] to make sure that you are using at least the version specified for each dependency
- If you update depencdency, be sure to save the updates with `godep`
- Always run `go fmt` and [`golint`][4] before submitting a pull request.

[2]:http://golang.org/doc/install
[3]:https://github.com/tools/godep
[4]:https://github.com/bytbox/golint
