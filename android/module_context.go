// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package android

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/blueprint"
	"github.com/google/blueprint/proptools"
)

// BuildParameters describes the set of potential parameters to build a Ninja rule.
// In general, these correspond to a Ninja concept.
type BuildParams struct {
	// A Ninja Rule that will be written to the Ninja file. This allows factoring out common code
	// among multiple modules to reduce repetition in the Ninja file of action requirements. A rule
	// can contain variables that should be provided in Args.
	Rule blueprint.Rule
	// Deps represents the depfile format. When using RuleBuilder, this defaults to GCC when depfiles
	// are used.
	Deps blueprint.Deps
	// Depfile is a writeable path that allows correct incremental builds when the inputs have not
	// been fully specified by the Ninja rule. Ninja supports a subset of the Makefile depfile syntax.
	Depfile WritablePath
	// A description of the build action.
	Description string
	// Output is an output file of the action. When using this field, references to $out in the Ninja
	// command will refer to this file.
	Output WritablePath
	// Outputs is a slice of output file of the action. When using this field, references to $out in
	// the Ninja command will refer to these files.
	Outputs WritablePaths
	// ImplicitOutput is an output file generated by the action. Note: references to `$out` in the
	// Ninja command will NOT include references to this file.
	ImplicitOutput WritablePath
	// ImplicitOutputs is a slice of output files generated by the action. Note: references to `$out`
	// in the Ninja command will NOT include references to these files.
	ImplicitOutputs WritablePaths
	// Input is an input file to the Ninja action. When using this field, references to $in in the
	// Ninja command will refer to this file.
	Input Path
	// Inputs is a slice of input files to the Ninja action. When using this field, references to $in
	// in the Ninja command will refer to these files.
	Inputs Paths
	// Implicit is an input file to the Ninja action. Note: references to `$in` in the Ninja command
	// will NOT include references to this file.
	Implicit Path
	// Implicits is a slice of input files to the Ninja action. Note: references to `$in` in the Ninja
	// command will NOT include references to these files.
	Implicits Paths
	// OrderOnly are Ninja order-only inputs to the action. When these are out of date, the output is
	// not rebuilt until they are built, but changes in order-only dependencies alone do not cause the
	// output to be rebuilt.
	OrderOnly Paths
	// Validation is an output path for a validation action. Validation outputs imply lower
	// non-blocking priority to building non-validation outputs.
	Validation Path
	// Validations is a slice of output path for a validation action. Validation outputs imply lower
	// non-blocking priority to building non-validation outputs.
	Validations Paths
	// Whether to skip outputting a default target statement which will be built by Ninja when no
	// targets are specified on Ninja's command line.
	Default bool
	// Args is a key value mapping for replacements of variables within the Rule
	Args map[string]string
}

type ModuleBuildParams BuildParams

