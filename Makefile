# This makefile is only for manual dev purposes

DEV_VOLUME = l2radar-dev-data
EXPORT_DIR = /var/lib/l2radar

all: build start

build:
	@docker build -t l2radar_dev -f probe/Dockerfile probe/
	@docker build -t l2radar-ui_dev -f ui/Dockerfile .

start:
	@docker run -d --rm --privileged --network=host \
		-v /sys/fs/bpf:/sys/fs/bpf 		\
		-v $(DEV_VOLUME):$(EXPORT_DIR)		\
		--name l2radar_dev l2radar_dev 		\
		--iface external --export-dir $(EXPORT_DIR)
	@docker run -d --rm				\
		-v $(DEV_VOLUME):$(EXPORT_DIR):ro	\
		--name l2radar-ui_dev			\
		-p 127.0.0.1:12553:443 l2radar-ui_dev
	@echo "UI at: https://localhost:12553 (admin:changeme)"
stop:
	@docker stop l2radar_dev || true
	@docker stop l2radar-ui_dev || true
	@docker volume rm $(DEV_VOLUME) 2>/dev/null || true
