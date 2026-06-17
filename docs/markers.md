# `+cobra:*` marker reference

`gen-cobra-flags` is driven by marker comments placed in the doc comment of a struct or of
its fields. Markers use the form:

```
// +cobra:<name>=<value>
```

Markers are read from the comment block immediately preceding a `type` declaration
(struct-level) or a field declaration (field-level). A struct is only processed if it carries
a `+cobra:flag` marker.

## Struct-level markers

These apply to the doc comment of the annotated struct. They control the single "config"
flag that accepts the whole struct as JSON/YAML.

| Marker                 | Required | Description                                                                                  |
| ---------------------- | -------- | -------------------------------------------------------------------------------------------- |
| `+cobra:flag`          | yes      | Name of the aggregate config flag. Presence of this marker marks the struct for generation.  |
| `+cobra:short`         | no       | Single-character shorthand for the aggregate flag.                                           |
| `+cobra:usage`         | no       | Usage/help string for the aggregate flag.                                                    |
| `+cobra:default`       | no       | Default value for the aggregate flag. Defaults to `""`.                                       |
| `+cobra:flag:adaptor`  | no       | Adaptor used to turn the flag string into the config struct. Defaults to `adaptors.JsonOrYamlToStruct[<Struct>Config]`. |

## Field-level markers

These apply to the doc comment of a struct field. A field is only turned into a flag if it
carries a `+cobra:flag` marker.

| Marker                  | Required | Description                                                                                       |
| ----------------------- | -------- | ------------------------------------------------------------------------------------------------- |
| `+cobra:flag`           | yes\*    | Flag name for this field. Without it, the field is not exposed as its own flag.                   |
| `+cobra:short`          | no       | Single-character shorthand for the field's flag.                                                  |
| `+cobra:usage`          | no       | Usage/help string for the field's flag.                                                           |
| `+cobra:default`        | no       | Default value for the flag. Defaults to the zero value of the flag type.                          |
| `+cobra:json`           | no       | JSON tag emitted on the generated config struct field.                                            |
| `+cobra:yaml`           | no       | YAML tag emitted on the generated config struct field.                                            |
| `+cobra:customTags`     | no       | Additional raw struct tags appended verbatim to the generated config field.                       |
| `+cobra:option`         | no       | Marks the field as an "option" (surfaced through the generated `<Struct>Options` type).           |
| `+cobra:flag:type`      | no       | Overrides the Cobra flag type (e.g. `String`, `Int64`). Defaults to the type inferred from the field. |
| `+cobra:flag:adaptor`   | no       | Custom adaptor converting the flag value into the config field value.                             |
| `+cobra:config:type`    | no       | Overrides the config struct field's Go type.                                                      |
| `+cobra:config:adaptor` | no       | Custom adaptor converting the config field value into the destination domain field value.         |

\* Required only for fields you want exposed as individual flags.

## Type and adaptor resolution

When no explicit `+cobra:flag:type` is given, the flag type is inferred from the field's Go
type. When no `+cobra:flag:adaptor` is given, an adaptor is chosen automatically:

- If the config type is a pointer to the flag's Go type, `adaptors.ToPtr` is used.
- Otherwise a conversion function is looked up by input/output type names in the runtime SDK
  (`sdk/pkg/adaptors`). If none is found, `adaptors.ToPtr` is used as a fallback.

Custom adaptors named via `+cobra:flag:adaptor` / `+cobra:config:adaptor` are referenced in
generated code as `adaptor<Name>`; you are expected to provide a function with that name in
the generated package.

## `+cobra:required`

`+cobra:required` is accepted (it appears in the example fixture) but is **not currently
consumed** by the generator — it has no effect on the generated output. It is documented here
for completeness; do not rely on it to enforce required flags.

## Example

```go
// CreateFooBarRequest represents a request to create a FooBar resource
// +cobra:flag=config
// +cobra:short=c
// +cobra:usage=Configuration for the server
type CreateFooBarRequest struct {
    // +cobra:flag=name
    // +cobra:short=N
    // +cobra:usage=Name of FooBar
    // +cobra:json=Name
    // +cobra:yaml=Name
    // +cobra:default=""
    // +cobra:customTags=validate:"true" example:"custom"
    Name string

    // +cobra:flag=a-conversion-of-types
    // +cobra:usage=A conversion of types
    // +cobra:default=1
    // +cobra:flag:type=Int64
    // +cobra:flag:adaptor=CustomInt64ToInt32
    // +cobra:config:type=int32
    // +cobra:config:adaptor=CustomInt32ToString
    AConversionOfTypes string
}
```

See [`example/config.go`](../example/config.go) for a complete, exercised example.
