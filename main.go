package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	GREEN  = "\033[1;32m"
	BLUE   = "\033[1;34m"
	PURPLE = "\033[1;35m"
	RED    = "\033[1;31m"
	RESET  = "\033[0;0m"

	banner = `
▗▄▄▄▖▗▖ ▗▖ ▗▄▄▖    ▗▖ ▗▖ ▄▄▄  █ ▄▄▄▄  ▗▞▀▚▖ ▄▄▄
▐▌   ▐▌▗▞▘▐▌       ▐▌ ▐▌█   █ █ █ █ █ ▐▛▀▀▘▀▄▄
▐▛▀▀▘▐▛▚▖  ▝▀▚▖    ▐▛▀▜▌▀▄▄▄▀ █ █   █ ▝▚▄▄▖▄▄▄▀
▐▙▄▄▖▐▌ ▐▌▗▄▄▞▘    ▐▌ ▐▌      █

    Author: Abhijeet Kumar
 Github: github.com/wand3rlust
`
	goodbye = `
▂▃▄▅▆▇█▓▒░G00dBye!░▒▓█▇▆▅▄▃▂
`
)

func showBanner() {
	if os.Getenv("NO_COLOR") != "" {
		fmt.Println(banner)
		return
	}
	fmt.Println(PURPLE + banner + RESET)
}

func colorize(text, color string) string {
	if os.Getenv("NO_COLOR") != "" {
		return text
	}
	return color + text + RESET
}

func showMenu() {
	fmt.Println(colorize("1. EKS API Server Enumeration", BLUE))
	fmt.Println(colorize("2. Generate Kubeconfig", BLUE))
	fmt.Println(colorize("0. Exit", BLUE))
	fmt.Print(colorize("\nSelect an option: ", BLUE))
}

func getUserInput() (int, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf(colorize("[X] Invalid input: "+input, RED))
	}
	return choice, nil
}

func main() {
	showBanner()
	fmt.Println("	     Version: 1.0.0\n\n")
	for {
		showMenu()
		choice, err := getUserInput()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		switch choice {
		case 1:
			eksAPIEnum()
		case 2:
			kubeconfigGenerator()
		case 0:
			fmt.Println(goodbye)
			os.Exit(0)
		default:
			fmt.Println(colorize("\n[X] Invalid option: "+strconv.Itoa(choice)+". Please try again.", RED))

		}
		fmt.Print(colorize("\n[+] Press Enter to continue...\n", BLUE))
		bufio.NewReader(os.Stdin).ReadString('\n')
	}
}
