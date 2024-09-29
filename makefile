
all:
	tinygo build -target=pico -o swiper.uf2
flash:
	tinygo flash -target=pico -o swiper.uf2 main.go
