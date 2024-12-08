package mygenetics

import (
	"context"
	"os"
	"testing"
)

const (
	TestCodeLab1 = "WN0000T"
	TestCodeLab2 = "DX0000T"
	TestCodeLab3 = "VM0000T"
)

func TestClient(t *testing.T) {
	var (
		TestEmail    = os.Getenv("MYGENETICS_EMAIL")
		TestPassword = os.Getenv("MYGENETICS_PASSWORD")
	)
	if TestEmail == "" || TestPassword == "" {
		t.Skip("MYGENETICS_EMAIL or MYGENETICS_PASSWORD not set")
	}

	ctx := context.Background()

	tokens, err := DefaultClient.Authenticate(ctx, TestEmail, TestPassword)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("tokens: %+v", tokens)

	codelabs, err := DefaultClient.FetchCodelabs(ctx, AccessToken(tokens))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("codelabs: %+v", codelabs)

	report1, err := DefaultClient.FetchFeatures(
		ctx,
		AccessToken(tokens),
		TestCodeLab1,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("report: %s", report1)

	report2, err := DefaultClient.FetchFeatures(
		ctx,
		AccessToken(tokens),
		TestCodeLab2,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("report: %s", report2)

	report3, err := DefaultClient.FetchFeatures(
		ctx,
		AccessToken(tokens),
		TestCodeLab3,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("report: %s", report3)

	// tokens, err = DefaultClient.Refresh(ctx, RefreshToken(tokens))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Logf("tokens: %+v", tokens)
}
