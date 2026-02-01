package config

import "testing"

func TestUserAccountBucketKey_DefaultsAndNormalizes(t *testing.T) {
	got := UserAccountBucketKey("app", "https://open.feishu.cn/open-apis/", "")
	want := "app|https://open.feishu.cn|default"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}

	gotDefault := UserAccountBucketKey("app", "https://open.feishu.cn", "default")
	if gotDefault != want {
		t.Fatalf("expected %q, got %q", want, gotDefault)
	}

	gotDefaultCase := UserAccountBucketKey("app", "https://open.feishu.cn", "Default")
	if gotDefaultCase != want {
		t.Fatalf("expected %q, got %q", want, gotDefaultCase)
	}

	gotCustomProfile := UserAccountBucketKey("app", "https://open.feishu.cn", "dev")
	wantCustom := "app|https://open.feishu.cn|dev"
	if gotCustomProfile != wantCustom {
		t.Fatalf("expected %q, got %q", wantCustom, gotCustomProfile)
	}
}
