// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type depConfig struct {
	bazelNameOverride string // Bazel style uses underscores not dashes, so we fix those if needed.
	needsBazelFile    bool
	patchCmds         []string
	patchCmdsWin      []string
}

// These are all C++ deps or Rust deps (with a compatible C++ FFI) used by the Bazel build.
// They are a subset of those listed in DEPS.
// The key is the name of the repo as specified in DEPS.
var deps = map[string]depConfig{
	"abseil-cpp":  {bazelNameOverride: "abseil_cpp"},
	"brotli":      {},
	"highway":     {},
	"spirv-tools": {bazelNameOverride: "spirv_tools"},
	// This name is important because spirv_tools expects @spirv_headers to exist by that name.
	"spirv-headers": {bazelNameOverride: "spirv_headers"},

	"dawn":           {needsBazelFile: true},
	"delaunator-cpp": {bazelNameOverride: "delaunator", needsBazelFile: true},
	"dng_sdk":        {needsBazelFile: true},
	"expat":          {needsBazelFile: true},
	"freetype":       {needsBazelFile: true},
	"harfbuzz":       {needsBazelFile: true},
	"icu": {
		needsBazelFile: true,
		patchCmds: []string{
			`"rm source/i18n/BUILD.bazel"`,
			`"rm source/common/BUILD.bazel"`,
			`"rm source/stubdata/BUILD.bazel"`,
		},
		patchCmdsWin: []string{
			`"del source/i18n/BUILD.bazel"`,
			`"del source/common/BUILD.bazel"`,
			`"del source/stubdata/BUILD.bazel"`,
		},
	},
	"icu4x":                    {needsBazelFile: true},
	"imgui":                    {needsBazelFile: true},
	"libavif":                  {needsBazelFile: true},
	"libgav1":                  {needsBazelFile: true},
	"libjpeg-turbo":            {bazelNameOverride: "libjpeg_turbo", needsBazelFile: true},
	"libjxl":                   {needsBazelFile: true},
	"libpng":                   {needsBazelFile: true},
	"libwebp":                  {needsBazelFile: true},
	"libyuv":                   {needsBazelFile: true},
	"spirv-cross":              {bazelNameOverride: "spirv_cross", needsBazelFile: true},
	"perfetto":                 {needsBazelFile: true},
	"piex":                     {needsBazelFile: true},
	"vello":                    {needsBazelFile: true},
	"vulkan-headers":           {bazelNameOverride: "vulkan_headers", needsBazelFile: true},
	"vulkan-tools":             {bazelNameOverride: "vulkan_tools", needsBazelFile: true},
	"vulkan-utility-libraries": {bazelNameOverride: "vulkan_utility_libraries", needsBazelFile: true},
	"vulkanmemoryallocator":    {needsBazelFile: true},
	"wuffs":                    {needsBazelFile: true},
	// Some other dependency downloads zlib but with their own rules
	"zlib": {bazelNameOverride: "zlib_skia", needsBazelFile: true},
}

