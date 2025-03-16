.PHONY: dev-voiper build-voiper

dev-voiper:
	mkdir -p web/dist
	touch web/dist/dummy.txt
	@cd cmd/voiper && wails dev -appargs "--base ../../configs"

build-voiper:
	mkdir -p web/dist
	touch web/dist/dummy.txt
	@cd cmd/voiper && wails build