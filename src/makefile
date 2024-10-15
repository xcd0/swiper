
all:
	make build
build:
	go generate
	@mkdir -p uf2
	tinygo build -target=pico -o uf2/swiper.uf2
flash:
	go generate
	tinygo flash -target=pico main.go

install:
	sudo apt-get install gcc-arm-linux-gnueabihf
	wget https://github.com/tinygo-org/tinygo/releases/download/v0.33.0/tinygo_0.33.0_amd64.deb
	sudo dpkg -i tinygo_0.33.0_amd64.deb
	rm tinygo_0.33.0_amd64.deb

flash-for-wsl:
	make build
	
	# BOOTモードに移行させる。
	make _bootmode-for-wsl
	
	# windows側のドライブをマウントする。
	# このドライブレターは各々異なるので書き換える必要がある。
	-@sudo mount -t drvfs E: /mnt/e
	# 書き込む
	-@cp uf2/swiper.uf2 /mnt/e

flash-and-monitor-for-wsl:
	make flash-for-wsl
	# モニターを起動する。
	make _monitor

monitor-for-wsl:
	make attach-for-wsl
	make _monitor

# _ がついているものは外から呼ぶこと想定していない。呼んでもよい。

_bootmode-for-wsl:
	# BOOTモードでない状態でwindows側につながっている前提で、BOOTモードに移行させる。
	# 既にSharedな状態かAttachedな状態である前提。
	#-@set -x ; p=`"/mnt/c/Program Files/usbipd-win/usbipd.exe" list -u | grep "Shared" | awk '{print $$1}'`; "/mnt/c/Program Files/usbipd-win/usbipd.exe" bind -b $$p
	# SharedとAttachedの部分に注意。
	# attach するとusbipd list -uは下記のようになる。
	# 9-3    2e8a:000a  Unknown device                                                Attached
	#-@set -x ; p=`"/mnt/c/Program Files/usbipd-win/usbipd.exe" list -u | grep -e "Shared" -e "Attached" | awk '{print $$1}'`; "/mnt/c/Program Files/usbipd-win/usbipd.exe" attach --wsl --busid $$p
	make attach-for-wsl
	# BOOTモードに移行させるために一旦tinygo flashを実行する。
	# ポートの番号は変わる可能性がある。
	# 現状wslからはtinygo flashで書き換えできない模様。linux kernelを再ビルドするとできるらしいがうまくいかなかった。
	-@tinygo flash -target=pico -port /dev/ttyACM`sudo dmesg | grep ttyACM | tail -n1 | sed 's/.*ttyACM//; s/:.*//'` >/dev/null &
	# AttachしたままBootモードに入っているはずなので detachしてwindows側に返す。
	-@set -x ; p=`"/mnt/c/Program Files/usbipd-win/usbipd.exe" list | grep "RP2 Boot" | awk '{print $$1}'`; "/mnt/c/Program Files/usbipd-win/usbipd.exe" detach --busid $$p
	# 待たないとwindows側が認識しない。5秒はいらないと思われる。
	sleep 3

_monitor:
	-@tinygo monitor -target=pico -baudrate=9600

attach-for-wsl:
	# 既にAttachedかもしれないが無視する。
	-@set -x ; p=`"/mnt/c/Program Files/usbipd-win/usbipd.exe" list -u | grep -e "Shared" | awk '{print $$1}'`; "/mnt/c/Program Files/usbipd-win/usbipd.exe" attach --wsl --busid $$p
dettach-for-wsl:
	-@set -x ; p=`"/mnt/c/Program Files/usbipd-win/usbipd.exe" list -u | grep -e "Attached" | awk '{print $$1}'`; "/mnt/c/Program Files/usbipd-win/usbipd.exe" detach --busid $$p

