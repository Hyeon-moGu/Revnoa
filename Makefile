APP_NAME=revnoa
DIST_DIR=dist
CONFIG_FILE=config.yaml

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(APP_NAME)_linux .
	cp $(CONFIG_FILE) $(DIST_DIR)/

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(DIST_DIR)/$(APP_NAME).exe .
	cp $(CONFIG_FILE) $(DIST_DIR)/

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/$(APP_NAME)_darwin .
	cp $(CONFIG_FILE) $(DIST_DIR)/

clean:
	rm -rf $(DIST_DIR)/*
