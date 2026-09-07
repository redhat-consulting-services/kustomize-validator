INTEGRATION_TEST_BASE_DIR := "./_tests"
INTEGRATION_TESTS_POSITIVE := success-basic
INTEGRATION_TESTS_NEGATIVE := failure-invalid-link failure-patch-me-check

integration-tests: integration-positive-tests integration-negative-tests

integration-positive-tests:
	@echo "Running integration positive tests..."
	for test in $(INTEGRATION_TESTS_POSITIVE); do \
		echo "Running $${test}..."; \
		go run main.go $(INTEGRATION_TEST_BASE_DIR)/$${test}; \
		exit_code=$$?; \
		if [ $$exit_code -eq 0 ]; then \
			echo "[OK] $${test} exited with code 0"; \
		else \
			echo "[ERROR] $${test} exited with code $${exit_code}"; \
			exit $$exit_code; \
		fi; \
	done

integration-negative-tests:
	@echo "Running integration negative tests..."
	for test in $(INTEGRATION_TESTS_NEGATIVE); do \
		echo "Running $${test}..."; \
		go run main.go $(INTEGRATION_TEST_BASE_DIR)/$${test}; \
		exit_code=$$?; \
		if [ $$exit_code -ne 0 ]; then \
			echo "[OK] $${test} exited with code $${exit_code}"; \
		else \
			echo "[ERROR] $${test} exited with code 0"; \
			exit 1; \
		fi; \
	done
