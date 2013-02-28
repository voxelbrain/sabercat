package main

import (
	"fmt"
	"strings"

	"labix.org/v2/mgo"
)

// The implementation of this type is huge because mgo uses an *unexported*
// type `mode` to set the connection mode. Since it is not exported, I cannot
// have variables of this type nor cast to it.

type MgoConsistencyMode int

func (mcm *MgoConsistencyMode) MarshalGoptions(val string) error {
	switch strings.ToUpper(val) {
	case "STRONG":
		*mcm = MgoConsistencyMode(mgo.Strong)
	case "MONOTONIC":
		*mcm = MgoConsistencyMode(mgo.Monotonic)
	case "EVENTUAL":
		*mcm = MgoConsistencyMode(mgo.Eventual)
	default:
		return fmt.Errorf("Unknown consistency type \"%s\"", val)
	}
	return nil
}

func (mcm *MgoConsistencyMode) String() string {
	switch MgoConsistencyMode(*mcm) {
	case MgoConsistencyMode(mgo.Strong):
		return "STRONG"
	case MgoConsistencyMode(mgo.Monotonic):
		return "MONOTONIC"
	case MgoConsistencyMode(mgo.Eventual):
		return "EVENTUAL"
	}
	return "INVALID"
}

func (mcm *MgoConsistencyMode) Apply(s *mgo.Session) {
	switch MgoConsistencyMode(*mcm) {
	case MgoConsistencyMode(mgo.Strong):
		s.SetMode(mgo.Strong, true)
	case MgoConsistencyMode(mgo.Monotonic):
		s.SetMode(mgo.Monotonic, true)
	case MgoConsistencyMode(mgo.Eventual):
		s.SetMode(mgo.Eventual, true)
	}
}
