# Makefile for running distributed cache servers

# Define the ports and peers for each server
SERVER1_PORT=:8060
SERVER1_PEERS=http://localhost:8061

SERVER2_PORT=:8061
SERVER2_PEERS=http://localhost:8060

SERVER3_PORT=:8062
SERVER3_PEERS=http://localhost:8060

# SERVER4_PORT=:8063
# SERVER4_PEERS=http://localhost:8060,http://localhost:8061,http://localhost:8062

# Target to run server 1
server1:
	go run main.go -port=$(SERVER1_PORT) -peers=$(SERVER1_PEERS)

# Target to run server 2
server2:
	go run main.go -port=$(SERVER2_PORT) -peers=$(SERVER2_PEERS)

# Target to run server 3
server3:
	go run main.go -port=$(SERVER3_PORT) -peers=$(SERVER3_PEERS)

# Target to run all servers
all: server1 server2 server3

# Target to clean up (if needed)
clean:
	@echo "Cleaning up..."
	# Add any cleanup commands here

.PHONY: server1 server2 server3 all clean
