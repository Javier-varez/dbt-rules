package core

type Pool struct {
	Name  string
	Depth uint
}

var ConsolePool Pool = Pool{
	Name: "console",
}

type Context interface {
	AddBuildStep(BuildStep)
	AddBuildStepWithRule(BuildStepWithRule)
	Cwd() OutPath
	BuildChild(c BuildInterface)

	// WithTrace calls the given function, with the given value added
	// to the trace.
	WithTrace(id string, f func(Context))

	// Trace returns the strings in the current trace (most recent last).
	Trace() []string

	// Rules to be collected in a compilation database. Their name must be unique
	RegisterCompDbRule(rule *BuildRule)

	// Obtains a registered rule (if it exists). The boolean is true if it exists
	GetCompDbRule(name string) (*BuildRule, bool)

	registerPool(pool Pool) error
}

// BuildStep represents one build step (i.e., one build command).
// Each BuildStep produces `Out` and `Outs` from `Ins` and `In` by running `Cmd`.
type BuildStep struct {
	Out     OutPath
	Outs    []OutPath
	In      Path
	Ins     []Path
	Depfile OutPath
	Cmd     string
	Script  string
	Data    string
	Descr   string
	Phony   bool
	Pool    *Pool
}

type BuildRule struct {
	Name      string
	Variables map[string]string
}

type BuildStepWithRule struct {
	Outs         []OutPath
	Ins          []Path
	ImplicitDeps []Path
	OrderDeps    []Path
	Variables    map[string]string
	Rule         BuildRule
	Phony        bool
	traces       [][]string
}

type TargetRule struct {
	Target    string
	Ins       []string
	Variables map[string]string
}

func (step *BuildStep) outs() []OutPath {
	if step.Out == nil {
		return step.Outs
	}
	return append(step.Outs, step.Out)
}

func (step *BuildStep) ins() []Path {
	if step.In == nil {
		return step.Ins
	}
	return append(step.Ins, step.In)
}

type BuildInterface interface {
	Build(ctx Context)
}

func AssertIsBuildableTarget(iface BuildInterface) {
	// Do nothing. This function is simply supposed to cause a compilation fail if the
	// type passed does not implement the interface
}

type outputsInterface interface {
	Outputs() []Path
}

type descriptionInterface interface {
	Description() string
}

type RunInterface interface {
	Run(args []string) string
}

func AssertIsRunnableTarget(iface RunInterface) {
	// Do nothing. This function is simply supposed to cause a compilation fail if the
	// type passed does not implement the interface
}

type ReportInterface interface {
	Report(allTargets []interface{}, selectedTargets []interface{}) BuildInterface
}

func AssertIsReportTarget(iface ReportInterface) {
	// Do nothing. This function is simply supposed to cause a compilation fail if the
	// type passed does not implement the interface
}

// An interface for runnables that depend on some other set of targets to run, but not to build
// For example, when using a test wrapper.
type ExtendedRunInterface interface {
	RunInterface
	RunDeps() []Path
}

type TestInterface interface {
	Test(args []string) string
}

// An interface for tests that depend on some other set of targets to run, but not to build
// For example, when using a test wrapper.
type ExtendedTestInterface interface {
	TestInterface
	TestDeps() []Path
}

func AssertIsTestableTarget(iface TestInterface) {
	// Do nothing. This function is simply supposed to cause a compilation fail if the
	// type passed does not implement the interface
}

type CoverageInterface interface {
	Test(args []string) string
	Binaries() []Path
	CoverageData() []OutPath
}

func AssertIsCoverageTarget(iface CoverageInterface) {
	// Do nothing. This function is simply supposed to cause a compilation fail if the
	// type passed does not implement the interface
}

type TranslationUnit struct {
	Source Path
	Object OutPath
	Flags  []string
}

// AnalyzeInterface is an interface for targets compatible with static analysis
type AnalyzeInterface interface {
	TranslationUnits(ctx Context) []TranslationUnit
	AnalysisDeps(ctx Context) []AnalyzeInterface
}

