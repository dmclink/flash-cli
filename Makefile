run:
	go run main.go

build:
	go build -o bin/flash-cli main.go

install-example-plugins:
	# prettier-render: Go renderer example
	@mkdir -p ~/.config/flash-cli/plugins/prettier-render
	go build -C examples/prettier-render -o ~/.config/flash-cli/plugins/prettier-render/main .
	cp examples/prettier-render/plugin.toml ~/.config/flash-cli/plugins/prettier-render/

	# python-simple-renderer: Python renderer example
	@mkdir -p ~/.config/flash-cli/plugins/python-simple-renderer
	rm -rf ~/.config/flash-cli/plugins/python-simple-renderer/gen
	rm -rf examples/python-simple-renderer/gen
	cp examples/python-simple-renderer/renderer.py ~/.config/flash-cli/plugins/python-simple-renderer/renderer.py
	cp examples/python-simple-renderer/main.py ~/.config/flash-cli/plugins/python-simple-renderer/main.py
	cp examples/python-simple-renderer/plugin.toml ~/.config/flash-cli/plugins/python-simple-renderer/
	@mkdir -p ~/.config/flash-cli/plugins/python-simple-renderer/gen
	cp -r gen/python ~/.config/flash-cli/plugins/python-simple-renderer/gen/
	@mkdir -p examples/python-simple-renderer/gen
	cp -r gen/python examples/python-simple-renderer/gen/
	chmod +x ~/.config/flash-cli/plugins/python-simple-renderer/main.py

	@echo "All example plugins successfully installed!"
