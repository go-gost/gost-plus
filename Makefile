# https://gioui.org/doc/install

NAME=gost-plus
BINDIR=bin
VERSION=$(shell cat version/version.go | grep 'Version =' | sed 's/.*\"\(.*\)\".*/\1/g')
GOBUILD=CGO_ENABLED=0 go build --ldflags="-s -w" -v -x -a
GOFILES=*.go

PLATFORM_LIST = \
	linux-amd64

WINDOWS_ARCH_LIST = \
	windows-amd64

linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build --ldflags="-s -w" -v -x -a -o $(BINDIR)/$(NAME)-$@ $(GOFILES)
    
darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

# https://github.com/tc-hib/go-winres
windows-amd64: winres
	go-winres make 
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -H windowsgui" -o $(BINDIR)/$(NAME)-$@.exe $(GOFILES)

# go install gioui.org/cmd/gogio@latest
android:
	gogio -x -work -target android -minsdk 22 -version 1 -appid gost.plus -o $(BINDIR)/$(NAME)-$(VERSION).apk .

gz_releases=$(addsuffix .gz, $(PLATFORM_LIST))
zip_releases=$(addsuffix .zip, $(WINDOWS_ARCH_LIST))

$(gz_releases): %.gz : %
	chmod +x $(BINDIR)/$(NAME)-$(basename $@)
	gzip -f -S -$(VERSION).gz $(BINDIR)/$(NAME)-$(basename $@)

$(zip_releases): %.zip : %
	zip -m -j $(BINDIR)/$(NAME)-$(basename $@)-$(VERSION).zip $(BINDIR)/$(NAME)-$(basename $@).exe

releases: $(gz_releases) $(zip_releases)

clean:
	rm $(BINDIR)/*