type ModuleContext interface {
	BaseModuleContext

	blueprintModuleContext() blueprint.ModuleContext

	// Deprecated: use ModuleContext.Build instead.
	ModuleBuild(pctx PackageContext, params ModuleBuildParams)

	// Returns a list of paths expanded from globs and modules referenced using ":module" syntax.  The property must
	// be tagged with `android:"path" to support automatic source module dependency resolution.
	//
	// Deprecated: use PathsForModuleSrc or PathsForModuleSrcExcludes instead.
	ExpandSources(srcFiles, excludes []string) Paths

	// Returns a single path expanded from globs and modules referenced using ":module" syntax.  The property must
	// be tagged with `android:"path" to support automatic source module dependency resolution.
	//
	// Deprecated: use PathForModuleSrc instead.
	ExpandSource(srcFile, prop string) Path

	ExpandOptionalSource(srcFile *string, prop string) OptionalPath

	// InstallExecutable creates a rule to copy srcPath to name in the installPath directory,
	// with the given additional dependencies.  The file is marked executable after copying.
	//
	// The installed file will be returned by FilesToInstall(), and the PackagingSpec for the
	// installed file will be returned by PackagingSpecs() on this module or by
	// TransitivePackagingSpecs() on modules that depend on this module through dependency tags
	// for which IsInstallDepNeeded returns true.
	InstallExecutable(installPath InstallPath, name string, srcPath Path, deps ...InstallPath) InstallPath

	// InstallFile creates a rule to copy srcPath to name in the installPath directory,
	// with the given additional dependencies.
	//
	// The installed file will be returned by FilesToInstall(), and the PackagingSpec for the
	// installed file will be returned by PackagingSpecs() on this module or by
	// TransitivePackagingSpecs() on modules that depend on this module through dependency tags
	// for which IsInstallDepNeeded returns true.
	InstallFile(installPath InstallPath, name string, srcPath Path, deps ...InstallPath) InstallPath

	// InstallFileWithExtraFilesZip creates a rule to copy srcPath to name in the installPath
	// directory, and also unzip a zip file containing extra files to install into the same
	// directory.
	//
	// The installed file will be returned by FilesToInstall(), and the PackagingSpec for the
	// installed file will be returned by PackagingSpecs() on this module or by
	// TransitivePackagingSpecs() on modules that depend on this module through dependency tags
	// for which IsInstallDepNeeded returns true.
	InstallFileWithExtraFilesZip(installPath InstallPath, name string, srcPath Path, extraZip Path, deps ...InstallPath) InstallPath

	// InstallSymlink creates a rule to create a symlink from src srcPath to name in the installPath
	// directory.
	//
	// The installed symlink will be returned by FilesToInstall(), and the PackagingSpec for the
	// installed file will be returned by PackagingSpecs() on this module or by
	// TransitivePackagingSpecs() on modules that depend on this module through dependency tags
	// for which IsInstallDepNeeded returns true.
	InstallSymlink(installPath InstallPath, name string, srcPath InstallPath) InstallPath

	// InstallAbsoluteSymlink creates a rule to create an absolute symlink from src srcPath to name
	// in the installPath directory.
	//
	// The installed symlink will be returned by FilesToInstall(), and the PackagingSpec for the
	// installed file will be returned by PackagingSpecs() on this module or by
	// TransitivePackagingSpecs() on modules that depend on this module through dependency tags
	// for which IsInstallDepNeeded returns true.
	InstallAbsoluteSymlink(installPath InstallPath, name string, absPath string) InstallPath

	// PackageFile creates a PackagingSpec as if InstallFile was called, but without creating
	// the rule to copy the file.  This is useful to define how a module would be packaged
	// without installing it into the global installation directories.
	//
	// The created PackagingSpec for the will be returned by PackagingSpecs() on this module or by
	// TransitivePackagingSpecs() on modules that depend on this module through dependency tags
	// for which IsInstallDepNeeded returns true.
	PackageFile(installPath InstallPath, name string, srcPath Path) PackagingSpec

	CheckbuildFile(srcPath Path)

	InstallInData() bool
	InstallInTestcases() bool
	InstallInSanitizerDir() bool
	InstallInRamdisk() bool
	InstallInVendorRamdisk() bool
	InstallInDebugRamdisk() bool
	InstallInRecovery() bool
	InstallInRoot() bool
	InstallInOdm() bool
	InstallInProduct() bool
	InstallInVendor() bool
	InstallForceOS() (*OsType, *ArchType)

	RequiredModuleNames() []string
	HostRequiredModuleNames() []string
	TargetRequiredModuleNames() []string

	ModuleSubDir() string
	SoongConfigTraceHash() string

	Variable(pctx PackageContext, name, value string)
	Rule(pctx PackageContext, name string, params blueprint.RuleParams, argNames ...string) blueprint.Rule
	// Similar to blueprint.ModuleContext.Build, but takes Paths instead of []string,
	// and performs more verification.
	Build(pctx PackageContext, params BuildParams)
	// Phony creates a Make-style phony rule, a rule with no commands that can depend on other
	// phony rules or real files.  Phony can be called on the same name multiple times to add
	// additional dependencies.
	Phony(phony string, deps ...Path)

	// GetMissingDependencies returns the list of dependencies that were passed to AddDependencies or related methods,
	// but do not exist.
	GetMissingDependencies() []string

	// LicenseMetadataFile returns the path where the license metadata for this module will be
	// generated.
	LicenseMetadataFile() Path

	// ModuleInfoJSON returns a pointer to the ModuleInfoJSON struct that can be filled out by
	// GenerateAndroidBuildActions.  If it is called then the struct will be written out and included in
	// the module-info.json generated by Make, and Make will not generate its own data for this module.
	ModuleInfoJSON() *ModuleInfoJSON
}

