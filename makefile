
all:
	go generate
	tinygo build -target=pico -o uf2/swiper.uf2
flash:
	go generate
	tinygo flash -target=pico -o uf2/swiper.uf2 main.go


install:
	sudo apt-get install gcc-arm-linux-gnueabihf
	wget https://github.com/tinygo-org/tinygo/releases/download/v0.33.0/tinygo_0.33.0_amd64.deb
	sudo dpkg -i tinygo_0.33.0_amd64.deb
	rm tinygo_0.33.0_amd64.deb

