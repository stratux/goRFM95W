all:
	go get -t -d -v ./...
	go build example_txrx.go
	go build broadcast_params_trials_rx.go broadcast_params.go
	go build broadcast_params_trials_tx.go broadcast_params.go
