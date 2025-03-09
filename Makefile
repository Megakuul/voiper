.PHONY: dev-voiper build-voiper

dev-voiper:
	@cd cmd/voiper && wails dev -appargs "--base ../../configs"

build-voiper:
	@cd cmd/voiper && wails build