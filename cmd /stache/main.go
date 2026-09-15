package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"stache/internal/engine"
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
	case "stats": 
		req := &protocol.Request{
			Cmd: protocol.CmdStats,
		}
		if err := protocol.WriteRequest(conn, req); err != nil { 
			fmt.Printf("Error sending command: %v\n", err)
			os.Exit(1)
		}

		resp, err := protocol.ReadResponse(conn)
		if err != nil { 
			fmt.Printf("Error reading response: %v\n", err)
		}

		reader := bytes.NewReader(resp.Value)
		
		var stats engine.Stats
		unmarshalErr := binary.Read(reader, binary.BigEndian, &stats)
		if unmarshalErr != nil { 
			fmt.Printf("Error reading stats: %v\n", err)
		}

		fmt.Printf("Hits : %d\n", stats.Hits)
		fmt.Printf("Misses : %d\n", stats.Misses)
		fmt.Printf("Item Count : %d\n", stats.ItemCount)
		fmt.Printf("Current Bytes : %d\n", stats.CurrentBytes)
		fmt.Printf("Maximum Memory : %d\n", stats.MaxMemory)


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