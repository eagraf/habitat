run:
	HABITAT_PATH=$(DEV_HABITAT_PATH) HABITAT_APP_PATH=$(DEV_HABITAT_APP_PATH) $(BINDIR)/habitat --hostname localhost

docker-build:
	docker build -t habitat .