type moduleContext struct {
	bp blueprint.ModuleContext
	baseModuleContext
	packagingSpecs  []PackagingSpec
	installFiles    InstallPaths
	checkbuildFiles Paths
	module          Module
	phonies         map[string]Paths

	katiInstalls []katiInstall
	katiSymlinks []katiInstall

	// For tests
	buildParams []BuildParams
	ruleParams  map[blueprint.Rule]blueprint.RuleParams
	variables   map[string]string
}

var _ ModuleContext = &moduleContext{}

func (m *moduleContext) ninjaError(params BuildParams, err error) (PackageContext, BuildParams) {
	return pctx, BuildParams{
		Rule:            ErrorRule,
		Description:     params.Description,
		Output:          params.Output,
		Outputs:         params.Outputs,
		ImplicitOutput:  params.ImplicitOutput,
		ImplicitOutputs: params.ImplicitOutputs,
		Args: map[string]string{
			"error": err.Error(),
		},
	}
}

func (m *moduleContext) ModuleBuild(pctx PackageContext, params ModuleBuildParams) {
	m.Build(pctx, BuildParams(params))
}

// Convert build parameters from their concrete Android types into their string representations,
// and combine the singular and plural fields of the same type (e.g. Output and Outputs).
func convertBuildParams(params BuildParams) blueprint.BuildParams {
	bparams := blueprint.BuildParams{
		Rule:            params.Rule,
		Description:     params.Description,
		Deps:            params.Deps,
		Outputs:         params.Outputs.Strings(),
		ImplicitOutputs: params.ImplicitOutputs.Strings(),
		Inputs:          params.Inputs.Strings(),
		Implicits:       params.Implicits.Strings(),
		OrderOnly:       params.OrderOnly.Strings(),
		Validations:     params.Validations.Strings(),
		Args:            params.Args,
		Optional:        !params.Default,
	}

	if params.Depfile != nil {
		bparams.Depfile = params.Depfile.String()
	}
	if params.Output != nil {
		bparams.Outputs = append(bparams.Outputs, params.Output.String())
	}
	if params.ImplicitOutput != nil {
		bparams.ImplicitOutputs = append(bparams.ImplicitOutputs, params.ImplicitOutput.String())
	}
	if params.Input != nil {
		bparams.Inputs = append(bparams.Inputs, params.Input.String())
	}
	if params.Implicit != nil {
		bparams.Implicits = append(bparams.Implicits, params.Implicit.String())
	}
	if params.Validation != nil {
		bparams.Validations = append(bparams.Validations, params.Validation.String())
	}

	bparams.Outputs = proptools.NinjaEscapeList(bparams.Outputs)
	bparams.ImplicitOutputs = proptools.NinjaEscapeList(bparams.ImplicitOutputs)
	bparams.Inputs = proptools.NinjaEscapeList(bparams.Inputs)
	bparams.Implicits = proptools.NinjaEscapeList(bparams.Implicits)
	bparams.OrderOnly = proptools.NinjaEscapeList(bparams.OrderOnly)
	bparams.Validations = proptools.NinjaEscapeList(bparams.Validations)
	bparams.Depfile = proptools.NinjaEscape(bparams.Depfile)

	return bparams
}

