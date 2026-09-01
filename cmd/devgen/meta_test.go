package main

import (
	"runtime/debug"
	"testing"
	"time"

	"github.com/xoctopus/confx/pkg/appx"
	. "github.com/xoctopus/x/testx"
)

func TestFillMetaFromBuildVCSSettings(t *testing.T) {
	m := appx.Meta{Name: "devgen"}
	fillMetaFromBuild(&m, "(devel)", []debug.BuildSetting{
		{Key: "vcs.revision", Value: "fb41384455793d83ff32b2e4740db15d8cd65cdb"},
		{Key: "vcs.time", Value: "2026-09-01T05:50:33Z"},
		{Key: "vcs.modified", Value: "true"},
	})

	Expect(t, m.Version, Equal(""))
	Expect(t, m.CommitID, Equal("fb4138445579-dirty"))
	Expect(t, m.CommitAt, Equal("20260901135033CST"))
}

func TestFormatTimestampCST(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2026-09-01T05:50:33Z")
	Expect(t, err, Succeed())
	Expect(t, formatTimestampCST(ts), Equal("20260901135033CST"))
	Expect(t, formatPseudoCommitAt("20260901055033"), Equal("20260901135033CST"))
}

func TestFillMetaFromBuildPseudoVersion(t *testing.T) {
	m := appx.Meta{Name: "devgen"}
	fillMetaFromBuild(&m, "v0.6.7-0.20260901055033-fb4138445579", nil)

	Expect(t, m.Version, Equal("v0.6.7"))
	Expect(t, m.CommitID, Equal("fb4138445579"))
	Expect(t, m.CommitAt, Equal("20260901135033CST"))
}

func TestFillMetaFromBuildPreservesLDFlags(t *testing.T) {
	m := appx.Meta{
		Name:     "devgen",
		Feature:  "main",
		Version:  "v1.0.0",
		CommitID: "abc1234",
		CommitAt: "20260101010101",
		BuildAt:  "20260102020202",
	}
	fillMetaFromBuild(&m, "v0.6.7-0.20260901055033-fb4138445579", []debug.BuildSetting{
		{Key: "vcs.revision", Value: "ffffffffffffffffffffffffffffffffffffffff"},
		{Key: "vcs.time", Value: "2026-09-01T05:50:33Z"},
	})

	Expect(t, m.Feature, Equal("main"))
	Expect(t, m.Version, Equal("v1.0.0"))
	Expect(t, m.CommitID, Equal("abc1234"))
	Expect(t, m.CommitAt, Equal("20260101010101"))
	Expect(t, m.BuildAt, Equal("20260102020202"))
}

func TestFillMetaFromBuildUntaggedPseudoVersion(t *testing.T) {
	m := appx.Meta{Name: "devgen"}
	fillMetaFromBuild(&m, "v0.0.0-20260901055033-fb4138445579", nil)

	Expect(t, m.Version, Equal("v0.0.0"))
	Expect(t, m.CommitID, Equal("fb4138445579"))
	Expect(t, m.CommitAt, Equal("20260901135033CST"))
}

func TestFillMetaFromBuildPseudoVersionWithDirtySuffix(t *testing.T) {
	m := appx.Meta{Name: "devgen"}
	fillMetaFromBuild(&m, "v0.6.7-0.20260901055033-fb4138445579+dirty", nil)

	Expect(t, m.Version, Equal("v0.6.7"))
	Expect(t, m.CommitID, Equal("fb4138445579"))
	Expect(t, m.CommitAt, Equal("20260901135033CST"))
}

func TestFillMetaFromBuildPlainModuleVersion(t *testing.T) {
	m := appx.Meta{Name: "devgen"}
	fillMetaFromBuild(&m, "v0.6.7", nil)

	Expect(t, m.Version, Equal("v0.6.7"))
}
