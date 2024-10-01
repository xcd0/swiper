
all:
	go generate
	tinygo build -target=pico -o uf2/swiper.uf2
flash:
	go generate
	tinygo flash -target=pico -o uf2/swiper.uf2 main.go
