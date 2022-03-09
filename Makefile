build:
	rm -f bin/*
	go build -a -tags netgo -installsuffix netgo --ldflags '-extldflags "-static"' -o bin/bootstrap
	zip -j bin/layer.zip bin/bootstrap