func (m *moduleContext) Variable(pctx PackageContext, name, value string) {
	if m.config.captureBuild {
		m.variables[name] = value
	}

	m.bp.Variable(pctx.PackageContext, name, value)
}

func (m *moduleContext) Rule(pctx PackageContext, name string, params blueprint.RuleParams,
	argNames ...string) blueprint.Rule {

	if m.config.UseRemoteBuild() {
		if params.Pool == nil {
			// When USE_GOMA=true or USE_RBE=true are set and the rule is not supported by goma/RBE, restrict
			// jobs to the local parallelism value
			params.Pool = localPool
		} else if params.Pool == remotePool {
			// remotePool is a fake pool used to identify rule that are supported for remoting. If the rule's
			// pool is the remotePool, replace with nil so that ninja runs it at NINJA_REMOTE_NUM_JOBS
			// parallelism.
			params.Pool = nil
		}
	}

	rule := m.bp.Rule(pctx.PackageContext, name, params, argNames...)

	if m.config.captureBuild {
		m.ruleParams[rule] = params
	}

	return rule
}

func (m *moduleContext) Build(pctx PackageContext, params BuildParams) {
	if params.Description != "" {
		params.Description = "${moduleDesc}" + params.Description + "${moduleDescSuffix}"
	}

	if missingDeps := m.GetMissingDependencies(); len(missingDeps) > 0 {
		pctx, params = m.ninjaError(params, fmt.Errorf("module %s missing dependencies: %s\n",
			m.ModuleName(), strings.Join(missingDeps, ", ")))
	}

	if m.config.captureBuild {
		m.buildParams = append(m.buildParams, params)
	}

	bparams := convertBuildParams(params)
	m.bp.Build(pctx.PackageContext, bparams)
}

func (m *moduleContext) Phony(name string, deps ...Path) {
	addPhony(m.config, name, deps...)
}

func (m *moduleContext) GetMissingDependencies() []string {
	var missingDeps []string
	missingDeps = append(missingDeps, m.Module().base().commonProperties.MissingDeps...)
	missingDeps = append(missingDeps, m.bp.GetMissingDependencies()...)
	missingDeps = FirstUniqueStrings(missingDeps)
	return missingDeps
}

func (m *moduleContext) GetDirectDepWithTag(name string, tag blueprint.DependencyTag) blueprint.Module {
	module, _ := m.getDirectDepInternal(name, tag)
	return module
}

func (m *moduleContext) ModuleSubDir() string {
	return m.bp.ModuleSubDir()
}

func (m *moduleContext) SoongConfigTraceHash() string {
	return m.module.base().commonProperties.SoongConfigTraceHash
}

func (m *moduleContext) InstallInData() bool {
	return m.module.InstallInData()
}

func (m *moduleContext) InstallInTestcases() bool {
	return m.module.InstallInTestcases()
}

func (m *moduleContext) InstallInSanitizerDir() bool {
	return m.module.InstallInSanitizerDir()
}

func (m *moduleContext) InstallInRamdisk() bool {
	return m.module.InstallInRamdisk()
}

func (m *moduleContext) InstallInVendorRamdisk() bool {
	return m.module.InstallInVendorRamdisk()
}

func (m *moduleContext) InstallInDebugRamdisk() bool {
	return m.module.InstallInDebugRamdisk()
}

func (m *moduleContext) InstallInRecovery() bool {
	return m.module.InstallInRecovery()
}

func (m *moduleContext) InstallInRoot() bool {
	return m.module.InstallInRoot()
}

func (m *moduleContext) InstallForceOS() (*OsType, *ArchType) {
	return m.module.InstallForceOS()
}

func (m *moduleContext) InstallInOdm() bool {
	return m.module.InstallInOdm()
}

func (m *moduleContext) InstallInProduct() bool {
	return m.module.InstallInProduct()
}

