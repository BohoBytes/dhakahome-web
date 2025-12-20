# DhakaHome Web – Documentation Hub

Your single place to find, update, and navigate DhakaHome Web documentation.

## Start Here
- Get running locally: [QUICK_START_GUIDE.md](QUICK_START_GUIDE.md)
- Understand the system: [PROJECT_ARCHITECTURE.md](PROJECT_ARCHITECTURE.md)
- Switch environments: [ENVIRONMENTS.md](ENVIRONMENTS.md)
- Work without the API: [Mocking Guide](MockingGuide/MOCK_MODE.md) and [Quick Start](MockingGuide/QUICK_START_MOCK.md)
- Nestlo API reference (vendor-managed): [docs/NestloAPI](NestloAPI/)

## Task Navigation
- Run the app locally → Quick Start (“Local setup”)
- Add or edit a page/component → Quick Start (“Creating a new page”) + Project Architecture (“Templates & partials”)
- Update search filters or data sources → Project Architecture (“Search experience” + “API client”)
- Build for release → Quick Start (“Build for deployment”)
- Troubleshoot API downtime → Mock Mode docs

## Document Map
- [QUICK_START_GUIDE.md](QUICK_START_GUIDE.md): Setup, run, create pages, troubleshooting
- [PROJECT_ARCHITECTURE.md](PROJECT_ARCHITECTURE.md): Routes, render flow, API integration, styling, best practices
- [ARCHITECTURE_DIAGRAMS.txt](ARCHITECTURE_DIAGRAMS.txt): Text diagrams for request/search/property flows
- [ENVIRONMENTS.md](ENVIRONMENTS.md): `.env` files, `ENV_FILE` usage, VS Code launch configs, safety rules
- [MockingGuide/MOCK_MODE.md](MockingGuide/MOCK_MODE.md): Full mock-mode reference (data set, filters, behavior)
- [MockingGuide/MOCK_MODE_SUMMARY.md](MockingGuide/MOCK_MODE_SUMMARY.md): High-level summary and implementation notes
- [MockingGuide/QUICK_START_MOCK.md](MockingGuide/QUICK_START_MOCK.md): 3-step offline runbook
- [NestloAPI/README.md](NestloAPI/README.md) & [DhakaHome-API-Integration-Guide.md](NestloAPI/DhakaHome-API-Integration-Guide.md): Official Nestlo API guides (do not edit)

## Maintenance Notes
- Keep secrets out of docs; use placeholders when referencing `.env` values.
- When adding routes, handlers, or partials, update the Quick Start and Architecture docs.
- Align diagrams with the current router and template structure after significant changes.

Last updated: 2025-12-19
