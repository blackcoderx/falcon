package tools

import (
	"github.com/blackcoderx/falcon/pkg/core"
	zapagent "github.com/blackcoderx/falcon/pkg/core/tools/agent"
	"github.com/blackcoderx/falcon/pkg/core/tools/data_driven_engine"
	"github.com/blackcoderx/falcon/pkg/core/tools/debugging"
	"github.com/blackcoderx/falcon/pkg/core/tools/functional_test_generator"
	"github.com/blackcoderx/falcon/pkg/core/tools/idempotency_verifier"
	"github.com/blackcoderx/falcon/pkg/core/tools/integration_orchestrator"
	"github.com/blackcoderx/falcon/pkg/core/tools/performance_engine"
	"github.com/blackcoderx/falcon/pkg/core/tools/persistence"
	"github.com/blackcoderx/falcon/pkg/core/tools/regression_watchdog"
	"github.com/blackcoderx/falcon/pkg/core/tools/security_scanner"
	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"github.com/blackcoderx/falcon/pkg/core/tools/smoke_runner"
	"github.com/blackcoderx/falcon/pkg/core/tools/spec_ingester"
	"github.com/blackcoderx/falcon/pkg/llm"
)

// Registry handles the initialization and registration of all Falcon tools.
// It uses a component-based approach to avoid a single monolithic registration function.
type Registry struct {
	Agent          *core.Agent
	LLMClient      llm.LLMClient
	WorkDir        string
	FalconDir      string
	MemStore       *core.MemoryStore
	ConfirmManager *shared.ConfirmationManager

	// Shared services
	ResponseManager *shared.ResponseManager
	VariableStore   *shared.VariableStore
	PersistManager  *persistence.PersistenceManager
	HTTPTool        *shared.HTTPTool // Shared HTTP tool instance
}

// NewRegistry creates a new tool registry with the necessary dependencies.
func NewRegistry(
	agent *core.Agent,
	llmClient llm.LLMClient,
	workDir string,
	falconDir string,
	memStore *core.MemoryStore,
	confirmManager *shared.ConfirmationManager,
) *Registry {
	return &Registry{
		Agent:          agent,
		LLMClient:      llmClient,
		WorkDir:        workDir,
		FalconDir:      falconDir,
		MemStore:       memStore,
		ConfirmManager: confirmManager,
	}
}

// RegisterAllTools initializes services and registers all tool categories.
func (r *Registry) RegisterAllTools() {
	r.initServices()
	r.registerSharedTools()
	r.registerDebuggingTools()
	r.registerPersistenceTools()
	r.registerAgentTools()
	r.registerSpecIngesterTools()
	r.registerFunctionalTestGeneratorTools()
	r.registerSecurityScannerTools()
	r.registerPerformanceEngineTools()
	r.registerModuleTools()
	r.registerWorkflowTools()
}

// initServices initializes shared services used by multiple tools.
func (r *Registry) initServices() {
	r.ResponseManager = shared.NewResponseManager()
	r.VariableStore = shared.NewVariableStore(r.FalconDir)
	r.PersistManager = persistence.NewPersistenceManager(r.FalconDir)
	r.HTTPTool = shared.NewHTTPTool(r.ResponseManager, r.VariableStore)
}

// registerSharedTools registers foundational tools (HTTP, Assertions, Auth, etc).
func (r *Registry) registerSharedTools() {
	// core HTTP tool - shared instance
	r.Agent.RegisterTool(r.HTTPTool)

	// assertions & extraction
	r.Agent.RegisterTool(shared.NewAssertTool(r.ResponseManager))
	r.Agent.RegisterTool(shared.NewExtractTool(r.ResponseManager, r.VariableStore))
	r.Agent.RegisterTool(shared.NewSchemaValidationTool(r.ResponseManager))
	r.Agent.RegisterTool(shared.NewCompareResponsesTool(r.ResponseManager, r.FalconDir))

	// utilities
	r.Agent.RegisterTool(shared.NewWaitTool())
	r.Agent.RegisterTool(shared.NewRetryTool(r.Agent))

	// unified auth tool (replaces auth_bearer, auth_basic, auth_oauth2, auth_helper)
	r.Agent.RegisterTool(shared.NewAuthTool(r.ResponseManager, r.VariableStore))

	// webhooks
	r.Agent.RegisterTool(shared.NewWebhookListenerTool(r.VariableStore))

	// test suites
	r.Agent.RegisterTool(shared.NewTestSuiteTool(
		r.HTTPTool,
		shared.NewAssertTool(r.ResponseManager),
		shared.NewExtractTool(r.ResponseManager, r.VariableStore),
		r.ResponseManager,
		r.VariableStore,
		r.FalconDir,
	))

	// .falcon-scoped read/write tools
	r.Agent.RegisterTool(shared.NewFalconWriteTool(r.FalconDir))
	r.Agent.RegisterTool(shared.NewFalconReadTool(r.FalconDir))
	r.Agent.RegisterTool(shared.NewSessionLogTool(r.FalconDir))
}

