package tools

import (
	"github.com/blackcoderx/zap/pkg/core"
	zapagent "github.com/blackcoderx/zap/pkg/core/tools/agent"
	"github.com/blackcoderx/zap/pkg/core/tools/debugging"
	"github.com/blackcoderx/zap/pkg/core/tools/persistence"
	"github.com/blackcoderx/zap/pkg/core/tools/shared"
	"github.com/blackcoderx/zap/pkg/core/tools/spec_ingester"
	"github.com/blackcoderx/zap/pkg/llm"
)

// Registry handles the initialization and registration of all ZAP tools.
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
	// Future: r.registerModuleTools()
}

// initServices initializes shared services used by multiple tools.
func (r *Registry) initServices() {
	r.ResponseManager = shared.NewResponseManager()
	r.VariableStore = shared.NewVariableStore(r.ZapDir)
	r.PersistManager = persistence.NewPersistenceManager(r.ZapDir)
}

// registerSharedTools registers foundational tools (HTTP, Assertions, Auth, etc).
func (r *Registry) registerSharedTools() {
	// core tools
	httpTool := shared.NewHTTPTool(r.ResponseManager, r.VariableStore)
	r.Agent.RegisterTool(httpTool)

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
	r.Agent.RegisterTool(shared.NewPerformanceTool(httpTool, r.VariableStore))

	// suites
	r.Agent.RegisterTool(shared.NewTestSuiteTool(
		httpTool,
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

	// orchestration dependencies
	httpTool := shared.NewHTTPTool(r.ResponseManager, r.VariableStore)
	assertTool := shared.NewAssertTool(r.ResponseManager)

	runTests := zapagent.NewRunTestsTool(httpTool, assertTool, r.VariableStore)
	r.Agent.RegisterTool(runTests)
	r.Agent.RegisterTool(zapagent.NewRunSingleTestTool(httpTool, assertTool, r.VariableStore))

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