func AssertIsAnalyzeTarget(iface AnalyzeInterface) {
	// Do nothing. This function is simply supposed to cause a compilation fail if the
	// type passed does not implement the interface
}

type context struct {
	cwd              OutPath
	nextRuleID       int
	trace            []string
	leafOutputs      map[Path]bool
	buildSteps       map[string]*BuildStepWithRule
	targetRules      []TargetRule
	compDbBuildRules map[string]*BuildRule
	nestedBuild      bool
	pools            map[string]uint
}

func newContext(vars map[string]interface{}) *context {
	ctx := &context{
		cwd:              outPath{""},
		leafOutputs:      map[Path]bool{},
		buildSteps:       map[string]*BuildStepWithRule{},
		compDbBuildRules: map[string]*BuildRule{},
		nestedBuild:      false,
	}
	return ctx
}

func (ctx *context) WithTrace(id string, f func(Context)) {
	ctx.trace = append(ctx.trace, id)
	defer func() {
		ctx.trace = ctx.trace[:len(ctx.trace)-1]
	}()
	f(ctx)
}

func (ctx *context) Trace() []string {
	// We return a copy of the trace, to avoid mutations.
	return append([]string{}, ctx.trace...)
}

// AddBuildStep adds a build step for the current target.
func (ctx *context) AddBuildStep(step BuildStep) {
}

// AddBuildStepWithRule adds a build step for the current target.
func (ctx *context) AddBuildStepWithRule(step BuildStepWithRule) {
}

// Cwd returns the build directory of the current target.
func (ctx *context) Cwd() OutPath {
	return ctx.cwd
}

func (ctx *context) BuildChild(c BuildInterface) {
	nb := ctx.nestedBuild
	ctx.nestedBuild = true
	c.Build(ctx)
	ctx.nestedBuild = nb
}

func (ctx *context) registerPool(pool Pool) error {
	return nil
}

func (ctx *context) handleTarget(targetPath string, target BuildInterface) {
}

func stepsAreEquivalent(a, b *BuildStepWithRule) error {
	if len(a.Ins) != len(b.Ins) {
		// return fmt.Errorf("different number of inputs")
	}
	for i := range a.Ins {
		if a.Ins[i] != b.Ins[i] {
			// return fmt.Errorf("different input at position %d: %s vs %s", i, a.Ins[i], b.Ins[i])
		}
	}

	if len(a.Outs) != len(b.Outs) {
		// return fmt.Errorf("different number of outputs")
	}
	for i := range a.Outs {
		if a.Outs[i] != b.Outs[i] {
			// return fmt.Errorf("different output at position %d: %s vs %s", i, a.Outs[i], b.Outs[i])
		}
	}

	if len(a.Variables) != len(b.Variables) {
		// return fmt.Errorf("different number of variables")
	}
	for name := range a.Variables {
		if a.Variables[name] != b.Variables[name] {
			// return fmt.Errorf("different value for variable '%s' (%s vs %s)", name, a.Variables[name], b.Variables[name])
		}
	}

	if a.Rule.Name != b.Rule.Name {
		// return fmt.Errorf("different build rule")
	}
	if len(a.Rule.Variables) != len(b.Rule.Variables) {
		// return fmt.Errorf("different number of variables in build rule")
	}
	for name := range a.Rule.Variables {
		if a.Rule.Variables[name] != b.Rule.Variables[name] {
			// return fmt.Errorf("different value for of variable '%s' in build rule", name)
		}
	}

	return nil
}

func (ctx *context) ninjaFile() string {
	return ""
}

func (ctx *context) RegisterCompDbRule(rule *BuildRule) {
	ctx.compDbBuildRules[rule.Name] = rule
}

func (ctx *context) GetCompDbRule(name string) (*BuildRule, bool) {
	buildRule, ok := ctx.compDbBuildRules[name]
	return buildRule, ok
}

func ninjaEscape(s string) string {
	return s
}
