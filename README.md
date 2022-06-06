# icemon

* normal operation, multiple streams from one server: `go run icemon.go 192.168.100.101 StreamA StreamB StreamC ...`
* dump metadata (lots of output!): `go run icemon.go -v 192.168.100.101 StreamA StreamB StreamC ...`
* same stream from different servers: `go run icemon.go -s StreamA 192.168.100.101 192.168.100.102 192.168.100.103 ...`

to monitor streams for some time run something such as the following in a screen session:

`go run icemon.go 192.168.100.101 StreamA StreamB StreamC 2>&1 | tee streams.log`
