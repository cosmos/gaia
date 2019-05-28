package main

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gaia/lcd_test"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	kb, err := keys.NewKeyBaseFromDir(lcdtest.InitClientHome(""))
	if err != nil {
		panic(err)
	}
	addr, _, err := lcdtest.CreateAddr("sender", "1234567890", kb)
	addr2, _, err := lcdtest.CreateAddr("receiver", "1234567890", kb)
	if err != nil {
		panic(err)
	}

	cleanup, _, _, _ := lcdtest.InitializeLCD(3, []sdk.AccAddress{addr, addr2}, true, "58645")
	defer cleanup()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, os.Interrupt, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGHUP)
	go func() {
		sig := <-sigs
		fmt.Println("Received", sig)
		done <- true
	}()

	fmt.Println("REST server running")
	<-done
	fmt.Println("exiting")
}
