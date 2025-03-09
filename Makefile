.PHONY: dev-voiper build-voiper

dev-voiper:
	@cd cmd/voiper && wails dev

build-voiper:
	@cd cmd/voiper && wails build