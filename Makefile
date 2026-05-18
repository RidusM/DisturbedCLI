.PHONY: build run test bench clean demo-local demo-distributed

BINARY := grep
ADDR1  := :5001
ADDR2  := :5002
ADDR3  := :5003
PEERS  := $(ADDR1),$(ADDR2),$(ADDR3)

build:
	go build -o $(BINARY) ./cmd/grep

test:
	go test ./internal/... -v -race

bench:
	go test ./benchmark/ -bench=. -benchtime=5s -benchmem

clean:
	rm -f $(BINARY)

demo-local: build
	echo "=== 4-worker local mode ===" && \
	seq 1 100000 | awk 'NR%7==0{print "ERROR line "NR} NR%7!=0{print "ok line "NR}' | \
	./$(BINARY) -workers 4 -n "ERROR" | tail -5

demo-distributed: build
	./$(BINARY) -addr $(ADDR1) & \
	./$(BINARY) -addr $(ADDR2) & \
	./$(BINARY) -addr $(ADDR3) & \
	sleep 0.3 && \
	echo "=== Distributed grep (quorum 2/3) ===" && \
	seq 1 10 | awk 'NR%3==0{print "ERROR item "NR} NR%3!=0{print "ok item "NR}' | \
	./$(BINARY) -peers $(PEERS) -quorum 2 "ERROR" ; \
	kill %1 %2 %3 2>/dev/null || true