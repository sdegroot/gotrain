package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rijdendetreinen/gotrain/receiver"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var archiverCommand = &cobra.Command{
	Use:   "archiver",
	Short: "Start archiver",
	Long:  `Start the GoTrain archiver. It receives data and pushes processed data to the archive queue.`,
	Run: func(cmd *cobra.Command, args []string) {
		startArchiver(cmd)
	},
}

func init() {
	RootCmd.AddCommand(archiverCommand)
}

var exitArchiverReceiverChannel = make(chan bool)

func startArchiver(cmd *cobra.Command) {
	initLogger(cmd)

	log.Infof("GoTrain archiver %v starting", Version.VersionStringLong())

	signalChan := make(chan os.Signal, 1)
	shutdownArchiverFinished := make(chan struct{})

	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		log.Warnf("Received signal: %+v, shutting down", sig)
		signal.Reset()
		shutdownArchiver()
		close(shutdownArchiverFinished)
	}()

	receiver.ProcessStores = false
	receiver.ArchiveServices = true

	go receiver.ReceiveData(exitArchiverReceiverChannel)

	<-shutdownArchiverFinished
	log.Warn("Exiting")
}

func shutdownArchiver() {
	log.Warn("Shutting down")

	exitArchiverReceiverChannel <- true

	<-exitArchiverReceiverChannel
}