func (m *moduleContext) InstallInVendor() bool {
	return m.module.InstallInVendor()
}

func (m *moduleContext) skipInstall() bool {
	if m.module.base().commonProperties.SkipInstall {
		return true
	}

	if m.module.base().commonProperties.HideFromMake {
		return true
	}

	// We'll need a solution for choosing which of modules with the same name in different
	// namespaces to install.  For now, reuse the list of namespaces exported to Make as the
	// list of namespaces to install in a Soong-only build.
	if !m.module.base().commonProperties.NamespaceExportedToMake {
		return true
	}

	return false
}

// Tells whether this module is installed to the full install path (ex:
// out/target/product/<name>/<partition>) or not. If this returns false, the install build rule is
// not created and this module can only be installed to packaging modules like android_filesystem.
func (m *moduleContext) requiresFullInstall() bool {
	if m.skipInstall() {
		return false
	}

	if proptools.Bool(m.module.base().commonProperties.No_full_install) {
		return false
	}

	return true
}

func (m *moduleContext) InstallFile(installPath InstallPath, name string, srcPath Path,
	deps ...InstallPath) InstallPath {
	return m.installFile(installPath, name, srcPath, deps, false, nil)
}

func (m *moduleContext) InstallExecutable(installPath InstallPath, name string, srcPath Path,
	deps ...InstallPath) InstallPath {
	return m.installFile(installPath, name, srcPath, deps, true, nil)
}

func (m *moduleContext) InstallFileWithExtraFilesZip(installPath InstallPath, name string, srcPath Path,
	extraZip Path, deps ...InstallPath) InstallPath {
	return m.installFile(installPath, name, srcPath, deps, false, &extraFilesZip{
		zip: extraZip,
		dir: installPath,
	})
}

func (m *moduleContext) PackageFile(installPath InstallPath, name string, srcPath Path) PackagingSpec {
	fullInstallPath := installPath.Join(m, name)
	return m.packageFile(fullInstallPath, srcPath, false)
}

func (m *moduleContext) getAconfigPaths() *Paths {
	return &m.module.base().aconfigFilePaths
}

func (m *moduleContext) packageFile(fullInstallPath InstallPath, srcPath Path, executable bool) PackagingSpec {
	licenseFiles := m.Module().EffectiveLicenseFiles()
	spec := PackagingSpec{
		relPathInPackage:      Rel(m, fullInstallPath.PartitionDir(), fullInstallPath.String()),
		srcPath:               srcPath,
		symlinkTarget:         "",
		executable:            executable,
		effectiveLicenseFiles: &licenseFiles,
		partition:             fullInstallPath.partition,
		skipInstall:           m.skipInstall(),
		aconfigPaths:          m.getAconfigPaths(),
	}
	m.packagingSpecs = append(m.packagingSpecs, spec)
	return spec
}

