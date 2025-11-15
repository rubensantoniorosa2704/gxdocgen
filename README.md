# GXDocGen

**Version:** 0.2.0

GXDocGen is a CLI tool that automatically generates documentation for GeneXus objects (Procedures, APIs, SDTs) from XPZ exports. It extracts structured comments and metadata to produce Markdown and OpenAPI documentation.

**✨ New in v0.2.0:** Refactored XML parser with intelligent fallbacks, automatic package detection, and comprehensive test coverage.

---

## Features

- 📦 **Smart Package Detection** - Automatically groups procedures by `@package`, parent module, or name inference
- 🎯 **Multi-Layer Parameter Extraction** - Extracts params from ParmRule, IsParm variables, or Parm() source
- ✍️ **Auto-Documentation** - Generates docs even without annotations using XML metadata
- 🔍 **XPath-Based Parsing** - Clean, maintainable code using xmlquery
- ✅ **Comprehensive Tests** - 20+ tests covering all fallback scenarios
- 🚀 **Fast & Lightweight** - <2ms per procedure, no CGO dependencies

---

## Annotation Standard

Annotations are **optional** but recommended for rich documentation. Without annotations, GXDocGen generates basic docs from XML metadata.

| Tag                 | Required | Description                                                                                                        |               
| ------------------- | -------- | -------------------------------------------------------------------------------------------------------------------|
| `@package`          | ⚙️       | Logical grouping (falls back to parent module or name inference).                    |                 
| `@summary`          | ⚙️       | Short summary (inferred from procedure name if missing).                                               |                
| `@description`      | ⚙️       | Extended explanation (auto-generated if missing).                                        |                
| `@author`           | ⚙️       | Developer responsible for creation.                                                                                |              
| `@created`          | ⚙️       | Date in ISO format (YYYY-MM-DD).                                                                                   |              
| `@param`            | ⚙️       | Describes a parameter (auto-extracted from XML if missing). Syntax: `@param name [IN|OUT] Type - Description`               |
| `@return`           | ⚙️       | Return type or SDT (used for Data Providers or functions).                                                         |
| `@example-request`  | ⚙️       | JSON block example for request body.                                                                               |               
| `@example-response` | ⚙️       | JSON block example for response body.                                                                              |                
| `@tag`              | ⚙️       | Optional OpenAPI tag for grouping endpoints.                                                                       |              
| `@deprecated`       | ⚙️       | Marks an object as deprecated (optional).                                                                          |              

---

## Folder Structure

```
gxdocgen/
├── cmd/
│   └── gxdocgen/          # main.go (CLI entry point)
├── internal/
│   ├── xpz/               # XPZ extraction & XML parsing (xmlquery-based)
│   ├── parser/            # Structured comment parser
│   ├── model/             # Core domain models (Procedure, Parameter, etc.)
│   ├── generator/         # Markdown and OpenAPI generators
│   ├── utils/             # Shared helpers (file ops, logging)
│   └── config/            # CLI config, env, flags
└── docs/                  # Generated docs output
```

---

## Modules

| Module         | Responsibility                                                                                     |
| -------------- | -------------------------------------------------------------------------------------------------- |
| **cmd/**       | CLI entry (flags, subcommands, input/output paths).                                                |
| **xpz/**       | Unzip `.xpz` → parse XML with XPath → extract metadata with intelligent fallbacks.                                  |
| **parser/**    | Extracts `/** ... */` comment blocks, identifies `@` tags, builds a structured `DocComment` model. |
| **model/**     | Defines entities: `GXObject`, `ProcedureDoc`, `ParameterDoc`, etc.                                 |
| **generator/** | Converts `DocComment` → Markdown and/or OpenAPI spec.                                              |
| **utils/**     | Logging, file management, error handling, JSON prettifying.                                        |

---

## License
See [LICENSE.md](./LICENSE.md)

## Contributing
See [CONTRIBUTING.md](./CONTRIBUTING.md)
