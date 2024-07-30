SRC_DIR = src/
DEST_DIR = /etc/dockerspy
BIN_DIR = /usr/local/bin

GREEN = \033[0;32m
YELLOW = \033[0;33m
BLUE = \033[0;34m
NC = \033[0m

all: build cleanup

build: copy-deps
	@echo "$(BLUE)[#] Building the dockerspy binary...$(NC)"
	sudo go build -o dockerspy
	@echo "$(BLUE)[#] Copying the dockerspy binary to $(BIN_DIR)...$(NC)"
	sudo cp dockerspy $(BIN_DIR)
	sudo chmod +x $(BIN_DIR)/dockerspy

copy-deps:
	@echo "$(YELLOW)[#] Copying configuration files...$(NC)"
	sudo mkdir -p $(DEST_DIR)
	sudo cp -R $(SRC_DIR)* $(DEST_DIR)

cleanup:
	@echo "$(BLUE)[#] Cleaning up...$(NC)"
	sudo rm -f dockerspy

.PHONY: all build copy-deps cleanup

all: build cleanup
	@echo "$(GREEN)[###] Build complete. You can now run dockerspy from the terminal.$(NC)"
