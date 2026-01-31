package main

import (
	"reflect"
	"testing"
)

func TestParseBaseRecordFieldsFromAssignments(t *testing.T) {
	fields, err := parseBaseRecordFields("", []string{"Title=Task", "Meta:={\"done\":true}"})
	if err != nil {
		t.Fatalf("parse fields: %v", err)
	}
	if fields["Title"] != "Task" {
		t.Fatalf("unexpected Title: %#v", fields["Title"])
	}
	meta, ok := fields["Meta"].(map[string]any)
	if !ok || meta["done"] != true {
		t.Fatalf("unexpected Meta: %#v", fields["Meta"])
	}
}

func TestParseBaseRecordFieldsFromNameValue(t *testing.T) {
	fields, err := parseBaseRecordFields("", []string{"name=City,value=Beijing"})
	if err != nil {
		t.Fatalf("parse fields: %v", err)
	}
	want := map[string]any{"City": "Beijing"}
	if !reflect.DeepEqual(fields, want) {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}

func TestParseBaseRecordFieldsRejectsMixedInputs(t *testing.T) {
	_, err := parseBaseRecordFields(`{"Title":"Task"}`, []string{"Title=Task"})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "fields-json and field are mutually exclusive" {
		t.Fatalf("unexpected error: %v", err)
	}
}
