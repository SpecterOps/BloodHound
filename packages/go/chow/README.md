# CHOW payload validator

The validator, schemas, CLI, and benchmark live in the BloodHound (BHCE) Go module. They were imported from [SpecterOps/CHOW](https://github.com/SpecterOps/chow) at commit `08cfc912e18557290734b326666a2f87d0ee87ca`.

Install the validator CLI from a released BloodHound revision:

```bash
go install github.com/specterops/bloodhound/cmd/chow@<revision>
```

To run it from the BloodHound (BHCE) module:

```bash
go run ./cmd/chow -output errors.txt payload.json
```

Run the benchmark from the same module with:

```bash
go run ./cmd/chowbench -runs 5 -warmup 1 payload-one.json payload-two.json
```

The benchmark reports invalid files while measuring them by default. Use `-strict` to make invalid files cause a non-zero exit status.

Raw OpenGraph JSON schemas are available in `SpecterOps/BloodHound/packages/go/chow/payload/jsonschema`.
