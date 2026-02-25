# chow (confirm hound output worthiness)
chow is a command line tool to validate BloodHound payloads

# Usage
```bash
chow -output errors.txt test.json
```
`-output` will redirect errors to an output file. Otherwise errors will be written to stdout

# Installation
```bash
go install github.com/specterops/bloodhound/packages/go/chow
```
If bloodhound is already installed locally, navigate to the bloodhound directory and run:
```bash
go install ./packages/go/chow
```