// registerDebuggingTools registers tools for code analysis and fixing.
func (r *Registry) registerDebuggingTools() {
	// file operations
	r.Agent.RegisterTool(debugging.NewReadFileTool(r.WorkDir))
	r.Agent.RegisterTool(debugging.NewListFilesTool(r.WorkDir))
	r.Agent.RegisterTool(debugging.NewWriteFileTool(r.WorkDir, r.ConfirmManager))
	r.Agent.RegisterTool(debugging.NewSearchCodeTool(r.WorkDir))

	// code analysis
	r.Agent.RegisterTool(debugging.NewFindHandlerTool(r.WorkDir))
	r.Agent.RegisterTool(debugging.NewAnalyzeEndpointTool(r.LLMClient))
	r.Agent.RegisterTool(debugging.NewAnalyzeFailureTool(r.LLMClient))
	r.Agent.RegisterTool(debugging.NewProposeFixTool(r.LLMClient))
	r.Agent.RegisterTool(debugging.NewCreateTestFileTool(r.LLMClient))
}

// registerPersistenceTools registers tools for saving state and requests.
func (r *Registry) registerPersistenceTools() {
	// variables
	r.Agent.RegisterTool(persistence.NewVariableTool(r.VariableStore))

	// unified request management (replaces save_request, load_request, list_requests)
	r.Agent.RegisterTool(persistence.NewRequestTool(r.PersistManager))

	// unified environment management (replaces set_environment, list_environments)
	r.Agent.RegisterTool(persistence.NewEnvironmentTool(r.PersistManager))
}

// registerAgentTools registers memory, reporting, and orchestration tools.
func (r *Registry) registerAgentTools() {
	r.Agent.RegisterTool(zapagent.NewMemoryTool(r.MemStore))

	assertTool := shared.NewAssertTool(r.ResponseManager)

	// run_tests now handles both single and bulk execution via optional scenario param
	runTests := zapagent.NewRunTestsTool(r.FalconDir, r.HTTPTool, assertTool, r.VariableStore)
	r.Agent.RegisterTool(runTests)

	// auto test orchestrator
	r.Agent.RegisterTool(zapagent.NewAutoTestTool(
		r.LLMClient,
		debugging.NewAnalyzeEndpointTool(r.LLMClient),
		runTests,
		debugging.NewAnalyzeFailureTool(r.LLMClient),
	))
}

// registerSpecIngesterTools registers spec-to-graph transformation tools.
func (r *Registry) registerSpecIngesterTools() {
	r.Agent.RegisterTool(spec_ingester.NewIngestSpecTool(r.LLMClient, r.FalconDir))
}

// registerFunctionalTestGeneratorTools registers spec-driven functional test generator.
func (r *Registry) registerFunctionalTestGeneratorTools() {
	assertTool := shared.NewAssertTool(r.ResponseManager)
	r.Agent.RegisterTool(functional_test_generator.NewFunctionalTestGeneratorTool(r.FalconDir, r.HTTPTool, assertTool))
}

// registerSecurityScannerTools registers the whole security scanner ecosystem.
func (r *Registry) registerSecurityScannerTools() {
	r.Agent.RegisterTool(security_scanner.NewSecurityScannerTool(r.FalconDir, r.HTTPTool))
}

// registerPerformanceEngineTools registers the multi-mode performance engine.
func (r *Registry) registerPerformanceEngineTools() {
	r.Agent.RegisterTool(performance_engine.NewPerformanceEngineTool(r.FalconDir, r.HTTPTool))
}

// registerModuleTools registers high-level capability modules.
func (r *Registry) registerModuleTools() {
	r.Agent.RegisterTool(smoke_runner.NewSmokeRunnerTool(r.FalconDir, r.HTTPTool))
	r.Agent.RegisterTool(idempotency_verifier.NewIdempotencyVerifierTool(r.FalconDir, r.HTTPTool))
	r.Agent.RegisterTool(data_driven_engine.NewDataDrivenEngineTool(r.FalconDir, r.HTTPTool))
}

// registerWorkflowTools registers integration and regression modules.
func (r *Registry) registerWorkflowTools() {
	r.Agent.RegisterTool(integration_orchestrator.NewIntegrationOrchestratorTool(r.FalconDir, r.HTTPTool))
	r.Agent.RegisterTool(regression_watchdog.NewRegressionWatchdogTool(r.FalconDir, r.HTTPTool))
}
