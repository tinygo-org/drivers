update: esp-wifi/README.md
	rm -rf blobs/headers
	rm -rf blobs/include
	rm -rf blobs/libs
	mkdir -p blobs/libs
	cp -rp esp-wifi/esp-wifi-sys/headers      blobs
	cp -rp esp-wifi/esp-wifi-sys/include      blobs
	cp -rp esp-wifi/esp-wifi-sys/libs/esp32c3 blobs/libs

esp-wifi/README.md:
	git clone https://github.com/esp-rs/esp-wifi
