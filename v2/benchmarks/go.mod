module benchmark

go 1.21.4

replace github.com/loopholelabs/polyglot/v2 => ../

require (
	github.com/loopholelabs/polyglot/v2 v2.0.2
	google.golang.org/grpc v1.66.0
	google.golang.org/protobuf v1.35.1
)

require (
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect; indirect..
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
)
