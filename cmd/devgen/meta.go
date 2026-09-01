package main

import (
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/xoctopus/confx/pkg/appx"
)

const (
	metaTimeLayout = "20060102150405"
	metaTimeSuffix = "CST"
	metaTimeZone   = "Asia/Shanghai"
)

var pseudoVersionRE = regexp.MustCompile(`^v(\d+\.\d+\.\d+)-(?:\d+\.)?(\d{14})-([0-9a-f]+)$`)

func buildMeta(name, feature, version, commitID, commitAt, buildAt string) appx.Meta {
	m := appx.Meta{
		Name:     name,
		Feature:  feature,
		Version:  version,
		CommitID: commitID,
		CommitAt: commitAt,
		BuildAt:  buildAt,
		Runtime:  appx.GetRuntime(),
	}
	if m.Name == "" {
		m.Name = "devgen"
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		fillMetaFromBuild(&m, info.Main.Version, info.Settings)
	}
	return m
}

func fillMetaFromBuild(m *appx.Meta, mainVersion string, settings []debug.BuildSetting) {
	mainVersion = stripModuleBuildSuffix(mainVersion)

	var vcsRev, vcsTime, vcsModified string
	for _, s := range settings {
		switch s.Key {
		case "vcs.revision":
			vcsRev = s.Value
		case "vcs.time":
			vcsTime = s.Value
		case "vcs.modified":
			vcsModified = s.Value
		}
	}

	if m.CommitID == "" && vcsRev != "" {
		m.CommitID = shortRevision(vcsRev)
	}
	if m.CommitAt == "" && vcsTime != "" {
		m.CommitAt = formatCommitAt(vcsTime)
	}

	applyPseudoVersion(m, mainVersion)

	if m.Version == "" && mainVersion != "" && mainVersion != "(devel)" {
		m.Version = mainVersion
	}

	if vcsModified == "true" && m.CommitID != "" && !strings.HasSuffix(m.CommitID, "-dirty") {
		m.CommitID += "-dirty"
	}
}

func applyPseudoVersion(m *appx.Meta, pseudo string) {
	sub := pseudoVersionRE.FindStringSubmatch(pseudo)
	if sub == nil {
		return
	}
	if m.Version == "" {
		m.Version = "v" + sub[1]
	}
	if m.CommitAt == "" {
		m.CommitAt = formatPseudoCommitAt(sub[2])
	}
	if m.CommitID == "" {
		m.CommitID = sub[3]
	}
}

func shortRevision(rev string) string {
	if len(rev) > 12 {
		return rev[:12]
	}
	return rev
}

func formatCommitAt(vcsTime string) string {
	t, err := time.Parse(time.RFC3339, vcsTime)
	if err != nil {
		return ""
	}
	return formatTimestampCST(t)
}

func formatPseudoCommitAt(ts string) string {
	t, err := time.ParseInLocation(metaTimeLayout, ts, time.UTC)
	if err != nil {
		return ts + metaTimeSuffix
	}
	return formatTimestampCST(t)
}

func formatTimestampCST(t time.Time) string {
	loc, err := time.LoadLocation(metaTimeZone)
	if err != nil {
		loc = time.FixedZone(metaTimeSuffix, 8*3600)
	}
	return t.In(loc).Format(metaTimeLayout) + metaTimeSuffix
}

func stripModuleBuildSuffix(version string) string {
	if before, _, ok := strings.Cut(version, "+"); ok {
		return before
	}
	return version
}
