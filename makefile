
all:
	tinygo build -target=pico -o uf2/swiper.uf2
flash:
	tinygo flash -target=pico -o uf2/swiper.uf2 main.go
