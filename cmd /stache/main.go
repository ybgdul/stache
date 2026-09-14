package main

import (
	"fmt"
	"net"
	"os"
	"stache/internal/protocol"
	"time"
)

func main() { 
	if len(os.Args) < 2 { 
		printUsage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	socketPath := "/tmp/stache.sock"

	conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)

	if err != nil {
		fmt.Printf("Error: couldn't connect to daemon at %s: %v\n", socketPath, err)
		os.Exit(1)
	}
	defer conn.Close()

	switch cmd {
	case "clear": 
		req := &protocol.Request{
			Cmd: protocol.CmdClear,
		}
		if err := protocol.WriteRequest(conn, req); err != nil { 
			fmt.Printf("Error sending command: %v\n", err)
			os.Exit(1)
		}

		resp, err := protocol.ReadResponse(conn)
		if err != nil { 
			fmt.Printf("Error reading response: %v\n", err)
			os.Exit(1)
		}

		if resp.Status == protocol.StatusOk {
			fmt.Println("Cache cleared succesfully")
		} else { 
			fmt.Printf("Failed to clear cache: %v\n", resp.Err)
		}

	case "status": 
		fmt.Println("stache daemon is running and reachable")
	default: 
		fmt.Println("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

}

func printUsage() { 
	fmt.Println("Helper:")
	fmt.Println("Usage: stache <command>")
	fmt.Println("Commands:")
	fmt.Println("  status  Check connectivity")
	fmt.Println("  clear   Clear all cached items")
}