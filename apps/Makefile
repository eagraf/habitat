SUBDIRS	=	ipfs-driver

APPS	=	ipfs-driver

include ../common.mk

install::
	for dir in $(APPS); do \
		mkdir -p $(DEV_APPDIR)/$$dir/bin ; \
		mkdir -p $(DEV_APPDIR)/$$dir/web ; \
		$(CP) $$dir/bin/$(CURARCH)/* $(DEV_APPDIR)/$$dir/bin ; \
		if [ -d $$dir/web ] ; then $(CP) -r $$dir/web/* $(DEV_APPDIR)/$$dir/web | true ; fi ; \
		$(CP) $$dir/habitat.yaml $(DEV_APPDIR)/$$dir ; \
	done
