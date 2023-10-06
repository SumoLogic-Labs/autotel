package main

import (
  log "github.com/sirupsen/logrus"
)

func test_logrus() {
  log.WithFields(log.Fields{
    "animal": "walrus",
  }).Info("A walrus appears")
}