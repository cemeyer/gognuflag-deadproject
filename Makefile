include $(GOROOT)/src/Make.$(GOARCH)

TARG=gnuflag
GOFILES=\
	gnuflag.go

include $(GOROOT)/src/Make.pkg
