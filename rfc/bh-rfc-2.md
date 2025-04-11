---
bh-rfc: 2
title: Conventional Commits Guide
authors: |
  [Holms, Alyx](aholms@specterops.io)
  [Rangel, Ulises](urangel@specterops.io)
status: DRAFT
created: 2025-02-13
---

# Conventional Commits Guide

## 1. Overview

This document provides a set of guidelines for formatting commit messages that contributors of the BloodHound repository are expected to follow. The formatting rules and options herein largely follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) standard while outlining additional details as they are contextually relevant to development within the BloodHound repository.

## 2. Motivation

- **Regular Messages** - Provide a standard format and set of options that developers should follow so that incoming code changes can be quickly and easily parsed for the level and scope of impact.
- **Semver Adherence** - Use commit messaging that mitigates ambiguity in versioning the application when new releases are tagged.
- **Improved Automations** - Parse commit messages with tools to automate the generation of changelogs.
- **Clear Communication** - Reviewers are able to navigate changesets with certainty. New contributors are able to clearly convey what their work entails.

## 3. Parts and Format

A conventional commit will have the following shape:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Simple example:

`feat: Adds table view to Explore page`

Detailed example:

```
feat(explore page): Adds a table view for graph results with only nodes

Since a result of only nodes is not particularly interesting to view as a graph, a table view has been
added to the explore page that allows for inspecting result details with a better and more streamlined experience.

Closes BED-5555
```

### 3.1 Types

All commit messages MUST have a type included.  
The following types should be used when appropriate in a conventional commit message:

| Type  | Description                                                                                                                                                                                    |
| :---- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| feat  | For introducing a new feature in the application. Denotes that the release including this commit will have a `MINOR` version bump if there are no other `MAJOR` version changes being applied. |
| fix   | For fixing a bug in the application. Denotes that the release including this commit will have a `PATCH` version bump if there are no other `MAJOR` or `MINOR` version changes being applied.   |
| chore | For miscellaneous changes that do not fit well as either a `feat` for `fix` type.                                                                                                              |
| wip   | A convenience type that for in progress work. This is NOT an acceptable type to use for a commit that will merge into the default branch.                                                      |

### 3.2 Scope

The scope of a conventional commit is OPTIONAL but may be included to provide more detail as to what part of the application is being touched on with the work.

Scopes may indicate what part of the codebase the work falls under such as:

- API
- UI

Scopes may indicate particular sections, views, components, endpoints, or other subsets that communicate more about the changeset, e.g.:

- Datapie
- Explore Page
- Ingest
- Post-processing
- V2 Audit Endpoint
- Migrations

### 3.3 Description

The description part of a conventional commit message MUST be included.

- The description of a conventional commit message should give a brief overview of the work committed.
- The description of a conventional commit message should be no longer than 72 characters. This character limit is aimed at improving readability in areas where commit messages are regularly viewed.
- The description of a conventional commit message MUST follow directly after the colon and a space of the type/scope prefix

### 3.4 Body

The body of a conventional commit message is OPTIONAL.

The body should be separated from the the type and description with an empty line.

The body is a good place to provide additional context, details, motivations, or other relevant information to the changeset. This text should wrap at 72 characters for readability.

### 3.5 Footer

The footer MUST be included as part of a conventional commit message

The footer MUST relate a ticket or issue number that the work is associated with. Some examples are included below and a reference to more details can be found [here](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue).

Examples:

`Closes #1111`

`Fixes BED-5555`

`resolves: #777`

`CLOSES: #1, resolves #2, Fix: BED-4444`

The footer should be separated from the body (or the type and description if no body is included) with an empty line.

The footer SHOULD be used to denote if there are breaking changes included in the change set. Including breaking changes denotes that the `MAJOR` version of the application should be bumped on the next release. The `BREAKING CHANGE` keyword should be used.

Example:

```
feat(auth): Introduces required facial recognition sign in

BREAKING CHANGE: The API endpoints for login and registration have changed.
Closes BED-5555
```

The footer can also be used to relate to the ticket or issue number the work is associated with. The keywords `closes` or `fixes` should be used in this case.

## 4. Best Practices

The following are best practices for using Conventional Commits:

- Always write clear and concise commit messages.
- Use the appropriate `type` for each commit based on the nature of the change.
- Avoid using vague `types` such as `chore` or `wip` if `feat` or `fix` can be appropriately applied.
- If a commit contains a breaking change, clearly document the change in the footer.
- Group related changes into a single commit instead of multiple small commits.

## 5. Exceptional Scenarios

It may not always be realistic to strictly adhere to guidelines presented for various reasons even though it should be pursued as best as possible.

The following scenarios provide common and historical examples of when this might be the case and attempt to provide some suggestions that can be followed in case the situation arises anew.

- The changeset includes both a bug fix and a new feature. Which type should I use?

  - In this case, the `feat` type should be used over `fix` as it would take precedence as it relates to semver
  - The work may be split up before code changes are undertaken so that these conflicts are minimized

- My changeset included various commits and messages. What should I do?

  - You may squash and summarize your commits into one commit
  - You can edit your previous commits so that they also align to conventional commit format

- The application needs to be bumped to the next major version even though no `MAJOR` or breaking changes have been introduced with the changeset

  - Coordinate with the drivers of the `MAJOR` version bump and include details in the body of the commit message

## 6. Resources

- [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
- [Semantic Versioning](https://semver.org/)
- [Commitlint](https://commitlint.js.org/)
- [semantic-release](https://semantic-release.gitbook.io/)

---
