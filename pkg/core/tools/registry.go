package tools

import (
	"github.com/blackcoderx/falcon/pkg/core"
	zapagent "github.com/blackcoderx/falcon/pkg/core/tools/agent"
	"github.com/blackcoderx/falcon/pkg/core/tools/api_drift_analyzer"
	"github.com/blackcoderx/falcon/pkg/core/tools/breaking_change_detector"
	"github.com/blackcoderx/falcon/pkg/core/tools/data_driven_engine"
	"github.com/blackcoderx/falcon/pkg/core/tools/debugging"
	"github.com/blackcoderx/falcon/pkg/core/tools/dependency_mapper"
	"github.com/blackcoderx/falcon/pkg/core/tools/documentation_validator"
	"github.com/blackcoderx/falcon/pkg/core/tools/functional_test_generator"
	"github.com/blackcoderx/falcon/pkg/core/tools/idempotency_verifier"
	"github.com/blackcoderx/falcon/pkg/core/tools/integration_orchestrator"
	"github.com/blackcoderx/falcon/pkg/core/tools/performance_engine"
	"github.com/blackcoderx/falcon/pkg/core/tools/persistence"
	"github.com/blackcoderx/falcon/pkg/core/tools/regression_watchdog"
	"github.com/blackcoderx/falcon/pkg/core/tools/schema_conformance"
	"github.com/blackcoderx/falcon/pkg/core/tools/security_scanner"
	"github.com/blackcoderx/falcon/pkg/core/tools/shared"
	"github.com/blackcoderx/falcon/pkg/core/tools/smoke_runner"
	"github.com/blackcoderx/falcon/pkg/core/tools/spec_ingester"
	"github.com/blackcoderx/falcon/pkg/core/tools/unit_test_scaffolder"
	"github.com/blackcoderx/falcon/pkg/llm"
)

