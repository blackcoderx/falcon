package prompt

import (
	"fmt"
)

// BuildContextSection generates dynamic context about the current session.
// This includes .zap folder state, active environment, and framework hints.
func BuildContextSection(zapFolder, framework, manifestSummary, memoryPreview string) string {
	var result string

	// .zap folder context
	result += "# SESSION CONTEXT\n\n"
	result += fmt.Sprintf("**Workspace**: %s\n", zapFolder)

	if manifestSummary != "" {
		result += fmt.Sprintf("**State**: %s\n", manifestSummary)
	}

	result += "\n**Folder Structure** (tool → path):\n"
	result += "```\n"
	result += ".zap/\n"
	result += "├── snapshots/      # ingest_spec → api-graph.json\n"
	result += "├── baselines/      # check_regression (read + write)\n"
	result += "├── requests/       # save_request / load_request / list_requests\n"
	result += "├── environments/   # set_environment / list_environments\n"
	result += "├── exports/        # export_results\n"
	result += "├── runs/           # run_tests / run_single_test\n"
	result += "└── state/          # variable(scope=global) + memory\n"
	result += "```\n\n"

	// Framework-specific hints (if configured)
	if framework != "" && framework != "other" {
		result += BuildFrameworkHints(framework)
	}

	// Memory preview (if available)
	if memoryPreview != "" {
		result += "## Long-Term Memory\n\n"
		result += memoryPreview + "\n\n"
	}

	return result
}

// BuildFrameworkHints returns compact, actionable framework-specific patterns.
func BuildFrameworkHints(framework string) string {
	result := fmt.Sprintf("## Framework: %s\n\n", framework)
	result += "**Search Patterns** (prioritize these when debugging):\n\n"

	switch framework {
	case "gin":
		result += "- Routes: `r.GET(`, `r.POST(`, `router.Group(`\n"
		result += "- Context: `c.JSON(`, `c.BindJSON(`, `c.Param(`\n"
		result += "- Errors: `c.AbortWithStatusJSON(`\n"

	case "echo":
		result += "- Routes: `e.GET(`, `e.POST(`, `e.Group(`\n"
		result += "- Context: `c.JSON(`, `c.Bind(`, `c.Param(`\n"
		result += "- Errors: `echo.NewHTTPError(`\n"

	case "chi":
		result += "- Routes: `r.Get(`, `r.Post(`, `r.Route(`\n"
		result += "- Context: `chi.URLParam(`, `render.JSON(`\n"

	case "fiber":
		result += "- Routes: `app.Get(`, `app.Post(`, `app.Group(`\n"
		result += "- Context: `c.JSON(`, `c.BodyParser(`, `c.Params(`\n"
		result += "- Errors: `fiber.NewError(`\n"

	case "fastapi":
		result += "- Routes: `@app.get(`, `@app.post(`, `@router.`\n"
		result += "- Models: `BaseModel`, `Field(`\n"
		result += "- Errors: `raise HTTPException(`, 422 = Pydantic validation\n"

	case "flask":
		result += "- Routes: `@app.route(`, `@blueprint.route(`\n"
		result += "- Request: `request.json`, `request.args.get(`\n"
		result += "- Errors: `@app.errorhandler(`\n"

	case "django":
		result += "- Views: `@api_view(`, `APIView`, `ViewSet`\n"
		result += "- Serializers: `serializers.`, `ModelSerializer`\n"
		result += "- Errors: `raise ValidationError(`\n"

	case "express":
		result += "- Routes: `app.get(`, `router.post(`\n"
		result += "- Request: `req.body`, `req.params`, `req.query`\n"
		result += "- Errors: `next(error)`, middleware error handler\n"

	case "nestjs":
		result += "- Controllers: `@Controller(`, `@Get(`, `@Post(`\n"
		result += "- DTOs: `@IsString`, `@IsNotEmpty`, `class-validator`\n"
		result += "- Errors: `throw new HttpException(`\n"

	case "hono":
		result += "- Routes: `app.get(`, `app.post(`\n"
		result += "- Context: `c.json(`, `c.req.json(`, `c.req.param(`\n"
		result += "- Errors: `throw new HTTPException(`\n"

	case "spring":
		result += "- Controllers: `@RestController`, `@GetMapping`, `@PostMapping`\n"
		result += "- Request: `@RequestBody`, `@PathVariable`, `@RequestParam`\n"
		result += "- Errors: `@ExceptionHandler`, `@ControllerAdvice`\n"

	case "laravel":
		result += "- Routes: `Route::get(`, `Route::post(`\n"
		result += "- Controllers: `public function`, `Request $request`\n"
		result += "- Errors: `ValidationException`, `abort(`\n"

	case "rails":
		result += "- Routes: `get`, `post`, `resources`\n"
		result += "- Controllers: `def index`, `params[:id]`, `render json:`\n"
		result += "- Errors: `status: :bad_request`\n"

	case "actix":
		result += "- Routes: `web::get()`, `web::resource(`\n"
		result += "- Extractors: `web::Path`, `web::Json`, `web::Query`\n"
		result += "- Errors: `impl ResponseError`\n"

	case "axum":
		result += "- Routes: `Router::new().route(`\n"
		result += "- Extractors: `Path`, `Json`, `Query`, `State`\n"
		result += "- Errors: `impl IntoResponse`\n"

	default:
		result += "- Search for route definitions and handler functions\n"
		result += "- Look for error handling patterns\n"
		result += "- Check validation logic\n"
	}

	result += "\n"
	return result
}
