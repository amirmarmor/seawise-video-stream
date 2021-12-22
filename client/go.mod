module www.seawise.com/client

go 1.16

require (
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/namsral/flag v1.7.4-pre
	gocv.io/x/gocv v0.28.0
	www.seawise.com/common v0.0.0-00000000000000-000000000000
)

replace www.seawise.com/common => ../common
