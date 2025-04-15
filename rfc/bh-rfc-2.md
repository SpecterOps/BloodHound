---
bh-rfc: 2
title: Early Return Pattern
authors: |
    [Holms, Alyx](aholms@specterops.io)
    [Pomeroy, Kaleb](kpomeroy@specterops.io)
status: DRAFT
created: 2025-04-15
---

# Early Return Pattern

Go encourages keeping blocks small, variables scoped tightly, and failure paths easy to follow. While Inline Initializer Statements help with scoping, they should not be combined with chained nesting patterns, which obscure control flow and increase complexity. Nesting hides the success case in the deepest block, when that’s actually the flow we care about most.

## 1. Overview

This proposal recommends adopting the Early Return Pattern as the default style for handling sequential workflows and error propagation in Go. Early returns flatten control flow, make error handling explicit, and align with Go’s idiomatic practices as described in Effective Go. This change improves readability, reduces indentation depth, and creates more maintainable and composable functions.

### 1.1 Example of Early Return Pattern

```
func flat() {
	a, err := stepA()
	if err != nil {
		fmt.Println("A had an error")
		return
	}

	b, err := stepB()
	if err != nil {
		fmt.Println("B had an error")
		return
	}
	c, err := stepC()
	if err != nil {
		fmt.Println("C had an error")
		return
	}

	fmt.Println(a + b + c)
}
```

### 1.2 Example of an Early Return Anti-pattern

```
func nesting() {
	if a, err := stepA(); err != nil {
		fmt.Println("A had an error")
	} else if b, err := stepB(); err != nil {
		fmt.Println("B had an error")
	} else if c, err := stepC(); err != nil {
		fmt.Println("C had an error")
	} else {
		fmt.Println(a + b + c)
	}
}
```

## 2. Motivation & Goals

The goal of this RFC is to establish a consistent and idiomatic approach to control flow and error handling in sequential Go functions. Specifically, it promotes the use of early returns over chained nested if statements when dealing with dependent operations that may return errors.

-   **Readability** Early returns keep the success path at the top level, making it easier to understand the primary logic of a function at a glance.
-   **Maintainability** Flattened control flow reduces indentation and visual noise, making future changes and code reviews easier.
-   **Idiomatic Go** This approach aligns with community conventions and guidance from Effective Go and widely adopted Go style guides.
-   **Error clarity** By separating failure paths early, each step’s error handling becomes clearer and less entangled with the main logic.
-   **Reduced coupling**: Avoids tightly chaining variable scopes, enabling greater reuse and flexibility of intermediate values.

This change supports writing code that is easier to reason about, debug, and refactor over time.

## 3. Existing Guidance

### 3.1 Bloodhound Style Guide

The existing guidance use an anti-pattern we want to avoid going forward. Outlined here are the reasons for this change.

`Chaining the scope of transient variables is an effective way to organize error handling for chained and interdependent function calls. Tight variable scopes also have the added benefit of providing strong refactoring seams.`

-   Chaining scopes creates a brittle coupling - a variable cannot be used independently without unwinding the chain
-   Go encourages functions to be short and scoped. Tight scopes for small functions are less helpful. If scope control is necessary, refactor into named functions instead of nested blocks.
-   Nesting gives the AST clean scoping, but that’s for the compiler’s benefit. For humans, deep nesting obscures the happy path and binds unrelated steps together visually. Early returns expose structure more clearly where it matters, at the level of code review, maintenance, and debugging

### 3.2 Considerations for existing patterns

There are valid reasons to use the nesting pattern as it is implemented historically. This section will contain a few justifications and counter points.

#### 3.2.1 Single Exit Rule

Encourage one clear exit point, which can simplify certain types of debugging, logging, or deferred cleanup logic (e.g. closing resources, unlocking mutexes) and can be placed after all the success paths in a single place.

Counterpoint: Go encourages multiple returns for clarity. Single exit rule is a holdover from older (C-like) languages where it mattered. In practice, early returns reduce indentation and make the success path obvious, while cleanup is better handled with defer, which works regardless of exit point.

#### 3.2.2 Inline error handling

Error checks are colocated with the logic they protect, which can help when steps are deeply dependent or when context needs to stay tightly grouped.

Counterpoint: Colocating error checks with logic sounds clean but quickly becomes cluttered, especially as error handling expands (logging, wrapping, metrics). Early returns separate failure paths from the happy path, which aligns with Go’s ‘fail fast’ idiom and improves scannability.

#### 3.2.3 Nested ifs as Scoped Guards:

Each nested block forms a well-scoped zone for a variable, ensuring it can’t accidentally be accessed elsewhere. Helps with tight, explicit control over variable lifetimes.
Counterpoint: Scoped lifetimes are a weak justification in small, linear functions. We could use anonymous blocks or extract to helper functions. Readability and maintainability matter more than narrowing variable lifespan.

Additionally, this does not guard us completely from scoping issues. If, for example stepC returned only a value (and not an error), the following two lines are now valid.

`} else if c := stepC(); err != nil {`
`c := stepC() {`

In a stack of nested if statements, this nuance would be easy to miss. The err from stepB is still in scope. While the bug would still compile in the idiomatic flat style, it’s more obvious during review because each variable declaration is isolated and visually prominent

## 4. Summary

Chained nesting offers real scoping guarantees at the compiler level, but it comes at a readability cost for humans. In Go, clarity and simplicity should take precedence. Early returns flatten logic, highlight the success path, and align with the language’s philosophy of small, focused functions and explicit control flow. The goal isn’t just correctness at the machine's level. it’s maintainability at the developer level. And for that, flat wins.
