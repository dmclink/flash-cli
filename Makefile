run:
	go run main.go

build:
	go build -o bin/flash-cli main.go

install-example-plugins:
	# prettier-render: Go renderer example
	@mkdir -p ~/.config/flash-cli/plugins/prettier-render
	go build -C examples/prettier-render -o ~/.config/flash-cli/plugins/prettier-render/renderer-plugin .
	cp examples/prettier-render/plugin.toml ~/.config/flash-cli/plugins/prettier-render/

	# python-simple-renderer: Python renderer example
	@mkdir -p ~/.config/flash-cli/plugins/python-simple-renderer
	cp examples/python-simple-renderer/renderer.py ~/.config/flash-cli/plugins/python-simple-renderer/renderer.py
	cp examples/python-simple-renderer/main.py ~/.config/flash-cli/plugins/python-simple-renderer/main.py
	cp examples/python-simple-renderer/plugin.toml ~/.config/flash-cli/plugins/python-simple-renderer/
	cp -r examples/python-simple-renderer/gen ~/.config/flash-cli/plugins/python-simple-renderer/
	chmod +x ~/.config/flash-cli/plugins/python-simple-renderer/main.py
	@echo "All example plugins successfully installed!"
