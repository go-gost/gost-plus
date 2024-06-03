# https://gioui.org/doc/install

NAME=gost.plus
BINDIR=bin
VERSION=$(shell cat version/version.go | grep 'Version =' | sed 's/.*\"\(.*\)\".*/\1/g')
GOBUILD=CGO_ENABLED=0 go build --ldflags="-s -w" -v -x -a
GOFILES=*.go

PLATFORM_LIST = \
	linux-amd64 \
	# linux-arm64

WINDOWS_ARCH_LIST = \
	windows-amd64 \
	windows-arm64

linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build --ldflags="-s -w" -v -x -a -o $(BINDIR)/$(NAME)-$(VERSION)-$@ $(GOFILES)

linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build --ldflags="-s -w" -v -x -a -o $(BINDIR)/$(NAME)-$(VERSION)-$@ $(GOFILES)
    
darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$(VERSION)-$@ $(GOFILES)

darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$(VERSION)-$@ $(GOFILES)

# https://github.com/tc-hib/go-winres
windows-amd64: 
	GOOS=windows GOARCH=amd64 go-winres make --in winres/winres.json --out winres/rsrc
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -H windowsgui" -o $(BINDIR)/$(NAME)-$(VERSION)-$@.exe $(GOFILES)

windows-arm64: 
	GOOS=windows GOARCH=arm64 go-winres make --in winres/winres.json --out winres/rsrc
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w -H windowsgui" -o $(BINDIR)/$(NAME)-$(VERSION)-$@.exe $(GOFILES)

# go install gioui.org/cmd/gogio@latest
android:
	gogio -x -work -target android -minsdk 22 -version $(VERSION).8 -name GOST+ -signkey build/sign.keystore -signpass android -appid gost.plus -o $(BINDIR)/$(NAME)-$(VERSION).aab .
	gogio -x -work -target android -minsdk 22 -version $(VERSION).8 -name GOST+ -signkey build/sign.keystore -signpass android -appid gost.plus -o $(BINDIR)/$(NAME)-$(VERSION).apk .

gz_releases=$(addsuffix .gz, $(PLATFORM_LIST))
zip_releases=$(addsuffix .zip, $(WINDOWS_ARCH_LIST))

$(gz_releases): %.gz : %
	chmod +x $(BINDIR)/$(NAME)-$(VERSION)-$(basename $@)
	gzip -f -S .gz $(BINDIR)/$(NAME)-$(VERSION)-$(basename $@)

$(zip_releases): %.zip : %
	zip -m -j $(BINDIR)/$(NAME)-$(VERSION)-$(basename $@).zip $(BINDIR)/$(NAME)-$(VERSION)-$(basename $@).exe

releases: $(gz_releases) $(zip_releases) android

clean:
	rm *.syso -f
	rm $(BINDIR)/* -rf
