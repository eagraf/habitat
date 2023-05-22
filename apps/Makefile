SUBDIRS	=	ipfs-driver
APPS	=	$(SUBDIRS)

include ../common.mk

install::
	for dir in $(APPS); do \
		mkdir -p $(DEV_APPDIR)/$$dir/bin/amd64-linux ; \
		mkdir -p $(DEV_APPDIR)/$$dir/web ; \
		$(CP) $$dir/bin/amd64-linux/* $(DEV_APPDIR)/$$dir/bin/amd64-linux ; \
		if [ -d $$dir/web ] ; then $(CP) -r $$dir/web/* $(DEV_APPDIR)/$$dir/web | true ; fi ; \
		$(CP) $$dir/habitat.yaml $(DEV_APPDIR)/$$dir ; \
	done
