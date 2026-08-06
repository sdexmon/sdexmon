.PHONY: readme readme-gen readme-check build fmt vet test

# View the colourised README: `make readme | less -R`
readme:
	@cat README.ansi

# Rebuild README.ansi (and the README.md preview block) from README.md
readme-gen:
	@python3 scripts/gen-readme-ansi.py

# Fail if README.ansi has drifted from README.md
readme-check:
	@python3 scripts/gen-readme-ansi.py --check

build:
	go build -o sdexmon ./cmd/sdexmon

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...
