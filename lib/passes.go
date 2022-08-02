package lib

const (
	contextPassFileSuffix         = "_pass_ctx.go"
	instrumentationPassFileSuffix = "_pass_tracing.go"
)

func ExecutePassesDumpIr(projectPath string,
	packagePattern string,
	rootFunctions []FuncDescriptor,
	funcDecls map[FuncDescriptor]bool,
	backwardCallGraph map[FuncDescriptor][]FuncDescriptor) {

	Instrument(projectPath,
		packagePattern,
		backwardCallGraph,
		rootFunctions,
		instrumentationPassFileSuffix)

	PropagateContext(projectPath,
		packagePattern,
		backwardCallGraph,
		rootFunctions,
		funcDecls,
		contextPassFileSuffix)

}

func ExecutePasses(projectPath string,
	packagePattern string,
	rootFunctions []FuncDescriptor,
	funcDecls map[FuncDescriptor]bool,
	backwardCallGraph map[FuncDescriptor][]FuncDescriptor) {

	Instrument(projectPath,
		packagePattern,
		backwardCallGraph,
		rootFunctions,
		"")

	PropagateContext(projectPath,
		packagePattern,
		backwardCallGraph,
		rootFunctions,
		funcDecls,
		"")

}