func main() {
	var (
		depsFile      = flag.String("deps_file", "DEPS", "The location of the DEPS file. Usually at the root of the repository")
		genBzlFile    = flag.String("gen_bzl_file", "bazel/deps.bzl", "The location of the .bzl file that has the generated Bazel repository rules.")
		workspaceFile = flag.String("workspace_file", "WORKSPACE.bazel", "The location of the WORKSPACE file that should be updated with dep names.")
		// https://bazel.build/docs/user-manual#running-executables
		repoDir        = flag.String("repo_dir", os.Getenv("BUILD_WORKSPACE_DIRECTORY"), "The root directory of the repo. Default set by BUILD_WORKSPACE_DIRECTORY env variable.")
		buildifierPath = flag.String("buildifier", "", "Where to find buildifier. Defaults to Bazel's location")
	)
	flag.Parse()

	if *repoDir == "" {
		fmt.Println(`Must set --repo_dir
This is done automatically via:
    bazel run //bazel/deps_parser`)
		os.Exit(1)
	}

	buildifier := *buildifierPath
	if buildifier == "" {
		// We don't know if this will be buildifier_linux_x64, buildifier_macos_arm64, etc
		bp, err := filepath.Glob("../buildifier*/file/buildifier")
		if err != nil || len(bp) != 1 {
			fmt.Printf("Could not find exactly one buildifier executable %s %v\n", err, bp)
			os.Exit(1)
		}
		buildifier = bp[0]
	}
	buildifier, err := filepath.Abs(buildifier)
	if err != nil {
		fmt.Printf("Abs path error %s\n", err)
		os.Exit(1)
	}

	fmt.Println(os.Environ())

	if *depsFile == "" || *genBzlFile == "" {
		fmt.Println("Must set --deps_file and --gen_bzl_file")
		flag.PrintDefaults()
	}

	if err := os.Chdir(*repoDir); err != nil {
		fmt.Printf("Could not cd to %s\n", *repoDir)
		os.Exit(1)
	}

	b, err := os.ReadFile(*depsFile)
	if err != nil {
		fmt.Printf("Could not open %s: %s\n", *depsFile, err)
		os.Exit(1)
	}
	contents := strings.Split(string(b), "\n")

	outputFile, count, err := parseDEPSFile(contents, *workspaceFile)
	if err != nil {
		fmt.Printf("Parsing error %s\n", err)
		os.Exit(1)
	}
	if err := exec.Command(buildifier, outputFile).Run(); err != nil {
		fmt.Printf("Buildifier error %s\n", err)
		os.Exit(1)
	}
	if err := moveWithCopyBackup(outputFile, *genBzlFile); err != nil {
		fmt.Printf("Could not write comments in workspace file: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Wrote %d deps\n", count)
}

func parseDEPSFile(contents []string, workspaceFile string) (string, int, error) {
	depsLine := regexp.MustCompile(`externals/(\S+)".+"(https.+)@([a-f0-9]+)"`)
	outputFile, err := os.CreateTemp("", "genbzl")
	if err != nil {
		return "", 0, fmt.Errorf("Could not create output file: %s\n", err)
	}
	defer outputFile.Close()

	if _, err := outputFile.WriteString(header); err != nil {
		return "", 0, fmt.Errorf("Could not write header to output file %s: %s\n", outputFile.Name(), err)
	}

	var nativeRepos []string
	var providedRepos []string

	count := 0
	for _, line := range contents {
		if match := depsLine.FindStringSubmatch(line); len(match) > 0 {
			id := match[1]
			repo := match[2]
			rev := match[3]

			cfg, ok := deps[id]
			if !ok {
				continue
			}
			if cfg.bazelNameOverride != "" {
				id = cfg.bazelNameOverride
			}
			if cfg.needsBazelFile {
				if err := writeNewGitRepositoryRule(outputFile, id, repo, rev, cfg.patchCmds, cfg.patchCmdsWin); err != nil {
					return "", 0, fmt.Errorf("Could not write to output file %s: %s\n", outputFile.Name(), err)
				}
				workspaceLine := fmt.Sprintf("# @%s - //bazel/external/%s:BUILD.bazel", id, id)
				providedRepos = append(providedRepos, workspaceLine)
			} else {
				if err := writeGitRepositoryRule(outputFile, id, repo, rev); err != nil {
					return "", 0, fmt.Errorf("Could not write to output file %s: %s\n", outputFile.Name(), err)
				}
				workspaceLine := fmt.Sprintf("# @%s - %s", id, repo)
				nativeRepos = append(nativeRepos, workspaceLine)
			}
			count++
		}
	}
	if count != len(deps) {
		return "", 0, fmt.Errorf("Not enough deps written. Maybe the deps dictionary needs a bazelNameOverride or an old dep needs to be removed?")
	}

	if _, err := outputFile.WriteString(footer); err != nil {
		return "", 0, fmt.Errorf("Could not write footer to output file %s: %s\n", outputFile.Name(), err)
	}

	if newWorkspaceFile, err := writeCommentsToWorkspace(workspaceFile, nativeRepos, providedRepos); err != nil {
		fmt.Printf("Could not parse workspace file %s: %s\n", workspaceFile, err)
		os.Exit(1)
	} else {
		if err := moveWithCopyBackup(newWorkspaceFile, workspaceFile); err != nil {
			fmt.Printf("Could not write comments in workspace file: %s\n", err)
			os.Exit(1)
		}
	}
	return outputFile.Name(), count, nil
}

func writeCommentsToWorkspace(workspaceFile string, nativeRepos, providedRepos []string) (string, error) {
	b, err := os.ReadFile(workspaceFile)
	if err != nil {
		return "", fmt.Errorf("Could not open %s: %s\n", workspaceFile, err)
	}
	newWorkspace, err := os.CreateTemp("", "workspace")
	if err != nil {
		return "", fmt.Errorf("Could not make tempfile: %s\n", err)
	}
	defer newWorkspace.Close()

	workspaceContents := strings.Split(string(b), "\n")

	sort.Strings(nativeRepos)
	sort.Strings(providedRepos)
	for _, line := range workspaceContents {
		if _, err := newWorkspace.WriteString(line + "\n"); err != nil {
			return "", err
		}
		if line == startListString {
			break
		}
	}
	for _, repoLine := range nativeRepos {
		if _, err := newWorkspace.WriteString(repoLine + "\n"); err != nil {
			return "", err
		}
	}
	if _, err := newWorkspace.WriteString("#\n"); err != nil {
		return "", err
	}
	for _, repoLine := range providedRepos {
		if _, err := newWorkspace.WriteString(repoLine + "\n"); err != nil {
			return "", err
		}
	}
	if _, err := newWorkspace.WriteString(endListString + "\n"); err != nil {
		return "", err
	}

	pastEnd := false
	// Skip the last line, which is blank. We don't want to end with two empty newlines.
	for _, line := range workspaceContents[:len(workspaceContents)-1] {
		if line == endListString {
			pastEnd = true
			continue
		}
		if !pastEnd {
			continue
		}
		if _, err := newWorkspace.WriteString(line + "\n"); err != nil {
			return "", err
		}
	}

	return newWorkspace.Name(), nil
}

const (
	startListString = `#### START GENERATED LIST OF THIRD_PARTY DEPS`
	endListString   = `#### END GENERATED LIST OF THIRD_PARTY DEPS`
)

const header = `"""
This file is auto-generated from //bazel/deps_parser
DO NOT MODIFY BY HAND.
Instead, do:
    bazel run //bazel/deps_parser
"""

load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository", "new_git_repository")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load("//bazel:download_config_files.bzl", "download_config_files")
load("//bazel:gcs_mirror.bzl", "gcs_mirror_url")

def c_plus_plus_deps(ws = "@skia"):
    """A list of native Bazel git rules to download third party git repositories

       These are in the order they appear in //DEPS.
        https://bazel.build/rules/lib/repo/git

    Args:
      ws: The name of the Skia Bazel workspace. The default, "@", may be when used from within the
          Skia workspace.
    """`

// If necessary, we can make a new map for bazel deps
const footer = `
def bazel_deps():
    maybe(
        http_archive,
        name = "bazel_skylib",
        sha256 = "c6966ec828da198c5d9adbaa94c05e3a1c7f21bd012a0b29ba8ddbccb2c93b0d",
        urls = gcs_mirror_url(
            sha256 = "c6966ec828da198c5d9adbaa94c05e3a1c7f21bd012a0b29ba8ddbccb2c93b0d",
            url = "https://github.com/bazelbuild/bazel-skylib/releases/download/1.1.1/bazel-skylib-1.1.1.tar.gz",
        ),
    )

    maybe(
        http_archive,
        name = "bazel_toolchains",
        sha256 = "e52789d4e89c3e2dc0e3446a9684626a626b6bec3fde787d70bae37c6ebcc47f",
        strip_prefix = "bazel-toolchains-5.1.1",
        urls = gcs_mirror_url(
            sha256 = "e52789d4e89c3e2dc0e3446a9684626a626b6bec3fde787d70bae37c6ebcc47f",
            url = "https://github.com/bazelbuild/bazel-toolchains/archive/refs/tags/v5.1.1.tar.gz",
        ),
    )

def header_based_configs():
    skia_revision = "d211141c45c9171437fa8e6e07989edb5bffa17a"
    maybe(
        download_config_files,
        name = "expat_config",
        skia_revision = skia_revision,
        files = {
            "BUILD.bazel": "third_party/expat/include/BUILD.bazel",
            "expat_config/expat_config.h": "third_party/expat/include/expat_config/expat_config.h",
        },
    )
    maybe(
        download_config_files,
        name = "freetype_config",
        skia_revision = skia_revision,
        files = {
            "BUILD.bazel": "third_party/freetype2/include/BUILD.bazel",
            "freetype-android/freetype/config/ftmodule.h": "third_party/freetype2/include/freetype-android/freetype/config/ftmodule.h",
            "freetype-android/freetype/config/ftoption.h": "third_party/freetype2/include/freetype-android/freetype/config/ftoption.h",
            "freetype-no-type1/freetype/config/ftmodule.h": "third_party/freetype2/include/freetype-no-type1/freetype/config/ftmodule.h",
            "freetype-no-type1/freetype/config/ftoption.h": "third_party/freetype2/include/freetype-no-type1/freetype/config/ftoption.h",
        },
    )
    maybe(
        download_config_files,
        name = "harfbuzz_config",
        skia_revision = skia_revision,
        files = {
            "BUILD.bazel": "third_party/harfbuzz/BUILD.bazel",
            "config-override.h": "third_party/harfbuzz/config-override.h",
        },
    )
    maybe(
        download_config_files,
        name = "icu_utils",
        skia_revision = skia_revision,
        files = {
            "BUILD.bazel": "third_party/icu/BUILD.bazel",
            "SkLoadICU.cpp": "third_party/icu/SkLoadICU.cpp",
            "SkLoadICU.h": "third_party/icu/SkLoadICU.h",
            "make_data_cpp.py": "third_party/icu/make_data_cpp.py",
        },
    )
`

func writeNewGitRepositoryRule(w io.StringWriter, bazelName, repo, rev string, patchCmds, patchCmdsWin []string) error {
	if len(patchCmds) == 0 {
		// TODO(kjlubick) In a newer version of Bazel, new_git_repository can be replaced with just
		// git_repository
		_, err := w.WriteString(fmt.Sprintf(`
    new_git_repository(
        name = "%s",
        build_file = ws + "//bazel/external/%s:BUILD.bazel",
        commit = "%s",
        remote = "%s",
    )
`, bazelName, bazelName, rev, repo))
		return err
	}
	patches := "[" + strings.Join(patchCmds, ",\n") + "]"
	patches_win := "[" + strings.Join(patchCmdsWin, ",\n") + "]"
	_, err := w.WriteString(fmt.Sprintf(`
    new_git_repository(
        name = "%s",
        build_file = ws + "//bazel/external/%s:BUILD.bazel",
        commit = "%s",
        remote = "%s",
        patch_cmds = %s,
		patch_cmds_win = %s,
    )
`, bazelName, bazelName, rev, repo, patches, patches_win))
	return err
}

func writeGitRepositoryRule(w io.StringWriter, bazelName, repo, rev string) error {
	_, err := w.WriteString(fmt.Sprintf(`
    git_repository(
        name = "%s",
        commit = "%s",
        remote = "%s",
    )
`, bazelName, rev, repo))
	return err
}

func moveWithCopyBackup(src, dst string) error {
	// Atomically rename temp file to workspace. This should minimize the chance of corruption
	// or writing a partial file if there is an error or the program is interrupted.
	if err := os.Rename(src, dst); err != nil {
		// Errors can happen if the temporary file is on a different partition than the Skia
		// codebase. In that case, do a manual read/write to copy the data. See
		// https://github.com/jenkins-x/jx/issues/449 for a similar issue
		if strings.Contains(err.Error(), "invalid cross-device link") {
			bytes, err := os.ReadFile(src)
			if err != nil {
				return fmt.Errorf("Could not do backup read from %s: %s\n", src, err)
			}
			if err := os.WriteFile(dst, bytes, 0644); err != nil {
				return fmt.Errorf("Could not do backup write of %d bytes to %s: %s\n", len(bytes), dst, err)
			}
			// Backup "move" successful
			return nil
		}
		return fmt.Errorf("Could not write %s -> %s: %s\n", src, dst, err)
	}
	return nil
}
