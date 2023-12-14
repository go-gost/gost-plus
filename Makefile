.PHONY: linux

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build
    
win:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-H windowsgui"

arm:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build

android:
	gogio -x -work -target android -minsdk 22 -version 1 -appid gost.plus github.com/go-gost/gost-plus

clean:
	rm gost-plus.exe gost-plus gost-plus.apk
