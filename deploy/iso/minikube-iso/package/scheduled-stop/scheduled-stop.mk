################################################################################
#
# minikube scheduled-stop
#
################################################################################

define SCHEDULED_STOP_INSTALL_INIT_SYSTEMD
	$(INSTALL) -D -m 644 \
		$(SCHEDULED_STOP_PKGDIR)/minikube-scheduled-stop.service \
		$(TARGET_DIR)/usr/lib/systemd/system/minikube-scheduled-stop.service

	mkdir -p $(TARGET_DIR)/etc/systemd/system/multi-user.target.wants
	ln -fs /usr/lib/systemd/system/minikube-scheduled-stop.service \
		$(TARGET_DIR)/etc/systemd/system/multi-user.target.wants/minikube-scheduled-stop.service
endef


$(eval $(generic-package))
