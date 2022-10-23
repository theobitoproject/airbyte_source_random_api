.PHONY: read
read:
	go run -race . read --config secrets/config.json --catalog sample_files/configured_catalog.json
