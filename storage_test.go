package main

import (
	"testing"
)

func TestNewStorage(t *testing.T) {
	target := "mongodb://localhost:27017"
	source := "mongodb://localhost:27018"
	s := newStorage(target, source)
	if s.targetURI != target {
		t.Errorf("expected targetURI %s, got %s", target, s.targetURI)
	}
	if s.sourceURI != source {
		t.Errorf("expected sourceURI %s, got %s", source, s.sourceURI)
	}
}

func TestGetTargetDatabases_Error(t *testing.T) {
	s := newStorage("mongodb://invalid:1234", "mongodb://invalid:1234")
	_, err := s.getTargetDatabases()
	if err == nil {
		t.Error("expected error for invalid target URI")
	}
}

func TestGetSourceDatabases_Error(t *testing.T) {
	s := newStorage("mongodb://invalid:1234", "mongodb://invalid:1234")
	_, err := s.getSourceDatabases()
	if err == nil {
		t.Error("expected error for invalid source URI")
	}
}

func TestGetTargetCollections_Error(t *testing.T) {
	s := newStorage("mongodb://invalid:1234", "mongodb://invalid:1234")
	_, err := s.getTargetCollections("testdb")
	if err == nil {
		t.Error("expected error for invalid target URI")
	}
}

func TestGetSourceCollections_Error(t *testing.T) {
	s := newStorage("mongodb://invalid:1234", "mongodb://invalid:1234")
	_, err := s.getSourceCollections("testdb")
	if err == nil {
		t.Error("expected error for invalid source URI")
	}
}

func TestCopy_Error(t *testing.T) {
	s := newStorage("mongodb://invalid:1234", "mongodb://invalid:1234")
	err := s.copy("srcCol", "tgtCol", "srcDB", "tgtDB")
	if err == nil {
		t.Error("expected error for invalid URIs in copy")
	}
}
