package authregistry

import (
	"reflect"
	"testing"
)

func TestRequirementsForCommandDrive(t *testing.T) {
	services, tokenTypes, offline, scopes, ok, err := RequirementsForCommand("drive list")
	if err != nil {
		t.Fatalf("RequirementsForCommand(drive list) err=%v", err)
	}
	if !ok {
		t.Fatalf("RequirementsForCommand(drive list) ok=false, want true")
	}

	if want := []string{"drive"}; !reflect.DeepEqual(services, want) {
		t.Fatalf("services=%v, want %v", services, want)
	}
	if want := []TokenType{TokenTenant, TokenUser}; !reflect.DeepEqual(tokenTypes, want) {
		t.Fatalf("tokenTypes=%v, want %v", tokenTypes, want)
	}
	if !offline {
		t.Fatalf("offline=false, want true")
	}
	if want := []string{"drive:drive"}; !reflect.DeepEqual(scopes, want) {
		t.Fatalf("scopes=%v, want %v", scopes, want)
	}
}

func TestRequirementsForCommandMail(t *testing.T) {
	services, tokenTypes, offline, scopes, ok, err := RequirementsForCommand("mail send")
	if err != nil {
		t.Fatalf("RequirementsForCommand(mail send) err=%v", err)
	}
	if !ok {
		t.Fatalf("RequirementsForCommand(mail send) ok=false, want true")
	}

	if want := []string{"mail-send"}; !reflect.DeepEqual(services, want) {
		t.Fatalf("services=%v, want %v", services, want)
	}
	if want := []TokenType{TokenUser}; !reflect.DeepEqual(tokenTypes, want) {
		t.Fatalf("tokenTypes=%v, want %v", tokenTypes, want)
	}
	if !offline {
		t.Fatalf("offline=false, want true")
	}
	if want := []string{"mail:user_mailbox.message:send"}; !reflect.DeepEqual(scopes, want) {
		t.Fatalf("scopes=%v, want %v", scopes, want)
	}
}

func TestRequirementsForCommandChatsAndMessagesMatch(t *testing.T) {
	svcs1, tts1, off1, scopes1, ok, err := RequirementsForCommand("chats list")
	if err != nil {
		t.Fatalf("RequirementsForCommand(chats list) err=%v", err)
	}
	if !ok {
		t.Fatalf("RequirementsForCommand(chats list) ok=false, want true")
	}

	svcs2, tts2, off2, scopes2, ok, err := RequirementsForCommand("messages send")
	if err != nil {
		t.Fatalf("RequirementsForCommand(messages send) err=%v", err)
	}
	if !ok {
		t.Fatalf("RequirementsForCommand(messages send) ok=false, want true")
	}

	if !reflect.DeepEqual(svcs1, []string{"im"}) {
		t.Fatalf("chats services=%v, want [im]", svcs1)
	}
	if !reflect.DeepEqual(tts1, []TokenType{TokenTenant}) {
		t.Fatalf("chats tokenTypes=%v, want [tenant]", tts1)
	}
	if off1 {
		t.Fatalf("chats offline=true, want false")
	}
	if len(scopes1) != 0 {
		t.Fatalf("chats scopes=%v, want empty", scopes1)
	}

	if !reflect.DeepEqual(svcs1, svcs2) || !reflect.DeepEqual(tts1, tts2) || off1 != off2 || !reflect.DeepEqual(scopes1, scopes2) {
		t.Fatalf("chats and messages requirements differ:\nchats:    services=%v tokenTypes=%v offline=%v scopes=%v\nmessages: services=%v tokenTypes=%v offline=%v scopes=%v", svcs1, tts1, off1, scopes1, svcs2, tts2, off2, scopes2)
	}

	// Backward-compatible alias support for auth explain / scripted calls.
	_, _, _, _, ok, err = RequirementsForCommand("msg send")
	if err != nil {
		t.Fatalf("RequirementsForCommand(msg send) err=%v", err)
	}
	if !ok {
		t.Fatalf("RequirementsForCommand(msg send) ok=false, want true")
	}
}

func TestRequirementsForCommandDeterministicSortedUnique(t *testing.T) {
	orig := commandServiceMap["drive"]
	commandServiceMap["drive"] = []string{"docs", "drive", "docs"}
	t.Cleanup(func() {
		commandServiceMap["drive"] = orig
	})

	services1, tokenTypes1, offline1, scopes1, ok, err := RequirementsForCommand("drive")
	if err != nil {
		t.Fatalf("RequirementsForCommand(drive) err=%v", err)
	}
	if !ok {
		t.Fatalf("RequirementsForCommand(drive) ok=false, want true")
	}
	services2, tokenTypes2, offline2, scopes2, ok, err := RequirementsForCommand("drive")
	if err != nil {
		t.Fatalf("RequirementsForCommand(drive) second call err=%v", err)
	}
	if !ok {
		t.Fatalf("RequirementsForCommand(drive) second call ok=false, want true")
	}

	wantServices := []string{"docs", "drive"}
	if !reflect.DeepEqual(services1, wantServices) {
		t.Fatalf("services=%v, want %v", services1, wantServices)
	}
	if want := []TokenType{TokenTenant, TokenUser}; !reflect.DeepEqual(tokenTypes1, want) {
		t.Fatalf("tokenTypes=%v, want %v", tokenTypes1, want)
	}
	if !offline1 {
		t.Fatalf("offline=false, want true")
	}
	if want := []string{"drive:drive"}; !reflect.DeepEqual(scopes1, want) {
		t.Fatalf("scopes=%v, want %v", scopes1, want)
	}

	if !reflect.DeepEqual(services1, services2) {
		t.Fatalf("services not deterministic: %v vs %v", services1, services2)
	}
	if !reflect.DeepEqual(tokenTypes1, tokenTypes2) {
		t.Fatalf("tokenTypes not deterministic: %v vs %v", tokenTypes1, tokenTypes2)
	}
	if offline1 != offline2 {
		t.Fatalf("offline not deterministic: %v vs %v", offline1, offline2)
	}
	if !reflect.DeepEqual(scopes1, scopes2) {
		t.Fatalf("scopes not deterministic: %v vs %v", scopes1, scopes2)
	}
}