func (m *moduleContext) installFile(installPath InstallPath, name string, srcPath Path, deps []InstallPath,
	executable bool, extraZip *extraFilesZip) InstallPath {

	fullInstallPath := installPath.Join(m, name)
	m.module.base().hooks.runInstallHooks(m, srcPath, fullInstallPath, false)

	if m.requiresFullInstall() {
		deps = append(deps, InstallPaths(m.module.base().installFilesDepSet.ToList())...)
		deps = append(deps, m.module.base().installedInitRcPaths...)
		deps = append(deps, m.module.base().installedVintfFragmentsPaths...)

		var implicitDeps, orderOnlyDeps Paths

		if m.Host() {
			// Installed host modules might be used during the build, depend directly on their
			// dependencies so their timestamp is updated whenever their dependency is updated
			implicitDeps = InstallPaths(deps).Paths()
		} else {
			orderOnlyDeps = InstallPaths(deps).Paths()
		}

		if m.Config().KatiEnabled() {
			// When creating the install rule in Soong but embedding in Make, write the rule to a
			// makefile instead of directly to the ninja file so that main.mk can add the
			// dependencies from the `required` property that are hard to resolve in Soong.
			m.katiInstalls = append(m.katiInstalls, katiInstall{
				from:          srcPath,
				to:            fullInstallPath,
				implicitDeps:  implicitDeps,
				orderOnlyDeps: orderOnlyDeps,
				executable:    executable,
				extraFiles:    extraZip,
			})
		} else {
			rule := Cp
			if executable {
				rule = CpExecutable
			}

			extraCmds := ""
			if extraZip != nil {
				extraCmds += fmt.Sprintf(" && ( unzip -qDD -d '%s' '%s' 2>&1 | grep -v \"zipfile is empty\"; exit $${PIPESTATUS[0]} )",
					extraZip.dir.String(), extraZip.zip.String())
				extraCmds += " || ( code=$$?; if [ $$code -ne 0 -a $$code -ne 1 ]; then exit $$code; fi )"
				implicitDeps = append(implicitDeps, extraZip.zip)
			}

			m.Build(pctx, BuildParams{
				Rule:        rule,
				Description: "install " + fullInstallPath.Base(),
				Output:      fullInstallPath,
				Input:       srcPath,
				Implicits:   implicitDeps,
				OrderOnly:   orderOnlyDeps,
				Default:     !m.Config().KatiEnabled(),
				Args: map[string]string{
					"extraCmds": extraCmds,
				},
			})
		}

		m.installFiles = append(m.installFiles, fullInstallPath)
	}

	m.packageFile(fullInstallPath, srcPath, executable)

	m.checkbuildFiles = append(m.checkbuildFiles, srcPath)

	return fullInstallPath
}

func (m *moduleContext) InstallSymlink(installPath InstallPath, name string, srcPath InstallPath) InstallPath {
	fullInstallPath := installPath.Join(m, name)
	m.module.base().hooks.runInstallHooks(m, srcPath, fullInstallPath, true)

	relPath, err := filepath.Rel(path.Dir(fullInstallPath.String()), srcPath.String())
	if err != nil {
		panic(fmt.Sprintf("Unable to generate symlink between %q and %q: %s", fullInstallPath.Base(), srcPath.Base(), err))
	}
	if m.requiresFullInstall() {

		if m.Config().KatiEnabled() {
			// When creating the symlink rule in Soong but embedding in Make, write the rule to a
			// makefile instead of directly to the ninja file so that main.mk can add the
			// dependencies from the `required` property that are hard to resolve in Soong.
			m.katiSymlinks = append(m.katiSymlinks, katiInstall{
				from: srcPath,
				to:   fullInstallPath,
			})
		} else {
			// The symlink doesn't need updating when the target is modified, but we sometimes
			// have a dependency on a symlink to a binary instead of to the binary directly, and
			// the mtime of the symlink must be updated when the binary is modified, so use a
			// normal dependency here instead of an order-only dependency.
			m.Build(pctx, BuildParams{
				Rule:        Symlink,
				Description: "install symlink " + fullInstallPath.Base(),
				Output:      fullInstallPath,
				Input:       srcPath,
				Default:     !m.Config().KatiEnabled(),
				Args: map[string]string{
					"fromPath": relPath,
				},
			})
		}

		m.installFiles = append(m.installFiles, fullInstallPath)
		m.checkbuildFiles = append(m.checkbuildFiles, srcPath)
	}

	m.packagingSpecs = append(m.packagingSpecs, PackagingSpec{
		relPathInPackage: Rel(m, fullInstallPath.PartitionDir(), fullInstallPath.String()),
		srcPath:          nil,
		symlinkTarget:    relPath,
		executable:       false,
		partition:        fullInstallPath.partition,
		skipInstall:      m.skipInstall(),
		aconfigPaths:     m.getAconfigPaths(),
	})

	return fullInstallPath
}

