TARGET = myra-dyn

GO      = go
GOLINT  = $(GOPATH)/bin/golint
GO_SUBPKGS = $(shell $(GO) list ./... | grep -v /vendor/ | sed -e "s!$$($(GO) list)!.!")

BINDIR = ./bin/

UPX := $(shell upx --version 2>/dev/null)

all: $(TARGET)

$(TARGET): build
ifdef UPX
	upx --brute $(BINDIR)$@
endif

build: clean $(BINDIR)
	$(GO) build -ldflags="-s -w" -o $(BINDIR)$(TARGET) ./*go

clean:
	rm -f $(BINDIR)*

test:
	$(GO) test -race $$($(GO) list ./...)

$(BINDIR):
	mkdir -p $(BINDIR)

lint: $(GOLINT)
	@for f in $(GO_SUBPKGS) ; do $(GOLINT) $$f ; done

.PHONY:test lint vendor 