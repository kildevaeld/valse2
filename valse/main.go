package main

import (
	"fmt"
	"net/http"
	"os"

	system "github.com/kildevaeld/go-system"
	"github.com/kildevaeld/valse2"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

func main() {

	if err := system.Run(wrappedMain); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}

}

func wrappedMain(kill system.KillChannel) error {

	address := pflag.StringP("address", "H", ":3000", "address")
	debug := pflag.BoolP("debug", "d", false, "debug")
	//workqueue := pflag.IntP("workqueue", "w", 10, "")

	pflag.Parse()

	if *debug {
		log, err := zap.NewDevelopment()
		if err != nil {
			return err
		}

		zap.ReplaceGlobals(log)

	}

	server := valse2.NewWithOptions(&valse2.Options{
		Debug: true,
	})

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	server.ServeFiles("/*filepath", http.Dir(cwd))
	zap.L().Info("Started server and listening", zap.String("address", *address))
	return server.Listen(*address)
}

// func main() {

// 	status, err := wrappedMain()

// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "%s\n", err)
// 	}
// 	os.Exit(status)
// }

// func wrappedMain() (int, error) {

// 	address := pflag.StringP("address", "H", ":3000", "address")
// 	debug := pflag.BoolP("debug", "d", false, "debug")
// 	workqueue := pflag.IntP("workqueue", "w", 10, "")

// 	pflag.Parse()

// 	if *debug {
// 		logrus.SetLevel(logrus.DebugLevel)
// 	}

// 	// log, err := zap.NewDevelopment()
// 	// if err != nil {
// 	// 	return 0, err
// 	// }

// 	// zap.ReplaceGlobals(log)

// 	server := valse2.NewWithOptions(&valse2.Options{
// 		Debug: true,
// 	})

// 	// server.Use(logger.Logger())

// 	l := lua.New(server, lua.LuaOptions{
// 		Path:      ".",
// 		WorkQueue: *workqueue,
// 	})

// 	if err := l.Open(); err != nil {
// 		return 200, nil
// 	}

// 	defer l.Close()

// 	if err := wait(server, *address); err != nil {
// 		return -1, err
// 	}

// 	return 0, nil
// }

// func wait(serv *valse2.Valse, addr string) error {

// 	signal_chan := make(chan os.Signal, 1)
// 	signal.Notify(signal_chan,
// 		syscall.SIGHUP,
// 		syscall.SIGINT,
// 		syscall.SIGTERM,
// 		syscall.SIGQUIT)

// 	exit_chan := make(chan error)

// 	go func() {
// 		logrus.Printf("Valse started on: '%s'", addr)
// 		exit_chan <- serv.Listen(addr)
// 	}()

// 	go func() {
// 		signal := <-signal_chan
// 		logrus.Printf("Signal %s. Existing...", signal)
// 		exit_chan <- nil //serv.Close()
// 	}()

// 	err := <-exit_chan

// 	signal.Stop(signal_chan)
// 	close(signal_chan)
// 	close(exit_chan)

// 	return err
// }