// installPath/name -> absPath where absPath might be a path that is available only at runtime
// (e.g. /apex/...)
func (m *moduleContext) InstallAbsoluteSymlink(installPath InstallPath, name string, absPath string) InstallPath {
	fullInstallPath := installPath.Join(m, name)
	m.module.base().hooks.runInstallHooks(m, nil, fullInstallPath, true)

	if m.requiresFullInstall() {
		if m.Config().KatiEnabled() {
			// When creating the symlink rule in Soong but embedding in Make, write the rule to a
			// makefile instead of directly to the ninja file so that main.mk can add the
			// dependencies from the `required` property that are hard to resolve in Soong.
			m.katiSymlinks = append(m.katiSymlinks, katiInstall{
				absFrom: absPath,
				to:      fullInstallPath,
			})
		} else {
			m.Build(pctx, BuildParams{
				Rule:        Symlink,
				Description: "install symlink " + fullInstallPath.Base() + " -> " + absPath,
				Output:      fullInstallPath,
				Default:     !m.Config().KatiEnabled(),
				Args: map[string]string{
					"fromPath": absPath,
				},
			})
		}

		m.installFiles = append(m.installFiles, fullInstallPath)
	}

	m.packagingSpecs = append(m.packagingSpecs, PackagingSpec{
		relPathInPackage: Rel(m, fullInstallPath.PartitionDir(), fullInstallPath.String()),
		srcPath:          nil,
		symlinkTarget:    absPath,
		executable:       false,
		partition:        fullInstallPath.partition,
		skipInstall:      m.skipInstall(),
		aconfigPaths:     m.getAconfigPaths(),
	})

	return fullInstallPath
}

func (m *moduleContext) CheckbuildFile(srcPath Path) {
	m.checkbuildFiles = append(m.checkbuildFiles, srcPath)
}

func (m *moduleContext) blueprintModuleContext() blueprint.ModuleContext {
	return m.bp
}

func (m *moduleContext) LicenseMetadataFile() Path {
	return m.module.base().licenseMetadataFile
}

func (m *moduleContext) ModuleInfoJSON() *ModuleInfoJSON {
	if moduleInfoJSON := m.module.base().moduleInfoJSON; moduleInfoJSON != nil {
		return moduleInfoJSON
	}
	moduleInfoJSON := &ModuleInfoJSON{}
	m.module.base().moduleInfoJSON = moduleInfoJSON
	return moduleInfoJSON
}

// Returns a list of paths expanded from globs and modules referenced using ":module" syntax.  The property must
// be tagged with `android:"path" to support automatic source module dependency resolution.
//
// Deprecated: use PathsForModuleSrc or PathsForModuleSrcExcludes instead.
func (m *moduleContext) ExpandSources(srcFiles, excludes []string) Paths {
	return PathsForModuleSrcExcludes(m, srcFiles, excludes)
}

// Returns a single path expanded from globs and modules referenced using ":module" syntax.  The property must
// be tagged with `android:"path" to support automatic source module dependency resolution.
//
// Deprecated: use PathForModuleSrc instead.
func (m *moduleContext) ExpandSource(srcFile, _ string) Path {
	return PathForModuleSrc(m, srcFile)
}

// Returns an optional single path expanded from globs and modules referenced using ":module" syntax if
// the srcFile is non-nil.  The property must be tagged with `android:"path" to support automatic source module
// dependency resolution.
func (m *moduleContext) ExpandOptionalSource(srcFile *string, _ string) OptionalPath {
	if srcFile != nil {
		return OptionalPathForPath(PathForModuleSrc(m, *srcFile))
	}
	return OptionalPath{}
}

func (m *moduleContext) RequiredModuleNames() []string {
	return m.module.RequiredModuleNames()
}

func (m *moduleContext) HostRequiredModuleNames() []string {
	return m.module.HostRequiredModuleNames()
}

func (m *moduleContext) TargetRequiredModuleNames() []string {
	return m.module.TargetRequiredModuleNames()
}
