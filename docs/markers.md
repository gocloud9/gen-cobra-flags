# `+cobra:*` marker reference

`gen-cobra-flags` is driven by marker comments placed in the doc comment of a struct or of
its fields. Markers use the form:

```
// +cobra:<name>=<value>
```

Markers are read from the comment block immediately preceding a `type` declaration
(struct-level) or a field declaration (field-level). A struct is only processed if it carries
a `+cobra:enabled` marker (see [Selection](#selection)).

## Selection

Which structs are generated is driven by the `+cobra:enabled` struct marker. Every struct
marked `+cobra:enabled` is generated, along with any child structs pulled in by
`+cobra:config:child` / `+cobra:subcommand` fields (see below). This replaces the former
`-struct` CLI flag as the selection mechanism, so a single invocation can generate every
enabled struct in a package:

```
//go:generate gen-cobra-flags -input ./ -output ./ -package types
```

When `-output` resolves to the same directory as `-input`, the generated code is emitted into
the source package, so no source-package import or qualifier is produced. When generating into
a different directory, the source package's import path is derived automatically from the input
directory's Go module (override it with `-source-import` if needed).

For backward compatibility, if no struct in the package declares `+cobra:enabled`, selection
falls back to the prior behavior (every parsed struct, optionally narrowed by `-struct`).

## Struct-level markers

These apply to the doc comment of the annotated struct. They control the single "config"
flag that accepts the whole struct as JSON/YAML.

| Marker                 | Required | Description                                                                                  |
| ---------------------- | -------- | -------------------------------------------------------------------------------------------- |
| `+cobra:enabled`       | yes      | Marks the struct for generation. `+cobra:enabled=false` disables it.                         |
| `+cobra:flag`          | yes      | Name of the aggregate config flag.                                                           |
| `+cobra:short`         | no       | Single-character shorthand for the aggregate flag.                                           |
| `+cobra:usage`         | no       | Usage/help string for the aggregate flag.                                                    |
| `+cobra:default`       | no       | Default value for the aggregate flag. Defaults to `""`.                                       |
| `+cobra:flag:adaptor`  | no       | Adaptor used to turn the flag string into the config struct. Defaults to `adaptors.JsonOrYamlToStruct[<Struct>Config]`. |
| `+cobra:subcommand:config:prefix` | no | Prefix applied to the struct's hoisted required field names in subcommand composed configs (e.g. `Vpc` renames `Name` → `VpcName`). See [`+cobra:subcommand`](#cobrasubcommand). |

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
| `+cobra:required`       | no       | Marks the field as required. Enforced after merging the config and individual flags; see below.   |
| `+cobra:flag:type`      | no       | Overrides the Cobra flag type (e.g. `String`, `Int64`). Defaults to the type inferred from the field. |
| `+cobra:flag:adaptor`   | no       | Custom adaptor converting the flag value into the config field value.                             |
| `+cobra:config:type`    | no       | Overrides the config struct field's Go type.                                                      |
| `+cobra:config:adaptor` | no       | Custom adaptor converting the config field value into the destination domain field value.         |
| `+cobra:config:child`   | no       | Generates a full child config (config struct + flags + `To<Child>`) for the field's element struct. See below. |
| `+cobra:subcommand`     | no       | Generates a composed subcommand config (child config + parent's hoisted required fields) for a repeated resource field. See below. |
| `+cobra:subcommand:config:flag` | no | Name of the composed-config aggregate flag for a subcommand field (defaults to `<child>-config`). See [`+cobra:subcommand`](#cobrasubcommand). |
| `+cobra:subcommand:config:short` | no | Short flag for the composed-config aggregate flag. |
| `+cobra:subcommand:config:usage` | no | Usage text for the composed-config aggregate flag. |
| `+cobra:subcommand:value:flag` | no | Name of the single value flag for a **scalar-slice** subcommand (a `[]string`-style field with no struct child). See [Scalar-slice subcommands](#scalar-slice-subcommands). |
| `+cobra:subcommand:value:short` | no | Short flag for the scalar-slice value flag. |
| `+cobra:subcommand:value:usage` | no | Usage text for the scalar-slice value flag. |

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

For each custom adaptor the generator emits a registration hook in `shared.go`: a package
variable `adaptor<Name>` of the resolved function type plus a `RegisterAdaptor<Name>` setter.
Provide your implementation at runtime (e.g. from an `init()` in a hand-written file in the
generated package) via `RegisterAdaptor<Name>(fn)`. The hook's input type is derived from the
field's (possibly overridden) Cobra flag type and the output type from the config field type.
For example a `map[string]string` config field with `+cobra:flag:type=StringArray` and
`+cobra:flag:adaptor=BuildTags` yields `var adaptorBuildTags func([]string) (map[string]string, error)`
— both `StringSlice` and `StringArray` flag types map to a `[]string` adaptor input.

## `+cobra:required`

`+cobra:required` marks a field as required. Because the generator produces both an
aggregate config flag (e.g. `--config`) and individual per-field flags, requiredness is
**not** enforced via cobra's `MarkFlagRequired` — that would force the individual flag to be
present even when the value was supplied through the aggregate config flag. Instead:

- The flag's usage/help string is prefixed with `[required] ` so the requirement is visible
  in `--help` output.
- After the aggregate config flag and individual flags are merged, the generated
  `<Struct>ConfigFromFlags` checks the field. If it is still its zero value, the function
  returns an error (`required flag <name> not set`). A value supplied through either the
  aggregate config flag or the individual flag satisfies the requirement.

Both `+cobra:required` and `+cobra:required=true` enable it; `+cobra:required=false`
disables it. The marker only applies to fields that also carry `+cobra:flag`.

## `+cobra:config:child`

`+cobra:config:child` placed on a field whose (element) type is a struct declared in the same
package requests that the child struct be generated as its own complete config: a
`<Child>Config` struct, an `Add<Child>Flags(cmd)` registrar, a `<Child>ConfigFromFlags(cmd)`
constructor, and a `To<Child>()` method — the same shape produced for a top-level
`+cobra:enabled` struct.

The child is pulled into generation automatically (it does not need `+cobra:enabled`), and the
markers it needs are synthesized when absent:

- If the child has no `+cobra:flag`, an aggregate flag named `<child-kebab>-config` is added.
- Each child field marked `+cobra:required` is exposed as an individual flag whose name is
  auto-derived from the field name (e.g. `PeerVpcId` → `peer-vpc-id`), unless the field
  already carries explicit markers (which are always respected). Child fields that are **not**
  marked `+cobra:required` are still present in the child config struct (settable via the
  aggregate JSON/YAML flag) but get no individual flag.

In the **parent's** config struct, a `+cobra:config:child` field is represented by the
generated child config type rather than the source domain type — preserving the field's
slice/pointer shape but swapping the element struct for its `<Child>Config`. For example a
field `AddPeers []*types.CreateVpcPeer` becomes `AddPeers []*CreateVpcPeerConfig`, and a
scalar `Peer *types.CreateVpcPeer` becomes `Peer *CreateVpcPeerConfig`. Consequently the
field's decode flag adaptor is `StringSliceToStructSlice[*<Child>Config]` (slice) or
`StringToStruct[*<Child>Config]` (scalar).

The parent's `To<Parent>()` method converts each child config element back to the domain
type by calling the generated `To<Child>()` method. For a slice it allocates the domain
slice and converts element-wise, propagating any conversion error; for a scalar it converts
the single value when non-nil. Struct-slice fields that are **not** marked
`+cobra:config:child` are unaffected and continue to use the domain type directly.

## `+cobra:subcommand`

`+cobra:subcommand` placed on a repeated field generates wiring for operating on a single
element of that collection (a "subcommand" in CLI terms). For a parent `P` with field `Fs`
(element type `C`), it generates in the parent's file a **composed config** that combines the
child config with the parent's required ("hoisted") fields:

- `P<C>Config` — a struct embedding the child's `CConfig` (tagged `json:",inline"`) and a
  per-parent `PChildRequiredFields` struct (tagged `yaml:",inline"`) that holds the parent's
  hoisted required fields. The type name is prefixed with the **parent message name verbatim**
  (e.g. parent `UpdateVpcRequest`, child `CreateVpcPeer` → `UpdateVpcRequestCreateVpcPeerConfig`).
- `PChildRequiredFields` — emitted once per parent (shared across all its subcommands), holding
  the parent's `+cobra:required` fields renamed by the `+cobra:subcommand:config:prefix` marker
  (see below). For example with `prefix=Vpc`, the parent's `Name` field becomes `VpcName`. Each
  hoisted field carries `json`/`yaml` struct tags matching its flag name (e.g.
  `json:"vpc-name" yaml:"vpc-name"`), so the hoisted fields are settable through the composed
  aggregate config flag as well as their individual flags.
- `AddP<C>ConfigFlags(cmd)` — registers a single **composed-config aggregate flag** (a
  JSON/YAML string flag, see `+cobra:subcommand:config:*` below), the child `C`'s individual
  flags (its `+cobra:required` fields), and the parent `P`'s `+cobra:required` flags (under
  their original names). The child's *own* aggregate config flag is intentionally **not**
  registered here; the composed-config aggregate flag supersedes it.
- `P<C>ConfigFromFlags(cmd) (*P, error)` — note the return type is the **parent domain type**
  `*P` (package-qualified when generating cross-package, e.g. `*types.UpdateVpcRequest`), not
  the composed config. It decodes the aggregate flag into the composed config first, then
  overlays any individual child and hoisted parent flags that were set, enforces the hoisted
  required fields post-merge, and finally **converts** the composed config into `*P`: each
  hoisted field is written back to its original parent field (e.g. `VpcName` → `r.Name`) and
  the child is converted via `C`'s generated `To<C>()` and placed onto the parent's singular
  field — appended as a one-element slice for a repeated field, or assigned directly for a
  scalar field.

`+cobra:subcommand` implies the child must be generated, so it is commonly paired with
`+cobra:config:child` on the same field. No `*cobra.Command` object is created; the caller
wires the returned flags/constructor into its own command tree.

### `+cobra:subcommand:config:flag` / `:short` / `:usage`

Placed on the **subcommand field**, these customize the composed-config aggregate flag that
`AddP<C>ConfigFlags` registers and `P<C>ConfigFromFlags` decodes:

- `+cobra:subcommand:config:flag` — the flag name (e.g. `peer-config`). When absent, defaults
  to a kebab-cased `<child>-config`.
- `+cobra:subcommand:config:short` — the single-character short flag (e.g. `c`). Optional.
- `+cobra:subcommand:config:usage` — the flag usage text. When absent, a generic description is
  used.

Because this aggregate flag represents the whole composed config, the child's own
`+cobra:flag` aggregate flag is usually removed from the child message when it is only consumed
as a subcommand child — values supplied through the composed aggregate flag are decoded into
the child portion directly, and individual child flags overlay them.

### `+cobra:subcommand:config:prefix`

Placed on the **parent** struct/message, this sets the prefix applied to the parent's hoisted
required field names in the `PChildRequiredFields` struct. The hoisted field name is
`<prefix><parentGoFieldName>` (e.g. `prefix=Vpc` + field `Name` → `VpcName`). The flag itself
keeps its original name (e.g. `vpc-name`); only the Go field in the composed config is renamed.
The prefix does **not** affect the generated *type* names, which always use the parent message
name verbatim. When the marker is absent, parent fields are hoisted under their original names.

Both `+cobra:config:child` and `+cobra:subcommand` accept an explicit `=false` to disable.

### Scalar-slice subcommands

`+cobra:subcommand` can also be placed on a repeated field of a **primitive** element type that
has no struct child — for example a `[]string` field such as `RemovePeers`. There is no child
struct to derive flags from, so instead of the child's individual flags the subcommand surfaces
a **single value flag** whose value is appended to the slice. This is useful for "remove by id"
style operations where the parent request carries a list of identifiers to remove.

For a parent `P` with scalar-slice field `Fs` (singularized to `F`, e.g. `RemovePeers` →
`RemovePeer`), it generates:

- `P<F>Config` — a composed config embedding the parent's `PChildRequiredFields`
  (tagged `yaml:",inline"`) plus a single named string field for the value (e.g.
  `RemovePeer string`), tagged with `json`/`yaml` tags matching the value flag name (e.g.
  `json:"peer-vpc-id" yaml:"peer-vpc-id"`). The type name uses the parent message name verbatim
  plus the singularized field name (e.g. `UpdateVpcRequestRemovePeerConfig`).
- `AddP<F>ConfigFlags(cmd)` — registers the composed-config aggregate flag (see
  `+cobra:subcommand:config:*`), the single value flag (see `+cobra:subcommand:value:*` below),
  and the parent `P`'s hoisted `+cobra:required` flags.
- `P<F>ConfigFromFlags(cmd) (*P, error)` — returns the **parent domain type** `*P`. It decodes
  the aggregate flag into the composed config, overlays the value flag and any hoisted parent
  flags that were set, enforces that both the value and the hoisted required fields are present
  post-merge, then writes the hoisted fields back to their original parent fields and **appends**
  the decoded value to the parent's slice field (`r.Fs = append(r.Fs, out.F)`).

#### `+cobra:subcommand:value:flag` / `:short` / `:usage`

Placed on the **scalar-slice subcommand field**, these customize the single value flag:

- `+cobra:subcommand:value:flag` — the value flag name (e.g. `peer-vpc-id`). Required to expose
  the value as an individual flag.
- `+cobra:subcommand:value:short` — the single-character short flag. Optional.
- `+cobra:subcommand:value:usage` — the value flag usage text. Optional.

The composed-config aggregate flag for a scalar-slice subcommand is configured with the same
`+cobra:subcommand:config:flag` / `:short` / `:usage` markers described above.

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
