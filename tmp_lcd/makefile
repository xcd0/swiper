
all:
	make build
build:
	go generate
	@mkdir -p uf2
	tinygo build -target=pico -o uf2/swiper.uf2
flash:
	go generate
	tinygo flash -target=pico

install:
	sudo apt-get install gcc-arm-linux-gnueabihf
	wget https://github.com/tinygo-org/tinygo/releases/download/v0.33.0/tinygo_0.33.0_amd64.deb
	sudo dpkg -i tinygo_0.33.0_amd64.deb
	rm tinygo_0.33.0_amd64.deb

win:
	"/mnt/c/WINDOWS/System32/WindowsPowerShell/v1.0/powershell.exe" -Command "cp ./uf2/swiper.uf2 ~/work/tmp"

#./lib/mbedtls/include/mbedtls/version.h

on-wsl:
	#p=`"/mnt/c/Program Files/usbipd-win/usbipd.exe" list | grep RP2 | awk '{print $$1}'`; \
	
	# 前提として、windows側で、raspberry pi picoのみをsharedにしておく。
	# わかりやすい方法は一度BOOTSELを押さずにwindowsにつないで、usbipd listを実行し、raspberry pi picoを探す。
	# 分からない場合、usbipd list -u を実行して unknownと書いていあるものを探す。多分それ。探したら、1列目の10-3のような文字列がbusid。下記の@を置き換えて実行する。
	# usbipd bind @
	# これでSharedになるはず。
	# usbipd attach --wsl --busid @
	# を実行するとwslから見えるようになるはず。sudo dmesgなどを実行すると接続されていれば見える。
	# usbipd detach --busid @
	# を実行するとwindows側に戻す。
	set -x ; \
		p=`"/mnt/c/Program Files/usbipd-win/usbipd.exe" list -u| grep "Shared" | awk '{print $$1}'`; \
		"/mnt/c/Program Files/usbipd-win/usbipd.exe" bind -b $$p ; \
		"/mnt/c/Program Files/usbipd-win/usbipd.exe" attach --wsl --busid $$p; \
		"/mnt/c/Program Files/usbipd-win/usbipd.exe" detach --busid $$p

	# BOOTモードでattachするとdmesgでは
	# [602532.717276] usb 1-1: New USB device found, idVendor=2e8a, idProduct=0003, bcdDevice= 1.00
	# [602532.717281] usb 1-1: New USB device strings: Mfr=1, Product=2, SerialNumber=3
	# [602532.717282] usb 1-1: Product: RP2 Boot
	# [602532.717284] usb 1-1: Manufacturer: Raspberry Pi
	# のように出るが、書き込めない。
	# usbipd listのように-uをつけずに実行するとBOOTモードなら下記のようにRP2 Bootの記述がある。
	# 9-3    2e8a:0003  USB 大容量記憶装置, RP2 Boot                                  Attached
	# BOOTモードでないとき、COM番号の記述がある。
	# 9-3    2e8a:000a  USB シリアル デバイス (COM8)                                  Shared


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

