# GXDocGen

GXDocGen is a CLI tool that automatically generates documentation for GeneXus objects (Procedures, APIs, SDTs) from XPZ exports. It extracts structured comments and metadata to produce Markdown and OpenAPI documentation.

---

## Annotation Standard

| Tag                 | Required | Description                                                                                                        |               
| ------------------- | -------- | -------------------------------------------------------------------------------------------------------------------|
| `@package`          | ✅        | Logical grouping (like Go packages or modules). Used to group Markdown files and OpenAPI tags.                    |                 
| `@summary`          | ✅        | Short summary (1–2 lines). Used as title and for OpenAPI `summary`.                                               |                
| `@description`      | ✅        | Extended explanation, Markdown-compatible. Used for OpenAPI `description`.                                        |                
| `@author`           | ⚙️       | Developer responsible for creation.                                                                                |              
| `@created`          | ⚙️       | Date in ISO format (YYYY-MM-DD).                                                                                   |              
| `@param`            | ⚙️       | Describes a procedure parameter (IN/OUT). Syntax: `@param name [IN|OUT] Type:TypeName - Description`               |
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
│   ├── xpz/               # XPZ extraction & XML parsing
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
| **xpz/**       | Unzip `.xpz` → read XML files → return parsed GX object metadata.                                  |
| **parser/**    | Extracts `/** ... */` comment blocks, identifies `@` tags, builds a structured `DocComment` model. |
| **model/**     | Defines entities: `GXObject`, `ProcedureDoc`, `ParameterDoc`, etc.                                 |
| **generator/** | Converts `DocComment` → Markdown and/or OpenAPI spec.                                              |
| **utils/**     | Logging, file management, error handling, JSON prettifying.                                        |

---

## License
See [LICENSE.md](./LICENSE.md)

## Contributing
See [CONTRIBUTING.md](./CONTRIBUTING.md)
