
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

install-usbipd-for-wsl:
	"/mnt/c/WINDOWS/System32/WindowsPowerShell/v1.0/powershell.exe" -Command "winget install usbipd"
	make list

bind-for-wsl:
	# 管理者権限が必要。
	# COMデバイスがraspberry pi pico 1つだけという前提のコード。
	p=`"/mnt/c/Program Files/usbipd-win/usbipd.exe" list | grep "COM" | awk '{print $$1}'`; \
	"/mnt/c/Program Files/usbipd-win/usbipd.exe" bind -b $$p
	make list

	# 前提として、windows側で、raspberry pi picoのみをsharedにしておく。
	# わかりやすい方法は一度BOOTSELを押さずにwindowsにつないで、usbipd listを実行し、raspberry pi picoを探す。
	# 分からない場合、usbipd list -u を実行して unknownと書いていあるものを探す。多分それ。探したら、1列目の10-3のような文字列がbusid。下記の@を置き換えて実行する。
	# usbipd bind @
	# これでSharedになるはず。
	# usbipd attach --wsl --busid @
	# を実行するとwslから見えるようになるはず。sudo dmesgなどを実行すると接続されていれば見える。
	# usbipd detach --busid @
	# を実行するとwindows側に戻す。

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

copy-and-monitor-for-wsl:
	make _copy-for-wsl
	make _monitor

flash-and-monitor-for-wsl:
	# 以降のコマンドのsudoでパスワードが聞かれないように入力しておく。
	sudo ls > /dev/null
	make build
	make copy-and-monitor-for-wsl

monitor-for-wsl:
	make _monitor

_copy-for-wsl:
	# BOOTモードに移行させる。
	make _bootmode-for-wsl
	make _mount-for-wsl
	d=`"/mnt/c/WINDOWS/System32/cmd.exe" /c "wmic logicaldisk get deviceid, volumename, description" 2>&1 | iconv -f SJIS -t UTF-8 | grep "RPI-RP2" | awk '{print $3}' | sed 's/://'`; \
		cp uf2/swiper.uf2 /mnt/$$d

_mount-for-wsl:
	# windows側のドライブをマウントする。
	# このドライブレターは各々異なるので書き換える必要がある。
	d=`"/mnt/c/WINDOWS/System32/cmd.exe" /c "wmic logicaldisk get deviceid, volumename, description" 2>&1 | iconv -f SJIS -t UTF-8 | grep "RPI-RP2" | awk '{print $$3}' | sed 's/://'` ; echo $$d; \
	d=`echo $$d | tr [:upper:] [:lower:]`; \
	sudo mount -t drvfs "$$d:" "/mnt/$$d"; \
	sleep 1; \
	sudo chmod -R 777 "/mnt/$$d"; \
	cp uf2/swiper.uf2 "/mnt/$$d"

	# 書き込む

# _ がついているものは外から呼ぶこと想定していない。呼んでもよい。

_bootmode-for-wsl:
	# BOOTモードでない状態でwindows側につながっている前提で、BOOTモードに移行させる。
	# 既にSharedな状態かAttachedな状態である前提。
	#-@set -x ; p=`"/mnt/c/Program Files/usbipd-win/usbipd.exe" list -u | grep "Shared" | awk '{print $$1}'`; "/mnt/c/Program Files/usbipd-win/usbipd.exe" bind -b $$p
	# SharedとAttachedの部分に注意。
	# attach するとusbipd list -uは下記のようになる。
	# 9-3    2e8a:000a  Unknown device                                                Attached
	#-@set -x ; p=`"/mnt/c/Program Files/usbipd-win/usbipd.exe" list -u | grep -e "Shared" -e "Attached" | awk '{print $$1}'`; "/mnt/c/Program Files/usbipd-win/usbipd.exe" attach --wsl --busid $$p
	-@make attach-for-wsl
	# BOOTモードに移行させるために一旦tinygo flashを実行する。
	# 2秒でタイムアウトする。
	# ポートの番号は変わる可能性がある。
	# 現状wslからはtinygo flashで書き換えできない模様。linux kernelを再ビルドするとできるらしいがうまくいかなかった。
	-@timeout 5 make _flash >/dev/null
	sleep 3
	# AttachしたままBootモードに入っているはずなので detachしてwindows側に返す。
	#-@set -x ; p=`"/mnt/c/Program Files/usbipd-win/usbipd.exe" list | grep "RP2 Boot" | awk '{print $$1}'`; "/mnt/c/Program Files/usbipd-win/usbipd.exe" detach --busid $$p
	-@make dettach-for-wsl
	# 待たないとwindows側が認識しない。5秒はいらないと思われる。
	sleep 5

cat:
	cat makefile | grep "^.*:$$"

_monitor:
	# モニターを起動する。
	make attach-for-wsl
	-@set -x ; sleep 1
	make attach-for-wsl
	-@set -x ; sleep 1
	-@set -x ; tinygo monitor -target=pico -baudrate=9600

_flash:
	-@make attach-for-wsl
	-@set -x ; tinygo flash -target=pico -port /dev/ttyACM`sudo dmesg | grep ttyACM | tail -n1 | sed 's/.*ttyACM//; s/:.*//'`

attach-for-wsl:
	############# attach-for-wsl #############
	-make list
	-set -x ; p=`"/mnt/c/Program Files/usbipd-win/usbipd.exe" list -u | grep -e "Shared" | awk '{print $$1}'`; if [ -n "$$p" ]; then "/mnt/c/Program Files/usbipd-win/usbipd.exe" attach --wsl --busid $$p; fi
	-make list
dettach-for-wsl:
	############# detach-for-wsl #############
	-@set -x ; "/mnt/c/Program Files/usbipd-win/usbipd.exe" list
	-set -x ; p=`"/mnt/c/Program Files/usbipd-win/usbipd.exe" list -u | grep -e "Attached" | awk '{print $$1}'`; if [ -n "$$p" ]; then "/mnt/c/Program Files/usbipd-win/usbipd.exe" detach --busid $$p; fi
	-sleep 2
	-@set -x ; "/mnt/c/Program Files/usbipd-win/usbipd.exe" list
list:
	-@set -x ; "/mnt/c/Program Files/usbipd-win/usbipd.exe" list
	make _list-grep
list-u:
	-@set -x ; "/mnt/c/Program Files/usbipd-win/usbipd.exe" list -u
	make _list-u-grep
_list-grep:
	@"/mnt/c/Program Files/usbipd-win/usbipd.exe" list | grep -e "Shared" -e "Attached"
_list-u-grep:
	@"/mnt/c/Program Files/usbipd-win/usbipd.exe" list -u | grep -e "Shared" -e "Attached"