// Registry handles the initialization and registration of all Falcon tools.
// It uses a component-based approach to avoid a single monolithic registration function.
type Registry struct {
	Agent          *core.Agent
	LLMClient      llm.LLMClient
	WorkDir        string
	ZapDir         string
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
	zapDir string,
	memStore *core.MemoryStore,
	confirmManager *shared.ConfirmationManager,
) *Registry {
	return &Registry{
		Agent:          agent,
		LLMClient:      llmClient,
		WorkDir:        workDir,
		ZapDir:         zapDir,
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
	r.VariableStore = shared.NewVariableStore(r.ZapDir)
	r.PersistManager = persistence.NewPersistenceManager(r.ZapDir)
	r.HTTPTool = shared.NewHTTPTool(r.ResponseManager, r.VariableStore) // Initialize once
}

// registerSharedTools registers foundational tools (HTTP, Assertions, Auth, etc).
func (r *Registry) registerSharedTools() {
	// core tools - use the shared instance
	r.Agent.RegisterTool(r.HTTPTool)

	// assertions & extraction
	r.Agent.RegisterTool(shared.NewAssertTool(r.ResponseManager))
	r.Agent.RegisterTool(shared.NewExtractTool(r.ResponseManager, r.VariableStore))
	r.Agent.RegisterTool(shared.NewSchemaValidationTool(r.ResponseManager))
	r.Agent.RegisterTool(shared.NewCompareResponsesTool(r.ResponseManager, r.ZapDir))

	// utilities
	r.Agent.RegisterTool(shared.NewWaitTool())
	r.Agent.RegisterTool(shared.NewRetryTool(r.Agent))

	// auth tools
	r.Agent.RegisterTool(shared.NewBearerTool(r.VariableStore))
	r.Agent.RegisterTool(shared.NewBasicTool(r.VariableStore))
	r.Agent.RegisterTool(shared.NewHelperTool(r.ResponseManager, r.VariableStore))
	r.Agent.RegisterTool(shared.NewOAuth2Tool(r.VariableStore))

	// webhooks & performance
	r.Agent.RegisterTool(shared.NewWebhookListenerTool(r.VariableStore))
	r.Agent.RegisterTool(shared.NewPerformanceTool(r.HTTPTool, r.VariableStore))

	// suites
	r.Agent.RegisterTool(shared.NewTestSuiteTool(
		r.HTTPTool,
		shared.NewAssertTool(r.ResponseManager),
		shared.NewExtractTool(r.ResponseManager, r.VariableStore),
		r.ResponseManager,
		r.VariableStore,
		r.ZapDir,
	))
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
	r.Agent.RegisterTool(debugging.NewGenerateTestsTool(r.LLMClient))
	r.Agent.RegisterTool(debugging.NewProposeFixTool(r.LLMClient))
	r.Agent.RegisterTool(debugging.NewCreateTestFileTool(r.LLMClient))
}

// registerPersistenceTools registers tools for saving state and requests.
func (r *Registry) registerPersistenceTools() {
	// variables (wrapped)
	r.Agent.RegisterTool(persistence.NewVariableTool(r.VariableStore))

	// request management
	r.Agent.RegisterTool(persistence.NewSaveRequestTool(r.PersistManager))
	r.Agent.RegisterTool(persistence.NewLoadRequestTool(r.PersistManager))
	r.Agent.RegisterTool(persistence.NewListRequestsTool(r.PersistManager))

	// environment management
	r.Agent.RegisterTool(persistence.NewSetEnvironmentTool(r.PersistManager))
	r.Agent.RegisterTool(persistence.NewListEnvironmentsTool(r.PersistManager))
}

// registerAgentTools registers memory, reporting, and orchestration tools.
func (r *Registry) registerAgentTools() {
	r.Agent.RegisterTool(zapagent.NewMemoryTool(r.MemStore))
	r.Agent.RegisterTool(zapagent.NewExportResultsTool(r.ZapDir))

	// orchestration dependencies - use shared instances
	assertTool := shared.NewAssertTool(r.ResponseManager)

	runTests := zapagent.NewRunTestsTool(r.HTTPTool, assertTool, r.VariableStore)
	r.Agent.RegisterTool(runTests)
	r.Agent.RegisterTool(zapagent.NewRunSingleTestTool(r.HTTPTool, assertTool, r.VariableStore))

	// auto test orchestrator
	r.Agent.RegisterTool(zapagent.NewAutoTestTool(
		debugging.NewAnalyzeEndpointTool(r.LLMClient),
		debugging.NewGenerateTestsTool(r.LLMClient),
		runTests,
		debugging.NewAnalyzeFailureTool(r.LLMClient),
	))
}

// registerSpecIngesterTools registers spec-to-graph transformation tools.
func (r *Registry) registerSpecIngesterTools() {
	r.Agent.RegisterTool(spec_ingester.NewIngestSpecTool(r.LLMClient, r.ZapDir))
}

// registerFunctionalTestGeneratorTools registers spec-driven functional test generator.
func (r *Registry) registerFunctionalTestGeneratorTools() {
	assertTool := shared.NewAssertTool(r.ResponseManager)
	r.Agent.RegisterTool(functional_test_generator.NewFunctionalTestGeneratorTool(r.ZapDir, r.HTTPTool, assertTool))
}

// registerSecurityScannerTools registers the whole security scanner ecosystem.
func (r *Registry) registerSecurityScannerTools() {
	r.Agent.RegisterTool(security_scanner.NewSecurityScannerTool(r.ZapDir, r.HTTPTool))
}

// registerPerformanceEngineTools registers the multi-mode performance engine.
func (r *Registry) registerPerformanceEngineTools() {
	r.Agent.RegisterTool(performance_engine.NewPerformanceEngineTool(r.ZapDir, r.HTTPTool))
}

// registerModuleTools registers the high-level capability modules (Smoke, Idempotency, etc).
func (r *Registry) registerModuleTools() {
	r.Agent.RegisterTool(smoke_runner.NewSmokeRunnerTool(r.ZapDir, r.HTTPTool))
	r.Agent.RegisterTool(unit_test_scaffolder.NewUnitTestScaffolderTool(r.LLMClient))
	r.Agent.RegisterTool(idempotency_verifier.NewIdempotencyVerifierTool(r.ZapDir, r.HTTPTool))
	r.Agent.RegisterTool(data_driven_engine.NewDataDrivenEngineTool(r.HTTPTool))

	// Sprint 9 registrations
	r.Agent.RegisterTool(schema_conformance.NewSchemaConformanceTool(r.ZapDir, r.HTTPTool))
	r.Agent.RegisterTool(breaking_change_detector.NewBreakingChangeDetectorTool(r.ZapDir))
	r.Agent.RegisterTool(dependency_mapper.NewDependencyMapperTool(r.ZapDir))
	r.Agent.RegisterTool(documentation_validator.NewDocumentationValidatorTool(r.ZapDir))
	r.Agent.RegisterTool(api_drift_analyzer.NewAPIDriftAnalyzerTool(r.ZapDir, r.HTTPTool))
}

// registerWorkflowTools registers integration and regression modules.
func (r *Registry) registerWorkflowTools() {
	r.Agent.RegisterTool(integration_orchestrator.NewIntegrationOrchestratorTool(r.ZapDir, r.HTTPTool))
	r.Agent.RegisterTool(regression_watchdog.NewRegressionWatchdogTool(r.ZapDir, r.HTTPTool))
}
