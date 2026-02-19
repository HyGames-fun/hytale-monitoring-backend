package domain

import (
	"testing"
)

func TestCreateServerDTO_Validation(t *testing.T) {
	ip := "127.0.0.1:25565"
	valid := CreateServerDTO{
		IP:          &ip,
		Name:        "Test Hytale Server",
		NameID:      "test-hytale-server",
		Description: "A great adventure server for all players to join and enjoy.",
		Tags:        []string{"PVP", "SURVIVAL"},
		Region:      "CIS",
	}

	errs := valid.Validate()
	if len(errs) > 0 {
		t.Fatalf("expected valid server to pass, got errors: %+v", errs)
	}

	invalid := CreateServerDTO{
		Name:        "  bad trimmed  ",
		NameID:      "Bad_Kebab_Case!",
		Description: "short",
		Tags:        []string{"UNKNOWN_TAG"},
		Region:      "MARS",
	}

	errs = invalid.Validate()
	if len(errs) == 0 {
		t.Fatal("expected validation errors for invalid server")
	}

	if _, ok := errs["_atLeastOne"]; !ok {
		t.Error("expected error for missing ip/domain")
	}
	if _, ok := errs["nameId"]; !ok {
		t.Error("expected error for invalid kebab-case nameId")
	}
	if _, ok := errs["region"]; !ok {
		t.Error("expected error for invalid region")
	}
	if _, ok := errs["tags"]; !ok {
		t.Error("expected error for invalid tag")
	}
}

func TestRegisterDTO_Validation(t *testing.T) {
	valid := RegisterDTO{
		Name:     "Snuslyk",
		Email:    "test@example.com",
		Password: "strongpassword123",
	}
	if errs := valid.Validate(); len(errs) > 0 {
		t.Fatalf("expected valid register DTO to pass, got: %+v", errs)
	}

	invalid := RegisterDTO{
		Name:     "ab",
		Email:    "notanemail",
		Password: "123",
	}
	errs := invalid.Validate()
	if len(errs) != 3 {
		t.Fatalf("expected 3 errors, got: %+v", errs)
	}
}
